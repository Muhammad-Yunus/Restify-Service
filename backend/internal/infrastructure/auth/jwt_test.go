package auth

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"
)

const testSecret = "test-secret-0123456789abcdef0123456789abcdef"

func TestJWTServiceGenerateAccessTokenProducesValidJWT(t *testing.T) {
	s := NewJWTService(testSecret, time.Hour)

	token, err := s.GenerateAccessToken("user-123", "user@example.com", []string{"admin"})
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	if token == "" {
		t.Fatal("GenerateAccessToken() returned empty token")
	}

	claims, err := s.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}

	if claims["sub"] != "user-123" {
		t.Errorf("sub = %v, want user-123", claims["sub"])
	}

	if claims["email"] != "user@example.com" {
		t.Errorf("email = %v, want user@example.com", claims["email"])
	}

	if claims["type"] != "access" {
		t.Errorf("type = %v, want access", claims["type"])
	}
}

func TestJWTServiceParseTokenValidatesCorrectToken(t *testing.T) {
	s := NewJWTService(testSecret, time.Hour)

	token, err := s.GenerateAccessToken("user-123", "user@example.com", nil)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	if _, err := s.ParseToken(token); err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
}

func TestJWTServiceParseTokenRejectsExpiredToken(t *testing.T) {
	s := NewJWTService(testSecret, -time.Minute)

	token, err := s.GenerateAccessToken("user-123", "user@example.com", nil)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	if _, err := s.ParseToken(token); err == nil {
		t.Fatal("ParseToken() expected error for expired token")
	}
}

func TestJWTServiceParseTokenRejectsTamperedToken(t *testing.T) {
	s := NewJWTService(testSecret, time.Hour)

	token, err := s.GenerateAccessToken("user-123", "user@example.com", nil)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("token has %d parts, want 3", len(parts))
	}

	rawPayload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}

	rawPayload[len(rawPayload)-1] ^= 0x01
	parts[1] = base64.RawURLEncoding.EncodeToString(rawPayload)
	tampered := strings.Join(parts, ".")

	if _, err := s.ParseToken(tampered); err == nil {
		t.Fatal("ParseToken() expected error for tampered token")
	}
}

func TestJWTServiceParseTokenRejectsForeignSignedToken(t *testing.T) {
	other := NewJWTService("other-secret-0123456789abcdef0123456789abcde", time.Hour)

	foreign, err := other.GenerateAccessToken("user-456", "other@example.com", nil)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	s := NewJWTService(testSecret, time.Hour)

	if _, err := s.ParseToken(foreign); err == nil {
		t.Fatal("ParseToken() expected error for token signed with different secret")
	}
}

func TestJWTServiceExtractUserIDReturnsCorrectUserID(t *testing.T) {
	s := NewJWTService(testSecret, time.Hour)

	token, err := s.GenerateAccessToken("user-123", "user@example.com", nil)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	userID, err := s.ExtractUserID(token)
	if err != nil {
		t.Fatalf("ExtractUserID() error = %v", err)
	}

	if userID != "user-123" {
		t.Errorf("ExtractUserID() = %q, want user-123", userID)
	}
}

func TestJWTServiceExtractRolesReturnsCorrectRoles(t *testing.T) {
	s := NewJWTService(testSecret, time.Hour)

	token, err := s.GenerateAccessToken("user-123", "user@example.com", []string{"admin", "viewer"})
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	roles, err := s.ExtractRoles(token)
	if err != nil {
		t.Fatalf("ExtractRoles() error = %v", err)
	}

	want := []string{"admin", "viewer"}
	if len(roles) != len(want) {
		t.Fatalf("ExtractRoles() = %v, want %v", roles, want)
	}

	for i := range want {
		if roles[i] != want[i] {
			t.Errorf("role[%d] = %q, want %q", i, roles[i], want[i])
		}
	}
}

func TestJWTServiceExtractRolesEmptyWhenNoRoles(t *testing.T) {
	s := NewJWTService(testSecret, time.Hour)

	token, err := s.GenerateRefreshToken("user-123")
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	roles, err := s.ExtractRoles(token)
	if err != nil {
		t.Fatalf("ExtractRoles() error = %v", err)
	}

	if len(roles) != 0 {
		t.Errorf("ExtractRoles() = %v, want empty", roles)
	}
}

func TestJWTServiceShortSecretGeneratesRandomKey(t *testing.T) {
	s1 := NewJWTService("short", time.Hour)
	s2 := NewJWTService("short", time.Hour)

	tok1, err := s1.GenerateAccessToken("user-123", "user@example.com", nil)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	tok2, err := s2.GenerateAccessToken("user-123", "user@example.com", nil)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	if tok1 == tok2 {
		t.Fatal("expected different tokens for independent random secrets")
	}
}

func TestJWTServiceRefreshTokenHasSevenDayExpiration(t *testing.T) {
	s := NewJWTService(testSecret, time.Hour)

	token, err := s.GenerateRefreshToken("user-123")
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	claims, err := s.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}

	iat := time.Unix(int64(claims["iat"].(float64)), 0)
	exp := time.Unix(int64(claims["exp"].(float64)), 0)

	if diff := exp.Sub(iat); diff != 7*24*time.Hour {
		t.Errorf("refresh expiry = %v, want 168h", diff)
	}
}
