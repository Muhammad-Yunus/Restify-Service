package repository

import (
	"context"
	"time"
)

// TokenBlacklistRepository manages revoked JWT tokens.
type TokenBlacklistRepository interface {
	// Add adds a token to the blacklist.
	Add(ctx context.Context, tokenHash string, expiresAt time.Time) error

	// IsBlacklisted checks if a token is blacklisted.
	IsBlacklisted(ctx context.Context, tokenHash string) (bool, error)

	// Cleanup removes expired blacklisted tokens.
	Cleanup(ctx context.Context) (int64, error)
}
