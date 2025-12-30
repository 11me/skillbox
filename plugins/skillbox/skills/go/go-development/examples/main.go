package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"myapp/internal/config"
	"myapp/internal/handler"
	"myapp/internal/services"
	"myapp/internal/storage"
	"myapp/pkg/postgres"
)

var ServiceVersion = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	// Load config
	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	// Setup logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLogLevel(cfg.LogLevel),
	}))
	slog.SetDefault(logger)

	logger.Info("starting service",
		slog.String("version", ServiceVersion),
		slog.String("app", cfg.AppName),
	)

	// Connect to database
	db, err := postgres.NewClient(ctx, cfg.DB.DSN(), cfg.DB.MaxConns, cfg.DB.MinConns)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer db.Close()

	logger.Info("connected to database")

	// Initialize storage and services
	store := storage.NewStorage(db)
	svcRegistry := services.NewRegistry(cfg, store)

	// Setup HTTP server
	h := handler.New(svcRegistry, logger)
	srv := &http.Server{
		Addr:         cfg.HTTP.Addr(),
		Handler:      h,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start server
	go func() {
		logger.Info("starting HTTP server", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.Error("server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	logger.Info("shutdown complete")
	return nil
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
