package service

import (
	"context"
	"errors"
	"time"

	"github.com/kbannyi/gophermart/internal/auth"
	"github.com/kbannyi/gophermart/internal/domain"
	"github.com/kbannyi/gophermart/internal/repository"
	"github.com/shopspring/decimal"
)

var ErrBelongToAnother = errors.New("this order id belongs to another user")

type OrderRepository interface {
	SaveNewOrder(context.Context, domain.Order) error
	Get(ctx context.Context, id string) (*domain.Order, error)
	GetOrders(ctx context.Context, userid string) ([]domain.Order, error)
}

type OrderFetcherSignaler interface {
	Activate()
}

type OrderService struct {
	repository OrderRepository
	fetcher    OrderFetcherSignaler
}

func NewOrderService(r OrderRepository, f OrderFetcherSignaler) OrderService {
	s := OrderService{
		repository: r,
		fetcher:    f,
	}

	return s
}

func (s OrderService) SaveNewOrder(ctx context.Context, id string) error {
	u, err := auth.FromContext(ctx)
	if err != nil {
		return err
	}
	o := domain.Order{
		ID:         id,
		Status:     domain.StatusNew,
		UserID:     u.UserID,
		Accrual:    decimal.NullDecimal{},
		CreatedUTC: time.Now(),
		UpdatedUTC: nil,
	}
	err = s.repository.SaveNewOrder(ctx, o)
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			existing, e := s.repository.Get(ctx, id)
			if e != nil {
				return e
			}
			if existing.UserID != o.UserID {
				return ErrBelongToAnother
			}
		}
		return err
	}

	s.fetcher.Activate()
	return nil
}

func (s OrderService) GetOrders(ctx context.Context) ([]domain.Order, error) {
	u, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	orders, err := s.repository.GetOrders(ctx, u.UserID)
	if err != nil {
		return nil, err
	}

	return orders, nil
}
