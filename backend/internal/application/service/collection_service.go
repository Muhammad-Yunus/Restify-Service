package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// CollectionService implements the service.CollectionService interface.
type CollectionService struct {
	repo   repository.CollectionRepository
	logger repository.Logger
}

// NewCollectionService creates a new collection service.
func NewCollectionService(gormDB *gorm.DB, logger repository.Logger) *CollectionService {
	return &CollectionService{
		repo:   &collectionRepositoryImpl{db: gormDB},
		logger: logger,
	}
}

func (s *CollectionService) Create(ctx context.Context, name, description string, workspaceID uuid.UUID) (*entity.Collection, error) {
	col := &entity.Collection{
		Name:        name,
		Description: &description,
		WorkspaceID: workspaceID,
	}

	if err := col.Validate(); err != nil {
		return nil, fmt.Errorf("validate collection: %w", err)
	}

	if err := s.repo.Create(ctx, col); err != nil {
		return nil, fmt.Errorf("create collection: %w", err)
	}

	s.logger.Info(ctx, "collection created", "collection_id", col.ID, "workspace_id", workspaceID)
	return col, nil
}

func (s *CollectionService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Collection, error) {
	col, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get collection %s: %w", id, err)
	}
	return col, nil
}

func (s *CollectionService) Update(ctx context.Context, id uuid.UUID, updates map[string]any) (*entity.Collection, error) {
	col, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name, ok := updates["name"].(string); ok && name != "" {
		col.Name = name
	}
	if desc, ok := updates["description"].(string); ok {
		col.Description = &desc
	}

	if err := col.Validate(); err != nil {
		return nil, fmt.Errorf("validate collection: %w", err)
	}

	if err := s.repo.Update(ctx, col); err != nil {
		return nil, fmt.Errorf("update collection %s: %w", id, err)
	}

	s.logger.Info(ctx, "collection updated", "collection_id", id)
	return col, nil
}

func (s *CollectionService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete collection %s: %w", id, err)
	}
	s.logger.Info(ctx, "collection deleted", "collection_id", id)
	return nil
}

func (s *CollectionService) List(ctx context.Context, workspaceID uuid.UUID) ([]*entity.Collection, error) {
	return s.repo.ListByWorkspace(ctx, workspaceID)
}
