package database

import (
	"database/sql"
	"github.com/ksusonic/gophermart/internal/api"
	"github.com/ksusonic/gophermart/internal/models"
)

func (d *DB) GetUserByLogin(login string) (*models.User, error) {
	user := &models.User{}
	tx := d.Orm.Where("login = ?", login).Limit(1).Find(user)
	err := tx.Error
	if err == nil && tx.RowsAffected == 0 {
		err = sql.ErrNoRows
	}
	return user, err
}

func (d *DB) GetOrderByID(id string) (*models.Order, error) {
	order := &models.Order{}
	tx := d.Orm.Model(&models.Order{}).Where("id = ?", id).Limit(1).Find(order)
	err := tx.Error
	if err == nil && tx.RowsAffected == 0 {
		err = sql.ErrNoRows
	}
	return order, err
}

func (d *DB) GetWithdrawnOrdersByUserID(userID uint) (*[]models.Order, error) {
	withdrawals := &[]models.Order{}
	tx := d.Orm.Model(&models.Order{}).Where("user_id = ? and withdraw is not null", userID).Find(withdrawals)
	err := tx.Error
	if err == nil && tx.RowsAffected == 0 {
		err = sql.ErrNoRows
	}
	return withdrawals, err
}

func (d *DB) GetOrdersByUserID(userID uint) (*[]models.Order, error) {
	orders := &[]models.Order{}
	tx := d.Orm.Model(&models.Order{}).Where("user_id = ?", userID).Find(orders)
	err := tx.Error
	if err == nil && tx.RowsAffected == 0 {
		err = sql.ErrNoRows
	}
	return orders, err
}

func (d *DB) CalculateUserStats(userID uint) (*api.UserInfo, error) {
	userInfo := &api.UserInfo{}
	err := d.Orm.
		Table("orders").
		Select("sum(accrual) as balance, sum(withdraw) as withdraw").
		Where("user_id = ?", userID).
		Scan(userInfo).
		Error
	userInfo.Balance -= userInfo.Withdraw
	return userInfo, err
}

func (d *DB) GetOrdersWithStatus(status ...models.OrderStatus) (*[]models.Order, error) {
	orders := &[]models.Order{}
	tx := d.Orm.Model(&models.Order{}).Where("status in ?", status).Find(orders)
	err := tx.Error
	if err == nil && tx.RowsAffected == 0 {
		err = sql.ErrNoRows
	}
	return orders, err
}
