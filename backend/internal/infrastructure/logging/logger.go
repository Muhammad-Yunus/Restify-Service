package logging

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"

	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// SLogger wraps slog.Logger with the repository.Logger interface.
type SLogger struct {
	logger *slog.Logger
}

// New creates a new structured logger. Format selects the handler:
// "json" for JSON output, "text" for plain text, anything else for tint.
func New(level string, format string, output io.Writer) repository.Logger {
	opts := &slog.HandlerOptions{
		Level: parseLevel(level),
	}

	switch format {
	case "json":
		return &SLogger{logger: slog.New(slog.NewJSONHandler(output, opts))}
	case "text":
		return &SLogger{logger: slog.New(slog.NewTextHandler(output, opts))}
	default:
		return &SLogger{logger: slog.New(tint.NewTextHandler(output, &tint.Options{
			Level:      opts.Level,
			TimeFormat: "2006-01-02T15:04:05.000Z",
		}))}
	}
}

// NewDefault creates a logger writing to stderr.
func NewDefault(level string) repository.Logger {
	return New(level, "json", os.Stderr)
}

// With adds key-value pairs to the logger context.
func (l *SLogger) With(keyValues ...any) repository.Logger {
	return &SLogger{logger: l.logger.With(keyValues...)}
}

// Debug logs a debug message.
func (l *SLogger) Debug(ctx context.Context, msg string, keyValues ...any) {
	l.logger.Log(ctx, slog.LevelDebug, msg, keyValues...)
}

// Info logs an info message.
func (l *SLogger) Info(ctx context.Context, msg string, keyValues ...any) {
	l.logger.Log(ctx, slog.LevelInfo, msg, keyValues...)
}

// Warn logs a warning message.
func (l *SLogger) Warn(ctx context.Context, msg string, keyValues ...any) {
	l.logger.Log(ctx, slog.LevelWarn, msg, keyValues...)
}

// Error logs an error message.
func (l *SLogger) Error(ctx context.Context, msg string, keyValues ...any) {
	l.logger.Log(ctx, slog.LevelError, msg, keyValues...)
}

// Logger returns the underlying *slog.Logger.
func (l *SLogger) Logger() *slog.Logger {
	return l.logger
}

func parseLevel(level string) slog.Level {
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

// Compile-time check.
var _ repository.Logger = (*SLogger)(nil)
