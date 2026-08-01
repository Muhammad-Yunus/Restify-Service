package di

import (
	"context"
	"errors"
	"time"

	"github.com/muhammadyunus/Restify-Service/internal/config"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

type cacheStub struct{}

func (s *cacheStub) Get(ctx context.Context, key string) (string, error) {
	return "", errStubNotImplemented
}

func (s *cacheStub) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return errStubNotImplemented
}

func (s *cacheStub) Delete(ctx context.Context, key string) error {
	return errStubNotImplemented
}

func (s *cacheStub) Exists(ctx context.Context, key string) (bool, error) {
	return false, errStubNotImplemented
}

func (s *cacheStub) Close(ctx context.Context) error {
	return nil
}

func initCache(cfg config.RedisConfig) (repository.Cache, error) {
	if cfg.URL == "" {
		return nil, errors.New("redis url is required")
	}

	return &cacheStub{}, nil
}
