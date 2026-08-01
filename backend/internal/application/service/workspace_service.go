package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// WorkspaceService implements the service.WorkspaceService interface.
type WorkspaceService struct {
	repo   repository.WorkspaceRepository
	logger repository.Logger
}

// NewWorkspaceService creates a new workspace service.
func NewWorkspaceService(gormDB *gorm.DB, logger repository.Logger) *WorkspaceService {
	return &WorkspaceService{
		repo:   &workspaceRepositoryImpl{db: gormDB},
		logger: logger,
	}
}

func (s *WorkspaceService) Create(ctx context.Context, name, description string, ownerID uuid.UUID) (*entity.Workspace, error) {
	ws := &entity.Workspace{
		Name:        name,
		Description: &description,
		OwnerID:     ownerID,
		IsPublic:    false,
	}

	if err := ws.Validate(); err != nil {
		return nil, fmt.Errorf("validate workspace: %w", err)
	}

	if err := s.repo.Create(ctx, ws); err != nil {
		return nil, fmt.Errorf("create workspace: %w", err)
	}

	s.logger.Info(ctx, "workspace created", "workspace_id", ws.ID, "owner_id", ownerID)
	return ws, nil
}

func (s *WorkspaceService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Workspace, error) {
	ws, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get workspace %s: %w", id, err)
	}
	return ws, nil
}

func (s *WorkspaceService) Update(ctx context.Context, id uuid.UUID, updates map[string]any) (*entity.Workspace, error) {
	ws, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name, ok := updates["name"].(string); ok && name != "" {
		ws.Name = name
	}
	if desc, ok := updates["description"].(string); ok {
		ws.Description = &desc
	}
	if pub, ok := updates["is_public"].(bool); ok {
		ws.IsPublic = pub
	}

	if err := ws.Validate(); err != nil {
		return nil, fmt.Errorf("validate workspace: %w", err)
	}

	if err := s.repo.Update(ctx, ws); err != nil {
		return nil, fmt.Errorf("update workspace %s: %w", id, err)
	}

	s.logger.Info(ctx, "workspace updated", "workspace_id", id)
	return ws, nil
}

func (s *WorkspaceService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete workspace %s: %w", id, err)
	}
	s.logger.Info(ctx, "workspace deleted", "workspace_id", id)
	return nil
}

func (s *WorkspaceService) List(ctx context.Context, ownerID uuid.UUID, page, pageSize int) ([]*entity.Workspace, int, error) {
	return s.repo.List(ctx, ownerID, page, pageSize)
}
