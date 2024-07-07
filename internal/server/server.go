package server

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/cronnoss/avitotech/internal/model"
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
	// GetBalance(ctx context.Context, userID int64, currency string) (decimal.Decimal, error)
	// GetTransactions(ctx context.Context, userID int64, currency string, sort string) ([]model.Transaction, error)
	TopUp(context.Context, *model.Balance) error
	// Debit(ctx context.Context, userID int64, amount decimal.Decimal, currency string) error
	// Transfer(ctx context.Context,
	// 	fromUserID int64, toUserID int64, amount decimal.Decimal, currency string) (model.TransferResult, error)
}

func Exitfail(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
