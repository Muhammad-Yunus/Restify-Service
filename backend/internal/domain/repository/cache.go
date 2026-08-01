package repository

import (
	"context"
	"time"
)

// Cache is the caching repository interface.
type Cache interface {
	// Get retrieves a value by key.
	Get(ctx context.Context, key string) (string, error)

	// Set stores a value with TTL.
	Set(ctx context.Context, key string, value string, ttl time.Duration) error

	// Delete removes a key.
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists.
	Exists(ctx context.Context, key string) (bool, error)

	// Close shuts down the cache connection.
	Close(ctx context.Context) error
}
