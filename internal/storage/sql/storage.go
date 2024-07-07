package sqlstorage

import (
	"context"
	"fmt"

	_ "github.com/jackc/pgx/stdlib" // pgx driver
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	dsn string
	db  *sqlx.DB
}

func New(dsn string) *Storage {
	return &Storage{dsn: dsn}
}

func (s *Storage) Connect(ctx context.Context) error {
	db, err := sqlx.Open("pgx", s.dsn)
	if err != nil {
		return fmt.Errorf("failed to load driver: %w", err)
	}
	s.db = db
	err = s.db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}
	return nil
}

func (s *Storage) Close(ctx context.Context) error {
	s.db.Close()
	ctx.Done()
	return nil
}
