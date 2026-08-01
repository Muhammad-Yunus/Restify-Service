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

// DistributedLockHandle represents an acquired lock.
type DistributedLockHandle struct {
	dl  *DistributedLock
	key string
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

// Release releases the lock held by this handle.
func (h *DistributedLockHandle) Release(ctx context.Context) error {
	return h.dl.Release(ctx, h.key)
}
