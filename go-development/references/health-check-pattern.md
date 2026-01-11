# Health Check Pattern

Kubernetes-compatible liveness and readiness probes.

## Endpoints

| Probe | Path | Purpose | Failure Action |
|-------|------|---------|----------------|
| **Liveness** | `/check/healthz/` | Is the app alive? | Restart pod |
| **Readiness** | `/check/readyz/` | Can it serve traffic? | Remove from LB |

**Key difference:** Liveness checks the process, readiness checks dependencies.

## Architecture

```
┌─────────────────────────────────────────┐
│              API Server (:8080)          │
│  - Business logic routes                 │
│  - No health checks                      │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│           Monitor Server (:8081)         │
│  - /check/healthz/ (liveness)           │
│  - /check/readyz/  (readiness)          │
│  - /metrics        (prometheus)          │
└─────────────────────────────────────────┘
```

## ReadyChecker Interface

```go
// ReadyChecker checks if a dependency is ready.
type ReadyChecker interface {
    CheckReady(ctx context.Context) error
}
```

All dependency checkers implement this interface.

## Healthz Handler (Liveness)

```go
const HealthzHandlerPathPrefix = "/check/healthz"

type HealthzHandler struct {
    http.Handler
}

func NewHealthzHandler() *HealthzHandler {
    router := chi.NewRouter()
    handler := &HealthzHandler{Handler: router}
    router.Get("/", handler.handleHealthz)
    return handler
}

func (HealthzHandler) handleHealthz(w http.ResponseWriter, _ *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
```

**Always returns 200 OK** — if the handler responds, the process is alive.

## Readyz Handler (Readiness)

```go
const ReadyzHandlerPathPrefix = "/check/readyz"

type ReadyzHandler struct {
    http.Handler
    checkers []ReadyChecker
}

func NewReadyzHandler(checkers ...ReadyChecker) *ReadyzHandler {
    router := chi.NewRouter()
    handler := &ReadyzHandler{
        Handler:  router,
        checkers: checkers,
    }
    router.Get("/", handler.handleReadyz)
    return handler
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
        json.NewEncoder(w).Encode(map[string]string{"error": errs[0].Error()})
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}
```

**Key features:**
- Runs all checkers **in parallel** (goroutines)
- Returns **503** if any checker fails
- Returns **200** only if all pass

## Dependency Checkers

### PostgreSQL Checker

```go
type PostgresChecker struct {
    pool          *pgxpool.Pool
    schemaVersion int64
    timeout       time.Duration
}

func NewPostgresChecker(pool *pgxpool.Pool, schemaVersion int64) *PostgresChecker {
    return &PostgresChecker{
        pool:          pool,
        schemaVersion: schemaVersion,
        timeout:       5 * time.Second,
    }
}

func (c *PostgresChecker) CheckReady(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, c.timeout)
    defer cancel()

    // Check connectivity
    if err := c.pool.Ping(ctx); err != nil {
        return fmt.Errorf("postgres ping: %w", err)
    }

    // Check schema version
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
            return fmt.Errorf("schema mismatch: want %d, got %d", c.schemaVersion, versionID)
        }
    }

    return nil
}
```

### HTTP Service Checker

```go
type HTTPChecker struct {
    client  *http.Client
    url     string
    timeout time.Duration
}

func NewHTTPChecker(url string) *HTTPChecker {
    return &HTTPChecker{
        client:  &http.Client{Timeout: 5 * time.Second},
        url:     url,
        timeout: 5 * time.Second,
    }
}

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
```

## Router Setup

```go
func setupMonitorServer(pool *pgxpool.Pool) *http.Server {
    router := chi.NewRouter()

    // Health checks
    healthz := health.NewHealthzHandler()
    readyz := health.NewReadyzHandler(
        health.NewPostgresChecker(pool, 20240115120000),
        // health.NewHTTPChecker("http://auth-service/check/healthz/"),
    )

    router.Mount(health.HealthzHandlerPathPrefix, healthz)
    router.Mount(health.ReadyzHandlerPathPrefix, readyz)

    // Metrics
    router.Handle("/metrics", promhttp.Handler())

    return &http.Server{
        Addr:    ":8081",
        Handler: router,
    }
}
```

## Kubernetes Configuration

```yaml
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
        - name: app
          ports:
            - name: http
              containerPort: 8080
            - name: monitor
              containerPort: 8081
          livenessProbe:
            httpGet:
              path: /check/healthz/
              port: monitor
            initialDelaySeconds: 5
            periodSeconds: 10
            timeoutSeconds: 5
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /check/readyz/
              port: monitor
            initialDelaySeconds: 5
            periodSeconds: 5
            timeoutSeconds: 5
            failureThreshold: 3
```

## Response Examples

### Healthz (Liveness)

```json
{"status": "healthy"}
```

### Readyz (Ready)

```json
{"status": "ready"}
```

### Readyz (Not Ready)

```json
{"error": "postgres ping: connection refused"}
```

## Best Practices

### DO:
- ✅ Keep `/check/healthz/` simple (just return 200)
- ✅ Check all critical dependencies in `/check/readyz/`
- ✅ Add timeouts to all checks
- ✅ Run checks in parallel
- ✅ Validate schema version in database check
- ✅ Use separate port for health/metrics

### DON'T:
- ❌ Add business logic to health checks
- ❌ Check non-critical dependencies
- ❌ Make health checks slow (>1s)
- ❌ Require authentication for health endpoints
- ❌ Return sensitive info in responses

## Related

- [tracing-pattern.md](tracing-pattern.md) — OpenTelemetry tracing
- [database-pattern.md](database-pattern.md) — Database patterns
