# Testing Pattern

Production-grade testing patterns with testcontainers and testify.

## Testing Philosophy

| Layer | Test Type | Database | Mock What |
|-------|-----------|----------|-----------|
| **Repository** | Integration | Real DB (testcontainers) | Nothing |
| **Service** | Unit | Mock repository | Repository |
| **Handler** | Unit | Mock service | Service |

## IMPORTANT: Repository Testing Rules

```
❌ DON'T mock SQL queries in repository tests
❌ DON'T use go-sqlmock for repository layer
❌ DON'T write fake implementations of database

✅ DO use testcontainers with real PostgreSQL
✅ DO test actual SQL queries against real database
✅ DO verify data is correctly stored and retrieved
```

**Why:** Repository layer's job is to talk to database. Mocking SQL queries tests the mock, not the repository logic. Real database tests catch:
- SQL syntax errors
- Type mismatches
- Constraint violations
- Transaction behavior
- Query performance issues

## TestMain with CI/Local Detection

```go
package storage_test

import (
    "context"
    "fmt"
    "log"
    "os"
    "testing"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

var pgConnURL string

func TestMain(m *testing.M) {
    var code int
    var pgCloser func()

    defer func() {
        if pgCloser != nil {
            pgCloser()
        }
        os.Exit(code)
    }()

    var err error
    if os.Getenv("CI") == "true" {
        // CI: use external PostgreSQL service
        pgConnURL = os.Getenv("DATABASE_URL")
        if pgConnURL == "" {
            log.Fatal("DATABASE_URL is required in CI")
        }
    } else {
        // Local: use testcontainers
        pgConnURL, pgCloser, err = runLocalPostgres()
        if err != nil {
            log.Fatalf("Failed to start PostgreSQL: %v", err)
        }
    }

    // Apply migrations
    applyMigrations(pgConnURL)

    code = m.Run()
}

func runLocalPostgres() (string, func(), error) {
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
        return "", nil, err
    }

    host, _ := container.Host(ctx)
    port, _ := container.MappedPort(ctx, "5432")

    url := fmt.Sprintf("postgres://test:test@%s:%s/test?sslmode=disable",
        host, port.Port())

    cleanup := func() { container.Terminate(ctx) }
    return url, cleanup, nil
}
```

## Test Helpers

```go
// connectDB creates a pool for a test with automatic cleanup.
func connectDB(t *testing.T) *pgxpool.Pool {
    t.Helper()

    pool, err := pgxpool.New(context.Background(), pgConnURL)
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

// createTestUser creates a user for testing.
func createTestUser(t *testing.T, pool *pgxpool.Pool) *models.User {
    t.Helper()

    repo := storage.NewUserRepository(pool)
    user, err := repo.Create(context.Background(), &models.User{
        Name:  "Test User",
        Email: fmt.Sprintf("test-%s@example.com", uuid.NewString()[:8]),
    })
    require.NoError(t, err)

    return user
}
```

## Repository Tests (Real Database)

```go
func TestUserRepository_Create(t *testing.T) {
    t.Parallel()

    pool := connectDB(t)
    repo := storage.NewUserRepository(pool)

    ctx := context.Background()
    user := &models.User{
        Name:  "Test User",
        Email: fmt.Sprintf("test-%s@example.com", uuid.NewString()[:8]),
    }

    // Create user
    created, err := repo.Create(ctx, user)
    require.NoError(t, err)
    assert.NotEmpty(t, created.ID)

    // Verify in database
    found, err := repo.GetByID(ctx, created.ID)
    require.NoError(t, err)
    assert.Equal(t, user.Name, found.Name)
}

func TestUserRepository_GetByID(t *testing.T) {
    t.Parallel()

    pool := connectDB(t)
    repo := storage.NewUserRepository(pool)

    ctx := context.Background()
    user := createTestUser(t, pool)

    tests := []struct {
        name    string
        id      string
        wantErr bool
    }{
        {name: "existing", id: user.ID, wantErr: false},
        {name: "non-existing", id: uuid.NewString(), wantErr: true},
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            found, err := repo.GetByID(ctx, tt.id)
            if tt.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
                assert.Equal(t, tt.id, found.ID)
            }
        })
    }
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
                    Return(nil, common.EntityNotFound("not found"))
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
            wantErr:   common.ValidationFailed("name is required"),
        },
        {
            name:      "email exists",
            inputName: "User",
            email:     "existing@example.com",
            setupMock: func(m *MockUserRepository) {
                m.On("GetByEmail", mock.Anything, "existing@example.com").
                    Return(&models.User{ID: uuid.NewString()}, nil)
            },
            wantErr: common.StateConflict("email already exists"),
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
                assert.Equal(t, tt.wantErr, err)
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
        wantBody   string
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
                    Return(nil, common.ValidationFailed("name is required"))
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
| Unique emails | Use UUID suffix for test data |
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
```
