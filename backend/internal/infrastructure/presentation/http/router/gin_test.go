package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewGinRouter(t *testing.T) {
	if NewGinRouter("test") == nil {
		t.Fatal("NewGinRouter returned nil")
	}
}

func TestHandleRegistersRoute(t *testing.T) {
	r := NewGinRouter("test")

	r.Handle(http.MethodGet, "/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if rec.Body.String() != "pong" {
		t.Errorf("body = %q, want pong", rec.Body.String())
	}
}

func TestHandleUnsupportedMethodPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for unsupported method, got none")
		}
	}()

	r := NewGinRouter("test")
	r.Handle("BREW", "/coffee", func(w http.ResponseWriter, r *http.Request) {})
}

func TestHandleMiddlewareApplies(t *testing.T) {
	r := NewGinRouter("test")

	mw := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Trace", "yes")
			next(w, r)
		}
	}

	r.Handle(http.MethodGet, "/mw", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, mw)

	req := httptest.NewRequest(http.MethodGet, "/mw", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Header().Get("X-Trace") != "yes" {
		t.Errorf("X-Trace header = %q, want yes", rec.Header().Get("X-Trace"))
	}
}

func TestGroupCreatesGroup(t *testing.T) {
	r := NewGinRouter("test")

	group := r.Group("/api", func(next http.HandlerFunc) http.HandlerFunc {
		return next
	})

	if group == nil {
		t.Fatal("Group returned nil")
	}

	r.Handle(http.MethodGet, "/api/v1/resource", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/resource", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestServeHTTPSupportsAllMethods(t *testing.T) {
	r := NewGinRouter("test")

	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions}

	for i, method := range methods {
		path := fmt.Sprintf("/route%d", i)

		r.Handle(method, path, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
	}

	for i, method := range methods {
		path := fmt.Sprintf("/route%d", i)

		req := httptest.NewRequest(method, path, nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("%s status = %d, want %d", method, rec.Code, http.StatusNoContent)
		}
	}
}
