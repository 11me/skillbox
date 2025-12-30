// Package pg provides PostgreSQL client with transaction support.
//
// This example shows:
// - Client as interface for easy mocking
// - Options pattern for configuration
// - Context-based transaction injection
// - Tx-aware Query/QueryRow/Exec methods
// - Retry with panic recovery
package pg

import (
	"context"
	"fmt"
	"strings"

	"github.com/avast/retry-go"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ---------- Types ----------

// TxFunc is a function that runs within a transaction.
// The context contains the transaction, so all queries
// using this context will automatically use the transaction.
type TxFunc func(context.Context) error

// Client is the database client interface.
// Using interface makes it easy to mock in tests.
type Client interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	WithTx(ctx context.Context, txFunc TxFunc, isoLvl pgx.TxIsoLevel) error
	Close()
}

// ---------- Configuration ----------

// Config holds database connection settings.
type Config struct {
	Host           string
	DBName         string
	User           string
	Password       string
	Port           int32
	SSLMode        string
	MaxConnections int
}

// Option configures the database client.
type Option func(*Config)

// WithHost sets the database host.
func WithHost(host string) Option {
	return func(c *Config) { c.Host = host }
}

// WithDBName sets the database name.
func WithDBName(name string) Option {
	return func(c *Config) { c.DBName = name }
}

// WithUser sets the database user.
func WithUser(user string) Option {
	return func(c *Config) { c.User = user }
}

// WithPassword sets the database password.
func WithPassword(password string) Option {
	return func(c *Config) { c.Password = password }
}

// WithPort sets the database port.
func WithPort(port int32) Option {
	return func(c *Config) { c.Port = port }
}

// WithSSLMode sets the SSL mode.
func WithSSLMode(mode string) Option {
	return func(c *Config) { c.SSLMode = mode }
}

// WithMaxConnections sets the maximum number of connections.
func WithMaxConnections(max int) Option {
	return func(c *Config) { c.MaxConnections = max }
}

// ---------- Client Implementation ----------

type client struct {
	pool *pgxpool.Pool
}

// NewClient creates a new database client.
func NewClient(ctx context.Context, opts ...Option) (Client, error) {
	cfg := &Config{
		Port:           5432,
		SSLMode:        "disable",
		MaxConnections: 100,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	connStr := fmt.Sprintf(
		"user=%s password=%s host=%s port=%d dbname=%s sslmode=%s pool_max_conns=%d",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode, cfg.MaxConnections,
	)

	poolCfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	return &client{pool: pool}, nil
}

// Close closes the database connection pool.
func (c *client) Close() {
	c.pool.Close()
}

// ---------- Transaction Injection ----------

type txCtxKey struct{}

func injectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txCtxKey{}, tx)
}

func extractTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txCtxKey{}).(pgx.Tx)
	return tx, ok
}

// ---------- Tx-Aware Query Methods ----------

// Query executes a query that returns rows.
// If a transaction exists in context, it uses the transaction.
func (c *client) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if tx, ok := extractTx(ctx); ok {
		return tx.Query(ctx, sql, args...)
	}
	return c.pool.Query(ctx, sql, args...)
}

// QueryRow executes a query that returns at most one row.
// If a transaction exists in context, it uses the transaction.
func (c *client) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if tx, ok := extractTx(ctx); ok {
		return tx.QueryRow(ctx, sql, args...)
	}
	return c.pool.QueryRow(ctx, sql, args...)
}

// Exec executes a query that doesn't return rows.
// If a transaction exists in context, it uses the transaction.
func (c *client) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if tx, ok := extractTx(ctx); ok {
		return tx.Exec(ctx, sql, args...)
	}
	return c.pool.Exec(ctx, sql, args...)
}

// ---------- Transaction with Retry ----------

// WithTx executes a function within a transaction.
// The transaction is automatically injected into the context,
// so all queries using that context will use the transaction.
//
// Features:
// - Automatic retry on transient failures (12 attempts)
// - Panic recovery to prevent connection leaks
// - Automatic rollback on error
func (c *client) WithTx(ctx context.Context, txFunc TxFunc, isoLvl pgx.TxIsoLevel) error {
	return retry.Do(
		func() (err error) {
			var conn *pgxpool.Conn

			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("recovered from panic: %v", r)
				}
				if conn != nil {
					conn.Release()
				}
			}()

			conn, err = c.pool.Acquire(ctx)
			if err != nil {
				return fmt.Errorf("acquire connection: %w", err)
			}

			tx, err := conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: isoLvl})
			if err != nil {
				return fmt.Errorf("begin transaction: %w", err)
			}
			defer tx.Rollback(ctx)

			ctx = injectTx(ctx, tx)

			if err = txFunc(ctx); err != nil {
				return err
			}

			if err = tx.Commit(ctx); err != nil {
				return fmt.Errorf("commit transaction: %w", err)
			}

			return nil
		},
		retry.Attempts(12),
		retry.Context(ctx),
		retry.RetryIf(isRetryable),
	)
}

func isRetryable(err error) bool {
	s := err.Error()
	return strings.Contains(s, "i/o timeout") ||
		strings.Contains(s, "unexpected EOF") ||
		strings.Contains(s, "SQLSTATE 08") || // Connection exception
		strings.Contains(s, "SQLSTATE 40") || // Transaction rollback
		strings.Contains(s, "SQLSTATE 53") || // Insufficient resources
		strings.Contains(s, "SQLSTATE 57") || // Operator intervention
		strings.Contains(s, "SQLSTATE 58") || // System error
		strings.Contains(s, "connection refused")
}

// ---------- Usage Example ----------

// Example usage:
//
//	func main() {
//	    ctx := context.Background()
//
//	    client, err := pg.NewClient(ctx,
//	        pg.WithHost("localhost"),
//	        pg.WithPort(5432),
//	        pg.WithDBName("myapp"),
//	        pg.WithUser("postgres"),
//	        pg.WithPassword("secret"),
//	    )
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    defer client.Close()
//
//	    // Simple query (no transaction)
//	    rows, _ := client.Query(ctx, "SELECT id, name FROM users")
//
//	    // With transaction
//	    err = client.WithTx(ctx, func(ctx context.Context) error {
//	        // All queries here automatically use the transaction!
//	        _, err := client.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "John")
//	        if err != nil {
//	            return err
//	        }
//	        _, err = client.Exec(ctx, "INSERT INTO audit_log (action) VALUES ($1)", "user_created")
//	        return err
//	    }, pgx.Serializable)
//	}
