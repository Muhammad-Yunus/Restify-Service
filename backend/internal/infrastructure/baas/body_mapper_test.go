package baas

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestMapBody_NoMapping(t *testing.T) {
	bm, err := NewBodyMapper(nil)
	if err != nil {
		t.Fatalf("NewBodyMapper: %v", err)
	}

	body := `{"name": "John", "age": 30}`
	r := &http.Request{
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}

	result, err := bm.MapBody(r)
	if err != nil {
		t.Fatalf("MapBody: %v", err)
	}

	if result["name"] != "John" {
		t.Errorf("expected 'John', got %v", result["name"])
	}
	if result["age"] != float64(30) {
		t.Errorf("expected 30, got %v", result["age"])
	}
}

func TestMapBody_WithTransform(t *testing.T) {
	mapping := []byte(`[
		{"source_field": "email", "target_param": "mail", "transform": "lowercase"},
		{"source_field": "name", "target_param": "full_name", "transform": "trim"}
	]`)

	bm, err := NewBodyMapper(mapping)
	if err != nil {
		t.Fatalf("NewBodyMapper: %v", err)
	}

	body := `{"email": "JOHN@EXAMPLE.COM", "name": "  Jane Doe  "}`
	r := &http.Request{
		Body:   io.NopCloser(strings.NewReader(body)),
	}

	result, err := bm.MapBody(r)
	if err != nil {
		t.Fatalf("MapBody: %v", err)
	}

	if result["mail"] != "john@example.com" {
		t.Errorf("expected 'john@example.com', got %v", result["mail"])
	}
	if result["full_name"] != "Jane Doe" {
		t.Errorf("expected 'Jane Doe', got %v", result["full_name"])
	}
}

func TestMapBody_UppercaseTransform(t *testing.T) {
	mapping := []byte(`[
		{"source_field": "code", "target_param": "CODE", "transform": "uppercase"}
	]`)

	bm, _ := NewBodyMapper(mapping)
	body := `{"code": "abcdef"}`
	r := &http.Request{Body: io.NopCloser(strings.NewReader(body))}

	result, _ := bm.MapBody(r)
	if result["CODE"] != "ABCDEF" {
		t.Errorf("expected 'ABCDEF', got %v", result["CODE"])
	}
}

func TestMapBody_UnknownFieldPassthrough(t *testing.T) {
	mapping := []byte(`[
		{"source_field": "name", "target_param": "full_name"}
	]`)

	bm, _ := NewBodyMapper(mapping)
	body := `{"name": "John", "email": "john@example.com"}`
	r := &http.Request{Body: io.NopCloser(strings.NewReader(body))}

	result, _ := bm.MapBody(r)
	if result["full_name"] != "John" {
		t.Errorf("expected mapped 'John', got %v", result["full_name"])
	}
	if result["email"] != "john@example.com" {
		t.Errorf("expected passthrough 'john@example.com', got %v", result["email"])
	}
}

func TestMapBody_InvalidJSON(t *testing.T) {
	bm, _ := NewBodyMapper(nil)
	r := &http.Request{Body: io.NopCloser(strings.NewReader("{\"invalid\"}"))}

	_, err := bm.MapBody(r)
	if err == nil {
		t.Error("expected error for invalid JSON body")
	}
}

func TestMapBody_MappingInvalidJSON(t *testing.T) {
	_, err := NewBodyMapper([]byte("{invalid"))
	if err == nil {
		t.Error("expected error for invalid mapping JSON")
	}
}

func TestMapBody_BodyRestored(t *testing.T) {
	bm, _ := NewBodyMapper(nil)
	body := `{"test": "data"}`
	r := &http.Request{Body: io.NopCloser(strings.NewReader(body))}

	bm.MapBody(r)

	// Restore and read again to verify body wasn't consumed
	restored, _ := io.ReadAll(r.Body)
	if string(restored) != body {
		t.Errorf("expected body to be restored, got %q", string(restored))
	}
}

func TestApplyTransform(t *testing.T) {
	if result := applyTransform("hello", "lowercase"); result != "hello" {
		t.Error("expected lowercase to have no effect on already lowercase string")
	}
	if result := applyTransform("HELLO", "lowercase"); result != "hello" {
		t.Errorf("expected 'hello', got %v", result)
	}
	if result := applyTransform("hello", "uppercase"); result != "HELLO" {
		t.Errorf("expected 'HELLO', got %v", result)
	}
	if result := applyTransform("  text  ", "trim"); result != "text" {
		t.Errorf("expected 'text', got %v", result)
	}
	if result := applyTransform(123, "lowercase"); result != 123 {
		t.Error("non-string should pass through transform")
	}
}
