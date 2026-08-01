package repository

import "context"

// DB is the primary database repository interface.
// All infrastructure adapters implement this.
type DB interface {
	// BeginTransaction starts a new database transaction.
	BeginTransaction(ctx context.Context) (Transaction, error)

	// WithTransaction executes fn within a transaction, committing on
	// success and rolling back on error.
	WithTransaction(ctx context.Context, fn func(tx Transaction) error) error

	// Raw executes raw SQL and scans into dest.
	Raw(ctx context.Context, query string, dest any, args ...any) error

	// Query executes a query and returns the resulting rows as maps.
	Query(ctx context.Context, query string, args ...any) ([]map[string]any, error)

	// Close closes the database connection.
	Close(ctx context.Context) error
}

// Transaction represents a database transaction.
type Transaction interface {
	DB
	Commit() error
	Rollback() error
}
