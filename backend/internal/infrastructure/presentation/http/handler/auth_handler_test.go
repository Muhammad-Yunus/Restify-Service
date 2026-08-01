package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/service"
	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/auth"
)

const handlerTestSecret = "handler-test-secret-0123456789abcdef"

type fakeAuthService struct {
	registerUser  *entity.User
	registerErr   error
	loginResult   *service.AuthResult
	loginErr      error
	refreshResult *service.AuthResult
	refreshErr    error
	logoutErr     error
	logoutCalled  bool
}

func (f *fakeAuthService) Register(_ context.Context, _ string, _ string, _ string) (*entity.User, error) {
	return f.registerUser, f.registerErr
}

func (f *fakeAuthService) Login(_ context.Context, _ string, _ string) (*service.AuthResult, error) {
	return f.loginResult, f.loginErr
}

func (f *fakeAuthService) RefreshToken(_ context.Context, _ string) (*service.AuthResult, error) {
	return f.refreshResult, f.refreshErr
}

func (f *fakeAuthService) Logout(_ context.Context, _ string) error {
	f.logoutCalled = true

	return f.logoutErr
}

func (f *fakeAuthService) GetCurrentUser(_ context.Context, _ string) (*entity.User, error) {
	return nil, nil
}

func (f *fakeAuthService) HasPermission(_ context.Context, _ uuid.UUID, _ string) bool {
	return false
}

func newTestUser() *entity.User {
	fullName := "John Doe"

	return &entity.User{
		ID:       uuid.New(),
		Email:    "user@example.com",
		FullName: &fullName,
		Roles:    []*entity.Role{{Name: "user"}},
	}
}

func newAuthRouter(h *AuthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/v1/auth/register", h.Register)
	r.POST("/api/v1/auth/login", h.Login)
	r.POST("/api/v1/auth/refresh", h.Refresh)
	r.POST("/api/v1/auth/logout", func(c *gin.Context) {
		c.Set("token", "tok-123")
	}, h.Logout)

	return r
}

func performJSON(r http.Handler, path string, body any) *httptest.ResponseRecorder {
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	return rec
}

func decodeAuthResponse(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	return body
}

func TestRegisterCreatesUserAndReturnsToken(t *testing.T) {
	fake := &fakeAuthService{registerUser: newTestUser()}
	h := NewAuthHandler(fake, auth.NewJWTService(handlerTestSecret, time.Hour))

	rec := performJSON(newAuthRouter(h), "/api/v1/auth/register", map[string]any{
		"email":     "user@example.com",
		"password":  "S3curePass!",
		"full_name": "John Doe",
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	body := decodeAuthResponse(t, rec)

	if body["access_token"] == "" {
		t.Error("access_token missing in response")
	}

	if body["refresh_token"] == "" {
		t.Error("refresh_token missing in response")
	}
}

func TestRegisterReturnsConflictWhenEmailExists(t *testing.T) {
	fake := &fakeAuthService{registerErr: errors.New("email already exists")}
	h := NewAuthHandler(fake, auth.NewJWTService(handlerTestSecret, time.Hour))

	rec := performJSON(newAuthRouter(h), "/api/v1/auth/register", map[string]any{
		"email":    "user@example.com",
		"password": "S3curePass!",
	})

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestRegisterValidatesInputFields(t *testing.T) {
	fake := &fakeAuthService{}
	h := NewAuthHandler(fake, auth.NewJWTService(handlerTestSecret, time.Hour))

	rec := performJSON(newAuthRouter(h), "/api/v1/auth/register", map[string]any{
		"email":    "not-an-email",
		"password": "short",
	})

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestLoginReturnsTokensForValidCredentials(t *testing.T) {
	fake := &fakeAuthService{
		loginResult: &service.AuthResult{
			User:         newTestUser(),
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			ExpiresIn:    86400,
		},
	}
	h := NewAuthHandler(fake, auth.NewJWTService(handlerTestSecret, time.Hour))

	rec := performJSON(newAuthRouter(h), "/api/v1/auth/login", map[string]any{
		"email":    "user@example.com",
		"password": "S3curePass!",
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := decodeAuthResponse(t, rec)

	if body["access_token"] != "access-token" {
		t.Errorf("access_token = %v, want access-token", body["access_token"])
	}
}

func TestLoginReturnsUnauthorizedForInvalidCredentials(t *testing.T) {
	fake := &fakeAuthService{loginErr: errors.New("invalid credentials")}
	h := NewAuthHandler(fake, auth.NewJWTService(handlerTestSecret, time.Hour))

	rec := performJSON(newAuthRouter(h), "/api/v1/auth/login", map[string]any{
		"email":    "user@example.com",
		"password": "WrongPass!",
	})

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRefreshReturnsNewTokensForValidRefreshToken(t *testing.T) {
	fake := &fakeAuthService{
		refreshResult: &service.AuthResult{
			User:         newTestUser(),
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			ExpiresIn:    86400,
		},
	}
	h := NewAuthHandler(fake, auth.NewJWTService(handlerTestSecret, time.Hour))

	rec := performJSON(newAuthRouter(h), "/api/v1/auth/refresh", map[string]any{
		"refresh_token": "old-refresh-token",
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := decodeAuthResponse(t, rec)

	if body["access_token"] != "new-access-token" {
		t.Errorf("access_token = %v, want new-access-token", body["access_token"])
	}
}

func TestLogoutCallsService(t *testing.T) {
	fake := &fakeAuthService{}
	h := NewAuthHandler(fake, auth.NewJWTService(handlerTestSecret, time.Hour))

	rec := performJSON(newAuthRouter(h), "/api/v1/auth/logout", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if !fake.logoutCalled {
		t.Fatal("Logout() did not call auth service")
	}
}
