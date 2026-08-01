package di

import (
	"os"

	"github.com/muhammadyunus/Restify-Service/internal/config"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/logging"
)

func initLogger(cfg config.LoggingConfig) repository.Logger {
	return logging.New(cfg.Level, cfg.Format, os.Stdout)
}
