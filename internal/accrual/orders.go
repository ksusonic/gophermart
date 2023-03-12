package accrual

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ksusonic/gophermart/internal/api"
)

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
