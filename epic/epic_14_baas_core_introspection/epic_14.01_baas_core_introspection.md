# Epic 14 — BaaS Core: Database Introspection

**Goal:** Implement PostgreSQL schema introspection — discover tables, functions, procedures, and column metadata.
**Dependencies:** Epic 03 (Database layer), Epic 04 (Domain entities), Epic 06 (DB adapter)
**Commit:** `feat: add PostgreSQL schema introspection engine`

---

## Step 14.01 — PostgreSQL Introspector Implementation

**Build:** Create `backend/internal/infrastructure/baas/postgres_introspector.go`:

```go
package baas

import (
    "context"
    "database/sql"
    "fmt"
    "strings"

    "github.com/jackc/pgx/v5"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// PostgreSQLIntrospector discovers database schema objects.
type PostgreSQLIntrospector struct {
    db *pgxpool.Pool
}

// NewPostgreSQLIntrospector creates a new introspector.
func NewPostgreSQLIntrospector(db repository.DB) repository.APIIntrospector {
    return &PostgreSQLIntrospector{db: db.(*PostgresDB).Pool()}
}

// DiscoverTables returns all user tables in a schema.
func (pi *PostgreSQLIntrospector) DiscoverTables(ctx context.Context, schema string) ([]repository.TableInfo, error) {
    query := `
        SELECT table_schema, table_name
        FROM information_schema.tables
        WHERE table_schema = $1
          AND table_type = 'BASE TABLE'
          AND table_name NOT LIKE 'pg_%'
          AND table_name NOT LIKE 'sql_%'
        ORDER BY table_name
    `
    rows, err := pi.db.Query(ctx, query, schema)
    if err != nil {
        return nil, fmt.Errorf("discover tables: %w", err)
    }
    defer rows.Close()

    var tables []repository.TableInfo
    for rows.Next() {
        var tbl repository.TableInfo
        if err := rows.Scan(&tbl.Schema, &tbl.Name); err != nil {
            return nil, fmt.Errorf("scan table: %w", err)
        }
        tables = append(tables, tbl)
    }
    return tables, nil
}

// DiscoverFunctions returns all user-defined functions.
func (pi *PostgreSQLIntrospector) DiscoverFunctions(ctx context.Context, schema string) ([]repository.FunctionInfo, error) {
    query := `
        SELECT routine_schema, routine_name, data_type
        FROM information_schema.routines
        WHERE routine_schema = $1
          AND routine_type = 'FUNCTION'
        ORDER BY routine_name
    `
    rows, err := pi.db.Query(ctx, query, schema)
    if err != nil {
        return nil, fmt.Errorf("discover functions: %w", err)
    }
    defer rows.Close()

    var funcs []repository.FunctionInfo
    for rows.Next() {
        var fn repository.FunctionInfo
        if err := rows.Scan(&fn.Schema, &fn.Name, &fn.ReturnType); err != nil {
            return nil, fmt.Errorf("scan function: %w", err)
        }
        funcs = append(funcs, fn)
    }
    return funcs, nil
}

// DiscoverProcedures returns all user-defined procedures.
func (pi *PostgreSQLIntrospector) DiscoverProcedures(ctx context.Context, schema string) ([]repository.ProcedureInfo, error) {
    // PostgreSQL 11+ supports PROCEDURES
    query := `
        SELECT routine_schema, routine_name
        FROM information_schema.routines
        WHERE routine_schema = $1
          AND routine_type = 'PROCEDURE'
        ORDER BY routine_name
    `
    rows, err := pi.db.Query(ctx, query, schema)
    if err != nil {
        // Fall back to pg_proc for older versions
        return pi.discoverProceduresFallback(ctx, schema)
    }
    defer rows.Close()

    var procs []repository.ProcedureInfo
    for rows.Next() {
        var proc repository.ProcedureInfo
        if err := rows.Scan(&proc.Schema, &proc.Name); err != nil {
            return nil, fmt.Errorf("scan procedure: %w", err)
        }
        procs = append(procs, proc)
    }
    return procs, nil
}

func (pi *PostgreSQLIntrospector) discoverProceduresFallback(ctx context.Context, schema string) ([]repository.ProcedureInfo, error) {
    query := `
        SELECT pronamespace::regnamespace::text, proname
        FROM pg_proc p
        JOIN pg_namespace n ON n.oid = p.pronamespace
        WHERE n.nspname = $1
          AND NOT p.proisagg
          AND NOT p.proiswindow
        ORDER BY proname
    `
    rows, err := pi.db.Query(ctx, query, schema)
    if err != nil {
        return nil, fmt.Errorf("discover procedures fallback: %w", err)
    }
    defer rows.Close()

    var procs []repository.ProcedureInfo
    for rows.Next() {
        var proc repository.ProcedureInfo
        if err := rows.Scan(&proc.Schema, &proc.Name); err != nil {
            return nil, fmt.Errorf("scan procedure: %w", err)
        }
        procs = append(procs, proc)
    }
    return procs, nil
}

// GetTableSchema returns column information for a table.
func (pi *PostgreSQLIntrospector) GetTableSchema(ctx context.Context, schema, table string) ([]repository.ColumnSchema, error) {
    query := `
        SELECT column_name, data_type, is_nullable, column_default
        FROM information_schema.columns
        WHERE table_schema = $1 AND table_name = $2
        ORDER BY ordinal_position
    `
    rows, err := pi.db.Query(ctx, query, schema, table)
    if err != nil {
        return nil, fmt.Errorf("get table schema: %w", err)
    }
    defer rows.Close()

    var columns []repository.ColumnSchema
    for rows.Next() {
        var col repository.ColumnSchema
        var defaultVal sql.NullString
        if err := rows.Scan(&col.Name, &col.Type, &col.IsNullable, &defaultVal); err != nil {
            return nil, fmt.Errorf("scan column: %w", err)
        }
        col.Default = defaultVal.StringPtr()
        columns = append(columns, col)
    }
    return columns, nil
}

// GetFunctionSignature returns parameter info for a function.
func (pi *PostgreSQLIntrospector) GetFunctionSignature(ctx context.Context, schema, name string) ([]repository.ParamSchema, error) {
    query := `
        SELECT arg.order, arg.name, arg.data_type, arg.mode
        FROM (
            SELECT p.argument_ORDINAL_POSITION AS "order",
                   p.argument_NAME AS name,
                   p.data_type AS data_type,
                   p.argument_mode AS mode
            FROM information_schema.parameters p
            WHERE p.specific_schema = $1
              AND p.specific_name = $2
        ) arg
        ORDER BY arg."order"
    `
    rows, err := pi.db.Query(ctx, query, schema, name)
    if err != nil {
        return nil, fmt.Errorf("get function signature: %w", err)
    }
    defer rows.Close()

    var params []repository.ParamSchema
    for rows.Next() {
        var p repository.ParamSchema
        if err := rows.Scan(&p.Name, &p.Type, &p.Mode); err != nil {
            return nil, fmt.Errorf("scan param: %w", err)
        }
        params = append(params, p)
    }
    return params, nil
}

// Compile-time check.
var _ repository.APIIntrospector = (*PostgreSQLIntrospector)(nil)
```

