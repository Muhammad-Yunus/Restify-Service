package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// EndpointService implements the service.EndpointService interface.
type EndpointService struct {
	repo   repository.EndpointRepository
	logger repository.Logger
}

// NewEndpointService creates a new endpoint service.
func NewEndpointService(gormDB *gorm.DB, logger repository.Logger) *EndpointService {
	return &EndpointService{
		repo:   &endpointRepositoryImpl{db: gormDB},
		logger: logger,
	}
}

func (s *EndpointService) Create(ctx context.Context, collectionID uuid.UUID, params map[string]any) (*entity.Endpoint, error) {
	ep := &entity.Endpoint{
		CollectionID: collectionID,
		Version:      "v1",
		Method:       "GET",
		IsActive:     true,
	}

	// Map params to endpoint fields
	if name, ok := params["name"].(string); ok {
		ep.Name = name
	}
	if desc, ok := params["description"].(string); ok {
		ep.Description = &desc
	}
	if path, ok := params["path"].(string); ok {
		ep.Path = path
	}
	if method, ok := params["method"].(string); ok {
		ep.Method = method
	}
	if dbType, ok := params["db_type"].(string); ok {
		ep.DBType = entity.EndpointType(dbType)
	}
	if schema, ok := params["schema"].(string); ok {
		ep.Schema = schema
	}
	if tableName, ok := params["table_name"].(string); ok {
		ep.TableName = tableName
	}
	if funcName, ok := params["func_name"].(string); ok {
		ep.FuncName = funcName
	}
	if ops, ok := params["operations"].([]any); ok {
		opsJSON, _ := json.Marshal(ops)
		ep.Operations = opsJSON
	}
	if security, ok := params["security"].(map[string]any); ok {
		secJSON, _ := json.Marshal(security)
		ep.SecurityPolicyJSON = secJSON
	}

	if err := ep.Validate(); err != nil {
		return nil, fmt.Errorf("validate endpoint: %w", err)
	}

	if err := s.repo.Create(ctx, ep); err != nil {
		return nil, fmt.Errorf("create endpoint: %w", err)
	}

	s.logger.Info(ctx, "endpoint created", "endpoint_id", ep.ID, "collection_id", collectionID)
	return ep, nil
}

func (s *EndpointService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Endpoint, error) {
	ep, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get endpoint %s: %w", id, err)
	}
	return ep, nil
}

func (s *EndpointService) Update(ctx context.Context, id uuid.UUID, updates map[string]any) (*entity.Endpoint, error) {
	ep, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name, ok := updates["name"].(string); ok {
		ep.Name = name
	}
	if desc, ok := updates["description"].(string); ok {
		ep.Description = &desc
	}
	if path, ok := updates["path"].(string); ok {
		ep.Path = path
	}
	if method, ok := updates["method"].(string); ok {
		ep.Method = method
	}
	if tableName, ok := updates["table_name"].(string); ok {
		ep.TableName = tableName
	}
	if funcName, ok := updates["func_name"].(string); ok {
		ep.FuncName = funcName
	}
	if schema, ok := updates["schema"].(string); ok {
		ep.Schema = schema
	}
	if dbType, ok := updates["db_type"].(string); ok {
		ep.DBType = entity.EndpointType(dbType)
	}
	if ops, ok := updates["operations"].([]any); ok {
		opsJSON, _ := json.Marshal(ops)
		ep.Operations = opsJSON
	}
	if security, ok := updates["security"].(map[string]any); ok {
		secJSON, _ := json.Marshal(security)
		ep.SecurityPolicyJSON = secJSON
	}

	if err := ep.Validate(); err != nil {
		return nil, fmt.Errorf("validate endpoint: %w", err)
	}

	if err := s.repo.Update(ctx, ep); err != nil {
		return nil, fmt.Errorf("update endpoint %s: %w", id, err)
	}

	s.logger.Info(ctx, "endpoint updated", "endpoint_id", id)
	return ep, nil
}

func (s *EndpointService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete endpoint %s: %w", id, err)
	}
	s.logger.Info(ctx, "endpoint deleted", "endpoint_id", id)
	return nil
}

func (s *EndpointService) List(ctx context.Context, collectionID uuid.UUID) ([]*entity.Endpoint, error) {
	return s.repo.ListByCollection(ctx, collectionID)
}

func (s *EndpointService) ToggleActive(ctx context.Context, id uuid.UUID, active bool) error {
	if err := s.repo.ToggleActive(ctx, id, active); err != nil {
		return fmt.Errorf("toggle endpoint %s: %w", id, err)
	}
	s.logger.Info(ctx, "endpoint toggled", "endpoint_id", id, "active", active)
	return nil
}

func (s *EndpointService) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entity.Endpoint, error) {
	return s.repo.ListByWorkspace(ctx, workspaceID)
}
