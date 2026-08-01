package di

import (
	"context"
	"errors"

	"github.com/muhammadyunus/Restify-Service/internal/config"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

type cacheStub struct{}

func (s *cacheStub) Close(ctx context.Context) error {
	return nil
}

func initCache(cfg config.RedisConfig) (repository.Cache, error) {
	if cfg.URL == "" {
		return nil, errors.New("redis url is required")
	}

	return &cacheStub{}, nil
}
