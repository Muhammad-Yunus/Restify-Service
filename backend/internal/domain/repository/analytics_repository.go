package repository

import "context"

// AnalyticsRepository is the API analytics repository interface.
type AnalyticsRepository interface {
	// Close shuts down any resources owned by the repository.
	Close(ctx context.Context) error
}
