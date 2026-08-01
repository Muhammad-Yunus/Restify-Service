package auth

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

type fakeCache struct {
	mu    sync.Mutex
	items map[string]string
}

func newFakeCache() *fakeCache {
	return &fakeCache{items: make(map[string]string)}
}

func (f *fakeCache) Get(_ context.Context, key string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	v, ok := f.items[key]
	if !ok {
		return "", fmt.Errorf("key %q not found", key)
	}

	return v, nil
}

func (f *fakeCache) Set(_ context.Context, key string, value string, _ time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.items[key] = value

	return nil
}

func (f *fakeCache) Delete(_ context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.items, key)

	return nil
}

func (f *fakeCache) Exists(_ context.Context, key string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	_, ok := f.items[key]

	return ok, nil
}

func (f *fakeCache) Close(_ context.Context) error {
	return nil
}

var _ repository.Cache = (*fakeCache)(nil)

func TestTokenBlacklistAddStoresHashedToken(t *testing.T) {
	c := newFakeCache()
	tb := NewTokenBlacklist(c)

	if err := tb.Add(context.Background(), "tok-123", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	key := tb.key(tb.hashToken("tok-123"))

	c.mu.Lock()
	_, ok := c.items[key]
	c.mu.Unlock()

	if !ok {
		t.Fatal("Add() did not store hashed token in cache")
	}
}

func TestTokenBlacklistIsBlacklistedTrue(t *testing.T) {
	c := newFakeCache()
	tb := NewTokenBlacklist(c)

	if err := tb.Add(context.Background(), "tok-123", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	blacklisted, err := tb.IsBlacklisted(context.Background(), "tok-123")
	if err != nil {
		t.Fatalf("IsBlacklisted() error = %v", err)
	}

	if !blacklisted {
		t.Fatal("IsBlacklisted() = false, want true")
	}
}

func TestTokenBlacklistIsBlacklistedFalse(t *testing.T) {
	tb := NewTokenBlacklist(newFakeCache())

	blacklisted, err := tb.IsBlacklisted(context.Background(), "never-added-token")
	if err != nil {
		t.Fatalf("IsBlacklisted() error = %v", err)
	}

	if blacklisted {
		t.Fatal("IsBlacklisted() = true, want false")
	}
}

func TestTokenBlacklistHashTokenConsistent(t *testing.T) {
	tb := NewTokenBlacklist(newFakeCache())

	h1 := tb.hashToken("some-token")
	h2 := tb.hashToken("some-token")
	h3 := tb.hashToken("other-token")

	if h1 != h2 {
		t.Fatalf("hashToken() inconsistent: %s != %s", h1, h2)
	}

	if h1 == h3 {
		t.Fatal("hashToken() produced identical hashes for different tokens")
	}
}

func TestTokenBlacklistExpiredTokenNotAdded(t *testing.T) {
	c := newFakeCache()
	tb := NewTokenBlacklist(c)

	if err := tb.Add(context.Background(), "expired-token", time.Now().Add(-time.Minute)); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.items) != 0 {
		t.Fatalf("Add() stored %d keys, want 0 for expired token", len(c.items))
	}
}
