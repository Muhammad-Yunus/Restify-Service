package baas

import (
	"net/http"
	"testing"
)

func TestExtractAuth(t *testing.T) {
	hm, err := NewHeaderMapper("Authorization", nil)
	if err != nil {
		t.Fatalf("NewHeaderMapper: %v", err)
	}

	r := &http.Request{Header: http.Header{"Authorization": []string{"Bearer test-token"}}}
	auth := hm.ExtractAuth(r)
	if auth != "Bearer test-token" {
		t.Errorf("expected 'Bearer test-token', got %q", auth)
	}
}

func TestExtractAuth_DefaultHeader(t *testing.T) {
	hm, _ := NewHeaderMapper("", nil)
	r := &http.Request{Header: http.Header{"Authorization": []string{"Basic abc"}}}
	if hm.ExtractAuth(r) != "Basic abc" {
		t.Error("default auth header should be 'Authorization'")
	}
}

func TestExtractParams(t *testing.T) {
	paramHeaders := []byte(`{"X-Request-Id": "request_id", "X-Custom-Header": "custom_val"}`)
	hm, err := NewHeaderMapper("Authorization", paramHeaders)
	if err != nil {
		t.Fatalf("NewHeaderMapper: %v", err)
	}

	r := &http.Request{Header: http.Header{
		"X-Request-Id": []string{"req-123"},
		"X-Custom-Header": []string{"custom-456"},
	}}
	params := hm.ExtractParams(r)

	if params["request_id"] != "req-123" {
		t.Errorf("expected 'req-123', got %q", params["request_id"])
	}
	if params["custom_val"] != "custom-456" {
		t.Errorf("expected 'custom-456', got %q", params["custom_val"])
	}
}

func TestExtractParams_EmptyMapping(t *testing.T) {
	hm, _ := NewHeaderMapper("Authorization", []byte{})
	r := &http.Request{Header: http.Header{"X-Test": []string{"value"}}}
	params := hm.ExtractParams(r)
	if params != nil {
		t.Error("expected nil params when no mapping is defined")
	}
}

func TestExtractAll(t *testing.T) {
	paramHeaders := []byte(`{"X-Tenant": "tenant_id"}`)
	hm, _ := NewHeaderMapper("Authorization", paramHeaders)

	r := &http.Request{Header: http.Header{
		"Authorization": []string{"Bearer token123"},
		 "X-Tenant": []string{"tenant-456"},
	}}

	auth, params := hm.ExtractAll(r)

	if auth != "Bearer token123" {
		t.Errorf("expected 'Bearer token123', got %q", auth)
	}
	if params["tenant_id"] != "tenant-456" {
		t.Errorf("expected 'tenant-456', got %q", params["tenant_id"])
	}
}

func TestNewHeaderMapper_InvalidJSON(t *testing.T) {
	_, err := NewHeaderMapper("Authorization", []byte("{invalid"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
