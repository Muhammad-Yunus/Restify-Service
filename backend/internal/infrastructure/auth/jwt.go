package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService handles token generation and validation.
type JWTService struct {
	secret     []byte
	expiration time.Duration
}

// NewJWTService creates a new JWT service.
func NewJWTService(secret string, expiration time.Duration) *JWTService {
	// Validate secret length
	if len(secret) < 32 {
		// Generate a random secret if too short
		secret = randomSecret()
	}

	return &JWTService{
		secret:     []byte(secret),
		expiration: expiration,
	}
}

// randomSecret generates a random 32-byte base64 URL-encoded secret.
func randomSecret() string {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		panic("auth: generate jwt secret: " + err.Error())
	}

	return base64.URLEncoding.EncodeToString(b)
}

// GenerateAccessToken creates a new access token for a user.
func (j *JWTService) GenerateAccessToken(userID string, email string, roles []string) (string, error) {
	now := time.Now()

	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"roles": roles,
		"type":  "access",
		"iat":   now.Unix(),
		"exp":   now.Add(j.expiration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(j.secret)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}

	return signed, nil
}

// GenerateRefreshToken creates a new refresh token with 7-day expiration.
func (j *JWTService) GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()

	claims := jwt.MapClaims{
		"sub":  userID,
		"type": "refresh",
		"iat":  now.Unix(),
		"exp":  now.Add(7 * 24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(j.secret)
	if err != nil {
		return "", fmt.Errorf("sign refresh token: %w", err)
	}

	return signed, nil
}

// ParseToken validates and parses a JWT token.
func (j *JWTService) ParseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return j.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// ExtractUserID extracts the user ID from a token.
func (j *JWTService) ExtractUserID(tokenString string) (string, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("user ID not found in token")
	}

	return sub, nil
}

// ExtractRoles extracts roles from a token.
func (j *JWTService) ExtractRoles(tokenString string) ([]string, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	rolesRaw, ok := claims["roles"]
	if !ok {
		return []string{}, nil
	}

	roles, ok := rolesRaw.([]any)
	if !ok {
		return nil, fmt.Errorf("roles has invalid type")
	}

	result := make([]string, len(roles))

	for i, r := range roles {
		result[i], ok = r.(string)
		if !ok {
			return nil, fmt.Errorf("role has invalid type")
		}
	}

	return result, nil
}
