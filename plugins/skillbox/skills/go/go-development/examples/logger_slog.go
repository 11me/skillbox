package logger

import (
	"log/slog"
	"os"
)

// New creates a new slog logger.
// Recommended for small projects.
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

// Usage example:
//
//	logger := logger.New("info")
//
//	logger.Info("user created",
//	    slog.String("user_id", user.ID.String()),
//	    slog.String("email", user.Email),
//	)
//
//	logger.Error("failed to create user",
//	    slog.String("error", err.Error()),
//	)
