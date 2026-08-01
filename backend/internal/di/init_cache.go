package di

import (
	"context"
	"fmt"
	"time"

	"github.com/muhammadyunus/Restify-Service/internal/config"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
	cacheinfra "github.com/muhammadyunus/Restify-Service/internal/infrastructure/cache"
	"github.com/redis/go-redis/v9"
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
		return &cacheStub{}, nil
	}

	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse redis URL: %w", err)
	}

	client := redis.NewClient(opt)

	if err := client.Ping(context.Background()).Err(); err != nil {
		// Return stub if Redis is not available, but log the error
		fmt.Printf("Warning: Redis not available, using stub cache: %v\n", err)
		return &cacheStub{}, nil
	}

	redisCache, err := cacheinfra.NewRedisCache(context.Background(), cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("create redis cache: %w", err)
	}

	// Also create distributed lock
	_ = cacheinfra.NewDistributedLock(client)

	return redisCache, nil
}
