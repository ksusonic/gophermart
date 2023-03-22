package accrual

import (
	"fmt"

	"github.com/ksusonic/gophermart/internal/models"
)

type DB interface {
	GetOrdersWithStatus(status ...models.OrderStatus) (*[]models.Order, error)
	UpdateOrder(order *models.Order) error
}

func (w *Worker) getOrdersToCheck() ([]models.Order, error) {
	orders, err := w.db.GetOrdersWithStatus(models.OrderStatusNew, models.OrderStatusProcessing)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}
	return *orders, nil
}
