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
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var pgConnURL string

// closer is a cleanup function that returns an error.
// Use this pattern for proper error handling during resource cleanup.
type closer func() error

var defCloser = func() error { return nil }

// TestMain sets up the test database.
// Uses testcontainers for local development, external PostgreSQL for CI.
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

		// Apply schema migrations
		if err := applyMigrations(pgConnURL); err != nil {
			log.Fatalf("Failed to apply migrations: %v", err)
		}

		// Apply test data fixtures (optional)
		if err := applyTestData(pgConnURL); err != nil {
			log.Fatalf("Failed to apply test data: %v", err)
		}

		code = m.Run()
	}()

	os.Exit(code)
}

// runLocalPostgres starts a PostgreSQL container using testcontainers.
// Returns connection URL, cleanup function, and error.
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

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
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

	url := fmt.Sprintf("postgres://test:test@%s:%s/test?sslmode=disable", host, port.Port())

	cleanup := func() error {
		log.Println("Terminating PostgreSQL container...")
		return container.Terminate(ctx)
	}

	return url, cleanup, nil
}

// applyMigrations runs database schema migrations using goose.
// Migrations are expected in ../../migrations relative to test file.
func applyMigrations(dbURL string) error {
	migrationsDir := "../../migrations"

	// Skip if migrations directory doesn't exist
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
		return fmt.Errorf("set goose dialect: %w", err)
	}

	// Use default goose table for schema migrations
	goose.SetTableName("goose_db_version")

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	log.Printf("Schema migrations applied from: %s", migrationsDir)
	return nil
}

// applyTestData runs test data fixtures using goose.
// Test fixtures are expected in ./testmigration relative to test file.
// Uses a separate goose table to track test data migrations.
func applyTestData(dbURL string) error {
	testDataDir := "testmigration"

	// Skip if testmigration directory doesn't exist
	if _, err := os.Stat(testDataDir); os.IsNotExist(err) {
		// No test data directory - this is fine, not all packages need fixtures
		return nil
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	goose.SetBaseFS(nil)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	// Use separate goose table for test data to avoid conflicts with schema migrations
	goose.SetTableName("goose_db_test_version")

	if err := goose.Up(db, testDataDir); err != nil {
		return fmt.Errorf("apply test data: %w", err)
	}

	log.Printf("Test data applied from: %s", testDataDir)
	return nil
}

// connectDB creates a new database connection pool for a test.
// Uses t.Cleanup to automatically close the pool after test.
// Configures pool with test-appropriate settings.
func connectDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx := context.Background()

	cfg, err := pgxpool.ParseConfig(pgConnURL)
	require.NoError(t, err)

	// Test-appropriate pool settings
	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 5 * time.Minute
	cfg.MaxConnIdleTime = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	require.NoError(t, err)

	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}

// truncateTable removes all data from the specified table.
// Useful for cleaning up between tests when modifying shared data.
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
