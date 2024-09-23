package domain

import "time"

type Order struct {
	ID         string      `db:"id"`
	Status     OrderStatus `db:"status"`
	UserId     string      `db:"user_id"`
	Accrual    *int        `db:"accrual"`
	CreatedUTC time.Time   `db:"created_utc"`
	UpdatedUTC *time.Time  `db:"updated_utc"`
}
