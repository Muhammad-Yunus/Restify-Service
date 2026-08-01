package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	domservice "github.com/muhammadyunus/Restify-Service/internal/domain/service"
)

// CachedWorkspaceService wraps WorkspaceService with caching.
type CachedWorkspaceService struct {
	inner domservice.WorkspaceService
	cache *CacheService
	TTL   time.Duration
}

// NewCachedWorkspaceService creates a cached workspace service.
func NewCachedWorkspaceService(inner domservice.WorkspaceService, cache *CacheService) *CachedWorkspaceService {
	return &CachedWorkspaceService{
		inner: inner,
		cache: cache,
		TTL:   5 * time.Minute,
	}
}

func (cs *CachedWorkspaceService) Create(ctx context.Context, name, description string, ownerID uuid.UUID) (*entity.Workspace, error) {
	ws, err := cs.inner.Create(ctx, name, description, ownerID)
	if err != nil {
		return nil, err
	}
	cs.cache.Delete(ctx, cs.listKey(ownerID))
	return ws, nil
}

func (cs *CachedWorkspaceService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Workspace, error) {
	key := cs.key(id)
	val, err := cs.cache.GetOrSet(ctx, key, cs.TTL, func() (string, error) {
		ws, err := cs.inner.GetByID(ctx, id)
		if err != nil {
			return "", err
		}
		data, err := json.Marshal(ws)
		if err != nil {
			return "", fmt.Errorf("marshal workspace: %w", err)
		}
		return string(data), ErrNotFound
	})
	if err == ErrNotFound {
		return nil, entity.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var ws entity.Workspace
	if err := json.Unmarshal([]byte(val), &ws); err != nil {
		return nil, fmt.Errorf("unmarshal workspace: %w", err)
	}
	return &ws, nil
}

func (cs *CachedWorkspaceService) List(ctx context.Context, ownerID uuid.UUID, page, pageSize int) ([]*entity.Workspace, int, error) {
	key := cs.listKey(ownerID)
	val, err := cs.cache.GetOrSet(ctx, key, cs.TTL, func() (string, error) {
		wss, total, err := cs.inner.List(ctx, ownerID, page, pageSize)
		if err != nil {
			return "", err
		}
		data, err := json.Marshal(map[string]any{"items": wss, "total": total})
		if err != nil {
			return "", fmt.Errorf("marshal workspace list: %w", err)
		}
		return string(data), nil
	})
	if err != nil {
		return nil, 0, err
	}
	var result struct {
		Items []*entity.Workspace `json:"items"`
		Total int                 `json:"total"`
	}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, 0, fmt.Errorf("unmarshal workspace list: %w", err)
	}
	return result.Items, result.Total, nil
}

func (cs *CachedWorkspaceService) Update(ctx context.Context, id uuid.UUID, updates map[string]any) (*entity.Workspace, error) {
	ws, err := cs.inner.Update(ctx, id, updates)
	if err != nil {
		return nil, err
	}
	cs.cache.Delete(ctx, cs.key(id))
	return ws, nil
}

func (cs *CachedWorkspaceService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := cs.inner.Delete(ctx, id); err != nil {
		return err
	}
	cs.cache.Delete(ctx, cs.key(id))
	return nil
}

func (cs *CachedWorkspaceService) key(id uuid.UUID) string {
	return fmt.Sprintf("workspace:%s", id)
}

func (cs *CachedWorkspaceService) listKey(ownerID uuid.UUID) string {
	return fmt.Sprintf("workspaces:list:%s", ownerID)
}

// CachedUserService wraps UserService with caching.
type CachedUserService struct {
	inner domservice.UserService
	cache *CacheService
	TTL   time.Duration
}

// NewCachedUserService creates a cached user service.
func NewCachedUserService(inner domservice.UserService, cache *CacheService) *CachedUserService {
	return &CachedUserService{
		inner: inner,
		cache: cache,
		TTL:   3 * time.Minute,
	}
}

func (cs *CachedUserService) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	key := cs.key(id)
	val, err := cs.cache.GetOrSet(ctx, key, cs.TTL, func() (string, error) {
		u, err := cs.inner.GetByID(ctx, id)
		if err != nil {
			return "", err
		}
		data, err := json.Marshal(u)
		if err != nil {
			return "", fmt.Errorf("marshal user: %w", err)
		}
		return string(data), ErrNotFound
	})
	if err == ErrNotFound {
		return nil, entity.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var u entity.User
	if err := json.Unmarshal([]byte(val), &u); err != nil {
		return nil, fmt.Errorf("unmarshal user: %w", err)
	}
	return &u, nil
}

func (cs *CachedUserService) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	key := cs.emailKey(email)
	val, err := cs.cache.GetOrSet(ctx, key, cs.TTL, func() (string, error) {
		u, err := cs.inner.GetByEmail(ctx, email)
		if err != nil {
			return "", err
		}
		data, err := json.Marshal(u)
		if err != nil {
			return "", fmt.Errorf("marshal user: %w", err)
		}
		return string(data), ErrNotFound
	})
	if err == ErrNotFound {
		return nil, entity.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var u entity.User
	if err := json.Unmarshal([]byte(val), &u); err != nil {
		return nil, fmt.Errorf("unmarshal user: %w", err)
	}
	return &u, nil
}

