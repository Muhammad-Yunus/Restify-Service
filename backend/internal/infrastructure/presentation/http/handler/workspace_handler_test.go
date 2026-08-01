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
)

// fakeWorkspaceService is a stub WorkspaceService for handler tests.
type fakeWorkspaceService struct {
	workspaces []*entity.Workspace
	listErr    error
}

func (f *fakeWorkspaceService) Create(_ context.Context, name, description string, ownerID uuid.UUID) (*entity.Workspace, error) {
	ws := &entity.Workspace{
		ID:          uuid.New(),
		Name:        name,
		Description: &description,
		OwnerID:     ownerID,
		Slug:        name + "-slug",
		IsPublic:    false,
	}
	f.workspaces = append(f.workspaces, ws)
	return ws, nil
}

func (f *fakeWorkspaceService) GetByID(_ context.Context, id uuid.UUID) (*entity.Workspace, error) {
	for _, ws := range f.workspaces {
		if ws.ID == id {
			return ws, nil
		}
	}
	return nil, errors.New("not found")
}

func (f *fakeWorkspaceService) Update(_ context.Context, id uuid.UUID, updates map[string]any) (*entity.Workspace, error) {
	for _, ws := range f.workspaces {
		if ws.ID == id {
			if name, ok := updates["name"].(string); ok && name != "" {
				ws.Name = name
			}
			if desc, ok := updates["description"].(string); ok {
				ws.Description = &desc
			}
			if pub, ok := updates["is_public"].(bool); ok {
				ws.IsPublic = pub
			}
			return ws, nil
		}
	}
	return nil, errors.New("not found")
}

func (f *fakeWorkspaceService) Delete(_ context.Context, id uuid.UUID) error {
	for i, ws := range f.workspaces {
		if ws.ID == id {
			f.workspaces = append(f.workspaces[:i], f.workspaces[i+1:]...)
			return nil
		}
	}
	return errors.New("not found")
}

func (f *fakeWorkspaceService) List(_ context.Context, ownerID uuid.UUID, page, pageSize int) ([]*entity.Workspace, int, error) {
	if f.listErr != nil {
		return nil, 0, f.listErr
	}
	return f.workspaces, len(f.workspaces), nil
}

var _ service.WorkspaceService = (*fakeWorkspaceService)(nil)

func newTestWorkspaceHandler(workspaces []*entity.Workspace, listErr error) *WorkspaceHandler {
	gin.SetMode(gin.TestMode)
	fs := &fakeWorkspaceService{workspaces: workspaces, listErr: listErr}
	return NewWorkspaceHandler(fs)
}

func newTestWorkspace() *entity.Workspace {
	desc := "Test workspace"
	return &entity.Workspace{
		ID:          uuid.New(),
		Name:        "Test Workspace",
		Description: &desc,
		Slug:        "test-workspace",
		OwnerID:     uuid.New(),
		IsPublic:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func performWorkspaceRequest(r http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	var payload *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		payload = bytes.NewReader(b)
	} else {
		payload = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, payload)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestCreateWorkspaceSuccessfully(t *testing.T) {
	h := newTestWorkspaceHandler(nil, nil)
	r := gin.New()
	r.POST("/workspaces", func(c *gin.Context) { h.Create(c) })

	reqBody := map[string]any{"name": "My Workspace", "description": "A test workspace"}
	rec := performWorkspaceRequest(r, http.MethodPost, "/workspaces", reqBody)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
}

func TestCreateWorkspaceReturns400OnMissingName(t *testing.T) {
	h := newTestWorkspaceHandler(nil, nil)
	r := gin.New()
	r.POST("/workspaces", func(c *gin.Context) { h.Create(c) })

	rec := performWorkspaceRequest(r, http.MethodPost, "/workspaces", map[string]any{"description": "No name"})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestGetByIDReturnsWorkspace(t *testing.T) {
	ws := newTestWorkspace()
	h := newTestWorkspaceHandler([]*entity.Workspace{ws}, nil)
	r := gin.New()
	r.GET("/workspaces/:id", func(c *gin.Context) { h.GetByID(c) })

	rec := performWorkspaceRequest(r, http.MethodGet, "/workspaces/"+ws.ID.String(), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestWorkspaceGetByIDReturns404WhenNotFound(t *testing.T) {
	h := newTestWorkspaceHandler(nil, nil)
	r := gin.New()
	r.GET("/workspaces/:id", func(c *gin.Context) { h.GetByID(c) })

	rec := performWorkspaceRequest(r, http.MethodGet, "/workspaces/"+uuid.New().String(), nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestWorkspaceGetByIDReturns400WhenInvalidUUID(t *testing.T) {
	h := newTestWorkspaceHandler(nil, nil)
	r := gin.New()
	r.GET("/workspaces/:id", func(c *gin.Context) { h.GetByID(c) })

	rec := performWorkspaceRequest(r, http.MethodGet, "/workspaces/not-a-uuid", nil)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestListReturnsPaginatedWorkspaces(t *testing.T) {
	ws1 := newTestWorkspace()
	ws2 := newTestWorkspace()
	h := newTestWorkspaceHandler([]*entity.Workspace{ws1, ws2}, nil)
	r := gin.New()
	r.GET("/workspaces", func(c *gin.Context) { h.List(c) })

	rec := performWorkspaceRequest(r, http.MethodGet, "/workspaces?page=1&page_size=10", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestUpdateUpdatesWorkspace(t *testing.T) {
	ws := newTestWorkspace()
	h := newTestWorkspaceHandler([]*entity.Workspace{ws}, nil)
	r := gin.New()
	r.PATCH("/workspaces/:id", func(c *gin.Context) { h.Update(c) })

	updates := map[string]any{"name": "Updated Workspace"}
	rec := performWorkspaceRequest(r, http.MethodPatch, "/workspaces/"+ws.ID.String(), updates)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestUpdateReturns404WhenNotFound(t *testing.T) {
	h := newTestWorkspaceHandler(nil, nil)
	r := gin.New()
	r.PATCH("/workspaces/:id", func(c *gin.Context) { h.Update(c) })

	updates := map[string]any{"name": "Updated"}
	rec := performWorkspaceRequest(r, http.MethodPatch, "/workspaces/"+uuid.New().String(), updates)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestDeleteSuccessfully(t *testing.T) {
	ws := newTestWorkspace()
	h := newTestWorkspaceHandler([]*entity.Workspace{ws}, nil)
	r := gin.New()
	r.DELETE("/workspaces/:id", func(c *gin.Context) { h.Delete(c) })

	rec := performWorkspaceRequest(r, http.MethodDelete, "/workspaces/"+ws.ID.String(), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestDeleteReturns404WhenNotFound(t *testing.T) {
	h := newTestWorkspaceHandler(nil, nil)
	r := gin.New()
	r.DELETE("/workspaces/:id", func(c *gin.Context) { h.Delete(c) })

	rec := performWorkspaceRequest(r, http.MethodDelete, "/workspaces/"+uuid.New().String(), nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
