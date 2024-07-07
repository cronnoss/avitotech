package model

import "github.com/shopspring/decimal"

type Balance struct {
	UserID   int64           `json:"userid" db:"userid"`
	Amount   decimal.Decimal `json:"amount" db:"amount"`
	Currency string          `json:"currency" db:"currency"`
}
