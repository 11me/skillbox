# Database Pattern

Using `pgx/v5` + `squirrel` with context-based transaction injection.

## Client Interface

```go
package pg

import (
    "context"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
)

type TxFunc func(context.Context) error

// Client is the database client interface.
// Using interface makes it easy to mock in tests.
type Client interface {
    Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
    QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
    Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
    WithTx(ctx context.Context, txFunc TxFunc, isoLvl pgx.TxIsoLevel) error
}
```

**Key feature:** Interface-based design for easy testing with mocks.

## Client Implementation

```go
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

type client struct {
    pool *pgxpool.Pool
}

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

func (c *client) Close() {
    c.pool.Close()
}
```

## Options Pattern

```go
type Config struct {
    Host           string
    DBName         string
    User           string
    Password       string
    Port           int32
    SSLMode        string
    MaxConnections int
}

type Option func(*Config)

func WithHost(host string) Option {
    return func(c *Config) { c.Host = host }
}

func WithDBName(name string) Option {
    return func(c *Config) { c.DBName = name }
}

func WithUser(user string) Option {
    return func(c *Config) { c.User = user }
}

func WithPassword(password string) Option {
    return func(c *Config) { c.Password = password }
}

func WithPort(port int32) Option {
    return func(c *Config) { c.Port = port }
}

func WithSSLMode(mode string) Option {
    return func(c *Config) { c.SSLMode = mode }
}

func WithMaxConnections(max int) Option {
    return func(c *Config) { c.MaxConnections = max }
}
```

## Transaction Injection

```go
type txCtxKey struct{}

func injectTx(ctx context.Context, tx pgx.Tx) context.Context {
    return context.WithValue(ctx, txCtxKey{}, tx)
}

func extractTx(ctx context.Context) (pgx.Tx, bool) {
    tx, ok := ctx.Value(txCtxKey{}).(pgx.Tx)
    return tx, ok
}
```

## Tx-Aware Query Methods

```go
func (c *client) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
    if tx, ok := extractTx(ctx); ok {
        return tx.Query(ctx, sql, args...)
    }
    return c.pool.Query(ctx, sql, args...)
}

func (c *client) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
    if tx, ok := extractTx(ctx); ok {
        return tx.QueryRow(ctx, sql, args...)
    }
    return c.pool.QueryRow(ctx, sql, args...)
}

func (c *client) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
    if tx, ok := extractTx(ctx); ok {
        return tx.Exec(ctx, sql, args...)
    }
    return c.pool.Exec(ctx, sql, args...)
}
```

**Key feature:** Methods automatically use transaction from context if present.

## WithTx with Retry and Panic Recovery

```go
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
```

**Key features:**
- Panic recovery prevents connection leaks
- Extended SQLSTATE codes for better resilience
- 12 retry attempts for transient failures

## Storage Layer Abstraction

```go
package storage

import (
    "context"

    "github.com/jackc/pgx/v5"
    "myapp/pkg/pg"
)

type Storage interface {
    Users() Users
    Orders() Orders

    ExecReadCommitted(context.Context, pg.TxFunc) error
    ExecRepeatableRead(context.Context, pg.TxFunc) error
    ExecSerializable(context.Context, pg.TxFunc) error
}

type store struct {
    client pg.Client
}

func NewStorage(ctx context.Context, opts ...pg.Option) (Storage, error) {
    client, err := pg.NewClient(ctx, opts...)
    if err != nil {
        return nil, err
    }
    return &store{client: client}, nil
}

func (s *store) Users() Users {
    return &userStorage{client: s.client}
}

func (s *store) Orders() Orders {
    return &orderStorage{client: s.client}
}

func (s *store) ExecReadCommitted(ctx context.Context, f pg.TxFunc) error {
    return s.client.WithTx(ctx, f, pgx.ReadCommitted)
}

func (s *store) ExecRepeatableRead(ctx context.Context, f pg.TxFunc) error {
    return s.client.WithTx(ctx, f, pgx.RepeatableRead)
}

func (s *store) ExecSerializable(ctx context.Context, f pg.TxFunc) error {
    return s.client.WithTx(ctx, f, pgx.Serializable)
}
```

## Repository Pattern

```go
type Users interface {
    Save(ctx context.Context, users ...*User) error
    FindByID(ctx context.Context, id uuid.UUID) (*User, error)
    Find(ctx context.Context, filter *UserFilter) ([]*User, error)
}

type userStorage struct {
    client pg.Client
}

func (s *userStorage) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
    sql, args, err := sq.
        Select("id", "name", "email", "created_at").
        From("users").
        Where(sq.Eq{"id": id}).
        PlaceholderFormat(sq.Dollar).
        ToSql()
    if err != nil {
        return nil, err
    }

    rows, err := s.client.Query(ctx, sql, args...)
    if err != nil {
        return nil, fmt.Errorf("query user: %w", err)
    }

    user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("collect user: %w", err)
    }

    return &user, nil
}
```

## Service Usage

```go
func (svc *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    var user *User

    err := svc.storage.ExecSerializable(ctx, func(ctx context.Context) error {
        // All queries inside this function automatically use the transaction!
        existing, err := svc.storage.Users().FindByEmail(ctx, req.Email)
        if err != nil && !errors.Is(err, ErrUserNotFound) {
            return fmt.Errorf("check existing: %w", err)
        }
        if existing != nil {
            return ErrEmailAlreadyExists
        }

        user = &User{
            ID:    uuid.New(),
            Name:  req.Name,
            Email: req.Email,
        }

        return svc.storage.Users().Save(ctx, user)
    })

    if err != nil {
        return nil, err
    }

    return user, nil
}
```

## Row Scanning with pgx.CollectRows

```go
// Single row
user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])

// Multiple rows
users, err := pgx.CollectRows(rows, pgx.RowToStructByName[User])

// With mapper (for encryption/transformation)
mappers, err := pgx.CollectRows(rows, pgx.RowToStructByName[UserMapper])
users := make([]*User, 0, len(mappers))
for _, m := range mappers {
    users = append(users, m.ToModel())
}
```

## Best Practices

### DO:
- ✅ Use `Client` interface for testability
- ✅ Use Options pattern for configuration
- ✅ Use `pgx.CollectRows` for row scanning
- ✅ Use squirrel with `PlaceholderFormat(sq.Dollar)`
- ✅ Wrap operations in transactions via `ExecSerializable`
- ✅ Handle panic in `WithTx` to prevent connection leaks

### DON'T:
- ❌ Expose `*pgxpool.Pool` directly
- ❌ Forget panic recovery in transaction wrapper
- ❌ Use `QueryRow().Scan()` — prefer `pgx.CollectOneRow`
- ❌ Mix transaction and non-transaction calls in same function

## Related

- [repository-pattern.md](repository-pattern.md) — Repository pattern
- [service-pattern.md](service-pattern.md) — Service layer
- [tracing-pattern.md](tracing-pattern.md) — Database tracing with otelpgx
