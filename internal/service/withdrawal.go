package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kbannyi/gophermart/internal/auth"
	"github.com/kbannyi/gophermart/internal/domain"
	"github.com/kbannyi/gophermart/internal/models"
	"github.com/shopspring/decimal"
)

type WithdrawalRepository interface {
	GetBalance(ctx context.Context, userid string) (*models.Balance, error)
	Withdraw(ctx context.Context, w domain.Withdrawal) error
	GetWithdrawals(ctx context.Context, userID string) ([]domain.Withdrawal, error)
}

type WithdrawalService struct {
	repository WithdrawalRepository
}

func NewWithdrawalService(r WithdrawalRepository) WithdrawalService {
	return WithdrawalService{
		repository: r,
	}
}

func (s WithdrawalService) Withdraw(ctx context.Context, orderID string, sum decimal.Decimal) error {
	u, err := auth.FromContext(ctx)
	if err != nil {
		return err
	}
	w := domain.Withdrawal{
		ID:         uuid.NewString(),
		UserId:     u.UserID,
		Amount:     sum,
		CreatedUTC: time.Now(),
		OrderID:    orderID,
	}
	err = s.repository.Withdraw(ctx, w)
	if err != nil {
		return err
	}

	return nil
}

func (s WithdrawalService) GetBalance(ctx context.Context) (*models.Balance, error) {
	u, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.repository.GetBalance(ctx, u.UserID)
}

func (s WithdrawalService) GetWithdrawals(ctx context.Context) ([]domain.Withdrawal, error) {
	u, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	withdrawals, err := s.repository.GetWithdrawals(ctx, u.UserID)
	if err != nil {
		return nil, err
	}

	return withdrawals, nil
}
