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

func (s *Storage) GetBalance(ctx context.Context, b *model.Balance) (*model.Balance, error) {
	var ans model.Balance
	query := `SELECT * FROM balances WHERE user_id = $1 AND currency = $2`
	err := s.db.QueryRowxContext(ctx, query, b.UserID, stringNull(b.Currency)).
		Scan(&ans.ID, &ans.UserID, &ans.Amount, &ans.Currency)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return &ans, nil
}

func (s *Storage) TopUp(
	ctx context.Context,
	userID int64,
	amount decimal.Decimal,
	currency string,
	by string,
) (*model.Balance, error) {
	var ans model.Balance

	var userExists bool
	err := s.db.QueryRowxContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&userExists)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !userExists {
		return nil, fmt.Errorf("user with ID %d does not exist", userID)
	}

	query := `
        INSERT INTO balances (user_id, amount, currency) 
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id) DO UPDATE 
        SET amount = balances.amount + EXCLUDED.amount, 
            currency = EXCLUDED.currency
        RETURNING id, user_id, amount, currency`
	err = s.db.QueryRowxContext(ctx, query, userID, amount, currency).
		Scan(&ans.ID, &ans.UserID, &ans.Amount, &ans.Currency)
	if err != nil {
		return nil, fmt.Errorf("failed to top up: %w", err)
	}

	operation := fmt.Sprintf("Top-up by %s %sRUB", by, amount.StringFixed(2))
	transactionQuery := `
        INSERT INTO transactions (user_id, amount, currency, operation) 
        VALUES ($1, $2, $3, $4)`
	_, err = s.db.ExecContext(ctx, transactionQuery, userID, amount, currency, operation)
	if err != nil {
		return nil, fmt.Errorf("failed to record transaction: %w", err)
	}

	return &ans, nil
}
