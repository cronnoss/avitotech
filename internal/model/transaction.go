package model

import "github.com/shopspring/decimal"

type Transaction struct {
	ID        int64           `json:"id"`
	UserID    int64           `json:"user_id"`
	Amount    decimal.Decimal `json:"amount"`
	Operation string          `json:"operation"`
	Date      string          `json:"date"`
}
