package di

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/muhammadyunus/Restify-Service/internal/config"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

type slogLogger struct {
	logger *slog.Logger
}

func (l *slogLogger) With(keyValues ...any) repository.Logger {
	return &slogLogger{logger: l.logger.With(keyValues...)}
}

func (l *slogLogger) Debug(ctx context.Context, msg string, keyValues ...any) {
	l.logger.Log(ctx, slog.LevelDebug, msg, keyValues...)
}

func (l *slogLogger) Info(ctx context.Context, msg string, keyValues ...any) {
	l.logger.Log(ctx, slog.LevelInfo, msg, keyValues...)
}

func (l *slogLogger) Warn(ctx context.Context, msg string, keyValues ...any) {
	l.logger.Log(ctx, slog.LevelWarn, msg, keyValues...)
}

func (l *slogLogger) Error(ctx context.Context, msg string, keyValues ...any) {
	l.logger.Log(ctx, slog.LevelError, msg, keyValues...)
}

func (l *slogLogger) Logger() *slog.Logger {
	return l.logger
}

func initLogger(cfg config.LoggingConfig) (repository.Logger, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, fmt.Errorf("parse log level %q: %w", cfg.Level, err)
	}

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return &slogLogger{logger: slog.New(handler)}, nil
}
