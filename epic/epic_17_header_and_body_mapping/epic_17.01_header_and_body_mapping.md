# Epic 17 — Header and Body Mapping

**Goal:** Implement HTTP header mapping (auth, params, custom) and request body mapping to database parameters.
**Dependencies:** Epic 15 (REST Generator), Epic 12 (Endpoint entity)
**Commit:** `feat: add header and body mapping for BaaS endpoints`

---

## Step 17.01 — Header Mapping Engine

**Build:** Create `backend/internal/infrastructure/baas/header_mapper.go`:

```go
// Package baas provides header and body mapping engines for BaaS endpoints.
package baas

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
)

// HeaderMapper extracts and maps HTTP headers to endpoint parameters.
type HeaderMapper struct {
    authHeader   string
    paramHeaders map[string]string // header name -> param name
}

// NewHeaderMapper creates a header mapper from endpoint config.
func NewHeaderMapper(ctx context.Context, authHeader string, paramHeaders []byte) (*HeaderMapper, error) {
    hm := &HeaderMapper{
        authHeader: authHeader,
    }

    if len(paramHeaders) > 0 {
        var headers map[string]string
        if err := json.Unmarshal(paramHeaders, &headers); err != nil {
            return nil, fmt.Errorf("parse param headers: %w", err)
        }
        hm.paramHeaders = headers
    }

    if hm.authHeader == "" {
        hm.authHeader = "Authorization"
    }

    return hm, nil
}

// ExtractAuth extracts the authentication value from the request.
func (hm *HeaderMapper) ExtractAuth(r *http.Request) string {
    return r.Header.Get(hm.authHeader)
}

// ExtractParams extracts parameter values from headers.
func (hm *HeaderMapper) ExtractParams(r *http.Request) map[string]string {
    params := make(map[string]string)
    for headerName, paramName := range hm.paramHeaders {
        value := r.Header.Get(headerName)
        if value != "" {
            params[paramName] = value
        }
    }
    return params
}

// ExtractAll extracts auth and params together.
func (hm *HeaderMapper) ExtractAll(r *http.Request) (string, map[string]string) {
    auth := hm.ExtractAuth(r)
    params := hm.ExtractParams(r)
    return auth, params
}
```

---

## Step 17.02 — Body Mapping Engine

**Build:** Create `backend/internal/infrastructure/baas/body_mapper.go`:

```go
// Package baas provides header and body mapping engines for BaaS endpoints.
package baas

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"
)

// BodyMapper maps HTTP request body to database parameters.
type BodyMapper struct {
    mappingJSON []byte // JSONB from endpoint
    mapping     map[string]BodyMappingRule
}

// BodyMappingRule defines how a request field maps to a DB parameter.
type BodyMappingRule struct {
    SourceField string `json:"source_field"`
    TargetParam string `json:"target_param"`
    Transform   string `json:"transform,omitempty"` // "lowercase", "uppercase", "trim"
}

// NewBodyMapper creates a body mapper from endpoint config.
func NewBodyMapper(ctx context.Context, mappingJSON []byte) (*BodyMapper, error) {
    bm := &BodyMapper{}

    if len(mappingJSON) == 0 {
        return bm, nil // no custom mapping
    }

    var rules []BodyMappingRule
    if err := json.Unmarshal(mappingJSON, &rules); err != nil {
        return nil, fmt.Errorf("parse body mapping: %w", err)
    }

    bm.mapping = make(map[string]BodyMappingRule, len(rules))
    for _, rule := range rules {
        bm.mapping[rule.SourceField] = rule
    }

    bm.mappingJSON = mappingJSON
    return bm, nil
}

// MapBody maps the request body to DB parameters.
func (bm *BodyMapper) MapBody(ctx context.Context, r *http.Request) (map[string]any, error) {
    bodyBytes, err := io.ReadAll(r.Body)
    if err != nil {
        return nil, fmt.Errorf("read body: %w", err)
    }
    defer r.Body.Close()

    var raw map[string]any
    if err := json.Unmarshal(bodyBytes, &raw); err != nil {
        return nil, fmt.Errorf("parse JSON body: %w", err)
    }

    if bm == nil || len(bm.mapping) == 0 {
        // No custom mapping — use body as-is
        return raw, nil
    }

    // Apply custom mapping
    result := make(map[string]any, len(raw))
    for sourceField, value := range raw {
        rule, hasRule := bm.mapping[sourceField]
        if hasRule {
            mappedValue := applyTransform(value, rule.Transform)
            result[rule.TargetParam] = mappedValue
        } else {
            // Unknown field — pass through or ignore based on config
            result[sourceField] = value
        }
    }

    return result, nil
}

func applyTransform(value any, transform string) any {
    switch transform {
    case "lowercase":
        if s, ok := value.(string); ok {
            return strings.ToLower(s)
        }
    case "uppercase":
        if s, ok := value.(string); ok {
            return strings.ToUpper(s)
        }
    case "trim":
        if s, ok := value.(string); ok {
            return strings.TrimSpace(s)
        }
    }
    return value
}
```

