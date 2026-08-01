package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
)

// APILogRepository defines the data access contract for API logs.
type APILogRepository interface {
	// Create inserts a log entry.
	Create(ctx context.Context, log *entity.APILog) error

	// CreateBatch inserts multiple log entries.
	CreateBatch(ctx context.Context, logs []*entity.APILog) error

	// FindByID returns a log entry by UUID.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.APILog, error)

	// Search returns paginated logs matching filters.
	Search(ctx context.Context, workspaceID, endpointID uuid.UUID,
		level entity.LogLevel, method, path string,
		from, to time.Time, page, pageSize int) ([]*entity.APILog, int, error)

	// DeleteOlderThan removes logs older than the given date.
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)

	// CountByWorkspace returns log count for a workspace in a time range.
	CountByWorkspace(ctx context.Context, workspaceID uuid.UUID, from, to time.Time) (int64, error)

	// Close shuts down any resources owned by the repository.
	Close(ctx context.Context) error
}
