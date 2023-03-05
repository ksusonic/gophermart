package models

import (
	"database/sql"
	"gorm.io/gorm"
)

const (
	StatusNew = "NEW"
	//StatusProcessing = "PROCESSING"
	//StatusInvalid    = "INVALID"
	//StatusProcessed  = "PROCESSED"
)

type Order struct {
	gorm.Model
	ID string `gorm:"primaryKey"` // Number

	UserID    uint
	Withdraws []Withdraw

	Status           string
	Accrual          sql.NullInt64
	AccrualAvailable int64
}

type Withdraw struct {
	gorm.Model

	OrderID string
	UserID  uint
	Sum     int64
}
