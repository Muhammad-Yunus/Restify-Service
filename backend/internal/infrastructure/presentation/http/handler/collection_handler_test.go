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

// fakeCollectionService is a stub CollectionService for handler tests.
type fakeCollectionService struct {
	collections []*entity.Collection
}

func (f *fakeCollectionService) Create(_ context.Context, name, description string, workspaceID uuid.UUID) (*entity.Collection, error) {
	col := &entity.Collection{
		ID:          uuid.New(),
		Name:        name,
		Description: &description,
		WorkspaceID: workspaceID,
		Slug:        name + "-slug",
	}
	f.collections = append(f.collections, col)
	return col, nil
}

func (f *fakeCollectionService) GetByID(_ context.Context, id uuid.UUID) (*entity.Collection, error) {
	for _, col := range f.collections {
		if col.ID == id {
			return col, nil
		}
	}
	return nil, errors.New("not found")
}

func (f *fakeCollectionService) Update(_ context.Context, id uuid.UUID, updates map[string]any) (*entity.Collection, error) {
	for _, col := range f.collections {
		if col.ID == id {
			if name, ok := updates["name"].(string); ok && name != "" {
				col.Name = name
			}
			if desc, ok := updates["description"].(string); ok {
				col.Description = &desc
			}
			return col, nil
		}
	}
	return nil, errors.New("not found")
}

func (f *fakeCollectionService) Delete(_ context.Context, id uuid.UUID) error {
	for i, col := range f.collections {
		if col.ID == id {
			f.collections = append(f.collections[:i], f.collections[i+1:]...)
			return nil
		}
	}
	return errors.New("not found")
}

func (f *fakeCollectionService) List(_ context.Context, workspaceID uuid.UUID) ([]*entity.Collection, error) {
	var result []*entity.Collection
	for _, col := range f.collections {
		if col.WorkspaceID == workspaceID {
			result = append(result, col)
		}
	}
	return result, nil
}

var _ service.CollectionService = (*fakeCollectionService)(nil)

func newTestCollectionHandler(cols []*entity.Collection) *CollectionHandler {
	gin.SetMode(gin.TestMode)
	fs := &fakeCollectionService{collections: cols}
	return NewCollectionHandler(fs)
}

func newTestCollection() *entity.Collection {
	desc := "Test collection"
	return &entity.Collection{
		ID:          uuid.New(),
		Name:        "Test Collection",
		Description: &desc,
		Slug:        "test-collection",
		WorkspaceID: uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func performCollectionRequest(r http.Handler, method, path string, body any) *httptest.ResponseRecorder {
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

func TestCreateCollectionSuccessfully(t *testing.T) {
	h := newTestCollectionHandler(nil)
	r := gin.New()
	r.POST("/workspaces/:ws_id/collections", func(c *gin.Context) { h.Create(c) })

	reqBody := map[string]any{"name": "My Collection", "description": "A test collection"}
	rec := performCollectionRequest(r, http.MethodPost, "/workspaces/"+uuid.New().String()+"/collections", reqBody)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
}

func TestCreateCollectionReturns400OnMissingName(t *testing.T) {
	h := newTestCollectionHandler(nil)
	r := gin.New()
	r.POST("/workspaces/:ws_id/collections", func(c *gin.Context) { h.Create(c) })

	rec := performCollectionRequest(r, http.MethodPost, "/workspaces/"+uuid.New().String()+"/collections", map[string]any{"description": "No name"})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestGetByIDReturnsCollection(t *testing.T) {
	col := newTestCollection()
	h := newTestCollectionHandler([]*entity.Collection{col})
	r := gin.New()
	r.GET("/collections/:id", func(c *gin.Context) { h.GetByID(c) })

	rec := performCollectionRequest(r, http.MethodGet, "/collections/"+col.ID.String(), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestCollectionGetByIDReturns404WhenNotFound(t *testing.T) {
	h := newTestCollectionHandler(nil)
	r := gin.New()
	r.GET("/collections/:id", func(c *gin.Context) { h.GetByID(c) })

	rec := performCollectionRequest(r, http.MethodGet, "/collections/"+uuid.New().String(), nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestListReturnsCollections(t *testing.T) {
	wsID := uuid.New()
	col1 := newTestCollection()
	col1.WorkspaceID = wsID
	col2 := newTestCollection()
	col2.WorkspaceID = wsID
	h := newTestCollectionHandler([]*entity.Collection{col1, col2})
	r := gin.New()
	r.GET("/workspaces/:ws_id/collections", func(c *gin.Context) { h.List(c) })

	rec := performCollectionRequest(r, http.MethodGet, "/workspaces/"+wsID.String()+"/collections", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestUpdateUpdatesCollection(t *testing.T) {
	col := newTestCollection()
	h := newTestCollectionHandler([]*entity.Collection{col})
	r := gin.New()
	r.PATCH("/collections/:id", func(c *gin.Context) { h.Update(c) })

	updates := map[string]any{"name": "Updated Collection"}
	rec := performCollectionRequest(r, http.MethodPatch, "/collections/"+col.ID.String(), updates)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestCollectionUpdateReturns404WhenNotFound(t *testing.T) {
	h := newTestCollectionHandler(nil)
	r := gin.New()
	r.PATCH("/collections/:id", func(c *gin.Context) { h.Update(c) })

	updates := map[string]any{"name": "Updated"}
	rec := performCollectionRequest(r, http.MethodPatch, "/collections/"+uuid.New().String(), updates)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestCollectionDeleteSuccessfully(t *testing.T) {
	col := newTestCollection()
	h := newTestCollectionHandler([]*entity.Collection{col})
	r := gin.New()
	r.DELETE("/collections/:id", func(c *gin.Context) { h.Delete(c) })

	rec := performCollectionRequest(r, http.MethodDelete, "/collections/"+col.ID.String(), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestCollectionDeleteReturns404WhenNotFound(t *testing.T) {
	h := newTestCollectionHandler(nil)
	r := gin.New()
	r.DELETE("/collections/:id", func(c *gin.Context) { h.Delete(c) })

	rec := performCollectionRequest(r, http.MethodDelete, "/collections/"+uuid.New().String(), nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
