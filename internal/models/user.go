package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Login        string `gorm:"not null;unique"`
	PasswordHash string `gorm:"not null"`

	Orders []Order
}
