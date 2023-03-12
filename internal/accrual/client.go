package accrual

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

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
