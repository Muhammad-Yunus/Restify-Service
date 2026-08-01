package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
)

// CollectionRepository defines the data access contract for collections.
type CollectionRepository interface {
	// Create inserts a new collection.
	Create(ctx context.Context, col *entity.Collection) error

	// FindByID returns a collection by UUID.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Collection, error)

	// FindBySlug returns a collection by slug within a workspace.
	FindBySlug(ctx context.Context, workspaceID uuid.UUID, slug string) (*entity.Collection, error)

	// ListByWorkspace returns all collections in a workspace.
	ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entity.Collection, error)

	// Update partially updates a collection.
	Update(ctx context.Context, col *entity.Collection) error

	// Delete removes a collection (and all its endpoints).
	Delete(ctx context.Context, id uuid.UUID) error

	// CountByWorkspace returns the number of collections in a workspace.
	CountByWorkspace(ctx context.Context, workspaceID uuid.UUID) (int, error)
}
