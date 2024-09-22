package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/kbannyi/gophermart/internal/domain"
	"github.com/kbannyi/gophermart/internal/dto"
	"github.com/kbannyi/gophermart/internal/logger"
	"github.com/kbannyi/gophermart/internal/models"
	"github.com/kbannyi/gophermart/internal/repository"
	"github.com/kbannyi/gophermart/internal/service/luhn"
)

type WithdrawalService interface {
	GetBalance(context.Context) (*models.Balance, error)
	Withdraw(ctx context.Context, orderID string, sum int) error
	GetWithdrawals(context.Context) ([]domain.Withdrawal, error)
}

type WithdrawalHandler struct {
	service WithdrawalService
}

func NewWithdrawalHandler(service WithdrawalService) WithdrawalHandler {
	return WithdrawalHandler{
		service: service,
	}
}

func (h WithdrawalHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	balance, err := h.service.GetBalance(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	err = encoder.Encode(dto.BalanceResponse{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	})
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
}

func (h WithdrawalHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	decoder := json.NewDecoder(r.Body)
	var req dto.WithdrawalRequest
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !luhn.Valid(req.OrderID) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	if req.Sum <= 0 {
		http.Error(w, "sum must be positive number", http.StatusBadRequest)
		return
	}
	err = h.service.Withdraw(ctx, req.OrderID, req.Sum)
	if err != nil {
		if errors.Is(err, repository.ErrNotEnoughPoints) {
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}
		if errors.Is(err, repository.ErrAlreadyExists) {
			http.Error(w, "order already paid", http.StatusBadRequest)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h WithdrawalHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	withdrawals, err := h.service.GetWithdrawals(ctx)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	respmodels := make([]dto.WithdrawalResponse, 0, len(withdrawals))
	for _, w := range withdrawals {
		respmodels = append(respmodels, dto.WithdrawalResponse{
			OrderID:     w.ID,
			Sum:         w.Amount,
			ProcessedAt: w.CreatedUTC.Format(time.RFC3339),
		})
	}

	encoder := json.NewEncoder(w)
	err = encoder.Encode(respmodels)
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
}
