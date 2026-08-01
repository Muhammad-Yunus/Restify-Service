package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// TokenBlacklist implements the repository.TokenBlacklistRepository interface.
type TokenBlacklist struct {
	cache  repository.Cache
	prefix string
}

// NewTokenBlacklist creates a new token blacklist service.
func NewTokenBlacklist(cache repository.Cache) *TokenBlacklist {
	return &TokenBlacklist{
		cache:  cache,
		prefix: "jwt:blacklist:",
	}
}

func (tb *TokenBlacklist) hashToken(token string) string {
	h := sha256.Sum256([]byte(token))

	return hex.EncodeToString(h[:])
}

func (tb *TokenBlacklist) key(tokenHash string) string {
	return tb.prefix + tokenHash
}

func (tb *TokenBlacklist) Add(ctx context.Context, token string, expiresAt time.Time) error {
	tokenHash := tb.hashToken(token)

	ttl := time.Until(expiresAt)

	if ttl <= 0 {
		return nil // token already expired
	}

	if err := tb.cache.Set(ctx, tb.key(tokenHash), "1", ttl); err != nil {
		return fmt.Errorf("add to blacklist: %w", err)
	}

	return nil
}

func (tb *TokenBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	exists, err := tb.cache.Exists(ctx, tb.key(tb.hashToken(token)))
	if err != nil {
		return false, fmt.Errorf("check blacklist: %w", err)
	}

	return exists, nil
}

func (tb *TokenBlacklist) Cleanup(ctx context.Context) (int64, error) {
	// Redis automatically expires keys; this is a no-op in practice.
	// Kept for interface compliance and future manual cleanup if needed.
	return 0, nil
}

// Compile-time check.
var _ repository.TokenBlacklistRepository = (*TokenBlacklist)(nil)
