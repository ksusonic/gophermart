package database

import "github.com/ksusonic/gophermart/internal/models"

func (d *DB) CreateUser(user *models.User) error {
	return d.Orm.Create(user).Error
}

func (d *DB) CreateOrder(order *models.Order) error {
	return d.Orm.Create(order).Error
}
