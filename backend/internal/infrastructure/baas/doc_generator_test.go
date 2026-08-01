package baas

import (
	"testing"

	"github.com/google/uuid"
	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
)

func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		panic(err)
	}
	return id
}

func TestGenerateOpenAPI_Basic(t *testing.T) {
	g := NewAPIDocGenerator(nil, nil)

	workspaceID := parseUUID("550e8400-e29b-41d4-a716-446655440005")
	col := &entity.Collection{
		ID:          parseUUID("550e8400-e29b-41d4-a716-446655440000"),
		Name:        "Users",
		Slug:        "users",
		WorkspaceID: workspaceID,
	}

	endpoints := []*entity.Endpoint{
		{
			ID:           parseUUID("550e8400-e29b-41d4-a716-446655440001"),
			CollectionID: col.ID,
			Name:         "Get all users",
			Path:         "users",
			Method:       "GET",
			Version:      "v1",
			DBType:       entity.EndpointTypeTable,
			TableName:    "users",
			IsActive:     true,
			Collection:   col,
		},
		{
			ID:           parseUUID("550e8400-e29b-41d4-a716-446655440002"),
			CollectionID: col.ID,
			Name:         "Create user",
			Path:         "users",
			Method:       "POST",
			Version:      "v1",
			DBType:       entity.EndpointTypeTable,
			TableName:    "users",
			IsActive:     true,
			Collection:   col,
		},
	}

	doc := g.GenerateOpenAPI(nil, endpoints, "Test API", "1.0.0", "http://localhost:3000")

	if doc.OpenAPI != "3.0.3" {
		t.Errorf("expected openapi version 3.0.3, got %s", doc.OpenAPI)
	}
	if doc.Info.Title != "Test API" {
		t.Errorf("expected title 'Test API', got %s", doc.Info.Title)
	}
	if len(doc.Paths) == 0 {
		t.Error("expected paths to be populated")
	}
	if len(doc.Tags) == 0 {
		t.Error("expected tags to be populated")
	}
	if len(doc.Servers) == 0 {
		t.Error("expected servers to be populated")
	}

	usersPath := doc.Paths["/users"]
	if usersPath == nil {
		t.Fatal("missing /users path")
	}
	if usersPath.Get == nil {
		t.Error("missing GET /users operation")
	}
	if usersPath.Post == nil {
		t.Error("missing POST /users operation")
	}
}

func TestGenerateOpenAPI_AuthRequired(t *testing.T) {
	g := NewAPIDocGenerator(nil, nil)

	authHeader := "X-API-Key"
	policyJSON := []byte(`{"auth_required":true,"allowed_roles":["admin"]}`)

	endpoints := []*entity.Endpoint{
		{
			ID:                 parseUUID("550e8400-e29b-41d4-a716-446655440003"),
			Name:               "Protected endpoint",
			Path:               "protected",
			Method:             "GET",
			Version:            "v1",
		 DBType:             entity.EndpointTypeTable,
			TableName:          "secret_data",
			IsActive:           true,
			AuthHeader:         authHeader,
			SecurityPolicyJSON: policyJSON,
		},
	}

	doc := g.GenerateOpenAPI(nil, endpoints, "Test API", "1.0.0", "http://localhost:3000")

	path := doc.Paths["/protected"]
	if path == nil || path.Get == nil {
		t.Fatal("expected /protected GET operation")
	}
	if path.Get.Security == nil {
		t.Error("expected security requirement on protected endpoint")
	}
}

func TestGenerateOpenAPI_FunctionEndpoint(t *testing.T) {
	g := NewAPIDocGenerator(nil, nil)

	endpoints := []*entity.Endpoint{
		{
			ID:       parseUUID("550e8400-e29b-41d4-a716-446655440004"),
			Name:     "Call hello",
			Path:     "hello",
			Method:   "POST",
			Version:  "v1",
			DBType:   entity.EndpointTypeFunction,
			FuncName: "hello_world",
			IsActive: true,
		},
	}

	doc := g.GenerateOpenAPI(nil, endpoints, "Test API", "1.0.0", "http://localhost:3000")

	path := doc.Paths["/hello"]
	if path == nil || path.Post == nil {
		t.Fatal("expected /hello POST operation")
	}
}

func TestBuildParamSchema(t *testing.T) {
	schema := buildParamSchema([]ParamSchema{
		{Name: "name", Type: "text", Mode: "IN"},
		{Name: "age", Type: "integer", Mode: "IN"},
		{Name: "result", Type: "text", Mode: "OUT"},
	})

	if schema == nil {
		t.Fatal("expected non-nil schema")
	}
	if schema.Type != "object" {
		t.Errorf("expected type 'object', got %s", schema.Type)
	}
	if len(schema.Properties) != 3 {
		t.Errorf("expected 3 properties (all params), got %d", len(schema.Properties))
	}
	if len(schema.Required) != 2 {
		t.Errorf("expected 2 required (IN params only), got %d", len(schema.Required))
	}
}

func TestOpenAPISchemaType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"text", "string"},
		{"varchar", "string"},
		{"integer", "integer"},
		{"boolean", "boolean"},
		{"jsonb", "object"},
		{"numeric", "number"},
		{"uuid", "string"},
		{"unknown", "string"},
	}

	for _, tt := range tests {
		got := openAPISchemaType(tt.input)
		if got != tt.expected {
			t.Errorf("openAPISchemaType(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"users", "/users"},
		{"/users", "/users"},
		{"", "/"},
	}

	for _, tt := range tests {
		got := sanitizePath(tt.input)
		if got != tt.expected {
			t.Errorf("sanitizePath(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestBuildParams(t *testing.T) {
	authHeader := "X-API-Key"
	// param_headers JSON: {"X-User-ID": "user_id"}
	paramHeaders := []byte(`{"X-User-ID":"user_id"}`)

	ep := &entity.Endpoint{
		AuthHeader:   authHeader,
		ParamHeaders: paramHeaders,
	}

	params := buildParams(ep)

	if len(params) == 0 {
		t.Fatal("expected params")
	}

	found := false
	for _, p := range params {
		if p.Name == "user_id" && p.In == "header" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected custom header param 'user_id'")
	}
}