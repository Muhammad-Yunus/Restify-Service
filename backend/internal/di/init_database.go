package di

import (
	"context"
	"errors"

	"github.com/muhammadyunus/Restify-Service/internal/config"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

type dbStub struct{}

func (s *dbStub) Close(ctx context.Context) error {
	return nil
}

func initDatabase(cfg config.DatabaseConfig) (repository.DB, error) {
	if cfg.URL == "" {
		return nil, errors.New("database url is required")
	}

	return &dbStub{}, nil
}
