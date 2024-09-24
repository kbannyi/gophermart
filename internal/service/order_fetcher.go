package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/kbannyi/gophermart/internal/config"
	"github.com/kbannyi/gophermart/internal/domain"
	"github.com/kbannyi/gophermart/internal/logger"
	"github.com/shopspring/decimal"
)

type OrderWorkerRepository interface {
	SelectForFetching(ctx context.Context, orders *[]domain.Order, pageSize int, page int) error
	BatchSave(ctx context.Context, orders []domain.Order) error
}

type OrderFetcher struct {
	accrualURL string
	repository OrderWorkerRepository
	wg         *sync.WaitGroup
	ctx        context.Context
	signalCh   chan struct{}
}

func NewOrderFetcher(r OrderWorkerRepository, ctx context.Context, wg *sync.WaitGroup, cfg config.Config) OrderFetcher {
	w := OrderFetcher{
		accrualURL: cfg.AccrualAddr,
		repository: r,
		wg:         wg,
		ctx:        ctx,
		signalCh:   make(chan struct{}, 1),
	}

	return w
}

func (w OrderFetcher) Activate() {
	select {
	case w.signalCh <- struct{}{}:
	default:
	}
}

func (w OrderFetcher) Run() {
	const pageSize = 50
	orders := make([]domain.Order, 0, pageSize)
	w.wg.Add(1)
	go func() {
		logger.Log.Debug("Fetching started")
		defer w.wg.Done()

		ctx := context.Background()
		page := 0
		for {
			orders = orders[:0]
			if err := w.repository.SelectForFetching(ctx, &orders, pageSize, page); err != nil {
				logger.Log.Error(err.Error())
				orders = orders[:0]
			}
			{
				var ids []string
				for _, o := range orders {
					ids = append(ids, o.ID)
				}
				logger.Log.Debug("Fetching:", "orders", ids)
			}

			selectedCount := len(orders)
			switch {
			case selectedCount == pageSize:
				page += 1
			case selectedCount == 0 && page == 0:
				logger.Log.Debug("No orders for fetching, entering sleeping mode")
				select {
				case <-w.ctx.Done():
					return
				case <-w.signalCh:
					continue
				case <-time.After(time.Second * 60):
					continue
				}
			default:
				page = 0
			}

			changed := make([]domain.Order, 0, pageSize)
			for _, o := range orders {
				if w.ctx.Err() != nil {
					return
				}

				resp, err := w.getAccrual(o.ID)
				if err == nil {
					err = w.processOrder(&o, resp)
				}
				if err != nil {
					switch {
					case errors.Is(err, errNoChanges):
						continue
					case errors.Is(err, errSameStatus):
						continue
					default:
						logger.Log.Error("fetch error", "orderid", o.ID, "error", err)
						continue
					}
				}

				changed = append(changed, o)
				logger.Log.Debug("Order update fetched", "data", resp)
			}

			if len(changed) != 0 {
				if err := w.repository.BatchSave(ctx, changed); err != nil {
					logger.Log.Error(fmt.Sprintf("couldn't save updated orders to db: %s", err))
				}
			}

			time.Sleep(time.Millisecond * 1000) // stop ddos
		}
	}()

}

type accrualResponse struct {
	Order   string          `json:"order"`
	Status  string          `json:"status"`
	Accrual decimal.Decimal `json:"accrual"`
}

const (
	AccrualStatusRegistered = "REGISTERED"
	AccrualStatusInvalid    = "INVALID"
	AccrualStatusProcessing = "PROCESSING"
	AccrualStatusProcessed  = "PROCESSED"
)

var errNoChanges = errors.New("accrual returned no changes")
var errSameStatus = errors.New("order status hasn't been changed")

func (w OrderFetcher) getAccrual(id string) (*accrualResponse, error) {
	ctx, cancel := context.WithTimeout(w.ctx, time.Second*5)
	defer cancel()

	accrualURL, err := url.JoinPath(w.accrualURL, "/api/orders/", id)
	if err != nil {
		return nil, fmt.Errorf("couldn't join path for order: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "GET", accrualURL, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create http request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("couldn't fetch order from accrual: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		return nil, errNoChanges
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected accrual response code: %d", resp.StatusCode)
	}
	decoder := json.NewDecoder(resp.Body)
	var reqmodel accrualResponse
	if err := decoder.Decode(&reqmodel); err != nil {
		return nil, fmt.Errorf("couldn't deserialize accrual response: %w", err)
	}

	return &reqmodel, nil
}

func (w OrderFetcher) processOrder(o *domain.Order, resp *accrualResponse) error {
	var newStatus domain.OrderStatus
	switch resp.Status {
	case AccrualStatusRegistered:
		newStatus = domain.StatusNew
	case AccrualStatusInvalid:
		newStatus = domain.StatusInvalid
	case AccrualStatusProcessing:
		newStatus = domain.StatusProcessing
	case AccrualStatusProcessed:
		newStatus = domain.StatusProcessed
	default:
		return fmt.Errorf("unexpected accrual order status: %s", resp.Status)
	}

	if o.Status == newStatus {
		return errSameStatus
	}

	err := o.SetStatus(newStatus)
	if err != nil {
		return err
	}
	if !resp.Accrual.IsZero() {
		o.Accrual = decimal.NewNullDecimal(resp.Accrual)
	}

	return nil
}
