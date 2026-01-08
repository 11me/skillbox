# Entry Point Pattern

Production-grade application entry point with graceful shutdown.

## Architecture

```
main()
├── Load Config (env vars)
├── Setup Logger (zap)
├── Create Backend
│   ├── Init Database
│   ├── Init Services
│   ├── Init Prometheus
│   └── Configure Servers
├── Signal Handler (SIGINT, SIGTERM)
├── Create errgroup
├── Start Servers + Jobs
├── Wait for signal
└── Graceful Shutdown (5s timeout)
```

## Main Function

```go
func main() {
    cfg := NewConfig()
    if err := cfg.Parse(); err != nil {
        log.Fatalln(err)
    }

    logger := setupLogger(cfg.LogLevel)
    defer logger.Sync()

    be := newBackend(cfg, logger)
    if err := be.init(); err != nil {
        logger.Fatal("init backend", zap.Error(err))
    }

    ctx, cancel := signal.NotifyContext(
        context.Background(),
        syscall.SIGINT, syscall.SIGTERM,
    )
    defer cancel()

    eg, ctx := errgroup.WithContext(ctx)

    logger.Info("starting application")

    eg.Go(be.startMonitorServer)
    eg.Go(be.startAPIServer)
    be.startJobs(ctx, eg)

    logger.Info("application started")

    <-ctx.Done()

    logger.Info("stopping application")

    ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    be.stop(ctx)

    if err := eg.Wait(); err != nil {
        logger.Error("error during shutdown", zap.Error(err))
    }

    logger.Info("application stopped")
}
```

## Backend Pattern

Backend struct aggregates all dependencies and manages lifecycle.

```go
type backend struct {
    cfg    *Config
    logger *zap.Logger

    // Infrastructure
    pgClient pg.Client
    registry *prometheus.Registry

    // Services
    userService    UserService
    orderService   OrderService

    // Servers
    apiServer     *http.Server
    monitorServer *http.Server

    // Background jobs
    jobs []BackgroundJob
}
```

### Constructor

```go
func newBackend(cfg *Config, logger *zap.Logger) *backend {
    return &backend{
        cfg:      cfg,
        logger:   logger,
        registry: prometheus.NewRegistry(),
    }
}
```

### Initialization

Initialize in dependency order:

```go
func (be *backend) init() error {
    if err := be.initDatabase(); err != nil {
        return fmt.Errorf("init database: %w", err)
    }

    be.initServices()
    be.initPrometheus()
    be.initServers()
    be.initJobs()

    return nil
}

func (be *backend) initDatabase() error {
    client, err := pg.New(
        pg.WithHost(be.cfg.Postgres.Host),
        pg.WithPort(be.cfg.Postgres.Port),
        pg.WithDBName(be.cfg.Postgres.DBName),
        pg.WithUser(be.cfg.Postgres.User),
        pg.WithPassword(be.cfg.Postgres.Password),
    )
    if err != nil {
        return err
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := client.Connect(ctx); err != nil {
        return err
    }

    be.pgClient = client
    return nil
}

func (be *backend) initServices() {
    userRepo := storage.NewUserRepository(be.pgClient)
    be.userService = services.NewUserService(userRepo, be.logger)

    orderRepo := storage.NewOrderRepository(be.pgClient)
    be.orderService = services.NewOrderService(orderRepo, be.logger)
}

func (be *backend) initPrometheus() {
    be.registry.MustRegister(
        collectors.NewGoCollector(),
        collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
    )
}
```

### Server Configuration

Two servers: API (main) + Monitor (metrics/health):

```go
func (be *backend) initServers() {
    // API server
    router := chi.NewRouter()
    router.Use(middleware.RequestID)
    router.Use(middleware.RealIP)
    router.Use(middleware.Recoverer)

    handler := handlers.NewUserHandler(be.userService)
    router.Route("/api/v1", func(r chi.Router) {
        r.Mount("/users", handler.Routes())
    })

    be.apiServer = &http.Server{
        Addr:         be.cfg.HTTP.Address,
        Handler:      router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Monitor server (metrics + health)
    monitorRouter := chi.NewRouter()
    monitorRouter.Get("/health", be.healthHandler)
    monitorRouter.Get("/ready", be.readyHandler)
    monitorRouter.Handle("/metrics", promhttp.HandlerFor(be.registry, promhttp.HandlerOpts{}))

    be.monitorServer = &http.Server{
        Addr:    be.cfg.Monitor.Address,
        Handler: monitorRouter,
    }
}
```

### Start/Stop Methods

