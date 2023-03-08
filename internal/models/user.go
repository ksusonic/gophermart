package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Login    string `gorm:"not null;unique"`
	Password string `gorm:"not null"`

	Orders []Order
}
