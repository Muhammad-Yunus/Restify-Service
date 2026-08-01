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

// ServiceEndpointRepo implements repository.EndpointRepository for DI injection.
type ServiceEndpointRepo struct {
	DB *gorm.DB
}

func (r *ServiceEndpointRepo) Create(ctx context.Context, ep *entity.Endpoint) error {
	if ep.ID == uuid.Nil {
		ep.ID = uuid.New()
	}
	if ep.Version == "" {
		ep.Version = "v1"
	}
	if ep.Method == "" {
		ep.Method = "GET"
	}
	if err := r.DB.WithContext(ctx).Create(ep).Error; err != nil {
		return fmt.Errorf("create endpoint: %w", err)
	}
	return nil
}

func (r *ServiceEndpointRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.Endpoint, error) {
	var ep entity.Endpoint
	err := r.DB.WithContext(ctx).Preload("Collection").First(&ep, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("find endpoint %s: %w", id, err)
	}
	return &ep, nil
}

func (r *ServiceEndpointRepo) ListByCollection(ctx context.Context, collectionID uuid.UUID) ([]*entity.Endpoint, error) {
	var eps []*entity.Endpoint
	if err := r.DB.WithContext(ctx).
		Where("collection_id = ?", collectionID).
		Order("created_at ASC").
		Find(&eps).Error; err != nil {
		return nil, fmt.Errorf("list endpoints for collection %s: %w", collectionID, err)
	}
	return eps, nil
}

func (r *ServiceEndpointRepo) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entity.Endpoint, error) {
	var eps []*entity.Endpoint
	if err := r.DB.WithContext(ctx).
		Table("endpoints").
		Select("endpoints.*").
		Joins("JOIN collections ON endpoints.collection_id = collections.id").
		Where("collections.workspace_id = ?", workspaceID).
		Order("endpoints.created_at ASC").
		Find(&eps).Error; err != nil {
		return nil, fmt.Errorf("list endpoints for workspace %s: %w", workspaceID, err)
	}
	return eps, nil
}

func (r *ServiceEndpointRepo) Update(ctx context.Context, ep *entity.Endpoint) error {
	if err := r.DB.WithContext(ctx).Model(ep).Updates(map[string]any{
		"name":                 ep.Name,
		"description":          ep.Description,
		"path":                 ep.Path,
		"method":               ep.Method,
		"version":              ep.Version,
		"default_method":       ep.Method,
		"db_type":              ep.DBType,
		"schema":               ep.Schema,
		"table_name":           ep.TableName,
		"func_name":            ep.FuncName,
		"params":               ep.Params,
		"operations":           ep.Operations,
		"security_policy_json": ep.SecurityPolicyJSON,
		"auth_header":          ep.AuthHeader,
		"param_headers":        ep.ParamHeaders,
		"body_mapping_json":    ep.BodyMappingJSON,
	}).Error; err != nil {
		return fmt.Errorf("update endpoint %s: %w", ep.ID, err)
	}
	return nil
}

func (r *ServiceEndpointRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.DB.WithContext(ctx).Delete(&entity.Endpoint{}, "id = ?", id)

	if result.Error != nil {
		return fmt.Errorf("delete endpoint %s: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *ServiceEndpointRepo) ToggleActive(ctx context.Context, id uuid.UUID, active bool) error {
	result := r.DB.WithContext(ctx).Model(&entity.Endpoint{}).
		Where("id = ?", id).
		Update("is_active", active)

	if result.Error != nil {
		return fmt.Errorf("toggle endpoint %s: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *ServiceEndpointRepo) FindByPath(ctx context.Context, path, version string) (*entity.Endpoint, error) {
	var ep entity.Endpoint
	err := r.DB.WithContext(ctx).
		Where("path = ? AND version = ? AND is_active = ?", path, version, true).
		First(&ep).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("find endpoint by path: %w", err)
	}
	return &ep, nil
}

func (r *ServiceEndpointRepo) CountByWorkspace(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	var count int64
	if err := r.DB.WithContext(ctx).
		Table("endpoints").
		Select("count(*)").
		Joins("JOIN collections ON endpoints.collection_id = collections.id").
		Where("collections.workspace_id = ?", workspaceID).
		Scan(&count).Error; err != nil {
		return 0, fmt.Errorf("count endpoints: %w", err)
	}
	return int(count), nil
}

func (r *ServiceEndpointRepo) ListAllActive(ctx context.Context) ([]*entity.Endpoint, error) {
	var eps []*entity.Endpoint
	if err := r.DB.WithContext(ctx).
		Where("is_active = ?", true).
		Order("created_at ASC").
		Find(&eps).Error; err != nil {
		return nil, fmt.Errorf("list all active endpoints: %w", err)
	}
	return eps, nil
}

// Compile-time check.
var _ repository.EndpointRepository = (*ServiceEndpointRepo)(nil)
