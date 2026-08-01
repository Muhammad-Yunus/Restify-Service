package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// TeamRepositoryImpl implements the repository.TeamRepository interface.
type TeamRepositoryImpl struct {
	db *gorm.DB
}

// NewTeamRepository creates a new team repository.
func NewTeamRepository(db *gorm.DB) repository.TeamRepository {
	return &TeamRepositoryImpl{db: db}
}

func (r *TeamRepositoryImpl) Create(ctx context.Context, team *entity.Team) error {
	if team.ID == uuid.Nil {
		team.ID = uuid.New()
	}
	if err := r.db.WithContext(ctx).Create(team).Error; err != nil {
		return fmt.Errorf("create team: %w", err)
	}
	return nil
}

func (r *TeamRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entity.Team, error) {
	var team entity.Team
	err := r.db.WithContext(ctx).Preload("Members.User.Roles").First(&team, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("find team %s: %w", id, err)
	}
	return &team, nil
}

func (r *TeamRepositoryImpl) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entity.Team, error) {
	var teams []*entity.Team
	if err := r.db.WithContext(ctx).
		Preload("Members.User.Roles").
		Where("workspace_id = ?", workspaceID).
		Find(&teams).Error; err != nil {
		return nil, fmt.Errorf("list teams for workspace %s: %w", workspaceID, err)
	}
	return teams, nil
}

func (r *TeamRepositoryImpl) Update(ctx context.Context, team *entity.Team) error {
	if err := r.db.WithContext(ctx).Model(team).Update("name", team.Name).Error; err != nil {
		return fmt.Errorf("update team %s: %w", team.ID, err)
	}
	return nil
}

func (r *TeamRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.Team{}, "id = ?", id)

	if result.Error != nil {
		return fmt.Errorf("delete team %s: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *TeamRepositoryImpl) AddMember(ctx context.Context, teamID, userID uuid.UUID, role string) error {
	member := &entity.TeamMember{
		TeamID: teamID,
		UserID: userID,
		Role:   role,
	}
	if err := r.db.WithContext(ctx).Create(member).Error; err != nil {
		return fmt.Errorf("add member to team: %w", err)
	}
	return nil
}

func (r *TeamRepositoryImpl) RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Delete(&entity.TeamMember{})

	if result.Error != nil {
		return fmt.Errorf("remove member: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *TeamRepositoryImpl) GetMember(ctx context.Context, teamID, userID uuid.UUID) (*entity.TeamMember, error) {
	var member entity.TeamMember
	err := r.db.WithContext(ctx).
		Preload("User.Roles").
		Where("team_id = ? AND user_id = ?", teamID, userID).
		First(&member).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("get team member: %w", err)
	}
	return &member, nil
}

func (r *TeamRepositoryImpl) ListMembers(ctx context.Context, teamID uuid.UUID) ([]*entity.TeamMember, error) {
	var members []*entity.TeamMember
	if err := r.db.WithContext(ctx).
		Preload("User.Roles").
		Where("team_id = ?", teamID).
		Find(&members).Error; err != nil {
		return nil, fmt.Errorf("list team members: %w", err)
	}
	return members, nil
}

func (r *TeamRepositoryImpl) AssignWorkspace(ctx context.Context, teamID, workspaceID uuid.UUID, role entity.TeamWorkspaceRole) error {
	// Uses workspace_teams join table
	type WorkspaceTeam struct {
		TeamID      uuid.UUID `gorm:"column:team_id;primaryKey"`
		WorkspaceID uuid.UUID `gorm:"column:workspace_id;primaryKey"`
		Role        string    `gorm:"Column:role;type:varchar(50)"`
	}
	wt := WorkspaceTeam{TeamID: teamID, WorkspaceID: workspaceID, Role: string(role)}
	if err := r.db.WithContext(ctx).Table("workspace_teams").
		Where("team_id = ? AND workspace_id = ?", teamID, workspaceID).
		First(&wt).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new assignment
		if err := r.db.WithContext(ctx).Table("workspace_teams").Create(&wt).Error; err != nil {
			return fmt.Errorf("assign workspace to team: %w", err)
		}
	} else {
		// Update existing assignment
		if err := r.db.WithContext(ctx).Table("workspace_teams").
			Where("team_id = ? AND workspace_id = ?", teamID, workspaceID).
			Update("role", string(role)).Error; err != nil {
			return fmt.Errorf("update workspace assignment: %w", err)
		}
	}
	return nil
}

func (r *TeamRepositoryImpl) GetWorkspaceAccess(ctx context.Context, teamID uuid.UUID) ([]*entity.Workspace, error) {
	var workspaces []*entity.Workspace
	if err := r.db.WithContext(ctx).
		Table("workspace_teams").
		Select("workspaces.*").
		Joins("LEFT JOIN workspaces ON workspace_teams.workspace_id = workspaces.id").
		Where("workspace_teams.team_id = ?", teamID).
		Find(&workspaces).Error; err != nil {
		return nil, fmt.Errorf("get workspace access: %w", err)
	}
	return workspaces, nil
}

// Compile-time check.
var _ repository.TeamRepository = (*TeamRepositoryImpl)(nil)
