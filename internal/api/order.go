package api

type Order struct {
	Number     string `json:"number"`
	Status     string `json:"status"`
	Accrual    int64  `json:"accrual,omitempty"`
	UploadedAt string `json:"uploaded_at"`
}

type WithdrawRequest struct {
	Order string `json:"order"`
	Sum   int64  `json:"sum"`
}

type WithdrawResponse struct {
	Order       string `json:"order"`
	Sum         int64  `json:"sum"`
	ProcessedAt string `json:"processed_at"`
}
