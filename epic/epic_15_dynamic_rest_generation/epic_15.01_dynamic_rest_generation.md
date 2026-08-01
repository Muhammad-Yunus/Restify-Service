# Epic 15 — Dynamic REST API Generation

**Goal:** Generate REST handlers dynamically from endpoint bindings to database objects (tables, functions, procedures).
**Dependencies:** Epic 14 (Introspection), Epic 12 (Endpoint entity), Epic 06 (DB adapter)
**Commit:** `feat: add dynamic REST API generator from DB bindings`

---

## Step 15.01 — REST Handler Factory

**Build:** Create `backend/internal/infrastructure/baas/rest_generator.go`:

```go
package baas

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "sort"
    "strconv"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const (
    dbContextKey contextKey = "db"
)

// SetDBInContext stores the database connection in the request context.
func SetDBInContext(ctx context.Context, db repository.DB) context.Context {
    return context.WithValue(ctx, dbContextKey, db)
}

// GetDBFromContext retrieves the database connection from the request context.
func GetDBFromContext(ctx context.Context) (repository.DB, bool) {
    db, ok := ctx.Value(dbContextKey).(repository.DB)
    return db, ok
}

// RESTGenerator generates HTTP handlers from endpoint configurations.
type RESTGenerator struct {
    introspector repository.APIIntrospector
    logger       repository.Logger
}

// NewRESTGenerator creates a new REST generator.
func NewRESTGenerator(introspector repository.APIIntrospector, logger repository.Logger) repository.RESTGenerator {
    return &RESTGenerator{
        introspector: introspector,
        logger:       logger,
    }
}

// GenerateHandler creates an HTTP handler for an endpoint.
func (rg *RESTGenerator) GenerateHandler(ctx context.Context, endpoint *entity.Endpoint) (http.HandlerFunc, error) {
    switch endpoint.DBType {
    case entity.EndpointTypeTable:
        return rg.generateTableHandler(endpoint), nil
    case entity.EndpointTypeFunction:
        return rg.generateFunctionHandler(endpoint), nil
    case entity.EndpointTypeProcedure:
        return rg.generateProcedureHandler(endpoint), nil
    default:
        return nil, fmt.Errorf("unsupported endpoint type: %s", endpoint.DBType)
    }
}

func (rg *RESTGenerator) validateBinding(ctx context.Context, endpoint *entity.Endpoint) error {
    switch endpoint.DBType {
    case entity.EndpointTypeTable:
        if endpoint.TableName == "" {
            return fmt.Errorf("table_name is required for table endpoints")
        }
        cols, err := rg.introspector.GetTableSchema(ctx, endpoint.Schema, endpoint.TableName)
        if err != nil {
            return fmt.Errorf("validate table schema: %w", err)
        }
        if len(cols) == 0 {
            return fmt.Errorf("table %s.%s not found", endpoint.Schema, endpoint.TableName)
        }
    case entity.EndpointTypeFunction, entity.EndpointTypeProcedure:
        if endpoint.FuncName == "" {
            return fmt.Errorf("func_name is required for function/procedure endpoints")
        }
    }
    return nil
}

// GenerateTableHandler creates a handler for table CRUD operations.
func (rg *RESTGenerator) generateTableHandler(endpoint *entity.Endpoint) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        pathParts := strings.Split(strings.Trim(endpoint.Path, "/"), "/")

        switch r.Method {
        case http.MethodGet:
            rg.handleTableSelect(w, r, endpoint, pathParts)
        case http.MethodPost:
            rg.handleTableInsert(w, r, endpoint)
        case http.MethodPut, http.MethodPatch:
            rg.handleTableUpdate(w, r, endpoint, pathParts)
        case http.MethodDelete:
            rg.handleTableDelete(w, r, endpoint, pathParts)
        default:
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        }
    }
}

func (rg *RESTGenerator) handleTableSelect(w http.ResponseWriter, r *http.Request, ep *entity.Endpoint, pathParts []string) {
    db, ok := GetDBFromContext(r.Context())
    if !ok {
        writeError(w, http.StatusInternalServerError, "database connection not available")
        return
    }
    schema := ep.Schema
    table := ep.TableName

    // Build query with optional ID filter
    query := fmt.Sprintf("SELECT * FROM %s.%s", schema, table)
    var args []any

    if len(pathParts) > 0 && pathParts[0] != "" {
        id, err := uuid.Parse(pathParts[0])
        if err != nil {
            writeError(w, http.StatusBadRequest, "invalid ID format")
            return
        }
        query += fmt.Sprintf(" WHERE id = $1")
        args = append(args, id)
    }

    // Pagination
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
    if page < 1 {
        page = 1
    }
    if pageSize < 1 || pageSize > 100 {
        pageSize = 20
    }
    offset := (page - 1) * pageSize
    query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
    args = append(args, pageSize, offset)

    rows, err := db.Query(r.Context(), query, args...)
    if err != nil {
        writeError(w, http.StatusInternalServerError, fmt.Sprintf("query error: %v", err))
        return
    }
    defer rows.Close()

    records, err := scanRows(rows)
    if err != nil {
        writeError(w, http.StatusInternalServerError, fmt.Sprintf("scan error: %v", err))
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(gin.H{"data": records, "page": page, "page_size": pageSize})
}

func (rg *RESTGenerator) handleTableInsert(w http.ResponseWriter, r *http.Request, ep *entity.Endpoint) {
    db, ok := GetDBFromContext(r.Context())
    if !ok {
        writeError(w, http.StatusInternalServerError, "database connection not available")
        return
    }
    schema := ep.Schema
    table := ep.TableName

    var record map[string]any
    if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
        writeError(w, http.StatusBadRequest, "invalid JSON body")
        return
    }

    // Build INSERT query
    cols := make([]string, 0, len(record))
    placeholders := make([]string, 0, len(record))
    args := make([]any, 0, len(record))
    for i, k := range orderedKeys(record) {
        cols = append(cols, k)
        placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
        args = append(args, record[k])
    }

    query := fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s) RETURNING *",
        schema, table, strings.Join(cols, ", "), strings.Join(placeholders, ", "))

    row := db.QueryRow(r.Context(), query, args...)
    columns, err := row.Columns()
    if err != nil {
        writeError(w, http.StatusInternalServerError, fmt.Sprintf("get columns: %v", err))
        return
    }
    values := make([]any, len(columns))
    valuePtrs := make([]any, len(columns))
    for i := range values {
        valuePtrs[i] = &values[i]
    }
    if err := row.Scan(valuePtrs...); err != nil {
        writeError(w, http.StatusInternalServerError, fmt.Sprintf("insert error: %v", err))
        return
    }

    result := make(map[string]any, len(columns))
    for i, col := range columns {
        result[col] = values[i]
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(gin.H{"data": result})
}

// handleTableUpdate, handleTableDelete similar patterns...

// GenerateFunctionHandler creates a handler for database functions.
func (rg *RESTGenerator) generateFunctionHandler(endpoint *entity.Endpoint) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        db, ok := GetDBFromContext(ctx)
        if !ok {
            writeError(w, http.StatusInternalServerError, "database connection not available")
            return
        }
        schema := endpoint.Schema
        funcName := endpoint.FuncName

        var args []any
        if r.Body != nil && r.ContentLength > 0 {
            var body map[string]any
            if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
                writeError(w, http.StatusBadRequest, "invalid JSON body")
                return
            }
            // Map body parameters to function args
            params, _ := endpoint.GetFunctionParams()
            for _, p := range params {
                if v, ok := body[p.Name]; ok {
                    args = append(args, v)
                }
            }
        }

        // Build CALL statement
        placeholders := make([]string, len(args))
        for i := range args {
            placeholders[i] = fmt.Sprintf("$%d", i+1)
        }
        query := fmt.Sprintf("SELECT * FROM %s.%s(%s)", schema, funcName, strings.Join(placeholders, ", "))

        rows, err := db.Query(ctx, query, args...)
        if err != nil {
            writeError(w, http.StatusInternalServerError, fmt.Sprintf("function call error: %v", err))
            return
        }
        defer rows.Close()

        records, err := scanRows(rows)
        if err != nil {
            writeError(w, http.StatusInternalServerError, fmt.Sprintf("scan error: %v", err))
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(gin.H{"data": records})
    }
}

// orderedKeys returns sorted keys from a map for deterministic query building.
func orderedKeys(m map[string]any) []string {
    keys := make([]string, 0, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    return keys
}

// scanRows converts pgx.Rows to []map[string]any.
func scanRows(rows pgx.Rows) ([]map[string]any, error) {
    fields := rows.FieldDescriptions()
    var records []map[string]any

    for rows.Next() {
        values, err := rows.Values()
        if err != nil {
            return nil, err
        }
        record := make(map[string]any, len(fields))
        for i, f := range fields {
            record[string(f.Name)] = values[i]
        }
        records = append(records, record)
    }
    return records, rows.Err()
}

func writeError(w http.ResponseWriter, status int, message string) {
    w.Header().Set("Content-Type", "application/problem+json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(gin.H{
        "type":   fmt.Sprintf("https://ForgeBase.api/errors/error"),
        "title":  http.StatusText(status),
        "status": status,
        "detail": message,
    })
}

// ValidateBinding checks if the endpoint's DB binding is valid.
func (rg *RESTGenerator) ValidateBinding(ctx context.Context, endpoint *entity.Endpoint) error {
    return rg.validateBinding(ctx, endpoint)
}

// Compile-time check.
var _ repository.RESTGenerator = (*RESTGenerator)(nil)
```

**Test cases:**
- [ ] Unit: `GenerateHandler()` creates handler for table endpoint
- [ ] Unit: `GenerateHandler()` creates handler for function endpoint
- [ ] Unit: `GenerateHandler()` returns error for unsupported type
- [ ] Unit: `ValidateBinding()` validates table binding
- [ ] Unit: `ValidateBinding()` validates function binding
- [ ] Integration: Generated handler executes SELECT correctly
- [ ] Integration: Generated handler executes INSERT correctly
- [ ] Integration: Generated handler calls function with params

---

## Commit Instruction

```bash
git add .
git commit -m "feat: add dynamic REST API generator from DB endpoint bindings"
```
