package model

import "github.com/shopspring/decimal"

type TransferResult struct {
	FromUserID int64
	ToUserID   int64
	Amount     decimal.Decimal
	Status     string
}
