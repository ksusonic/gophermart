package api

import (
	"github.com/ksusonic/gophermart/internal/models"
)

type Order struct {
	Number     string             `json:"number"`
	Status     models.OrderStatus `json:"status"`
	Accrual    float64            `json:"accrual,omitempty"`
	UploadedAt string             `json:"uploaded_at"`
}

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type WithdrawResponse []Withdraw
type Withdraw struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}
