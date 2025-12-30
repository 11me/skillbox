package logger

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewZap creates a new zap logger.
// Recommended for large projects.
func NewZap(level string) (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	switch level {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	return config.Build()
}

// Decorator wrapper pattern for service logging with timing

type User struct {
	ID    uuid.UUID
	Name  string
	Email string
}

type UserService interface {
	GetUser(ctx context.Context, id uuid.UUID) (*User, error)
	CreateUser(ctx context.Context, name, email string) (*User, error)
}

type UserServiceLogger struct {
	wrapped UserService
	logger  *zap.Logger
}

func NewUserServiceLogger(svc UserService, logger *zap.Logger) *UserServiceLogger {
	return &UserServiceLogger{
		wrapped: svc,
		logger:  logger.Named("user_service"),
	}
}

func (s *UserServiceLogger) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
	start := time.Now()

	user, err := s.wrapped.GetUser(ctx, id)

	elapsed := time.Since(start)
	fields := []zap.Field{
		zap.Stringer("user_id", id),
		zap.Duration("elapsed", elapsed),
		zap.Error(err),
	}

	if err != nil {
		s.logger.Error("get user failed", fields...)
	} else {
		s.logger.Debug("get user", fields...)
	}

	return user, err
}

func (s *UserServiceLogger) CreateUser(ctx context.Context, name, email string) (*User, error) {
	start := time.Now()

	user, err := s.wrapped.CreateUser(ctx, name, email)

	elapsed := time.Since(start)
	fields := []zap.Field{
		zap.String("name", name),
		zap.String("email", email),
		zap.Duration("elapsed", elapsed),
		zap.Error(err),
	}

	if err != nil {
		s.logger.Error("create user failed", fields...)
	} else {
		s.logger.Info("user created",
			zap.Stringer("user_id", user.ID),
			zap.Duration("elapsed", elapsed),
		)
	}

	return user, err
}
