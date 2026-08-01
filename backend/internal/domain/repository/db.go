package repository

import "context"

// DB is the primary database repository interface.
type DB interface {
	// Close shuts down the database connection pool.
	Close(ctx context.Context) error
}
