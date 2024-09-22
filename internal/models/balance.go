package models

import "github.com/shopspring/decimal"

type Balance struct {
	UserID    string
	Current   decimal.Decimal
	Withdrawn decimal.Decimal
}