**Test cases:**
- [ ] Unit: `DiscoverTables()` returns tables for given schema
- [ ] Unit: `DiscoverTables()` excludes system tables
- [ ] Unit: `DiscoverFunctions()` returns user-defined functions
- [ ] Unit: `GetTableSchema()` returns column definitions
- [ ] Unit: `GetFunctionSignature()` returns parameter list
- [ ] Integration: Full introspection against test PostgreSQL instance

---

## Step 14.02 — Introspector Service

**Build:** Create `backend/internal/application/service/introspector_service.go`:

```go
package service

import (
    "context"
    "fmt"

    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// IntrospectorService wraps the APIIntrospector with caching and validation.
type IntrospectorService struct {
    introspector repository.APIIntrospector
}

// NewIntrospectorService creates a new introspector service.
func NewIntrospectorService(introspector repository.APIIntrospector) repository.APIIntrospector {
    return &IntrospectorService{introspector: introspector}
}

// DiscoverTables returns all user tables.
func (s *IntrospectorService) DiscoverTables(ctx context.Context, schema string) ([]repository.TableInfo, error) {
    tables, err := s.introspector.DiscoverTables(ctx, schema)
    if err != nil {
        return nil, fmt.Errorf("discover tables: %w", err)
    }
    // Enrich with column info
    for i := range tables {
        cols, err := s.introspector.GetTableSchema(ctx, schema, tables[i].Name)
        if err != nil {
            continue // skip table if schema query fails
        }
        tables[i].Columns = cols
    }
    return tables, nil
}

// DiscoverFunctions returns all user-defined functions.
func (s *IntrospectorService) DiscoverFunctions(ctx context.Context, schema string) ([]repository.FunctionInfo, error) {
    funcs, err := s.introspector.DiscoverFunctions(ctx, schema)
    if err != nil {
        return nil, fmt.Errorf("discover functions: %w", err)
    }
    // Enrich with parameter info
    for i := range funcs {
        params, err := s.introspector.GetFunctionSignature(ctx, schema, funcs[i].Name)
        if err != nil {
            continue
        }
        funcs[i].Params = params
    }
    return funcs, nil
}

// DiscoverProcedures returns all user-defined procedures.
func (s *IntrospectorService) DiscoverProcedures(ctx context.Context, schema string) ([]repository.ProcedureInfo, error) {
    procs, err := s.introspector.DiscoverProcedures(ctx, schema)
    if err != nil {
        return nil, fmt.Errorf("discover procedures: %w", err)
    }
    for i := range procs {
        params, err := s.introspector.GetFunctionSignature(ctx, schema, procs[i].Name)
        if err != nil {
            continue
        }
        procs[i].Params = params
    }
    return procs, nil
}
```

