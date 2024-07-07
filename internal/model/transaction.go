package model

import "github.com/shopspring/decimal"

type Transaction struct {
	ID        int64           `json:"id"`
	UserID    int64           `json:"userid"`
	Amount    decimal.Decimal `json:"amount"`
	Currency  string          `json:"currency"`
	Operation string          `json:"operation"`
	Date      string          `json:"date"`
}
