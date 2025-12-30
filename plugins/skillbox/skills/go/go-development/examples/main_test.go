package storage_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var pgConnURL string

// TestMain sets up the test database.
// Uses testcontainers for local development, external PostgreSQL for CI.
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
		// CI: use external PostgreSQL service (GitHub Actions, GitLab CI, etc.)
		pgConnURL = os.Getenv("DATABASE_URL")
		if pgConnURL == "" {
			log.Fatal("DATABASE_URL is required in CI environment")
		}
	} else {
		// Local: use testcontainers
		pgConnURL, pgCloser, err = runLocalPostgres()
		if err != nil {
			log.Fatalf("Failed to start PostgreSQL container: %v", err)
		}
	}

	// Apply migrations
	if err := applyMigrations(pgConnURL); err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	code = m.Run()
}

// runLocalPostgres starts a PostgreSQL container using testcontainers.
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
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return "", nil, fmt.Errorf("start container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return "", nil, fmt.Errorf("get host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		return "", nil, fmt.Errorf("get port: %w", err)
	}

	url := fmt.Sprintf("postgres://test:test@%s:%s/test?sslmode=disable", host, port.Port())

	cleanup := func() {
		if err := container.Terminate(ctx); err != nil {
			log.Printf("Failed to terminate container: %v", err)
		}
	}

	return url, cleanup, nil
}

// applyMigrations runs database migrations using goose.
func applyMigrations(dbURL string) error {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	// Ping to ensure connection is ready
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	// Get underlying *sql.DB for goose
	db := pool.Config().ConnConfig

	goose.SetBaseFS(nil) // Use filesystem migrations
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	// Migrations are typically in ../../migrations relative to test file
	// Adjust path as needed for your project structure
	migrationsDir := "../../migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		// Skip migrations if directory doesn't exist (example file)
		log.Printf("Migrations directory not found: %s (skipping)", migrationsDir)
		return nil
	}

	// Note: For actual projects, use goose.Up with proper sql.DB connection
	_ = db // Placeholder - actual implementation depends on project structure
	log.Printf("Migrations would be applied from: %s", migrationsDir)

	return nil
}

// connectDB creates a new database connection for a test.
// Uses t.Cleanup to automatically close the connection after test.
func connectDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, pgConnURL)
	require.NoError(t, err)

	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}

// truncateTable removes all data from the specified table.
// Useful for cleaning up between tests.
func truncateTable(t *testing.T, pool *pgxpool.Pool, table string) {
	t.Helper()

	ctx := context.Background()
	_, err := pool.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
	require.NoError(t, err)
}

// truncateTables removes all data from multiple tables.
func truncateTables(t *testing.T, pool *pgxpool.Pool, tables ...string) {
	t.Helper()

	for _, table := range tables {
		truncateTable(t, pool, table)
	}
}
