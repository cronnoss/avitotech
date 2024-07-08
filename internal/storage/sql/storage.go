package sqlstorage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cronnoss/avitotech/internal/model"
	_ "github.com/jackc/pgx/stdlib" // pgx driver
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
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

func stringNull(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func (s *Storage) GetBalance(ctx context.Context, b *model.Balance) (decimal.Decimal, error) {
	var totalAmount decimal.Decimal
	query := `SELECT SUM(amount) AS total_amount FROM balances WHERE user_id = $1 AND currency = $2 
                                                 GROUP BY user_id, currency`
	err := s.db.QueryRowxContext(ctx, query, b.UserID, stringNull(b.Currency)).Scan(&totalAmount)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get balance: %w", err)
	}

	return totalAmount, nil
}

func (s *Storage) TopUp(ctx context.Context, b *model.Balance) error {
	var userExists bool
	err := s.db.QueryRowxContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", b.UserID).Scan(&userExists)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !userExists {
		return fmt.Errorf("user with ID %d does not exist", b.UserID)
	}

	query := `INSERT INTO balances (user_id, amount, currency) VALUES ($1, $2, $3) RETURNING id`
	rows, err := s.db.QueryxContext(ctx, query, b.UserID, b.Amount, stringNull(b.Currency))
	if err != nil {
		return fmt.Errorf("failed to top up: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&b.UserID)
		if err != nil {
			return fmt.Errorf("failed to rows.Scan: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to rows.Err: %w", err)
	}

	return nil
}
