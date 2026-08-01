package cache

import (
	"context"
	"strings"
	"testing"
)

func TestNewRedisCacheInvalidURL(t *testing.T) {
	_, err := NewRedisCache(context.Background(), "not-a-url")
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}

	if !strings.Contains(err.Error(), "parse redis URL") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewRedisCacheEmptyURL(t *testing.T) {
	_, err := NewRedisCache(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}
}