```go
func (be *backend) startAPIServer() error {
    be.logger.Info("starting API server", zap.String("addr", be.cfg.HTTP.Address))
    if err := be.apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        return err
    }
    return nil
}

func (be *backend) startMonitorServer() error {
    be.logger.Info("starting monitor server", zap.String("addr", be.cfg.Monitor.Address))
    if err := be.monitorServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        return err
    }
    return nil
}

func (be *backend) startJobs(ctx context.Context, eg *errgroup.Group) {
    for _, job := range be.jobs {
        job := job
        eg.Go(func() error {
            return job.Run(ctx)
        })
    }
}

func (be *backend) stop(ctx context.Context) {
    if err := be.apiServer.Shutdown(ctx); err != nil {
        be.logger.Error("shutdown API server", zap.Error(err))
    }

    if err := be.monitorServer.Shutdown(ctx); err != nil {
        be.logger.Error("shutdown monitor server", zap.Error(err))
    }

    be.pgClient.Close()
}
```

## Signal Handling

Use `signal.NotifyContext` for clean cancellation:

```go
ctx, cancel := signal.NotifyContext(
    context.Background(),
    syscall.SIGINT,  // Ctrl+C
    syscall.SIGTERM, // Docker/K8s stop
)
defer cancel()

// ... start servers ...

<-ctx.Done() // Blocks until signal received
```

**Why this pattern:**
- Context automatically cancelled on signal
- Works with errgroup context chaining
- Clean cancellation propagation to all goroutines

## Errgroup Pattern

Use `errgroup` for concurrent server management:

```go
eg, ctx := errgroup.WithContext(ctx)

// Start servers concurrently
eg.Go(be.startMonitorServer)
eg.Go(be.startAPIServer)

// Start background jobs
be.startJobs(ctx, eg)

// Wait for signal
<-ctx.Done()

// Shutdown
be.stop(shutdownCtx)

// Wait for all goroutines
if err := eg.Wait(); err != nil {
    logger.Error("shutdown error", zap.Error(err))
}
```

**Benefits:**
- Single error channel for all goroutines
- Context cancellation on first error
- Clean synchronization on shutdown

## Graceful Shutdown

Shutdown with timeout to prevent hanging:

```go
<-ctx.Done() // Signal received

logger.Info("stopping application")

// Create shutdown context with timeout
ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Stop servers gracefully
be.stop(ctx)

// Wait for background goroutines
eg.Wait()

logger.Info("application stopped")
```

**Shutdown sequence:**
1. Stop accepting new connections
2. Wait for in-flight requests (up to timeout)
3. Close database connections
4. Wait for background jobs to finish

## Logger Setup

```go
func setupLogger(level string) *zap.Logger {
    cfg := zap.NewProductionConfig()

    // Parse level
    var lvl zapcore.Level
    if err := lvl.UnmarshalText([]byte(level)); err != nil {
        lvl = zapcore.InfoLevel
    }
    cfg.Level = zap.NewAtomicLevelAt(lvl)

    // Disable stacktrace for non-error
    cfg.DisableStacktrace = true

    logger, err := cfg.Build()
    if err != nil {
        log.Fatalln("build logger:", err)
    }

    // Name the logger
    logger = logger.Named("app")

    // Replace global logger
    zap.ReplaceGlobals(logger)

    return logger
}
```

## Background Jobs Interface

```go
type BackgroundJob interface {
    Run(ctx context.Context) error
}

// Example job
type cleanupJob struct {
    interval time.Duration
    service  CleanupService
    logger   *zap.Logger
}

func (j *cleanupJob) Run(ctx context.Context) error {
    ticker := time.NewTicker(j.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return nil
        case <-ticker.C:
            if err := j.service.Cleanup(ctx); err != nil {
                j.logger.Error("cleanup failed", zap.Error(err))
            }
        }
    }
}
```

## Health/Ready Handlers

```go
func (be *backend) healthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func (be *backend) readyHandler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
    defer cancel()

    // Check database
    if err := be.pgClient.Ping(ctx); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        w.Write([]byte("database not ready"))
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}
```

## File Structure

```
cmd/
└── app/
    ├── main.go      # Entry point
    ├── config.go    # Configuration
    └── backend.go   # Dependency initialization
```

## Dependencies

```bash
go get golang.org/x/sync/errgroup@latest
go get go.uber.org/zap@latest
go get github.com/go-chi/chi/v5@latest
go get github.com/prometheus/client_golang@latest
```

## Summary

| Component | Pattern |
|-----------|---------|
| Config | `NewConfig()` + `Parse()` |
| Logger | `zap` with production config |
| Backend | Struct with init/start/stop |
| Signals | `signal.NotifyContext()` |
| Concurrency | `errgroup.WithContext()` |
| Servers | API + Monitor (dual server) |
| Shutdown | 5s timeout, graceful stop |
| Jobs | Interface with context cancellation |
