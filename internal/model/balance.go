package model

import "github.com/shopspring/decimal"

type Balance struct {
	ID     int64           `json:"id" db:"id"`
	UserID int64           `json:"user_id" db:"user_id"`
	Amount decimal.Decimal `json:"amount" db:"amount"`
}
