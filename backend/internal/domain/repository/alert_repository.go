package repository

import "context"

// AlertRepository is the alert persistence repository interface.
type AlertRepository interface {
	// Close shuts down any resources owned by the repository.
	Close(ctx context.Context) error
}
