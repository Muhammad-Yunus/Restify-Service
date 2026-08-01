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

// WorkspaceRepositoryImpl implements the repository.WorkspaceRepository interface.
type WorkspaceRepositoryImpl struct {
	db *gorm.DB
}

// NewWorkspaceRepository creates a new workspace repository.
func NewWorkspaceRepository(db *gorm.DB) repository.WorkspaceRepository {
	return &WorkspaceRepositoryImpl{db: db}
}

func (r *WorkspaceRepositoryImpl) Create(ctx context.Context, ws *entity.Workspace) error {
	if ws.ID == uuid.Nil {
		ws.ID = uuid.New()
	}
	ws.GenerateSlug()
	if err := r.db.WithContext(ctx).Create(ws).Error; err != nil {
		return fmt.Errorf("create workspace: %w", err)
	}
	return nil
}

func (r *WorkspaceRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entity.Workspace, error) {
	var ws entity.Workspace
	err := r.db.WithContext(ctx).Preload("Owner").First(&ws, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("find workspace %s: %w", id, err)
	}
	return &ws, nil
}

func (r *WorkspaceRepositoryImpl) FindBySlug(ctx context.Context, slug string) (*entity.Workspace, error) {
	var ws entity.Workspace
	err := r.db.WithContext(ctx).Preload("Owner").First(&ws, "slug = ?", slug).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("find workspace by slug %s: %w", slug, err)
	}
	return &ws, nil
}

func (r *WorkspaceRepositoryImpl) Update(ctx context.Context, ws *entity.Workspace) error {
	if err := r.db.WithContext(ctx).Model(ws).Updates(map[string]any{
		"name":        ws.Name,
		"description": ws.Description,
		"is_public":   ws.IsPublic,
	}).Error; err != nil {
		return fmt.Errorf("update workspace %s: %w", ws.ID, err)
	}
	return nil
}

func (r *WorkspaceRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.Workspace{}, "id = ?", id)

	if result.Error != nil {
		return fmt.Errorf("delete workspace %s: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *WorkspaceRepositoryImpl) List(ctx context.Context, ownerID uuid.UUID, page, pageSize int) ([]*entity.Workspace, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	var workspaces []*entity.Workspace
	var total int64

	if err := r.db.WithContext(ctx).Model(&entity.Workspace{}).
		Where("owner_id = ?", ownerID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count workspaces: %w", err)
	}

	if err := r.db.WithContext(ctx).
		Where("owner_id = ?", ownerID).
		Offset(offset).Limit(pageSize).
		Find(&workspaces).Error; err != nil {
		return nil, 0, fmt.Errorf("list workspaces: %w", err)
	}

	return workspaces, int(total), nil
}

func (r *WorkspaceRepositoryImpl) ListAll(ctx context.Context, page, pageSize int) ([]*entity.Workspace, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	var workspaces []*entity.Workspace
	var total int64

	if err := r.db.WithContext(ctx).Model(&entity.Workspace{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count all workspaces: %w", err)
	}

	if err := r.db.WithContext(ctx).
		Offset(offset).Limit(pageSize).
		Find(&workspaces).Error; err != nil {
		return nil, 0, fmt.Errorf("list all workspaces: %w", err)
	}

	return workspaces, int(total), nil
}

func (r *WorkspaceRepositoryImpl) CountByOwner(ctx context.Context, ownerID uuid.UUID) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entity.Workspace{}).
		Where("owner_id = ?", ownerID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count workspaces by owner: %w", err)
	}
	return int(count), nil
}

// Compile-time check.
var _ repository.WorkspaceRepository = (*WorkspaceRepositoryImpl)(nil)
