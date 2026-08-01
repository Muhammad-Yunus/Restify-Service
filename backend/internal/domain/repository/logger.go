package repository

import (
	"context"
	"log/slog"
)

// Logger is the logging repository interface.
type Logger interface {
	// With adds key-value pairs to the logger context.
	With(keyValues ...any) Logger

	// Debug logs a debug message.
	Debug(ctx context.Context, msg string, keyValues ...any)

	// Info logs an info message.
	Info(ctx context.Context, msg string, keyValues ...any)

	// Warn logs a warning message.
	Warn(ctx context.Context, msg string, keyValues ...any)

	// Error logs an error message.
	Error(ctx context.Context, msg string, keyValues ...any)

	// Logger returns the underlying *slog.Logger.
	Logger() *slog.Logger
}
