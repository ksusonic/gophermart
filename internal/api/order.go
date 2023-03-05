package api

import "time"

const (
	StatusRegistered = "REGISTERED"
	StatusInvalid    = "INVALID"
	StatusProcessing = "PROCESSING"
	StatusProcessed  = "PROCESSED"
)

type Order struct {
	Number     string `json:"number"`
	Status     string `json:"status"`
	Accrual    int64  `json:"accrual,omitempty"`
	UploadedAt string `json:"uploaded_at"`
}

type Withdraw struct {
	Order       string    `json:"order"`
	Sum         int64     `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
