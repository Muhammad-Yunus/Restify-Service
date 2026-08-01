package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
)

// EndpointRepository defines the data access contract for endpoints.
type EndpointRepository interface {
	// Create inserts a new endpoint.
	Create(ctx context.Context, ep *entity.Endpoint) error

	// FindByID returns an endpoint by UUID.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Endpoint, error)

	// ListByCollection returns all endpoints in a collection.
	ListByCollection(ctx context.Context, collectionID uuid.UUID) ([]*entity.Endpoint, error)

	// ListByWorkspace returns all endpoints across all collections in a workspace.
	ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entity.Endpoint, error)

	// Update partially updates an endpoint.
	Update(ctx context.Context, ep *entity.Endpoint) error

	// Delete removes an endpoint.
	Delete(ctx context.Context, id uuid.UUID) error

	// ToggleActive enables or disables an endpoint.
	ToggleActive(ctx context.Context, id uuid.UUID, active bool) error

	// FindByPath returns an endpoint matching a path and version.
	FindByPath(ctx context.Context, path, version string) (*entity.Endpoint, error)

	// CountByWorkspace returns the number of endpoints in a workspace.
	CountByWorkspace(ctx context.Context, workspaceID uuid.UUID) (int, error)

	// ListAllActive returns all active endpoints across all workspaces.
	ListAllActive(ctx context.Context) ([]*entity.Endpoint, error)
}
