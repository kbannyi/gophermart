package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type Withdrawal struct {
	ID         string          `db:"id"`
	UserID     string          `db:"user_id"`
	Amount     decimal.Decimal `db:"amount"`
	CreatedUTC time.Time       `db:"created_utc"`
	OrderID    string          `db:"order_id"`
}
