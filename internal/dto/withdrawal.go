package dto

type WithdrawalRequest struct {
	OrderID string `json:"order"`
	Sum     int    `json:"sum"`
}

type BalanceResponse struct {
	Current   int `json:"current"`
	Withdrawn int `json:"withdrawn"`
}

type WithdrawalResponse struct {
	OrderID     string `json:"order"`
	Sum         int    `json:"sum"`
	ProcessedAt string `json:"processed_at"`
}
