package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// BackgroundJob is the interface for background workers.
type BackgroundJob interface {
	Run(ctx context.Context) error
}

// backend aggregates all application dependencies.
type backend struct {
	cfg    *Config
	logger *zap.Logger

	// Infrastructure
	pool     *pgxpool.Pool
	registry *prometheus.Registry

	// Services (add your services here)
	// userService    services.UserService
	// orderService   services.OrderService

	// Servers
	apiServer     *http.Server
	monitorServer *http.Server

	// Background jobs
	jobs []BackgroundJob
}

// newBackend creates a new backend instance.
func newBackend(cfg *Config, logger *zap.Logger) *backend {
	return &backend{
		cfg:      cfg,
		logger:   logger,
		registry: prometheus.NewRegistry(),
		jobs:     make([]BackgroundJob, 0),
	}
}

// init initializes all dependencies in order.
func (be *backend) init(ctx context.Context) error {
	if err := be.initDatabase(ctx); err != nil {
		return fmt.Errorf("init database: %w", err)
	}

	be.initServices()
	be.initPrometheus()
	be.initServers()
	be.initJobs()

	return nil
}

// initDatabase establishes database connection.
func (be *backend) initDatabase(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, be.cfg.Postgres.DSN())
	if err != nil {
		return fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	be.pool = pool
	be.logger.Info("database connected")

	return nil
}

// initServices initializes application services.
func (be *backend) initServices() {
	// Initialize repositories
	// userRepo := storage.NewUserRepository(be.pool)
	// orderRepo := storage.NewOrderRepository(be.pool)

	// Initialize services
	// be.userService = services.NewUserService(userRepo, be.logger)
	// be.orderService = services.NewOrderService(orderRepo, be.logger)

	be.logger.Info("services initialized")
}

// initPrometheus registers Prometheus collectors.
func (be *backend) initPrometheus() {
	be.registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	be.logger.Info("prometheus collectors registered")
}

// initServers configures HTTP servers.
func (be *backend) initServers() {
	// API server
	apiRouter := chi.NewRouter()
	apiRouter.Use(middleware.RequestID)
	apiRouter.Use(middleware.RealIP)
	apiRouter.Use(middleware.Recoverer)
	apiRouter.Use(middleware.Timeout(30 * time.Second))

	// Mount your handlers here
	// userHandler := handlers.NewUserHandler(be.userService)
	// apiRouter.Route("/api/v1", func(r chi.Router) {
	//     r.Mount("/users", userHandler.Routes())
	// })

	apiRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	be.apiServer = &http.Server{
		Addr:         be.cfg.HTTP.Address,
		Handler:      apiRouter,
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
		Addr:         be.cfg.Monitor.Address,
		Handler:      monitorRouter,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	be.logger.Info("servers configured")
}

// initJobs registers background jobs.
func (be *backend) initJobs() {
	// Add your background jobs here
	// be.jobs = append(be.jobs, NewCleanupJob(be.cleanupService, be.logger))
	// be.jobs = append(be.jobs, NewSyncJob(be.syncService, be.logger))

	be.logger.Info("jobs initialized", zap.Int("count", len(be.jobs)))
}

// startAPIServer starts the API HTTP server.
func (be *backend) startAPIServer() error {
	be.logger.Info("starting API server", zap.String("addr", be.cfg.HTTP.Address))

	if err := be.apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("api server: %w", err)
	}

	return nil
}

// startMonitorServer starts the monitor HTTP server.
func (be *backend) startMonitorServer() error {
	be.logger.Info("starting monitor server", zap.String("addr", be.cfg.Monitor.Address))

	if err := be.monitorServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("monitor server: %w", err)
	}

	return nil
}

// startJobs starts all background jobs in the errgroup.
func (be *backend) startJobs(ctx context.Context, eg *errgroup.Group) {
	for _, job := range be.jobs {
		job := job // capture for goroutine
		eg.Go(func() error {
			return job.Run(ctx)
		})
	}
}

// stop gracefully shuts down all components.
func (be *backend) stop(ctx context.Context) {
	// Stop API server
	if err := be.apiServer.Shutdown(ctx); err != nil {
		be.logger.Error("shutdown API server", zap.Error(err))
	}

	// Stop monitor server
	if err := be.monitorServer.Shutdown(ctx); err != nil {
		be.logger.Error("shutdown monitor server", zap.Error(err))
	}

	// Close database
	if be.pool != nil {
		be.pool.Close()
	}

	be.logger.Info("backend stopped")
}

// healthHandler returns 200 if the service is alive.
func (be *backend) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// readyHandler returns 200 if the service is ready to accept traffic.
func (be *backend) readyHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	// Check database connection
	if err := be.pool.Ping(ctx); err != nil {
		be.logger.Warn("readiness check failed", zap.Error(err))
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("database not ready"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
