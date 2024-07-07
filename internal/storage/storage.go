package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/cronnoss/avitotech/internal/model"
	sqlstorage "github.com/cronnoss/avitotech/internal/storage/sql"
)

type Conf struct {
	DB  string `toml:"db"`
	DSN string `toml:"dsn"`
}

type Storage interface {
	Connect(context.Context) error
	Close(context.Context) error
	TopUp(context.Context, *model.Balance) error
}

func NewStorage(conf Conf) Storage {
	if conf.DB == "sql" {
		return sqlstorage.New(conf.DSN)
	}

	fmt.Fprintln(os.Stderr, "wrong DB")
	os.Exit(1)
	return nil
}
