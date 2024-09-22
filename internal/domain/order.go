package domain

import (
	"errors"
	"slices"
	"time"

	"github.com/shopspring/decimal"
)

type Order struct {
	ID         string              `db:"id"`
	Status     OrderStatus         `db:"status"`
	UserId     string              `db:"user_id"`
	Accrual    decimal.NullDecimal `db:"accrual"`
	CreatedUTC time.Time           `db:"created_utc"`
	UpdatedUTC *time.Time          `db:"updated_utc"`
}

var ErrStatusChange = errors.New("invalid status change attempt")

var allowedChanges = map[OrderStatus][]OrderStatus{
	StatusNew:        {StatusProcessing, StatusProcessed, StatusInvalid},
	StatusProcessing: {StatusProcessed, StatusInvalid},
	StatusProcessed:  {},
	StatusInvalid:    {},
}

func (o *Order) SetStatus(s OrderStatus) error {
	if !slices.Contains(allowedChanges[o.Status], s) {
		return ErrStatusChange
	}

	o.Status = s

	return nil
}
