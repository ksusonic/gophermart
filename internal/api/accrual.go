package api

type AccrualStatus string

const (
	AccrualStatusRegistered AccrualStatus = "REGISTERED"
	AccrualStatusInvalid    AccrualStatus = "INVALID"
	AccrualStatusProcessing AccrualStatus = "PROCESSING"
	AccrualStatusProcessed  AccrualStatus = "PROCESSED"
)

type AccrualResponse struct {
	OrderNumber string        `json:"order"`
	Status      AccrualStatus `json:"status"`
	Accrual     *int64        `json:"accrual"`
}
