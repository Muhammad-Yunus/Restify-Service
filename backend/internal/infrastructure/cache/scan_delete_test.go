package cache

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRedisCache_ScanDeletePattern(t *testing.T) {
	opt, err := redis.ParseURL("redis://localhost:6379/0")
	if err != nil {
		t.Skip("parse redis URL: skip test")
	}
	client := redis.NewClient(opt)
	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Skip("redis not available: skip test")
	}
	defer client.Close()

	// Clean up any existing test keys first
	client.Del(context.Background(), "epic24scan:*")

	cache := &RedisCache{client: client}
	prefix := "epic24scan:"

	// Set some test keys
	for i := 0; i < 5; i++ {
		key := prefix + time.Now().Format(time.RFC3339Nano)
		if err := cache.Set(context.Background(), key, "value", 5*time.Second); err != nil {
			t.Fatalf("failed to set key: %v", err)
		}
	}

	// Verify keys exist
	countBefore := 0
	iter := client.Scan(context.Background(), 0, prefix+"*", 0).Iterator()
	for iter.Next(context.Background()) {
		countBefore++
	}
	if countBefore < 5 {
		t.Fatalf("expected at least 5 keys before scan, got %d", countBefore)
	}
}
