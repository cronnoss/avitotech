package model

import "github.com/shopspring/decimal"

type Transfer struct {
	FromID int64           `json:"user_id" db:"user_id"`
	ToID   int64           `json:"to_id" db:"user_id"`
	Amount decimal.Decimal `json:"amount" db:"amount"`
}
