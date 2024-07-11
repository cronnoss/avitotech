package sqlstorage

import (
	"context"
	"testing"

	"github.com/cronnoss/avitotech/internal/model"
	"github.com/shopspring/decimal"
	sqlmock "github.com/zhashkevych/go-sqlxmock"
)

const testDSN = "sqlmock_db_0"

func TestStorage_GetBalance(t *testing.T) {
	s := New(testDSN)
	if s == nil {
		t.Error("New() should not return nil")
	}
	db, mock, err := sqlmock.Newx()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	// Set the mock database to the Storage object
	s.db = db
	defer db.Close()

	type args struct {
		b *model.Balance
	}

	type mockBehavior func(args args)

	tests := []struct {
		name    string
		mock    mockBehavior
		input   args
		want    *model.Balance
		wantErr bool
	}{
		{
			name: "OK",
			mock: func(args args) {
				// Mocking the balance retrieval
				mock.ExpectQuery("^SELECT \\* FROM balances WHERE user_id = \\$1$").
					WithArgs(args.b.UserID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "amount"}).
						AddRow(1, args.b.UserID, decimal.NewFromFloat(100.00)))
			},
			input: args{
				b: &model.Balance{
					UserID: 1,
				},
			},
			want: &model.Balance{
				ID:     1,
				UserID: 1,
				Amount: decimal.NewFromFloat(100.00),
			},
			wantErr: false,
		},
		{
			name: "Not found",
			mock: func(args args) {
				// Mocking the balance retrieval
				mock.ExpectQuery("^SELECT \\* FROM balances WHERE user_id = \\$1$").
					WithArgs(args.b.UserID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "amount"}).
						AddRow(nil, nil, nil))
			},
			input: args{
				b: &model.Balance{
					UserID: 1,
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock(tt.input)
			got, err := s.GetBalance(context.Background(), tt.input.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.GetBalance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if got.Amount.Cmp(tt.want.Amount) != 0 {
					t.Errorf("Storage.GetBalance() = %v, want %v", got, tt.want)
				}
			}

			// Ensure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestStorage_TopUp(t *testing.T) {
	s := New(testDSN)
	if s == nil {
		t.Error("New() should not return nil")
	}
	db, mock, err := sqlmock.Newx()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	s.db = db
	defer db.Close()

	type args struct {
		userID int64
		amount decimal.Decimal
	}

	type mockBehavior func(args args)

	tests := []struct {
		name    string
		mock    mockBehavior
		input   args
		want    *model.Balance
		wantErr bool
	}{
		{
			name: "OK",
			mock: func(args args) {
				// Mocking the user existence check
				mock.ExpectQuery("SELECT EXISTS").
					WithArgs(args.userID).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

				// Mocking the balance update
				mock.ExpectQuery("INSERT INTO balances").
					WithArgs(args.userID, args.amount).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "amount"}).
						AddRow(1, args.userID, args.amount))

				// Mocking the transaction recording
				mock.ExpectExec("INSERT INTO transactions").
					WithArgs(args.userID, args.amount, sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			input: args{
				userID: 1,
				amount: decimal.NewFromFloat(100.00),
			},
			want: &model.Balance{
				ID:     1,
				UserID: 1,
				Amount: decimal.NewFromFloat(100.00),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock(tt.input)
			got, err := s.TopUp(context.Background(), tt.input.userID, tt.input.amount, "some by")
			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.TopUp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if got.ID != tt.want.ID {
					t.Errorf("Storage.TopUp() ID = %v, want %v", got.ID, tt.want.ID)
				}
				if got.UserID != tt.want.UserID {
					t.Errorf("Storage.TopUp() UserID = %v, want %v", got.UserID, tt.want.UserID)
				}
				if got.Amount.Cmp(tt.want.Amount) != 0 {
					t.Errorf("Storage.TopUp() Amount = %v, want %v", got.Amount, tt.want.Amount)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestStorage_Debit(t *testing.T) {
	s := New(testDSN)
	if s == nil {
		t.Error("New() should not return nil")
	}
	db, mock, err := sqlmock.Newx()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	s.db = db
	defer db.Close()

	type args struct {
		userID int64
		amount decimal.Decimal
	}

	type mockBehavior func(args args)

	tests := []struct {
		name    string
		mock    mockBehavior
		input   args
		want    *model.Balance
		wantErr bool
	}{
		{
			name: "OK",
			mock: func(args args) {
				// Mocking the user existence check
				mock.ExpectQuery("SELECT EXISTS").
					WithArgs(args.userID).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

				// Mocking the balance retrieval
				mock.ExpectQuery("^SELECT \\* FROM balances WHERE user_id = \\$1$").
					WithArgs(args.userID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "amount"}).
						AddRow(1, args.userID, decimal.NewFromFloat(10)))

				// Mocking the balance update
				mock.ExpectQuery("UPDATE balances").
					WithArgs(args.userID, decimal.NewFromFloat(5)).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "amount"}).
						AddRow(1, args.userID, decimal.NewFromFloat(5)))

				// Mocking the transaction recording
				mock.ExpectExec("INSERT INTO transactions").
					WithArgs(args.userID, args.amount, sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			input: args{
				userID: 1,
				amount: decimal.NewFromFloat(-5), // Debiting 5 from the balance
			},
			want: &model.Balance{
				ID:     1,
				UserID: 1,
				Amount: decimal.NewFromFloat(5),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock(tt.input)
			got, err := s.Debit(context.Background(), tt.input.userID, tt.input.amount, "some by")
			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.Debit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if got.ID != tt.want.ID {
					t.Errorf("Storage.Debit() ID = %v, want %v", got.ID, tt.want.ID)
				}
				if got.UserID != tt.want.UserID {
					t.Errorf("Storage.Debit() UserID = %v, want %v", got.UserID, tt.want.UserID)
				}
				if got.Amount.Cmp(tt.want.Amount) != 0 {
					t.Errorf("Storage.Debit() Amount = %v, want %v", got.Amount, tt.want.Amount)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestStorage_GetTransactions(t *testing.T) {
	s := New(testDSN)
	if s == nil {
		t.Error("New() should not return nil")
	}
	db, mock, err := sqlmock.Newx()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	s.db = db
	defer db.Close()

	type args struct {
		userID int64
	}

	type mockBehavior func(args args)

	tests := []struct {
		name    string
		mock    mockBehavior
		input   args
		want    []model.Transaction
		wantErr bool
	}{
		{
			name: "OK",
			mock: func(args args) {
				// Mocking the transaction retrieval
				mock.ExpectQuery("^SELECT \\* FROM transactions WHERE user_id = \\$1$").
					WithArgs(args.userID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "amount", "operation"}).
						AddRow(1, args.userID, decimal.NewFromFloat(10), "Top-up by some by 10.00RUB"))
			},
			input: args{
				userID: 1,
			},
			want: []model.Transaction{
				{
					ID:        1,
					UserID:    1,
					Amount:    decimal.NewFromFloat(10),
					Operation: "Top-up by some by 10.00RUB",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock(tt.input)
			got, err := s.GetTransactions(context.Background(), tt.input.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.GetTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if len(got) != len(tt.want) {
					t.Errorf("Storage.GetTransactions() = %v, want %v", got, tt.want)
				}
				for i := range got {
					switch {
					case got[i].ID != tt.want[i].ID:
						t.Errorf("Storage.GetTransactions() ID = %v, want %v", got[i].ID, tt.want[i].ID)
					case got[i].UserID != tt.want[i].UserID:
						t.Errorf("Storage.GetTransactions() UserID = %v, want %v", got[i].UserID, tt.want[i].UserID)
					case got[i].Amount.Cmp(tt.want[i].Amount) != 0:
						t.Errorf("Storage.GetTransactions() Amount = %v, want %v", got[i].Amount, tt.want[i].Amount)
					case got[i].Operation != tt.want[i].Operation:
						t.Errorf("Storage.GetTransactions() Operation = %v, want %v", got[i].Operation, tt.want[i].Operation)
					}
				}
			}
		})
	}
}
