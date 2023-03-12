package accrual

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ksusonic/gophermart/internal/api"
	"github.com/ksusonic/gophermart/internal/models"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const OrdersHandler = "/api/orders/"

type Worker struct {
	accrualAddress string
	db             DB
	logger         *zap.SugaredLogger

	updateRate time.Duration
	client     *http.Client
}

func NewWorker(accrualAddress string, db DB, logger *zap.SugaredLogger) *Worker {
	return &Worker{
		accrualAddress: accrualAddress,
		db:             db,
		logger:         logger,

		updateRate: time.Second * 3,
		client:     &http.Client{}, // for client customization
	}
}

func (w *Worker) Run(ctx context.Context) {
	w.logger.Infof("Started accrual worker")
	ticker := time.NewTicker(w.updateRate)
	select {
	case <-ticker.C:
		if err := w.processAccrual(); err != nil {
			w.logger.Errorf("error processing accrual: %v", err)
		}
	case <-ctx.Done():
		return
	}
}

func (w *Worker) processAccrual() error {
	orders, err := w.getOrdersToCheck()
	if err != nil {
		return fmt.Errorf("could not get orders from db: %w", err)
	}

	eg := &errgroup.Group{}
	for i := range orders {
		response, err := w.getOrderInfo(orders[i].ID)
		if err != nil {
			return err
		}
		order := orders[i]
		eg.Go(func() error {
			return w.processOrder(response, &order)
		})
	}
	if err := eg.Wait(); err != nil {
		w.logger.Errorf("error processing order: %v", err)
		return err
	}
	return nil
}

func (w *Worker) processOrder(response *api.AccrualResponse, order *models.Order) error {
	switch response.Status {
	case api.AccrualStatusProcessed:
		order.Status = models.OrderStatusProcessed
		if err := order.Accrual.Scan(response.Accrual); err != nil {
			return fmt.Errorf("error in scan int64 value %d: %v", response.Accrual, err)
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
