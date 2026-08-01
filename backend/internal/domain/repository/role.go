package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
)

// RoleRepository defines the data access contract for roles.
type RoleRepository interface {
	// Create inserts a new role.
	Create(ctx context.Context, role *entity.Role) error

	// FindByID returns a role by UUID.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Role, error)

	// FindByName returns a role by name.
	FindByName(ctx context.Context, name string) (*entity.Role, error)

	// List returns all roles.
	List(ctx context.Context) ([]*entity.Role, error)

	// Update partially updates a role.
	Update(ctx context.Context, role *entity.Role) error

	// Delete removes a role (only if not system role).
	Delete(ctx context.Context, id uuid.UUID) error

	// AssignUser assigns a role to a user.
	AssignUser(ctx context.Context, userID, roleID uuid.UUID) error

	// RemoveUser removes a role assignment from a user.
	RemoveUser(ctx context.Context, userID, roleID uuid.UUID) error

	// GetUserRoles returns all roles for a user.
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*entity.Role, error)
}
