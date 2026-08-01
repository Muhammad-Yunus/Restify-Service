package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/auth"
)

const middlewareTestSecret = "middleware-test-secret-0123456789abcdef"

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

func newTestAuthMiddleware(t *testing.T, roles []string) (*AuthMiddleware, *fakeCache, string) {
	t.Helper()

	js := auth.NewJWTService(middlewareTestSecret, time.Hour)
	fc := newFakeCache()
	bl := auth.NewTokenBlacklist(fc)

	token, err := js.GenerateAccessToken("user-123", "user@example.com", roles)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	return NewAuthMiddleware(js, bl), fc, token
}

func performRequest(r http.Handler, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	return rec
}

func newEchoRouter(middlewares ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middlewares...)
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id": c.GetString("user_id"),
			"email":   c.GetString("email"),
		})
	})

	return r
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	return body
}

func TestRequireAuthRejectsRequestWithoutAuthorizationHeader(t *testing.T) {
	mw, _, _ := newTestAuthMiddleware(t, nil)

	rec := performRequest(newEchoRouter(mw.RequireAuth()), "")

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuthRejectsInvalidTokenFormat(t *testing.T) {
	mw, _, _ := newTestAuthMiddleware(t, nil)

	rec := performRequest(newEchoRouter(mw.RequireAuth()), "Basic dXNlcjpwYXNz")

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuthAcceptsValidTokenAndSetsUserContext(t *testing.T) {
	mw, _, token := newTestAuthMiddleware(t, nil)

	rec := performRequest(newEchoRouter(mw.RequireAuth()), token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := decodeBody(t, rec)
	if body["user_id"] != "user-123" {
		t.Errorf("user_id = %v, want user-123", body["user_id"])
	}

	if body["email"] != "user@example.com" {
		t.Errorf("email = %v, want user@example.com", body["email"])
	}
}

func TestRequireAuthRejectsBlacklistedToken(t *testing.T) {
	mw, fc, token := newTestAuthMiddleware(t, nil)

	bl := auth.NewTokenBlacklist(fc)
	if err := bl.Add(context.Background(), token, time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("seed blacklist: %v", err)
	}

	rec := performRequest(newEchoRouter(mw.RequireAuth()), token)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireRoleAllowsRequestWithMatchingRole(t *testing.T) {
	mw, _, token := newTestAuthMiddleware(t, []string{"admin"})

	rec := performRequest(newEchoRouter(mw.RequireAuth(), mw.RequireRole("admin")), token)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRequireRoleRejectsRequestWithoutMatchingRole(t *testing.T) {
	mw, _, token := newTestAuthMiddleware(t, []string{"viewer"})

	rec := performRequest(newEchoRouter(mw.RequireAuth(), mw.RequireRole("admin")), token)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestOptionalAuthAllowsRequestWithoutToken(t *testing.T) {
	mw, _, _ := newTestAuthMiddleware(t, nil)

	rec := performRequest(newEchoRouter(mw.OptionalAuth()), "")

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestOptionalAuthSetsContextWhenValidTokenPresent(t *testing.T) {
	mw, _, token := newTestAuthMiddleware(t, nil)

	rec := performRequest(newEchoRouter(mw.OptionalAuth()), token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := decodeBody(t, rec)
	if body["user_id"] != "user-123" {
		t.Errorf("user_id = %v, want user-123", body["user_id"])
	}
}
