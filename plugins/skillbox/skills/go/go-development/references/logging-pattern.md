# Logging Pattern

Choose based on project size: **slog** (small) or **zap** (large).

## slog (Small Projects)

```go
package logger

import (
    "log/slog"
    "os"
)

func New(level string) *slog.Logger {
    var lvl slog.Level
    switch level {
    case "debug":
        lvl = slog.LevelDebug
    case "warn":
        lvl = slog.LevelWarn
    case "error":
        lvl = slog.LevelError
    default:
        lvl = slog.LevelInfo
    }

    return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: lvl,
    }))
}
```

### slog Usage

```go
logger := logger.New("info")

logger.Info("user created",
    slog.String("user_id", user.ID.String()),
    slog.String("email", user.Email),
)

logger.Error("failed to create user",
    slog.String("error", err.Error()),
)
```

## zap (Large Projects)

```go
package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func New(level string) (*zap.Logger, error) {
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
```

### zap Usage

```go
logger, _ := logger.New("info")
defer logger.Sync()

logger.Info("user created",
    zap.String("user_id", user.ID.String()),
    zap.String("email", user.Email),
)

logger.Error("failed to create user",
    zap.Error(err),
)
```

## Decorator Wrapper Pattern

For structured logging with timing:

```go
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
```

## When to Use

| Criteria | slog | zap |
|----------|------|-----|
| Team size | < 5 | > 5 |
| Microservices | Few | Many |
| Dependencies | Minimal | More tooling |
| Performance | Good | Best |
| Structured fields | Yes | Yes |
