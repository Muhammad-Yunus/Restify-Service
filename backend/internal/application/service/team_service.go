package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// TeamService implements the service.TeamService interface.
type TeamService struct {
	repo   repository.TeamRepository
	logger repository.Logger
}

// NewTeamService creates a new team service.
func NewTeamService(gormDB *gorm.DB, logger repository.Logger) *TeamService {
	return &TeamService{
		repo:   &teamRepositoryImpl{db: gormDB},
		logger: logger,
	}
}

func (s *TeamService) Create(ctx context.Context, name string, workspaceID uuid.UUID) (*entity.Team, error) {
	team := &entity.Team{
		Name:        name,
		WorkspaceID: workspaceID,
	}

	if err := team.Validate(); err != nil {
		return nil, fmt.Errorf("validate team: %w", err)
	}

	if err := s.repo.Create(ctx, team); err != nil {
		return nil, fmt.Errorf("create team: %w", err)
	}

	s.logger.Info(ctx, "team created", "team_id", team.ID, "workspace_id", workspaceID)
	return team, nil
}

func (s *TeamService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Team, error) {
	team, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get team %s: %w", id, err)
	}
	return team, nil
}

func (s *TeamService) AddMember(ctx context.Context, teamID, userID uuid.UUID, role string) error {
	if err := s.repo.AddMember(ctx, teamID, userID, role); err != nil {
		return fmt.Errorf("add member to team: %w", err)
	}
	s.logger.Info(ctx, "team member added", "team_id", teamID, "user_id", userID)
	return nil
}

func (s *TeamService) RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error {
	if err := s.repo.RemoveMember(ctx, teamID, userID); err != nil {
		return fmt.Errorf("remove member from team: %w", err)
	}
	s.logger.Info(ctx, "team member removed", "team_id", teamID, "user_id", userID)
	return nil
}

func (s *TeamService) ListMembers(ctx context.Context, teamID uuid.UUID) ([]*entity.TeamMember, error) {
	return s.repo.ListMembers(ctx, teamID)
}

func (s *TeamService) AssignToWorkspace(ctx context.Context, teamID, workspaceID uuid.UUID, wsRole entity.TeamWorkspaceRole) error {
	if err := s.repo.AssignWorkspace(ctx, teamID, workspaceID, wsRole); err != nil {
		return fmt.Errorf("assign team to workspace: %w", err)
	}
	s.logger.Info(ctx, "team assigned to workspace", "team_id", teamID, "workspace_id", workspaceID, "role", wsRole)
	return nil
}
