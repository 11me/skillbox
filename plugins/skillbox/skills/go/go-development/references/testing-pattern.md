# Testing Pattern

4-struct composition pattern with sqlmock and testcontainers.

## Table-Driven Test Pattern

```go
type meta struct {
    name    string
    enabled bool
}

type fields struct {
    setupDatabase func(dbMock sqlmock.Sqlmock)
}

type args struct {
    ctx  context.Context
    user *User
}

type wants struct {
    err error
}

func TestService_Create(t *testing.T) {
    t.Parallel()

    tests := []struct {
        meta   meta
        fields fields
        args   args
        wants  wants
    }{
        {
            meta: meta{name: "success", enabled: true},
            fields: fields{
                setupDatabase: func(dbMock sqlmock.Sqlmock) {
                    dbMock.ExpectBegin()
                    dbMock.ExpectQuery(regexp.QuoteMeta(insertQuery)).
                        WithArgs(sqlmock.AnyArg(), "test", "test@example.com").
                        WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
                    dbMock.ExpectCommit()
                },
            },
            args: args{
                ctx:  context.Background(),
                user: &User{Name: "test", Email: "test@example.com"},
            },
            wants: wants{err: nil},
        },
        {
            meta: meta{name: "validation error", enabled: true},
            fields: fields{
                setupDatabase: func(dbMock sqlmock.Sqlmock) {},
            },
            args: args{
                ctx:  context.Background(),
                user: &User{Name: "", Email: "test@example.com"},
            },
            wants: wants{err: ValidationFailed("name is required")},
        },
    }

    for _, tt := range tests {
        tt := tt  // Capture for parallel
        t.Run(tt.meta.name, func(t *testing.T) {
            t.Parallel()
            if !tt.meta.enabled {
                t.SkipNow()
            }

            db, dbMock, err := sqlmock.New()
            require.NoError(t, err)
            tt.fields.setupDatabase(dbMock)

            svc := NewService(NewMockStorage(db))
            err = svc.Create(tt.args.ctx, tt.args.user)

            assert.Equal(t, tt.wants.err, err)
            assert.NoError(t, dbMock.ExpectationsWereMet())
        })
    }
}
```

## Database Integration Tests (TestMain)

```go
package repository_test

import (
    "context"
    "fmt"
    "log"
    "os"
    "testing"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

var pgConnURL string

func TestMain(m *testing.M) {
    var code int
    func() {
        url, cleanup, err := setupPostgres()
        if err != nil {
            log.Fatalf("Setup: %v", err)
        }
        defer cleanup()

        pgConnURL = url
        if err := applyMigrations(url); err != nil {
            log.Fatalf("Migrations: %v", err)
        }
        code = m.Run()
    }()
    os.Exit(code)
}

func setupPostgres() (string, func(), error) {
    ctx := context.Background()
    req := testcontainers.ContainerRequest{
        Image:        "postgres:alpine",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_USER":     "test",
            "POSTGRES_PASSWORD": "test",
            "POSTGRES_DB":       "test",
        },
        WaitingFor: wait.ForListeningPort("5432/tcp"),
    }

    container, err := testcontainers.GenericContainer(ctx,
        testcontainers.GenericContainerRequest{
            ContainerRequest: req,
            Started:          true,
        })
    if err != nil {
        return "", nil, err
    }

    port, _ := container.MappedPort(ctx, "5432")
    url := fmt.Sprintf("postgres://test:test@127.0.0.1:%d/test", port.Int())

    cleanup := func() { container.Terminate(ctx) }
    return url, cleanup, nil
}

func connectDB(t *testing.T) *pgxpool.Pool {
    t.Helper()
    pool, err := pgxpool.New(context.Background(), pgConnURL)
    require.NoError(t, err)
    t.Cleanup(func() { pool.Close() })
    return pool
}
```

## Mock Interface

```go
type MockStorage struct {
    db *sql.DB
}

func NewMockStorage(db *sql.DB) *MockStorage {
    return &MockStorage{db: db}
}

func (m *MockStorage) Users() UserRepository {
    return &mockUserRepo{db: m.db}
}
```

## Best Practices

- `t.Parallel()` — always for independent tests
- `tt := tt` — capture loop variable
- `t.Helper()` — in helper functions
- `t.Cleanup()` — for resource cleanup
- `require.NoError()` — for critical checks
- `assert.Equal()` — for assertions
- `regexp.QuoteMeta()` — for SQL queries
- `enabled` field — skip tests without deletion
