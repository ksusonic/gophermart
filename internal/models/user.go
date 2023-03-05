package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Login    string `gorm:"not null;unique"`
	Password string `gorm:"not null"`

	Current   float64
	Withdrawn int64

	Orders      []Order
	Withdrawals []Withdraw
}