---

## Step 17.03 — Integration with REST Generator

**Build:** Update `internal/infrastructure/baas/rest_generator.go` to use mappers:

```go
// In handleTableInsert:
func (rg *RESTGenerator) handleTableInsert(ctx context.Context, w http.ResponseWriter, r *http.Request, ep *entity.Endpoint) {
    bodyMapper, err := NewBodyMapper(ctx, ep.BodyMappingJSON)
    if err != nil {
        writeError(w, http.StatusBadRequest, fmt.Sprintf("body mapping error: %v", err))
        return
    }
    mappedBody, err := bodyMapper.MapBody(ctx, r)
    if err != nil {
        writeError(w, http.StatusBadRequest, fmt.Sprintf("body mapping error: %v", err))
        return
    }
    // ... use mappedBody instead of raw body
}

// In handleFunctionCall:
func (rg *RESTGenerator) handleFunctionCall(ctx context.Context, w http.ResponseWriter, r *http.Request, ep *entity.Endpoint) {
    headerMapper, err := NewHeaderMapper(ctx, ep.AuthHeader, ep.ParamHeaders)
    if err != nil {
        writeError(w, http.StatusBadRequest, fmt.Sprintf("header mapping error: %v", err))
        return
    }
    _, params := headerMapper.ExtractAll(r)

    // Merge header params with body params
    bodyMapper, err := NewBodyMapper(ctx, ep.BodyMappingJSON)
    if err != nil {
        writeError(w, http.StatusBadRequest, fmt.Sprintf("body mapping error: %v", err))
        return
    }
    bodyParams, err := bodyMapper.MapBody(ctx, r)
    if err != nil {
        writeError(w, http.StatusBadRequest, fmt.Sprintf("body mapping error: %v", err))
        return
    }

    // Merge: header params take precedence
    for k, v := range params {
        bodyParams[k] = v
    }
    // ... use merged params for function call
}
```

---

## Step 17.04 — Header Mapping Tests

**Test cases:**
- [ ] Unit: `ExtractAuth()` reads Authorization header
- [ ] Unit: `ExtractParams()` reads custom header mappings
- [ ] Unit: `MapBody()` applies transform rules
- [ ] Unit: `MapBody()` passes through when no mapping defined
- [ ] Unit: `MapBody()` handles nil mapper gracefully
- [ ] Unit: `applyTransform()` lowercases strings
- [ ] Unit: `applyTransform()` uppercases strings
- [ ] Unit: `applyTransform()` trims strings
- [ ] Integration: Full request with headers + body mapping
- [ ] Integration: Error returned when body mapping config is invalid JSON

---

## Commit Instruction

```bash
git add .
git commit -m "feat: add header and body mapping engines for BaaS endpoints"
```
