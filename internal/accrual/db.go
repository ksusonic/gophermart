package accrual

import (
	"fmt"

	"github.com/ksusonic/gophermart/internal/models"
)

type DB interface {
	GetOrdersWithStatus(orders *[]models.Order, status ...models.OrderStatus) error
	UpdateOrder(order *models.Order) error
}

func (w *Worker) getOrdersToCheck() ([]models.Order, error) {
	var orders []models.Order
	err := w.db.GetOrdersWithStatus(&orders, models.OrderStatusNew, models.OrderStatusProcessing)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}
	return orders, nil
}
