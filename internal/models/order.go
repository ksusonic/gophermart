package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	StatusRegistered = "REGISTERED"
	StatusInvalid    = "INVALID"
	StatusProcessing = "PROCESSING"
	StatusProcessed  = "PROCESSED"
)

type Order struct {
	gorm.Model
	Number    string    `gorm:"unique" json:"number"`
	Status    string    `json:"status"`
	Accrual   int64     `json:"accrual,omitempty"`
	CreatedAt time.Time `json:"uploaded_at"`
}

type Withdraw struct {
	Order       string     `json:"order"`
	Sum         int64      `json:"sum"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}
