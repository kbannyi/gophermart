package domain

import "time"

type Withdrawal struct {
	ID         string    `db:"id"`
	UserId     string    `db:"user_id"`
	Amount     int       `db:"amount"`
	CreatedUTC time.Time `db:"created_utc"`
	OrderID    string    `db:"order_id"`
}
