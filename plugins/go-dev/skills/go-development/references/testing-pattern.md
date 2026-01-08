# Testing Pattern

Production-grade testing patterns with testcontainers and testify.

## Testing Philosophy

| Layer | Test Type | Database | Mock What | Test Data |
|-------|-----------|----------|-----------|-----------|
| **Repository** | Integration | Real DB (testcontainers) | Nothing | SQL fixtures (testmigration/) |
| **Service** | Unit | Mock repository | Repository | In-test setup |
| **Handler** | Unit | Mock service | Service | In-test setup |

## IMPORTANT: Repository Testing Rules

```
❌ DON'T mock SQL queries in repository tests
❌ DON'T use go-sqlmock for repository layer
❌ DON'T write fake implementations of database

✅ DO use testcontainers with real PostgreSQL
✅ DO test actual SQL queries against real database
✅ DO use SQL fixtures for complex test scenarios
✅ DO verify data is correctly stored and retrieved
```

**Why:** Repository layer's job is to talk to database. Mocking SQL queries tests the mock, not the repository logic. Real database tests catch:
- SQL syntax errors
- Type mismatches
- Constraint violations
- Transaction behavior
- Query performance issues

## Typed Closer Pattern

Use typed `closer` for proper error handling during resource cleanup:

```go
// closer is a cleanup function that returns an error.
type closer func() error

var defCloser = func() error { return nil }
```

**Why:** Container termination can fail. Typed closer allows proper error logging.

## TestMain with CI/Local Detection

```go
package storage_test

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "os"
    "testing"
    "time"

    _ "github.com/jackc/pgx/v5/stdlib" // pgx driver for database/sql
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/pressly/goose/v3"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

var pgConnURL string

type closer func() error
var defCloser = func() error { return nil }

func TestMain(m *testing.M) {
    var code int

    func() {
        var (
            pgCloser closer = defCloser
            err      error
        )

        defer func() {
            if err := pgCloser(); err != nil {
                log.Printf("Failed to close postgres: %v", err)
            }
        }()

        if os.Getenv("CI") == "true" {
            pgConnURL = os.Getenv("DATABASE_URL")
            if pgConnURL == "" {
                log.Fatal("DATABASE_URL is required in CI")
            }
        } else {
            pgConnURL, pgCloser, err = runLocalPostgres()
            if err != nil {
                log.Fatalf("Failed to start PostgreSQL: %v", err)
            }
        }

        if err := applyMigrations(pgConnURL); err != nil {
            log.Fatalf("Failed to apply migrations: %v", err)
        }

        if err := applyTestData(pgConnURL); err != nil {
            log.Fatalf("Failed to apply test data: %v", err)
        }

        code = m.Run()
    }()

    os.Exit(code)
}

func runLocalPostgres() (string, closer, error) {
    ctx := context.Background()

    req := testcontainers.ContainerRequest{
        Image:        "postgres:16-alpine",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_USER":     "test",
            "POSTGRES_PASSWORD": "test",
            "POSTGRES_DB":       "test",
        },
        WaitingFor: wait.ForListeningPort("5432/tcp").
            WithStartupTimeout(60 * time.Second),
    }

    container, err := testcontainers.GenericContainer(ctx,
        testcontainers.GenericContainerRequest{
            ContainerRequest: req,
            Started:          true,
        })
    if err != nil {
        return "", defCloser, fmt.Errorf("start container: %w", err)
    }

    host, err := container.Host(ctx)
    if err != nil {
        _ = container.Terminate(ctx)
        return "", defCloser, fmt.Errorf("get host: %w", err)
    }

    port, err := container.MappedPort(ctx, "5432")
    if err != nil {
        _ = container.Terminate(ctx)
        return "", defCloser, fmt.Errorf("get port: %w", err)
    }

    url := fmt.Sprintf("postgres://test:test@%s:%s/test?sslmode=disable",
        host, port.Port())

    cleanup := func() error {
        log.Println("Terminating PostgreSQL container...")
        return container.Terminate(ctx)
    }

    return url, cleanup, nil
}
```

## Test Migrations (testmigration/)

SQL fixtures for repository tests. Uses goose with a separate version table.

### Directory Structure

```
internal/storage/
├── main_test.go
├── user_test.go
├── order_test.go
└── testmigration/
    ├── 100001_users_dataset.up.sql
    ├── 100001_users_dataset.down.sql
    ├── 100002_orders_dataset.up.sql
    └── 100002_orders_dataset.down.sql
```

### Naming Convention

