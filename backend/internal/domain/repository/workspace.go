package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
)

// WorkspaceRepository defines the data access contract for workspaces.
type WorkspaceRepository interface {
	// Create inserts a new workspace and returns it with generated ID.
	Create(ctx context.Context, ws *entity.Workspace) error

	// FindByID returns a workspace by UUID.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Workspace, error)

	// FindBySlug returns a workspace by URL slug.
	FindBySlug(ctx context.Context, slug string) (*entity.Workspace, error)

	// Update partially updates a workspace.
	Update(ctx context.Context, ws *entity.Workspace) error

	// Delete removes a workspace.
	Delete(ctx context.Context, id uuid.UUID) error

	// List returns paginated workspaces owned by a user.
	List(ctx context.Context, ownerID uuid.UUID, page, pageSize int) ([]*entity.Workspace, int, error)

	// ListAll returns all workspaces (admin only).
	ListAll(ctx context.Context, page, pageSize int) ([]*entity.Workspace, int, error)

	// CountByOwner returns how many workspaces a user owns.
	CountByOwner(ctx context.Context, ownerID uuid.UUID) (int, error)
}
