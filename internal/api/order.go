package api

import "time"

const (
	StatusRegistered = "REGISTERED"
	StatusInvalid    = "INVALID"
	StatusProcessing = "PROCESSING"
	StatusProcessed  = "PROCESSED"
)

type Order struct {
	Order      string     `json:"order"`
	Status     string     `json:"status"`
	Accrual    int64      `json:"accrual,omitempty"`
	UploadedAt *time.Time `json:"uploaded_at,omitempty"`
}

type Withdraw struct {
	Order       string     `json:"order"`
	Sum         int64      `json:"sum"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}