| Range | Purpose | Goose Table |
|-------|---------|-------------|
| 000001-099999 | Production schema | `goose_db_version` |
| 100001-199999 | Test data fixtures | `goose_db_test_version` |

### Apply Migrations

```go
func applyMigrations(dbURL string) error {
    migrationsDir := "../../migrations"

    if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
        log.Printf("Migrations directory not found: %s (skipping)", migrationsDir)
        return nil
    }

    db, err := sql.Open("pgx", dbURL)
    if err != nil {
        return fmt.Errorf("open database: %w", err)
    }
    defer db.Close()

    goose.SetBaseFS(nil)
    if err := goose.SetDialect("postgres"); err != nil {
        return fmt.Errorf("set dialect: %w", err)
    }

    goose.SetTableName("goose_db_version")
    return goose.Up(db, migrationsDir)
}

func applyTestData(dbURL string) error {
    testDataDir := "testmigration"

    if _, err := os.Stat(testDataDir); os.IsNotExist(err) {
        return nil // No fixtures - this is fine
    }

    db, err := sql.Open("pgx", dbURL)
    if err != nil {
        return fmt.Errorf("open database: %w", err)
    }
    defer db.Close()

    goose.SetBaseFS(nil)
    if err := goose.SetDialect("postgres"); err != nil {
        return fmt.Errorf("set dialect: %w", err)
    }

    // Separate table for test data
    goose.SetTableName("goose_db_test_version")
    return goose.Up(db, testDataDir)
}
```

### Example Fixture

```sql
-- testmigration/100001_users_dataset.up.sql
INSERT INTO users (id, name, email, status, created_at) VALUES
    ('11111111-1111-1111-1111-111111111111', 'Alice', 'alice@test.local', 'active', NOW()),
    ('22222222-2222-2222-2222-222222222222', 'Bob', 'bob@test.local', 'active', NOW()),
    ('33333333-3333-3333-3333-333333333333', 'Inactive', 'inactive@test.local', 'inactive', NOW());

-- testmigration/100001_users_dataset.down.sql
DELETE FROM users WHERE email LIKE '%@test.local';
```

## Test Helpers

```go
// connectDB creates a pool with test-appropriate settings.
func connectDB(t *testing.T) *pgxpool.Pool {
    t.Helper()

    cfg, err := pgxpool.ParseConfig(pgConnURL)
    require.NoError(t, err)

    cfg.MaxConns = 10
    cfg.MinConns = 2
    cfg.MaxConnLifetime = 5 * time.Minute
    cfg.MaxConnIdleTime = 1 * time.Minute

    pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
    require.NoError(t, err)

    t.Cleanup(func() { pool.Close() })

    return pool
}

// truncateTable removes all data from a table.
func truncateTable(t *testing.T, pool *pgxpool.Pool, table string) {
    t.Helper()

    _, err := pool.Exec(context.Background(),
        fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
    require.NoError(t, err)
}
```

## Repository Tests with Fixtures

```go
// Tests use known IDs from fixtures
func TestUserRepository_GetByID(t *testing.T) {
    t.Parallel()

    pool := connectDB(t)
    repo := storage.NewUserRepository(pool)

    tests := []struct {
        name    string
        id      string
        wantErr bool
    }{
        {name: "alice", id: "11111111-1111-1111-1111-111111111111", wantErr: false},
        {name: "bob", id: "22222222-2222-2222-2222-222222222222", wantErr: false},
        {name: "non-existing", id: "99999999-9999-9999-9999-999999999999", wantErr: true},
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            found, err := repo.GetByID(context.Background(), tt.id)
            if tt.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
                assert.Equal(t, tt.id, found.ID)
            }
        })
    }
}

func TestUserRepository_FindByStatus(t *testing.T) {
    t.Parallel()

    pool := connectDB(t)
    repo := storage.NewUserRepository(pool)

    // Known from fixtures: 2 active, 1 inactive
    users, err := repo.FindByStatus(context.Background(), "active")
    require.NoError(t, err)
    assert.Len(t, users, 2)

    inactive, err := repo.FindByStatus(context.Background(), "inactive")
    require.NoError(t, err)
    assert.Len(t, inactive, 1)
    assert.Equal(t, "33333333-3333-3333-3333-333333333333", inactive[0].ID)
}
```

## Service Tests (Testify Mock)

