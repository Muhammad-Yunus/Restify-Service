package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// RedisCache implements the repository.Cache interface.
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache connection.
func NewRedisCache(ctx context.Context, url string) (*RedisCache, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

// Get retrieves a value by key.
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("redis get %s: %w", key, err)
	}

	return val, nil
}

// Set stores a value with TTL.
func (r *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("redis set %s: %w", key, err)
	}

	return nil
}

// Delete removes a key.
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis del %s: %w", key, err)
	}

	return nil
}

// Exists checks if a key exists.
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists %s: %w", key, err)
	}

	return count > 0, nil
}

// Close shuts down the cache connection.
func (r *RedisCache) Close(ctx context.Context) error {
	if err := r.client.Close(); err != nil {
		return fmt.Errorf("close redis: %w", err)
	}

	return nil
}

// Compile-time check.
var _ repository.Cache = (*RedisCache)(nil)
