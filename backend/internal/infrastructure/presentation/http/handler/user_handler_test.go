package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/service"
)

// fakeUserService is a stub UserService for handler tests.
type fakeUserService struct {
	users   []*entity.User
	listErr error
}

func (f *fakeUserService) GetByID(_ context.Context, id uuid.UUID) (*entity.User, error) {
	for _, u := range f.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, errors.New("not found")
}

func (f *fakeUserService) GetByEmail(_ context.Context, email string) (*entity.User, error) {
	for _, u := range f.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("not found")
}

func (f *fakeUserService) Update(_ context.Context, id uuid.UUID, updates map[string]any) (*entity.User, error) {
	for _, u := range f.users {
		if u.ID == id {
			if name, ok := updates["full_name"].(string); ok {
				u.FullName = &name
			}
			if active, ok := updates["is_active"].(bool); ok {
				u.IsActive = active
			}
			return u, nil
		}
	}
	return nil, errors.New("not found")
}

func (f *fakeUserService) Delete(_ context.Context, id uuid.UUID) error {
	for i, u := range f.users {
		if u.ID == id {
			f.users = append(f.users[:i], f.users[i+1:]...)
			return nil
		}
	}
	return errors.New("not found")
}

func (f *fakeUserService) List(_ context.Context, page, pageSize int) ([]*entity.User, int, error) {
	if f.listErr != nil {
		return nil, 0, f.listErr
	}
	return f.users, len(f.users), nil
}

var _ service.UserService = (*fakeUserService)(nil)

func newUserHandler(users []*entity.User, listErr error) *UserHandler {
	gin.SetMode(gin.TestMode)
	fs := &fakeUserService{users: users, listErr: listErr}
	return NewUserHandler(fs)
}

func newTestUserForHandler() *entity.User {
	fullName := "John Doe"
	return &entity.User{
		ID:        uuid.New(),
		Email:     "john@example.com",
		FullName:  &fullName,
		IsActive:  true,
		Roles:     []*entity.Role{{Name: "user"}},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestGetByIDReturnsUser(t *testing.T) {
	user := newTestUserForHandler()
	h := newUserHandler([]*entity.User{user}, nil)
	r := gin.New()
	r.GET("/users/:id", func(c *gin.Context) { h.GetByID(c) })

	req := httptest.NewRequest(http.MethodGet, "/users/"+user.ID.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestGetByIDReturns404WhenNotFound(t *testing.T) {
	h := newUserHandler(nil, nil)
	r := gin.New()
	r.GET("/users/:id", func(c *gin.Context) { h.GetByID(c) })

	req := httptest.NewRequest(http.MethodGet, "/users/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetByIDReturns400WhenInvalidUUID(t *testing.T) {
	h := newUserHandler(nil, nil)
	r := gin.New()
	r.GET("/users/:id", func(c *gin.Context) { h.GetByID(c) })

	req := httptest.NewRequest(http.MethodGet, "/users/not-a-uuid", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestListReturnsPaginatedUsers(t *testing.T) {
	user1 := newTestUserForHandler()
	user2 := newTestUserForHandler()
	h := newUserHandler([]*entity.User{user1, user2}, nil)
	r := gin.New()
	r.GET("/users", func(c *gin.Context) { h.List(c) })

	req := httptest.NewRequest(http.MethodGet, "/users?page=1&page_size=10", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestDeleteSoftDeletesUser(t *testing.T) {
	user := newTestUserForHandler()
	h := newUserHandler([]*entity.User{user}, nil)
	r := gin.New()
	r.DELETE("/users/:id", func(c *gin.Context) { h.Delete(c) })

	req := httptest.NewRequest(http.MethodDelete, "/users/"+user.ID.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
