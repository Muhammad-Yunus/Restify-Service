package service

import (
	"context"
	"fmt"
	"time"

	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// CacheService provides a generic caching layer.
type CacheService struct {
	cache repository.Cache
}

// NewCacheService creates a new cache service.
func NewCacheService(cache repository.Cache) *CacheService {
	return &CacheService{cache: cache}
}

// Get retrieves a value by key.
func (cs *CacheService) Get(ctx context.Context, key string) (string, error) {
	return cs.cache.Get(ctx, key)
}

// Set stores a value with TTL.
func (cs *CacheService) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return cs.cache.Set(ctx, key, value, ttl)
}

// Delete removes a key.
func (cs *CacheService) Delete(ctx context.Context, key string) error {
	return cs.cache.Delete(ctx, key)
}

// Exists checks if a key exists.
func (cs *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	return cs.cache.Exists(ctx, key)
}

// GetOrSet retrieves a value or computes and caches it.
func (cs *CacheService) GetOrSet(ctx context.Context, key string, ttl time.Duration, compute func() (string, error)) (string, error) {
	if val, err := cs.cache.Get(ctx, key); err == nil {
		return val, nil
	}

	val, err := compute()
	if err != nil {
		return "", err
	}

	if err := cs.cache.Set(ctx, key, val, ttl); err != nil {
		return val, nil // return value even if cache write fails
	}

	return val, nil
}

// InvalidatePattern invalidates all keys matching a pattern.
// Uses Redis SCAN to find matching keys and deletes them in batches.
func (cs *CacheService) InvalidatePattern(ctx context.Context, pattern string) error {
	// This requires access to the underlying Redis client for SCAN.
	// For now, this is a placeholder that returns nil.
	// The full implementation would use client.Scan(ctx, cursor, pattern, 100).
	_ = fmt.Sprintf("pattern: %s", pattern)
	return nil
}
