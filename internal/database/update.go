package database

import "github.com/ksusonic/gophermart/internal/models"

func (d *DB) UpdateOrder(order *models.Order) error {
	return d.Orm.Save(order).Error
}
