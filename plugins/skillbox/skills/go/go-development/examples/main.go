package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
)

var Version = "dev"

func main() {
	// Load configuration
	cfg := NewConfig()
	if err := cfg.Parse(); err != nil {
		log.Fatalln("parse config:", err)
	}

	// Setup logger
	logger := setupLogger(cfg.LogLevel)
	defer logger.Sync()

	logger.Info("starting application",
		zap.String("version", Version),
		zap.String("app", cfg.AppName),
	)

	// Create and initialize backend
	be := newBackend(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := be.init(ctx); err != nil {
		logger.Fatal("init backend", zap.Error(err))
	}

	// Setup signal handling
	ctx, cancel = signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	// Create errgroup for concurrent execution
	eg, ctx := errgroup.WithContext(ctx)

	logger.Info("starting servers")

	// Start servers concurrently
	eg.Go(be.startMonitorServer)
	eg.Go(be.startAPIServer)

	// Start background jobs
	be.startJobs(ctx, eg)

	logger.Info("application started")

	// Wait for shutdown signal
	<-ctx.Done()

	logger.Info("stopping application")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	be.stop(shutdownCtx)

	// Wait for all goroutines to finish
	if err := eg.Wait(); err != nil {
		logger.Error("shutdown error", zap.Error(err))
	}

	logger.Info("application stopped")
}

// setupLogger creates a production-ready zap logger.
func setupLogger(level string) *zap.Logger {
	cfg := zap.NewProductionConfig()

	// Parse log level
	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = zapcore.InfoLevel
	}
	cfg.Level = zap.NewAtomicLevelAt(lvl)

	// Disable stacktrace for non-error levels
	cfg.DisableStacktrace = true

	// ISO8601 time format
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := cfg.Build()
	if err != nil {
		log.Fatalln("build logger:", err)
	}

	// Replace global logger
	zap.ReplaceGlobals(logger)

	return logger
}
