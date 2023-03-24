package models

import (
	"database/sql"

	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	gorm.Model
	ID string `gorm:"primaryKey"` // Number

	UserID uint `gorm:"not null"`

	Status   OrderStatus
	Accrual  sql.NullInt64
	Withdraw sql.NullInt64
}
