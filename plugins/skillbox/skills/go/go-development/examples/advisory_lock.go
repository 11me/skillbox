package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

// ============================================================================
// Advisory Lock Pattern Example
// ============================================================================
// This example demonstrates how to use PostgreSQL advisory locks to prevent
// serializable transaction conflicts. The pattern is useful for operations
// that modify shared resources concurrently (wallet transfers, counter
// increments, auction bidding, etc.)
// ============================================================================

// WalletRepository demonstrates advisory lock integration.
type WalletRepository interface {
	GetByUserID(ctx context.Context, userID string) (*Wallet, error)
	Update(ctx context.Context, wallet *Wallet) error
	Serialize(ctx context.Context, label string) error // Advisory lock method
}

type walletRepository struct {
	db QueryExecer
}

func newWalletRepository(db QueryExecer) *walletRepository {
	return &walletRepository{db: db}
}

// Serialize acquires an advisory lock for serializable transactions.
// Use inside WithTx with pgx.Serializable to prevent serialization conflicts.
//
// The lock is automatically released when the transaction commits or rolls back.
// Multiple calls with the same label will block until the lock is released.
func (r *walletRepository) Serialize(ctx context.Context, label string) error {
	query, _, err := squirrel.
		Select("pg_advisory_xact_lock(hashtext(?))").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("build lock query: %w", err)
	}

	if _, err := r.db.Exec(ctx, query, label); err != nil {
		return fmt.Errorf("acquire lock: %w", err)
	}
	return nil
}

func (r *walletRepository) GetByUserID(ctx context.Context, userID string) (*Wallet, error) {
	query, args, err := squirrel.
		Select("id", "user_id", "balance", "currency").
		From("wallets").
		Where(squirrel.Eq{"user_id": userID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	row := r.db.QueryRow(ctx, query, args...)
	w := &Wallet{}
	if err := row.Scan(&w.ID, &w.UserID, &w.Balance, &w.Currency); err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}
	return w, nil
}

func (r *walletRepository) Update(ctx context.Context, wallet *Wallet) error {
	query, args, err := squirrel.
		Update("wallets").
		Set("balance", wallet.Balance).
		Where(squirrel.Eq{"id": wallet.ID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}

// ============================================================================
// Service Layer Example
// ============================================================================

type WalletService struct {
	db      Client // Database client with WithTx
	wallets WalletRepository
}

// TransferFunds demonstrates the Lock-First pattern.
// CRITICAL: Always acquire the lock BEFORE reading data.
func (s *WalletService) TransferFunds(ctx context.Context, fromUserID, toUserID string, amount int64) error {
	return s.db.WithTx(ctx, func(ctx context.Context) error {
		// 1. Acquire advisory lock FIRST
		// Label format: "Operation:resource1:resource2"
		label := fmt.Sprintf("TransferFunds:%s:%s", fromUserID, toUserID)
		if err := s.wallets.Serialize(ctx, label); err != nil {
			return fmt.Errorf("serialize: %w", err)
		}

		// 2. Read current state (AFTER lock acquired)
		from, err := s.wallets.GetByUserID(ctx, fromUserID)
		if err != nil {
			return fmt.Errorf("get sender wallet: %w", err)
		}

		to, err := s.wallets.GetByUserID(ctx, toUserID)
		if err != nil {
			return fmt.Errorf("get recipient wallet: %w", err)
		}

		// 3. Validate business rules
		if from.Balance < amount {
			return errors.New("insufficient funds")
		}

		// 4. Perform updates
		from.Balance -= amount
		to.Balance += amount

		if err := s.wallets.Update(ctx, from); err != nil {
			return fmt.Errorf("update sender: %w", err)
		}
		if err := s.wallets.Update(ctx, to); err != nil {
			return fmt.Errorf("update recipient: %w", err)
		}

		return nil
	}, pgx.Serializable) // Use Serializable isolation
}

// IncrementCounter demonstrates global sequence locking.
// Used when generating sequential IDs (e.g., token IDs, invoice numbers).
func (s *WalletService) IncrementCounter(ctx context.Context, counterName string) (int64, error) {
	var newValue int64

	err := s.db.WithTx(ctx, func(ctx context.Context) error {
		// Lock the counter namespace
		if err := s.wallets.Serialize(ctx, fmt.Sprintf("Counter:%s", counterName)); err != nil {
			return fmt.Errorf("serialize: %w", err)
		}

		// Get current value and increment
		// ... counter logic here ...
		newValue = 42 // placeholder

		return nil
	}, pgx.Serializable)

	return newValue, err
}

// ============================================================================
// Models
// ============================================================================

type Wallet struct {
	ID       string
	UserID   string
	Balance  int64
	Currency string
}

// ============================================================================
// Interfaces (from database-pattern.md)
// ============================================================================

type QueryExecer interface {
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) Row
	Exec(ctx context.Context, sql string, args ...any) (CommandTag, error)
}

type Client interface {
	QueryExecer
	WithTx(ctx context.Context, txFunc func(context.Context) error, isoLvl pgx.TxIsoLevel) error
}

// Placeholder interfaces for example compilation
type Rows interface {
	Close()
	Next() bool
	Scan(dest ...any) error
}

type Row interface {
	Scan(dest ...any) error
}

type CommandTag interface{}
