package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
)

// TeamRepository defines the data access contract for teams.
type TeamRepository interface {
	// Create inserts a new team.
	Create(ctx context.Context, team *entity.Team) error

	// FindByID returns a team by UUID.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Team, error)

	// ListByWorkspace returns all teams in a workspace.
	ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entity.Team, error)

	// Update partially updates a team.
	Update(ctx context.Context, team *entity.Team) error

	// Delete removes a team.
	Delete(ctx context.Context, id uuid.UUID) error

	// AddMember adds a user to a team.
	AddMember(ctx context.Context, teamID, userID uuid.UUID, role string) error

	// RemoveMember removes a user from a team.
	RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error

	// GetMember returns a team member's info.
	GetMember(ctx context.Context, teamID, userID uuid.UUID) (*entity.TeamMember, error)

	// ListMembers returns all members of a team.
	ListMembers(ctx context.Context, teamID uuid.UUID) ([]*entity.TeamMember, error)

	// AssignWorkspace assigns a team to a workspace with a role.
	AssignWorkspace(ctx context.Context, teamID, workspaceID uuid.UUID, role entity.TeamWorkspaceRole) error

	// GetWorkspaceAccess returns all workspaces a team has access to.
	GetWorkspaceAccess(ctx context.Context, teamID uuid.UUID) ([]*entity.Workspace, error)
}
