package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// collectionRepositoryImpl implements the repository.CollectionRepository interface.
type collectionRepositoryImpl struct {
	db *gorm.DB
}

func (r *collectionRepositoryImpl) Create(ctx context.Context, col *entity.Collection) error {
	if col.ID == uuid.Nil {
		col.ID = uuid.New()
	}
	col.GenerateSlug()
	if err := r.db.WithContext(ctx).Create(col).Error; err != nil {
		return fmt.Errorf("create collection: %w", err)
	}
	return nil
}

func (r *collectionRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entity.Collection, error) {
	var col entity.Collection
	err := r.db.WithContext(ctx).Preload("Endpoints").First(&col, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("find collection %s: %w", id, err)
	}
	return &col, nil
}

func (r *collectionRepositoryImpl) FindBySlug(ctx context.Context, workspaceID uuid.UUID, slug string) (*entity.Collection, error) {
	var col entity.Collection
	err := r.db.WithContext(ctx).Preload("Endpoints").
		Where("workspace_id = ? AND slug = ?", workspaceID, slug).
		First(&col).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("find collection by slug %s in workspace %s: %w", slug, workspaceID, err)
	}
	return &col, nil
}

func (r *collectionRepositoryImpl) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entity.Collection, error) {
	var collections []*entity.Collection
	if err := r.db.WithContext(ctx).
		Preload("Endpoints").
		Where("workspace_id = ?", workspaceID).
		Find(&collections).Error; err != nil {
		return nil, fmt.Errorf("list collections for workspace %s: %w", workspaceID, err)
	}
	return collections, nil
}

func (r *collectionRepositoryImpl) Update(ctx context.Context, col *entity.Collection) error {
	if err := r.db.WithContext(ctx).Model(col).Updates(map[string]any{
		"name":        col.Name,
		"description": col.Description,
	}).Error; err != nil {
		return fmt.Errorf("update collection %s: %w", col.ID, err)
	}
	return nil
}

func (r *collectionRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	// Delete all endpoints first (cascade)
	if err := r.db.WithContext(ctx).Delete(&entity.Endpoint{}, "collection_id = ?", id).Error; err != nil {
		return fmt.Errorf("delete endpoints: %w", err)
	}
	result := r.db.WithContext(ctx).Delete(&entity.Collection{}, "id = ?", id)

	if result.Error != nil {
		return fmt.Errorf("delete collection %s: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *collectionRepositoryImpl) CountByWorkspace(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entity.Collection{}).
		Where("workspace_id = ?", workspaceID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count collections: %w", err)
	}
	return int(count), nil
}

// Compile-time check.
var _ repository.CollectionRepository = (*collectionRepositoryImpl)(nil)
