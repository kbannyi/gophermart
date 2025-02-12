package dto

type WithdrawalRequest struct {
	OrderID string  `json:"order"`
	Sum     float64 `json:"sum"`
}

type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type WithdrawalResponse struct {
	OrderID     string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}
