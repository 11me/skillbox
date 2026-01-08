// Package health provides health check handlers for Kubernetes probes.
//
// This example shows:
// - HealthzHandler for liveness probes
// - ReadyzHandler for readiness probes with parallel checks
// - ReadyChecker interface for dependency checks
// - Database and HTTP service checkers
package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ---------- Path Constants ----------

const (
	HealthzHandlerPathPrefix = "/check/healthz"
	ReadyzHandlerPathPrefix  = "/check/readyz"
)

// ---------- ReadyChecker Interface ----------

// ReadyChecker checks if a dependency is ready.
type ReadyChecker interface {
	CheckReady(ctx context.Context) error
}

// ---------- Healthz Handler (Liveness) ----------

// HealthzHandler handles liveness probe requests.
// Always returns 200 OK if the process is running.
type HealthzHandler struct {
	http.Handler
}

// NewHealthzHandler creates a new liveness handler.
func NewHealthzHandler() *HealthzHandler {
	router := chi.NewRouter()
	handler := &HealthzHandler{Handler: router}

	router.Get("/", handler.handleHealthz)

	return handler
}

type healthzResponse struct {
	Status string `json:"status"`
}

func (HealthzHandler) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(healthzResponse{Status: "healthy"})
}

// ---------- Readyz Handler (Readiness) ----------

// ReadyzHandler handles readiness probe requests.
// Returns 200 OK only if all checkers pass.
type ReadyzHandler struct {
	http.Handler
	checkers []ReadyChecker
}

// NewReadyzHandler creates a new readiness handler with checkers.
func NewReadyzHandler(checkers ...ReadyChecker) *ReadyzHandler {
	router := chi.NewRouter()
	handler := &ReadyzHandler{
		Handler:  router,
		checkers: checkers,
	}

	router.Get("/", handler.handleReadyz)

	return handler
}

type readyzResponse struct {
	Status string `json:"status"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *ReadyzHandler) handleReadyz(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Run all checkers in parallel
	errCh := make(chan error, len(h.checkers))

	for _, checker := range h.checkers {
		go func(checker ReadyChecker) {
			errCh <- checker.CheckReady(ctx)
		}(checker)
	}

	// Collect errors
	var errs []error
	for i := 0; i < len(h.checkers); i++ {
		if err := <-errCh; err != nil {
			errs = append(errs, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")

	if len(errs) > 0 {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(errorResponse{Error: errs[0].Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(readyzResponse{Status: "ready"})
}

// ---------- PostgreSQL Checker ----------

// PostgresChecker checks PostgreSQL connectivity and migration version.
type PostgresChecker struct {
	pool          *pgxpool.Pool
	schemaVersion int64
	timeout       time.Duration
}

// NewPostgresChecker creates a new PostgreSQL checker.
func NewPostgresChecker(pool *pgxpool.Pool, schemaVersion int64) *PostgresChecker {
	return &PostgresChecker{
		pool:          pool,
		schemaVersion: schemaVersion,
		timeout:       5 * time.Second,
	}
}

// CheckReady checks database connectivity and schema version.
func (c *PostgresChecker) CheckReady(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Check connectivity
	if err := c.pool.Ping(ctx); err != nil {
		return fmt.Errorf("postgres ping: %w", err)
	}

	// Check schema version (optional, for goose migrations)
	if c.schemaVersion > 0 {
		var versionID int64
		var dirty bool

		err := c.pool.QueryRow(ctx, `
			SELECT version_id, dirty
			FROM goose_db_version
			ORDER BY id DESC
			LIMIT 1
		`).Scan(&versionID, &dirty)

		if err != nil {
			return fmt.Errorf("postgres schema check: %w", err)
		}

		if dirty {
			return fmt.Errorf("postgres schema is dirty")
		}

		if versionID != c.schemaVersion {
			return fmt.Errorf("postgres schema version mismatch: want %d, got %d",
				c.schemaVersion, versionID)
		}
	}

	return nil
}

// ---------- HTTP Service Checker ----------

// HTTPChecker checks an HTTP service health endpoint.
type HTTPChecker struct {
	client  *http.Client
	url     string
	timeout time.Duration
}

// NewHTTPChecker creates a new HTTP service checker.
func NewHTTPChecker(url string) *HTTPChecker {
	return &HTTPChecker{
		client:  &http.Client{Timeout: 5 * time.Second},
		url:     url,
		timeout: 5 * time.Second,
	}
}

// CheckReady checks if the HTTP service is available.
func (c *HTTPChecker) CheckReady(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return fmt.Errorf("http checker: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("http checker: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("http checker: status %d", resp.StatusCode)
	}

	return nil
}

// ---------- Usage Example ----------

// Example setup:
//
//	func setupHealthChecks(pool *pgxpool.Pool) (http.Handler, http.Handler) {
//	    // Liveness - always healthy if process runs
//	    healthz := health.NewHealthzHandler()
//
//	    // Readiness - check dependencies
//	    readyz := health.NewReadyzHandler(
//	        health.NewPostgresChecker(pool, 20240115120000),
//	        // health.NewHTTPChecker("http://auth-service/check/healthz/"),
//	    )
//
//	    return healthz, readyz
//	}
//
//	// Router setup:
//	router := chi.NewRouter()
//	router.Mount(health.HealthzHandlerPathPrefix, healthz)
//	router.Mount(health.ReadyzHandlerPathPrefix, readyz)
//
//	// Kubernetes probes:
//	// livenessProbe:
//	//   httpGet:
//	//     path: /check/healthz/
//	//     port: 8081
//	// readinessProbe:
//	//   httpGet:
//	//     path: /check/readyz/
//	//     port: 8081