**Test cases:**
- [ ] Unit: `DiscoverTables()` enriches tables with column info
- [ ] Unit: `DiscoverFunctions()` enriches functions with params
- [ ] Unit: Gracefully skips tables/functions with schema errors

---

## Step 14.03 — Introspection HTTP Handler

**Build:** Create `backend/internal/presentation/http/handler/introspect_handler.go`:

```go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/muhammadyunus/ForgeBase/internal/application/service"
    "github.com/muhammadyunus/ForgeBase/internal/presentation/http/dto"
)

// IntrospectHandler handles schema introspection requests.
type IntrospectHandler struct {
    introspector service.APIIntrospector
}

// NewIntrospectHandler creates a new introspect handler.
func NewIntrospectHandler(introspector service.APIIntrospector) *IntrospectHandler {
    return &IntrospectHandler{introspector: introspector}
}

// DiscoverTables handles GET /api/v1/introspect/tables?schema=public
func (h *IntrospectHandler) DiscoverTables(c *gin.Context) {
    schema := c.DefaultQuery("schema", "public")

    tables, err := h.introspector.DiscoverTables(c.Request.Context(), schema)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": toTableListDTO(tables)})
}

// DiscoverFunctions handles GET /api/v1/introspect/functions?schema=public
func (h *IntrospectHandler) DiscoverFunctions(c *gin.Context) {
    schema := c.DefaultQuery("schema", "public")

    funcs, err := h.introspector.DiscoverFunctions(c.Request.Context(), schema)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": toFunctionListDTO(funcs)})
}

// GetTableSchema handles GET /api/v1/introspect/tables/:name/schema
func (h *IntrospectHandler) GetTableSchema(c *gin.Context) {
    schema := c.DefaultQuery("schema", "public")
    tableName := c.Param("name")

    columns, err := h.introspector.GetTableSchema(c.Request.Context(), schema, tableName)
    if err != nil {
        c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "Table not found", err.Error()))
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": toColumnListDTO(columns)})
}

func toTableListDTO(tables []repository.TableInfo) []gin.H {
    result := make([]gin.H, len(tables))
    for i, t := range tables {
        cols := make([]gin.H, len(t.Columns))
        for j, c := range t.Columns {
            cols[j] = gin.H{
                "name":       c.Name,
                "type":       c.Type,
                "nullable":   c.IsNullable == "YES",
                "default":    c.Default,
            }
        }
        result[i] = gin.H{
            "schema":  t.Schema,
            "name":    t.Name,
            "columns": cols,
        }
    }
    return result
}

func toFunctionListDTO(funcs []repository.FunctionInfo) []gin.H {
    result := make([]gin.H, len(funcs))
    for i, f := range funcs {
        params := make([]gin.H, len(f.Params))
        for j, p := range f.Params {
            params[j] = gin.H{"name": p.Name, "type": p.Type, "mode": p.Mode}
        }
        result[i] = gin.H{
            "schema":     f.Schema,
            "name":       f.Name,
            "return_type": f.ReturnType,
            "params":     params,
        }
    }
    return result
}

func toColumnListDTO(cols []repository.ColumnSchema) []gin.H {
    result := make([]gin.H, len(cols))
    for i, c := range cols {
        result[i] = gin.H{
            "name":     c.Name,
            "type":     c.Type,
            "nullable": c.IsNullable == "YES",
            "default":  c.Default,
        }
    }
    return result
}
```

---

## Commit Instruction

```bash
git add .
git commit -m "feat: add PostgreSQL schema introspection engine for BaaS core"
```
