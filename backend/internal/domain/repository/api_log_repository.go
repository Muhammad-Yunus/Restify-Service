package repository

import "context"

// APILogRepository is the API request log repository interface.
type APILogRepository interface {
	// Close shuts down any resources owned by the repository.
	Close(ctx context.Context) error
}
