package accrual

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ksusonic/gophermart/internal/api"
	"github.com/ksusonic/gophermart/internal/models"
)

const OrdersHandler = "/api/orders/"

func (w *Worker) getOrderInfo(number string) (*api.AccrualResponse, error) {
	response, err := w.client.Get(w.accrualAddress + OrdersHandler + number)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusNoContent:
		return nil, fmt.Errorf("order %s not registered in accrual system", number)
	case http.StatusTooManyRequests:
		return nil, fmt.Errorf("ratelimited")
	case http.StatusOK:
		bytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("could not read response body: %w", err)
		}
		var accrualResponse api.AccrualResponse
		err = json.Unmarshal(bytes, &accrualResponse)
		if err != nil {
			return nil, fmt.Errorf("could not parse response body: %s error: %w", string(bytes), err)
		}
		return &accrualResponse, nil
	default:
		return nil, fmt.Errorf("unknown status: %s", response.Status)
	}
}

func (w *Worker) processOrder(response *api.AccrualResponse, order *models.Order) error {
	switch response.Status {
	case api.AccrualStatusProcessed:
		order.Status = models.OrderStatusProcessed
		order.Accrual = sql.NullFloat64{
			Float64: response.Accrual,
			Valid:   true,
		}
		err := w.db.UpdateOrder(order)
		if err != nil {
			return fmt.Errorf("error updating order: %v", err)
		}
		w.logger.Infof("order %s is processed!", order.ID)
	case api.AccrualStatusProcessing:
		order.Status = models.OrderStatusInvalid
		err := w.db.UpdateOrder(order)
		if err != nil {
			return fmt.Errorf("error updating order: %v", err)
		}
		w.logger.Debugf("order %s is processing", order.ID)
	case api.AccrualStatusRegistered:
		w.logger.Debugf("order %s is registered", order.ID)
	case api.AccrualStatusInvalid:
		order.Status = models.OrderStatusInvalid
		err := w.db.UpdateOrder(order)
		if err != nil {
			return fmt.Errorf("error updating order: %v", err)
		}
	default:
		w.logger.Warn("unknown status from accrual: %s", response.Status)
	}
	return nil
}
