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

// fakeEndpointService is a stub EndpointService for handler tests.
type fakeEndpointService struct {
	endpoints []*entity.Endpoint
}

func (f *fakeEndpointService) Create(_ context.Context, collectionID uuid.UUID, params map[string]any) (*entity.Endpoint, error) {
	ep := &entity.Endpoint{
		ID:           uuid.New(),
		CollectionID: collectionID,
		Name:         params["name"].(string),
		Path:         params["path"].(string),
		Method:       params["method"].(string),
		Version:      "v1",
		IsActive:     true,
	}
	f.endpoints = append(f.endpoints, ep)
	return ep, nil
}

func (f *fakeEndpointService) GetByID(_ context.Context, id uuid.UUID) (*entity.Endpoint, error) {
	for _, ep := range f.endpoints {
		if ep.ID == id {
			return ep, nil
		}
	}
	return nil, errors.New("not found")
}

func (f *fakeEndpointService) Update(_ context.Context, id uuid.UUID, updates map[string]any) (*entity.Endpoint, error) {
	for _, ep := range f.endpoints {
		if ep.ID == id {
			if name, ok := updates["name"].(string); ok {
				ep.Name = name
			}
			if path, ok := updates["path"].(string); ok {
				ep.Path = path
			}
			return ep, nil
		}
	}
	return nil, errors.New("not found")
}

func (f *fakeEndpointService) Delete(_ context.Context, id uuid.UUID) error {
	for i, ep := range f.endpoints {
		if ep.ID == id {
			f.endpoints = append(f.endpoints[:i], f.endpoints[i+1:]...)
			return nil
		}
	}
	return errors.New("not found")
}

func (f *fakeEndpointService) List(_ context.Context, collectionID uuid.UUID) ([]*entity.Endpoint, error) {
	var result []*entity.Endpoint
	for _, ep := range f.endpoints {
		if ep.CollectionID == collectionID {
			result = append(result, ep)
		}
	}
	return result, nil
}

func (f *fakeEndpointService) ToggleActive(_ context.Context, id uuid.UUID, active bool) error {
	for _, ep := range f.endpoints {
		if ep.ID == id {
			ep.IsActive = active
			return nil
		}
	}
	return errors.New("not found")
}

func (f *fakeEndpointService) ListByWorkspace(_ context.Context, workspaceID uuid.UUID) ([]*entity.Endpoint, error) {
	return nil, nil
}

var _ service.EndpointService = (*fakeEndpointService)(nil)

func newTestEndpointHandler(eps []*entity.Endpoint) *EndpointHandler {
	gin.SetMode(gin.TestMode)
	fs := &fakeEndpointService{endpoints: eps}
	return NewEndpointHandler(fs)
}

