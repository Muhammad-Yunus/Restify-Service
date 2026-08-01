package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// PasswordService handles password hashing and verification.
type PasswordService struct {
	cost int
}

// NewPasswordService creates a new password service.
func NewPasswordService(cost int) *PasswordService {
	// Validate cost
	if cost < 10 {
		cost = 12 // default minimum
	}

	return &PasswordService{cost: cost}
}

// Hash generates a bcrypt hash from plaintext password.
func (p *PasswordService) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), p.cost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	return string(bytes), nil
}

// Verify checks if plaintext password matches hash.
func (p *PasswordService) Verify(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// NeedsRehash checks if the hash cost needs to be upgraded.
func (p *PasswordService) NeedsRehash(hash string) bool {
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		return false
	}

	return cost < p.cost
}
