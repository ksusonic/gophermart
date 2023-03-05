package models

import (
	"database/sql"
	"gorm.io/gorm"
	"time"
)

const (
	StatusRegistered = "REGISTERED"
	StatusInvalid    = "INVALID"
	StatusProcessing = "PROCESSING"
	StatusProcessed  = "PROCESSED"
)

type Order struct {
	gorm.Model
	Number string `gorm:"unique"`
	UserID uint

	Status    string
	Accrual   sql.NullInt64
	CreatedAt time.Time
}

type Withdraw struct {
	Order       string     `json:"order"`
	Sum         int64      `json:"sum"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}
