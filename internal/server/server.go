package server

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/cronnoss/avitotech/internal/model"
	"github.com/shopspring/decimal"
)

var (
	ErrUserID   = errors.New("wrong UserID")
	ErrCurrency = errors.New("wrong Currency")
)

//go:generate mockery --name Logger
type Logger interface {
	Fatalf(format string, a ...interface{})
	Errorf(format string, a ...interface{})
	Warningf(format string, a ...interface{})
	Infof(format string, a ...interface{})
	Debugf(format string, a ...interface{})
}

//go:generate mockery --name Application
type Application interface {
	GetBalance(context.Context, *model.Balance) (*model.Balance, error)
	TopUp(context.Context, int64, decimal.Decimal) (*model.Balance, error)
	Debit(context.Context, int64, decimal.Decimal) (*model.Balance, error)
	GetTransactions(context.Context, int64, string) ([]model.Transaction, error)
	// Transfer(ctx context.Context,
	// 	fromUserID int64, toUserID int64, amount decimal.Decimal, currency string) (model.TransferResult, error)
	ConvertBalance(context.Context, *model.Balance, string) (*model.Balance, error)
}

func Exitfail(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
