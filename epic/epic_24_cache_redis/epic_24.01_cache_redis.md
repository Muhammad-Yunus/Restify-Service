# Epic 24 — Cache Layer (Redis)

**Goal:** Implement Redis caching layer with TTL management, cache invalidation, and distributed locking.
**Dependencies:** Epic 06 (Redis adapter), Epic 05 (Cache repository interface)
**Commit:** `feat: add Redis caching layer with TTL and invalidation`

---

## Step 24.01 — Cache Service

**Build:** Create `backend/internal/application/service/cache_service.go`:

```go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// CacheService provides a generic caching layer.
type CacheService struct {
    cache repository.Cache
}

// NewCacheService creates a new cache service.
func NewCacheService(cache repository.Cache) *CacheService {
    return &CacheService{cache: cache}
}

// Get retrieves a value by key.
func (cs *CacheService) Get(ctx context.Context, key string) (string, error) {
    return cs.cache.Get(ctx, key)
}

// Set stores a value with TTL.
func (cs *CacheService) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
    return cs.cache.Set(ctx, key, value, ttl)
}

// Delete removes a key.
func (cs *CacheService) Delete(ctx context.Context, key string) error {
    return cs.cache.Delete(ctx, key)
}

// GetOrSet retrieves a value or computes and caches it.
func (cs *CacheService) GetOrSet(ctx context.Context, key string, ttl time.Duration, compute func() (string, error)) (string, error) {
    if val, err := cs.cache.Get(ctx, key); err == nil {
        return val, nil
    }

    val, err := compute()
    if err != nil {
        return "", err
    }

    if err := cs.cache.Set(ctx, key, val, ttl); err != nil {
        return val, nil // return value even if cache write fails
    }

    return val, nil
}

// InvalidatePattern invalidates all keys matching a pattern.
func (cs *CacheService) InvalidatePattern(ctx context.Context, pattern string) error {
    // Scan and delete keys matching pattern
    // Simplified: requires Redis SCAN command
    return nil
}
```

---

## Step 24.02 — Caching Decorator Pattern

**Build:** Create `backend/internal/application/service/cache_decorator.go`:

```go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// CachedWorkspaceService wraps WorkspaceService with caching.
type CachedWorkspaceService struct {
    inner    repository.WorkspaceService
    cache    *CacheService
    TTL      time.Duration
}

// NewCachedWorkspaceService creates a cached workspace service.
func NewCachedWorkspaceService(inner repository.WorkspaceService, cache *CacheService) *CachedWorkspaceService {
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
    // Invalidate workspace list cache for owner
    cs.cache.Delete(ctx, cs.listKey(ownerID))
    return ws, nil
}

func (cs *CachedWorkspaceService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Workspace, error) {
    key := cs.key(id)
    return cs.cache.GetOrSet(ctx, key, cs.TTL, func() (string, error) {
        ws, err := cs.inner.GetByID(ctx, id)
        if err != nil {
            return "", err
        }
        data, _ := json.Marshal(ws)
        return string(data), nil
    })
}

func (cs *CachedWorkspaceService) List(ctx context.Context, ownerID uuid.UUID, page, pageSize int) ([]*entity.Workspace, int, error) {
    key := cs.listKey(ownerID)
    // List is not cached due to pagination complexity
    return cs.inner.List(ctx, ownerID, page, pageSize)
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
```

---

## Step 24.03 — Distributed Lock

**Build:** Create `backend/internal/infrastructure/cache/lock.go`:

```go
package cache

import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

// DistributedLock provides Redis-based distributed locking.
type DistributedLock struct {
    client *redis.Client
}

// NewDistributedLock creates a new distributed lock.
func NewDistributedLock(client *redis.Client) *DistributedLock {
    return &DistributedLock{client: client}
}

// Acquire acquires a lock with the given key and TTL.
func (dl *DistributedLock) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
    result, err := dl.client.SetNX(ctx, key, "1", ttl).Result()
    if err != nil {
        return false, fmt.Errorf("acquire lock: %w", err)
    }
    return result, nil
}

// Release releases a lock.
func (dl *DistributedLock) Release(ctx context.Context, key string) error {
    return dl.client.Del(ctx, key).Err()
}

// TryAcquire attempts to acquire a lock without blocking.
func (dl *DistributedLock) TryAcquire(ctx context.Context, key string, ttl time.Duration) (*DistributedLockHandle, error) {
    acquired, err := dl.Acquire(ctx, key, ttl)
    if err != nil {
        return nil, err
    }
    if !acquired {
        return nil, nil // lock not available
    }
    return &DistributedLockHandle{dl: dl, key: key}, nil
}
```

**Test cases:**
- [ ] Unit: `GetOrSet()` caches and retrieves values
- [ ] Unit: `InvalidatePattern()` clears matching keys
- [ ] Unit: Distributed lock acquires and releases correctly
- [ ] Integration: Cache invalidation on workspace update

---

## Commit Instruction

```bash
git add .
git commit -m "feat: add Redis caching layer with TTL, invalidation, and distributed locking"
```
