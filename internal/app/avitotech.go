package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/cronnoss/avitotech/internal/logger"
	"github.com/cronnoss/avitotech/internal/server"
	"github.com/cronnoss/avitotech/internal/storage"
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
}

type Server interface {
	Start(context.Context) error
	Stop(context.Context) error
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