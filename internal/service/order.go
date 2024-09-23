package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/kbannyi/gophermart/internal/auth"
	"github.com/kbannyi/gophermart/internal/domain"
	"github.com/kbannyi/gophermart/internal/logger"
	"github.com/kbannyi/gophermart/internal/repository"
)

var ErrBelongToAnother = errors.New("this order id belongs to another user")

type OrderRepository interface {
	SaveNewOrder(context.Context, domain.Order) error
	Get(ctx context.Context, id string) (*domain.Order, error)
	GetOrders(ctx context.Context, userid string) ([]domain.Order, error)
}

type OrderService struct {
	repository OrderRepository
}

func NewOrderService(r OrderRepository) OrderService {
	s := OrderService{
		repository: r,
	}

	return s
}

func (s OrderService) RunBackgroundFetch(done <-chan struct{}, wg *sync.WaitGroup) {
	go func() {
		wg.Add(1)
		logger.Log.Info("RunBackgroundFetch is started")
		<-done
		logger.Log.Info("RunBackgroundFetch is finished")
		wg.Done()
	}()
}

func (s OrderService) SaveNewOrder(ctx context.Context, id string) error {
	u, err := auth.FromContext(ctx)
	if err != nil {
		return err
	}
	o := domain.Order{
		ID:         id,
		Status:     domain.StatusNew,
		UserId:     u.UserID,
		Accrual:    nil,
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
			if existing.UserId != o.UserId {
				return ErrBelongToAnother
			}
		}
		return err
	}

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
