package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/cronnoss/avitotech/internal/logger"
	"github.com/cronnoss/avitotech/internal/model"
	"github.com/cronnoss/avitotech/internal/server"
	"github.com/cronnoss/avitotech/internal/storage"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

type AvitotechConf struct {
	Logger  logger.Conf  `toml:"logger"`
	Storage storage.Conf `toml:"storage"`
	HTTP    struct {
		Host string `toml:"host"`
		Port string `toml:"port"`
	} `toml:"http-server"`
}

type Avitotech struct {
	conf    AvitotechConf
	log     server.Logger
	storage Storage
}

type Storage interface {
	Connect(context.Context) error
	Close(context.Context) error
	GetBalance(context.Context, *model.Balance) (*model.Balance, error)
	TopUp(context.Context, int64, decimal.Decimal, string) (*model.Balance, error)
	Debit(context.Context, int64, decimal.Decimal, string) (*model.Balance, error)
	GetTransactions(context.Context, int64) ([]model.Transaction, error)
	GetTransactionsByDate(context.Context, int64) ([]model.Transaction, error)
	GetTransactionsByAmount(context.Context, int64) ([]model.Transaction, error)
}

type Server interface {
	Start(context.Context) error
	Stop(context.Context) error
}

var (
	purchase = "purchase"
	bankCard = "bank_card"
	transfer = "transfer"
)

func (a *Avitotech) GetBalance(ctx context.Context, b *model.Balance) (*model.Balance, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return a.storage.GetBalance(ctx, b)
}

func (a *Avitotech) TopUp(ctx context.Context, userID int64, amount decimal.Decimal) (*model.Balance, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return a.storage.TopUp(ctx, userID, amount, bankCard)
}

func (a *Avitotech) Debit(ctx context.Context, userID int64, amount decimal.Decimal) (*model.Balance, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return a.storage.Debit(ctx, userID, amount.Neg(), purchase)
}

func (a *Avitotech) GetTransactions(ctx context.Context, userID int64, sort string) ([]model.Transaction, error) {
	switch sort {
	case "":
		return a.storage.GetTransactions(ctx, userID)
	case "date":
		return a.storage.GetTransactionsByDate(ctx, userID)
	case "amount":
		return a.storage.GetTransactionsByAmount(ctx, userID)
	default:
		return nil, errors.New("wrong sort")
	}
}

func (a *Avitotech) Transfer(ctx context.Context, fromID, toID int64, amount decimal.Decimal) (*model.Balance, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("amount must be greater than zero")
	}

	balance, err := a.storage.GetBalance(ctx, &model.Balance{UserID: fromID})
	if err != nil {
		return nil, err
	}

	if balance.Amount.LessThan(amount) {
		return nil, errors.New("insufficient funds")
	}

	_, err = a.storage.Debit(ctx, fromID, amount.Neg(), transfer)
	if err != nil {
		return nil, err
	}

	ans, err := a.storage.TopUp(ctx, toID, amount, transfer)
	if err != nil {
		return nil, err
	}

	return ans, nil
}

func (a *Avitotech) ConvertBalance(ctx context.Context, b *model.Balance, currency string) (*model.Balance, error) {
	type Currency struct {
		Base  string `json:"base"` // base = EUR
		Rates struct {
			Rub float64 `json:"RUB"`
			Usd float64 `json:"USD"`
		} `json:"rates"`
	}

	var cur Currency

	endpoint := "http://api.exchangeratesapi.io/v1/latest?access_key=e532701035ed3f4040b2660e6b7a8a3d"
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&cur)
	if err != nil {
		return nil, errors.New("failed to decode response from exchangeratesapi.io")
	}

	// balance in eur, because it is the base in api.exchangeratesapi.io
	beur := b.Amount.Div(decimal.NewFromFloat(cur.Rates.Rub)) // b.Amount / cur.Rates.Rub

	switch currency {
	case "USD":
		busd := beur.Mul(decimal.NewFromFloat(cur.Rates.Usd)) // beur * cur.Rates.Usd
		b.Amount = busd
	case "EUR":
		b.Amount = beur
	}
	return b, nil
}

func (a *Avitotech) Close(ctx context.Context) error {
	a.log.Infof("App closed\n")
	return a.storage.Close(ctx)
}

func NewAvitotech(log server.Logger, conf AvitotechConf, storage Storage) *Avitotech {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := storage.Connect(ctx); err != nil {
		server.Exitfail(fmt.Sprintf("Can't connect to storage:%v", err))
	}

	return &Avitotech{log: log, conf: conf, storage: storage}
}

func (a Avitotech) Run(httpsrv Server) {
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	g, ctxEG := errgroup.WithContext(ctx)

	func1 := func() error {
		return httpsrv.Start(ctxEG)
	}

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := httpsrv.Stop(ctx); err != nil {
			if !errors.Is(err, http.ErrServerClosed) &&
				!errors.Is(err, context.Canceled) {
				a.log.Errorf("failed to stop HTTP-server:%v\n", err)
			}
		}

		if err := a.storage.Close(ctx); err != nil {
			a.log.Errorf("failed to close db:%v\n", err)
		}
	}()

	g.Go(func1)

	if err := g.Wait(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) &&
			!errors.Is(err, context.Canceled) {
			a.log.Errorf("%v\n", err)
		}
	}
}
