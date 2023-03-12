package database

import "github.com/ksusonic/gophermart/internal/models"

func (d *DB) GetUserByLogin(login string, user *models.User) error {
	return d.Orm.Where("login = ?", login).Limit(1).Find(&user).Error
}

func (d *DB) GetOrderByID(id string, order *models.Order) (int64, error) {
	tx := d.Orm.Model(&models.Order{}).Where("id = ?", id).Limit(1).Find(&order)
	return tx.RowsAffected, tx.Error
}

func (d *DB) GetUserWithdraws(userID uint, withdrawals *[]models.Order) error {
	return d.Orm.Model(&models.Order{}).Where("user_id = ? and withdraw is not null", userID).Find(&withdrawals).Error
}

func (d *DB) GetOrdersByUserID(userID uint, orders *[]models.Order) error {
	return d.Orm.Model(&models.Order{}).Where("user_id = ?", userID).Find(&orders).Error
}

func (d *DB) CalculateUserStats(userID uint) (userInfo struct{ Balance, Withdraw int64 }, err error) {
	err = d.Orm.
		Table("orders").
		Select("sum(accrual)-sum(withdraw) as balance, sum(withdraw) as withdraw").
		Where("user_id=?", userID).
		Scan(&userInfo).
		Error
	return userInfo, err
}
