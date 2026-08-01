package repository

import "context"

// Cache is the caching repository interface.
type Cache interface {
	// Close shuts down the cache connection.
	Close(ctx context.Context) error
}