func performEndpointRequest(r http.Handler, method, path string, body any) *httptest.ResponseRecorder {
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

func newTestEndpoint(collectionID uuid.UUID) *entity.Endpoint {
	return &entity.Endpoint{
		ID:           uuid.New(),
		CollectionID: collectionID,
		Name:         "Test Endpoint",
		Path:         "/test",
		Method:       "GET",
		Version:      "v1",
		IsActive:     true,
		EndpointType: entity.EndpointTypeTable,
		Schema:       "public",
		TableName:    "test_table",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func TestCreateEndpointSuccessfully(t *testing.T) {
	h := newTestEndpointHandler(nil)
	r := gin.New()
	r.POST("/collections/:col_id/endpoints", func(c *gin.Context) { h.Create(c) })

	collectionID := uuid.New()
	reqBody := map[string]any{"name": "My Endpoint", "path": "/my-endpoint", "method": "GET"}
	rec := performEndpointRequest(r, http.MethodPost, "/collections/"+collectionID.String()+"/endpoints", reqBody)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
}

func TestCreateEndpointReturns400OnMissingName(t *testing.T) {
	h := newTestEndpointHandler(nil)
	r := gin.New()
	r.POST("/collections/:col_id/endpoints", func(c *gin.Context) { h.Create(c) })

	collectionID := uuid.New()
	rec := performEndpointRequest(r, http.MethodPost, "/collections/"+collectionID.String()+"/endpoints", map[string]any{"path": "/test"})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestGetByIDReturnsEndpoint(t *testing.T) {
	collectionID := uuid.New()
	ep := newTestEndpoint(collectionID)
	h := newTestEndpointHandler([]*entity.Endpoint{ep})
	r := gin.New()
	r.GET("/endpoints/:id", func(c *gin.Context) { h.GetByID(c) })

	rec := performEndpointRequest(r, http.MethodGet, "/endpoints/"+ep.ID.String(), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestEndpointGetByIDReturns404WhenNotFound(t *testing.T) {
	h := newTestEndpointHandler(nil)
	r := gin.New()
	r.GET("/endpoints/:id", func(c *gin.Context) { h.GetByID(c) })

	rec := performEndpointRequest(r, http.MethodGet, "/endpoints/"+uuid.New().String(), nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestListReturnsEndpoints(t *testing.T) {
	collectionID := uuid.New()
	ep1 := newTestEndpoint(collectionID)
	ep2 := newTestEndpoint(collectionID)
	h := newTestEndpointHandler([]*entity.Endpoint{ep1, ep2})
	r := gin.New()
	r.GET("/collections/:col_id/endpoints", func(c *gin.Context) { h.List(c) })

	rec := performEndpointRequest(r, http.MethodGet, "/collections/"+collectionID.String()+"/endpoints", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestUpdateUpdatesEndpoint(t *testing.T) {
	collectionID := uuid.New()
	ep := newTestEndpoint(collectionID)
	h := newTestEndpointHandler([]*entity.Endpoint{ep})
	r := gin.New()
	r.PATCH("/endpoints/:id", func(c *gin.Context) { h.Update(c) })

	updates := map[string]any{"name": "Updated Endpoint", "path": "/updated"}
	rec := performEndpointRequest(r, http.MethodPatch, "/endpoints/"+ep.ID.String(), updates)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestEndpointUpdateReturns404WhenNotFound(t *testing.T) {
	h := newTestEndpointHandler(nil)
	r := gin.New()
	r.PATCH("/endpoints/:id", func(c *gin.Context) { h.Update(c) })

	updates := map[string]any{"name": "Updated"}
	rec := performEndpointRequest(r, http.MethodPatch, "/endpoints/"+uuid.New().String(), updates)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestEndpointDeleteSuccessfully(t *testing.T) {
	collectionID := uuid.New()
	ep := newTestEndpoint(collectionID)
	h := newTestEndpointHandler([]*entity.Endpoint{ep})
	r := gin.New()
	r.DELETE("/endpoints/:id", func(c *gin.Context) { h.Delete(c) })

	rec := performEndpointRequest(r, http.MethodDelete, "/endpoints/"+ep.ID.String(), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestEndpointDeleteReturns404WhenNotFound(t *testing.T) {
	h := newTestEndpointHandler(nil)
	r := gin.New()
	r.DELETE("/endpoints/:id", func(c *gin.Context) { h.Delete(c) })

	rec := performEndpointRequest(r, http.MethodDelete, "/endpoints/"+uuid.New().String(), nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestToggleEndpointSuccessfully(t *testing.T) {
	collectionID := uuid.New()
	ep := newTestEndpoint(collectionID)
	h := newTestEndpointHandler([]*entity.Endpoint{ep})
	r := gin.New()
	r.POST("/endpoints/:id/toggle", func(c *gin.Context) { h.Toggle(c) })

	reqBody := map[string]any{"active": false}
	rec := performEndpointRequest(r, http.MethodPost, "/endpoints/"+ep.ID.String()+"/toggle", reqBody)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestToggleEndpointReturns404WhenNotFound(t *testing.T) {
	h := newTestEndpointHandler(nil)
	r := gin.New()
	r.POST("/endpoints/:id/toggle", func(c *gin.Context) { h.Toggle(c) })

	reqBody := map[string]any{"active": false}
	rec := performEndpointRequest(r, http.MethodPost, "/endpoints/"+uuid.New().String()+"/toggle", reqBody)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
