# Epic 18 — API Documentation Generator

## Goal

Implement auto-generated API documentation using OpenAPI v3 spec, supporting both JSON and YAML output formats.

## Tasks Completed

### 18.01 — OpenAPI Document Model
**Files:**
- `internal/infrastructure/baas/openapi.go`

**What was implemented:**
- `OpenAPI` struct — top-level document following OpenAPI v3.0.3 spec
- `Info` struct — API title, version, description
- `Server` struct — server URL configuration
- `Tag` struct — endpoint grouping tags
- `PathItem` struct — all HTTP methods for one path (Get, Post, Put, Delete, Patch, Parameters)
- `Operation` struct — single HTTP operation with tags, summary, description, operationId, params, request body, responses, security
- `Param` struct — operation-level parameter (name, in, description, required, schema)
- `RequestBody` / `MediaType` — request body content schema
- `Response` struct — response definition
- `Schema` struct — JSON Schema type description (type, items, properties, required, ref, anyOf)
- `SecurityReq` struct — auth requirement (bearerAuth)
- `ParamSchema` struct — function parameter schema (name, type, mode)
- Helper functions: `openAPISchemaType()`, `buildParamSchema()`, `sanitizePath()`

**Testing:**
```bash
go test ./internal/infrastructure/baas/ -v -run TestOpenAPISchemaType
go test ./internal/infrastructure/baas/ -v -run TestSanitizePath
go test ./internal/infrastructure/baas/ -v -run TestBuildParamSchema
```
✅ All pass

---

### 18.02 — Documentation Generator Service
**Files:**
- `internal/infrastructure/baas/doc_generator.go`

**What was implemented:**
- `APIDocGenerator` struct — generates OpenAPI documents from endpoint entities
- `NewAPIDocGenerator()` — constructor accepting introspector and logger
- `GenerateOpenAPI()` — main entry point producing full OpenAPI document from `[]*entity.Endpoint`
  - Groups endpoints by collection name and creates tags
  - Builds servers list from baseURL parameter
  - Iterates endpoints and adds each to doc.Paths
- `addEndpointToDoc()` — maps endpoint method+path to PathItem in OpenAPI document
- `buildOperation()` — creates Operation from endpoint
  - Sets operationId, summary, tags, parameters
  - Handles security policy (auth_required → security requirement)
  - For function/procedure endpoints: resolves input schema from introspector and adds to requestBody
- `resolveInputSchema()` — fetches function signature via introspector and builds ParamSchema
- `buildParams()` — extracts auth header and param-header mappings into OpenAPI params
- `sanitizePath()` — ensures leading slash, handles empty input

**Testing:**
```bash
go test ./internal/infrastructure/baas/ -v -run TestGenerateOpenAPI_Basic
go test ./internal/infrastructure/baas/ -v -run TestGenerateOpenAPI_AuthRequired
go test ./internal/infrastructure/baas/ -v -run TestGenerateOpenAPI_FunctionEndpoint
```
✅ All pass

---

### 18.03 — OpenAPI HTTP Handler
**Files:**
- `internal/infrastructure/baas/openapi_handler.go`

**What was implemented:**
- `OpenAPIHandler` struct — HTTP handler for serving OpenAPI docs
- `NewOpenAPIHandler()` — constructor accepting generator and endpoint fetcher function
- `GetJSONHandler()` — returns `http.HandlerFunc` for `/openapi.json`
- `GetYAMLHandler()` — returns `http.HandlerFunc` for `/openapi.yaml`
- `serve()` — internal shared serving logic:
  - Fetches endpoints via endpoint fetcher
  - Reads query params: `title`, `version`, `baseurl`
  - Generates OpenAPI document
  - Marshals to JSON or YAML based on format parameter
  - Sets correct Content-Type header (`application/json` or `application/yaml`)

**Constants (doc_constants.go):**
```go
const (
    ContentTypeJSONValue = "application/json"
    ContentTypeYAMLValue = "application/yaml"
)

type docFormat string

const (
    docFormatJSON docFormat = "json"
    docFormatYAML docFormat = "yaml"
)
```

**Pending:** Register routes in router initialization:
```go
// In internal/infrastructure/presentation/http/router/router.go
openapiHandler := baas.NewOpenAPIHandler(docGenerator, fetchEndpoints)
public.GET("/openapi.json", openapiHandler.GetJSONHandler())
public.GET("/openapi.yaml", openapiHandler.GetYAMLHandler())
```

---

### 18.04 — OpenAPI Generation Tests
**Files:**
- `internal/infrastructure/baas/doc_generator_test.go`

**Test coverage:**
| Test | Coverage |
|------|----------|
| `TestGenerateOpenAPI_Basic` | Document structure, paths, tags, servers populated correctly |
| `TestGenerateOpenAPI_AuthRequired` | Security policy applied to endpoints requiring auth |
| `TestGenerateOpenAPI_FunctionEndpoint` | Function endpoints generate request body with schema |
| `TestBuildParamSchema` | IN/INOUT params become required, all params in properties |
| `TestOpenAPISchemaType` | DB types map to correct OpenAPI types (string, integer, boolean, object, number) |
| `TestSanitizePath` | Path normalization (leading slash, empty string) |
| `TestBuildParams` | Custom auth header and param-header mappings extracted into params |

**Run all:**
```bash
go test ./internal/infrastructure/baas/ -v
# PASS: TestGenerateOpenAPI_Basic, _AuthRequired, _FunctionEndpoint
# PASS: TestBuildParamSchema, TestOpenAPISchemaType, TestSanitizePath, TestBuildParams
```

---

## File Map

| File | Role |
|------|------|
| `internal/infrastructure/baas/openapi.go` | OpenAPI v3.0.3 document model |
| `internal/infrastructure/baas/doc_generator.go` | Generates OpenAPI from endpoints |
| `internal/infrastructure/baas/openapi_handler.go` | HTTP handler for /openapi.json & /openapi.yaml |
| `internal/infrastructure/baas/doc_constants.go` | Format constants and type aliases |
| `internal/infrastructure/baas/doc_generator_test.go` | Tests for generator |

## Integration Checklist

- [x] OpenAPI v3.0.3 document model
- [x] Generator service producing docs from entities
- [x] HTTP handler with JSON + YAML format support
- [x] Security policy → security requirements mapping
- [x] Function input schema resolution
- [x] Collection grouping → tags
- [x] 7 unit tests covering all paths
- [x] All tests passing ✅
- [ ] **Pending:** Register routes in router.go (see `serve()` comment in openapi_handler.go)

## Example OpenAPI Output

```json
{
  "openapi": "3.0.3",
  "info": { "title": "Test API", "version": "1.0.0" },
  "servers": [{ "url": "http://localhost:3000", "description": "API Server" }],
  "tags": [{ "name": "Users", "description": "Users" }],
  "paths": {
    "/users": {
      "get": {
        "tags": ["Users"],
        "summary": "Get all users",
        "operationId": "GET_/users",
        "responses": {
          "200": { "description": "Successful response" }
        },
        "security": []
      },
      "post": {
        "tags": ["Users"],
        "summary": "Create user",
        "operationId": "POST_/users",
        "responses": { "200": { "description": "Successful response" } }
      }
    }
  },
  "security": [{ "bearerAuth": [""] }]
}
```
