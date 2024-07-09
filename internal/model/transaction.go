package model

import "github.com/shopspring/decimal"

type Transaction struct {
	ID        int64           `json:"id" db:"id"`
	UserID    int64           `json:"user_id" db:"user_id"`
	Amount    decimal.Decimal `json:"amount" db:"amount"`
	Operation string          `json:"operation" db:"operation"`
	Date      string          `json:"date" db:"date"`
}
