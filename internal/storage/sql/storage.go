package sqlstorage

import (
	"context"
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

func (s *Storage) GetBalance(ctx context.Context, b *model.Balance) (*model.Balance, error) {
	var ans model.Balance
	query := `SELECT * FROM balances WHERE user_id = $1`
	err := s.db.QueryRowxContext(ctx, query, b.UserID).
		Scan(&ans.ID, &ans.UserID, &ans.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return &ans, nil
}

func (s *Storage) TopUp(ctx context.Context, userID int64, amount decimal.Decimal, by string) (*model.Balance, error) {
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
        INSERT INTO balances (user_id, amount) 
        VALUES ($1, $2)
        ON CONFLICT (user_id) DO UPDATE 
        SET amount = balances.amount + EXCLUDED.amount
        RETURNING id, user_id, amount`
	err = s.db.QueryRowxContext(ctx, query, userID, amount).
		Scan(&ans.ID, &ans.UserID, &ans.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to top up: %w", err)
	}

	operation := fmt.Sprintf("Top-up by %s %sRUB", by, amount.StringFixed(2))
	transactionQuery := `
        INSERT INTO transactions (user_id, amount, operation) 
        VALUES ($1, $2, $3)`
	_, err = s.db.ExecContext(ctx, transactionQuery, userID, amount, operation)
	if err != nil {
		return nil, fmt.Errorf("failed to record transaction: %w", err)
	}

	return &ans, nil
}

func (s *Storage) Debit(ctx context.Context, userID int64, amount decimal.Decimal, by string) (*model.Balance, error) {
	var ans model.Balance

	var userExists bool
	err := s.db.QueryRowxContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&userExists)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !userExists {
		return nil, fmt.Errorf("user with ID %d does not exist", userID)
	}

	balance, err := s.GetBalance(ctx, &model.Balance{UserID: userID})
	if err != nil {
		return nil, fmt.Errorf("user has no balance: %w", err)
	}

	newBalance := balance.Amount.Add(amount) // amount is negative, so we add it
	if newBalance.LessThan(decimal.Zero) {
		return nil, fmt.Errorf("insufficient funds")
	}

	query := `
		UPDATE balances 
		SET amount = $2
		WHERE user_id = $1
		RETURNING id, user_id, amount`
	err = s.db.QueryRowxContext(ctx, query, userID, newBalance).
		Scan(&ans.ID, &ans.UserID, &ans.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to debit: %w", err)
	}

	operation := fmt.Sprintf("Debit by %s %sRUB", by, amount.StringFixed(2))
	transactionQuery := `
		INSERT INTO transactions (user_id, amount, operation) 
		VALUES ($1, $2, $3)`
	_, err = s.db.ExecContext(ctx, transactionQuery, userID, amount, operation)
	if err != nil {
		return nil, fmt.Errorf("failed to record transaction: %w", err)
	}

	return &ans, nil
}

func (s *Storage) GetTransactions(ctx context.Context, userID int64) ([]model.Transaction, error) {
	var ans []model.Transaction

	// sorted from oldest to newest
	query := `SELECT * FROM transactions WHERE user_id = $1`
	err := s.db.SelectContext(ctx, &ans, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	return ans, nil
}

func (s *Storage) GetTransactionsByDate(ctx context.Context, userID int64) ([]model.Transaction, error) {
	var ans []model.Transaction

	// sorted from newest to oldest
	query := `SELECT * FROM transactions WHERE user_id = $1 ORDER BY date DESC`
	err := s.db.SelectContext(ctx, &ans, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	return ans, nil
}

func (s *Storage) GetTransactionsByAmount(ctx context.Context, userID int64) ([]model.Transaction, error) {
	var ans []model.Transaction

	// sorted from highest to lowest
	query := `SELECT * FROM transactions WHERE user_id = $1 ORDER BY amount DESC`
	err := s.db.SelectContext(ctx, &ans, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	return ans, nil
}