func (cs *CachedUserService) Update(ctx context.Context, id uuid.UUID, updates map[string]any) (*entity.User, error) {
	u, err := cs.inner.Update(ctx, id, updates)
	if err != nil {
		return nil, err
	}
	cs.cache.Delete(ctx, cs.key(id))
	return u, nil
}

func (cs *CachedUserService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := cs.inner.Delete(ctx, id); err != nil {
		return err
	}
	cs.cache.Delete(ctx, cs.key(id))
	return nil
}

func (cs *CachedUserService) List(ctx context.Context, page, pageSize int) ([]*entity.User, int, error) {
	return cs.inner.List(ctx, page, pageSize)
}

func (cs *CachedUserService) key(id uuid.UUID) string {
	return fmt.Sprintf("user:%s", id)
}

func (cs *CachedUserService) emailKey(email string) string {
	return fmt.Sprintf("user:email:%s", email)
}

// CachedEndpointService wraps EndpointService with caching.
type CachedEndpointService struct {
	inner domservice.EndpointService
	cache *CacheService
	TTL   time.Duration
}

// NewCachedEndpointService creates a cached endpoint service.
func NewCachedEndpointService(inner domservice.EndpointService, cache *CacheService) *CachedEndpointService {
	return &CachedEndpointService{
		inner: inner,
		cache: cache,
		TTL:   2 * time.Minute,
	}
}

func (cs *CachedEndpointService) Create(ctx context.Context, collectionID uuid.UUID, params map[string]any) (*entity.Endpoint, error) {
	return cs.inner.Create(ctx, collectionID, params)
}

func (cs *CachedEndpointService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Endpoint, error) {
	key := cs.key(id)
	val, err := cs.cache.GetOrSet(ctx, key, cs.TTL, func() (string, error) {
		ep, err := cs.inner.GetByID(ctx, id)
		if err != nil {
			return "", err
		}
		data, err := json.Marshal(ep)
		if err != nil {
			return "", fmt.Errorf("marshal endpoint: %w", err)
		}
		return string(data), ErrNotFound
	})
	if err == ErrNotFound {
		return nil, entity.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var ep entity.Endpoint
	if err := json.Unmarshal([]byte(val), &ep); err != nil {
		return nil, fmt.Errorf("unmarshal endpoint: %w", err)
	}
	return &ep, nil
}

func (cs *CachedEndpointService) Update(ctx context.Context, id uuid.UUID, updates map[string]any) (*entity.Endpoint, error) {
	ep, err := cs.inner.Update(ctx, id, updates)
	if err != nil {
		return nil, err
	}
	cs.cache.Delete(ctx, cs.key(id))
	return ep, nil
}

func (cs *CachedEndpointService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := cs.inner.Delete(ctx, id); err != nil {
		return err
	}
	cs.cache.Delete(ctx, cs.key(id))
	return nil
}

func (cs *CachedEndpointService) List(ctx context.Context, collectionID uuid.UUID) ([]*entity.Endpoint, error) {
	key := cs.listKey(collectionID)
	val, err := cs.cache.GetOrSet(ctx, key, cs.TTL, func() (string, error) {
		eps, err := cs.inner.List(ctx, collectionID)
		if err != nil {
			return "", err
		}
		data, err := json.Marshal(eps)
		if err != nil {
			return "", fmt.Errorf("marshal endpoints: %w", err)
		}
		return string(data), nil
	})
	if err != nil {
		return nil, err
	}
	var eps []*entity.Endpoint
	if err := json.Unmarshal([]byte(val), &eps); err != nil {
		return nil, fmt.Errorf("unmarshal endpoints: %w", err)
	}
	return eps, nil
}

func (cs *CachedEndpointService) ToggleActive(ctx context.Context, id uuid.UUID, active bool) error {
	if err := cs.inner.ToggleActive(ctx, id, active); err != nil {
		return err
	}
	cs.cache.Delete(ctx, cs.key(id))
	return nil
}

func (cs *CachedEndpointService) key(id uuid.UUID) string {
	return fmt.Sprintf("endpoint:%s", id)
}

func (cs *CachedEndpointService) listKey(collectionID uuid.UUID) string {
	return fmt.Sprintf("endpoints:list:%s", collectionID)
}

// ErrNotFound is a sentinel error for cache miss indicating not found.
var ErrNotFound = fmt.Errorf("not found in cache")

// Ensure CachedWorkspaceService implements WorkspaceService
var _ domservice.WorkspaceService = (*CachedWorkspaceService)(nil)

// Ensure CachedUserService implements UserService
var _ domservice.UserService = (*CachedUserService)(nil)

// Ensure CachedEndpointService implements EndpointService
var _ domservice.EndpointService = (*CachedEndpointService)(nil)
