package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
)

// UserRepository defines the data access contract for users.
type UserRepository interface {
	// Create inserts a new user and returns it with generated ID.
	Create(ctx context.Context, user *entity.User) error

	// FindByID returns a user by UUID, returns entity.ErrNotFound if not found.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)

	// FindByEmail returns a user by email, returns entity.ErrNotFound if not found.
	FindByEmail(ctx context.Context, email string) (*entity.User, error)

	// Update partially updates a user (only non-nil fields).
	Update(ctx context.Context, user *entity.User) error

	// Delete marks a user as inactive (soft delete).
	Delete(ctx context.Context, id uuid.UUID) error

	// List returns paginated users.
	List(ctx context.Context, page, pageSize int) ([]*entity.User, int, error)

	// CountActive returns the number of active users.
	CountActive(ctx context.Context) (int, error)
}
