package cache

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestDistributedLock_AcquireAndRelease(t *testing.T) {
	// This test requires a running Redis instance.
	// Skip if Redis is not available.
	opt, err := redis.ParseURL("redis://localhost:6379/0")
	if err != nil {
		t.Skip("parse redis URL: skip test")
	}
	client := redis.NewClient(opt)
	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Skip("redis not available: skip test")
	}
	defer client.Close()

	dl := NewDistributedLock(client)
	key := "test-lock:" + time.Now().Format(time.RFC3339)

	// Acquire the lock
	h, err := dl.TryAcquire(context.Background(), key, 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h == nil {
		t.Fatal("expected to acquire lock")
	}

	// Try to acquire again (should fail)
	h2, err := dl.TryAcquire(context.Background(), key, 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h2 != nil {
		t.Fatal("expected second acquire to fail")
	}

	// Release the first lock
	if err := h.Release(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be able to acquire now
	h3, err := dl.TryAcquire(context.Background(), key, 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h3 == nil {
		t.Fatal("expected to acquire lock after release")
	}
	if err := h3.Release(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDistributedLock_TTLExpires(t *testing.T) {
	opt, err := redis.ParseURL("redis://localhost:6379/0")
	if err != nil {
		t.Skip("parse redis URL: skip test")
	}
	client := redis.NewClient(opt)
	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Skip("redis not available: skip test")
	}
	defer client.Close()

	dl := NewDistributedLock(client)
	key := "test-lock-ttl:" + time.Now().Format(time.RFC3339)

	h, err := dl.TryAcquire(context.Background(), key, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h == nil {
		t.Fatal("expected to acquire lock")
	}

	// Wait for TTL to expire
	time.Sleep(200 * time.Millisecond)

	// Should be able to acquire again after TTL
	h2, err := dl.TryAcquire(context.Background(), key, 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h2 == nil {
		t.Fatal("expected to acquire lock after TTL expiry")
	}
	if err := h2.Release(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