```go
// MockUserRepository implements UserRepository using testify/mock.
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
    args := m.Called(ctx, user)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
    args := m.Called(ctx, email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.User), args.Error(1)
}

func TestUserService_Create(t *testing.T) {
    t.Parallel()

    ctx := context.Background()

    tests := []struct {
        name      string
        inputName string
        email     string
        setupMock func(*MockUserRepository)
        wantErr   error
    }{
        {
            name:      "success",
            inputName: "Test User",
            email:     "test@example.com",
            setupMock: func(m *MockUserRepository) {
                m.On("GetByEmail", mock.Anything, "test@example.com").
                    Return(nil, errs.ErrNotFound)
                m.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).
                    Return(&models.User{ID: uuid.NewString()}, nil)
            },
            wantErr: nil,
        },
        {
            name:      "validation error",
            inputName: "",
            email:     "test@example.com",
            setupMock: func(m *MockUserRepository) {},
            wantErr:   errs.ErrValidation,
        },
        {
            name:      "email exists",
            inputName: "User",
            email:     "existing@example.com",
            setupMock: func(m *MockUserRepository) {
                m.On("GetByEmail", mock.Anything, "existing@example.com").
                    Return(&models.User{ID: uuid.NewString()}, nil)
            },
            wantErr: errs.ErrConflict,
        },
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            mockRepo := new(MockUserRepository)
            tt.setupMock(mockRepo)

            svc := services.NewUserService(mockRepo, nil)
            user, err := svc.Create(ctx, tt.inputName, tt.email)

            if tt.wantErr != nil {
                require.ErrorIs(t, err, tt.wantErr)
            } else {
                require.NoError(t, err)
                require.NotNil(t, user)
            }

            mockRepo.AssertExpectations(t)
        })
    }
}
```

## Handler Tests (Table-Driven)

```go
func TestUserHandler_Create(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name       string
        body       string
        setupMock  func(*MockUserService)
        wantStatus int
    }{
        {
            name: "success",
            body: `{"name":"Test","email":"test@example.com"}`,
            setupMock: func(m *MockUserService) {
                m.On("Create", mock.Anything, "Test", "test@example.com").
                    Return(&models.User{ID: uuid.NewString()}, nil)
            },
            wantStatus: http.StatusCreated,
        },
        {
            name: "validation error",
            body: `{"name":"","email":"test@example.com"}`,
            setupMock: func(m *MockUserService) {
                m.On("Create", mock.Anything, "", "test@example.com").
                    Return(nil, errs.ErrValidation)
            },
            wantStatus: http.StatusBadRequest,
        },
        {
            name:       "invalid json",
            body:       `invalid`,
            setupMock:  func(m *MockUserService) {},
            wantStatus: http.StatusBadRequest,
        },
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            mockSvc := new(MockUserService)
            tt.setupMock(mockSvc)

            handler := handlers.NewUserHandler(mockSvc)

            req := httptest.NewRequest(http.MethodPost, "/users",
                strings.NewReader(tt.body))
            req.Header.Set("Content-Type", "application/json")

            rec := httptest.NewRecorder()
            handler.Create(rec, req)

            assert.Equal(t, tt.wantStatus, rec.Code)
            mockSvc.AssertExpectations(t)
        })
    }
}
```

## Best Practices

| Practice | Description |
|----------|-------------|
| `t.Parallel()` | Run independent tests concurrently |
| `tt := tt` | Capture loop variable for parallel tests |
| `t.Helper()` | Mark test helper functions |
| `t.Cleanup()` | Automatic cleanup after test |
| `require.NoError()` | Fail test immediately on error |
| `assert.Equal()` | Continue test on assertion failure |
| Deterministic IDs | Use known UUIDs in fixtures |
| `@test.local` | Use test domain for emails |
| Table-driven | Single test function, multiple cases |

## CI Configuration

### GitHub Actions

```yaml
services:
  postgres:
    image: postgres:16-alpine
    env:
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
      POSTGRES_DB: test
    ports:
      - 5432:5432
    options: >-
      --health-cmd pg_isready
      --health-interval 10s
      --health-timeout 5s
      --health-retries 5

env:
  CI: "true"
  DATABASE_URL: postgres://test:test@localhost:5432/test?sslmode=disable
```

### GitLab CI

```yaml
services:
  - name: postgres:16-alpine
    alias: postgres

variables:
  CI: "true"
  POSTGRES_USER: test
  POSTGRES_PASSWORD: test
  POSTGRES_DB: test
  DATABASE_URL: postgres://test:test@postgres:5432/test?sslmode=disable
```

## Dependencies

```bash
go get github.com/stretchr/testify@latest
go get github.com/testcontainers/testcontainers-go@latest
go get github.com/pressly/goose/v3@latest
```
