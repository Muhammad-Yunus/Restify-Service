# Epic 12 — Endpoint Service & Repository

**Goal:** Implement endpoint repository, service, and HTTP handlers for endpoint management.
**Dependencies:** Epic 04 (Endpoint entity), Epic 05 (Repository interfaces), Epic 06 (DB adapter), Epic 11 (Collection service)
**Commit:** `feat: implement endpoint repository, service, and HTTP handlers`

---

## Step 12.01 — Endpoint Repository Implementation

**Build:** Create `backend/internal/application/repository/endpoint_repository.go`:

```go
package repository

import (
    "context"
    "errors"
    "fmt"

    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
    "gorm.io/gorm"
)

// EndpointRepositoryImpl implements the repository.EndpointRepository interface.
type EndpointRepositoryImpl struct {
    db *gorm.DB
}

// NewEndpointRepository creates a new endpoint repository.
func NewEndpointRepository(db repository.DB) repository.EndpointRepository {
    return &EndpointRepositoryImpl{db: gormDB(db)}
}

func (r *EndpointRepositoryImpl) Create(ctx context.Context, ep *entity.Endpoint) error {
    if ep.ID == uuid.Nil {
        ep.ID = uuid.New()
    }
    if ep.Version == "" {
        ep.Version = "v1"
    }
    if ep.Method == "" {
        ep.Method = "GET"
    }
    if err := r.db.WithContext(ctx).Create(ep).Error; err != nil {
        return fmt.Errorf("create endpoint: %w", err)
    }
    return nil
}

func (r *EndpointRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entity.Endpoint, error) {
    var ep entity.Endpoint
    err := r.db.WithContext(ctx).Preload("Collection").First(&ep, "id = ?", id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, entity.ErrNotFound
        }
        return nil, fmt.Errorf("find endpoint %s: %w", id, err)
    }
    return &ep, nil
}

func (r *EndpointRepositoryImpl) ListByCollection(ctx context.Context, collectionID uuid.UUID) ([]*entity.Endpoint, error) {
    var eps []*entity.Endpoint
    if err := r.db.WithContext(ctx).
        Where("collection_id = ?", collectionID).
        Order("created_at ASC").
        Find(&eps).Error; err != nil {
        return nil, fmt.Errorf("list endpoints for collection %s: %w", collectionID, err)
    }
    return eps, nil
}

func (r *EndpointRepositoryImpl) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entity.Endpoint, error) {
    var eps []*entity.Endpoint
    if err := r.db.WithContext(ctx).
        Table("endpoints").
        Select("endpoints.*").
        Joins("JOIN collections ON endpoints.collection_id = collections.id").
        Where("collections.workspace_id = ?", workspaceID).
        Order("endpoints.created_at ASC").
        Find(&eps).Error; err != nil {
        return nil, fmt.Errorf("list endpoints for workspace %s: %w", workspaceID, err)
    }
    return eps, nil
}

func (r *EndpointRepositoryImpl) Update(ctx context.Context, ep *entity.Endpoint) error {
    if err := r.db.WithContext(ctx).Model(ep).Updates(map[string]any{
        "name":                 ep.Name,
        "description":          ep.Description,
        "path":                 ep.Path,
        "method":               ep.Method,
        "version":              ep.Version,
        "db_type":              ep.DBType,
        "schema":               ep.Schema,
        "table_name":           ep.TableName,
        "func_name":            ep.FuncName,
        "params":               ep.Params,
        "operations":           ep.Operations,
        "security_policy_json": ep.SecurityPolicyJSON,
        "auth_header":          ep.AuthHeader,
        "param_headers":        ep.ParamHeaders,
        "body_mapping_json":    ep.BodyMappingJSON,
    }).Error; err != nil {
        return fmt.Errorf("update endpoint %s: %w", ep.ID, err)
    }
    return nil
}

func (r *EndpointRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
    result := r.db.WithContext(ctx).Delete(&entity.Endpoint{}, "id = ?", id)
    if result.Error != nil {
        return fmt.Errorf("delete endpoint %s: %w", id, result.Error)
    }
    if result.RowsAffected == 0 {
        return entity.ErrNotFound
    }
    return nil
}

func (r *EndpointRepositoryImpl) ToggleActive(ctx context.Context, id uuid.UUID, active bool) error {
    result := r.db.WithContext(ctx).Model(&entity.Endpoint{}).
        Where("id = ?", id).
        Update("is_active", active)
    if result.Error != nil {
        return fmt.Errorf("toggle endpoint %s: %w", id, result.Error)
    }
    if result.RowsAffected == 0 {
        return entity.ErrNotFound
    }
    return nil
}

func (r *EndpointRepositoryImpl) FindByPath(ctx context.Context, path, version string) (*entity.Endpoint, error) {
    var ep entity.Endpoint
    err := r.db.WithContext(ctx).
        Where("path = ? AND version = ? AND is_active = ?", path, version, true).
        First(&ep).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, entity.ErrNotFound
        }
        return nil, fmt.Errorf("find endpoint by path: %w", err)
    }
    return &ep, nil
}

func (r *EndpointRepositoryImpl) CountByWorkspace(ctx context.Context, workspaceID uuid.UUID) (int, error) {
    var count int64
    if err := r.db.WithContext(ctx).
        Table("endpoints").
        Select("count(*)").
        Joins("JOIN collections ON endpoints.collection_id = collections.id").
        Where("collections.workspace_id = ?", workspaceID).
        Scan(&count).Error; err != nil {
        return 0, fmt.Errorf("count endpoints: %w", err)
    }
    return int(count), nil
}
```

**Test cases:**
- [ ] Unit: `Create()` creates endpoint with defaults (v1, GET)
- [ ] Unit: `FindByID()` returns endpoint with collection
- [ ] Unit: `ListByCollection()` returns all endpoints
- [ ] Unit: `ToggleActive()` enables/disables endpoint
- [ ] Unit: `FindByPath()` finds active endpoint by path+version
- [ ] Integration: Full endpoint lifecycle

---

## Step 12.02 ��� Endpoint Service

**Build:** Create `backend/internal/application/service/endpoint_service.go`:

```go
package service

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// EndpointService implements the service.EndpointService interface.
type EndpointService struct {
    repo   repository.EndpointRepository
    logger repository.Logger
}

// NewEndpointService creates a new endpoint service.
func NewEndpointService(repo repository.EndpointRepository, logger repository.Logger) repository.EndpointService {
    return &EndpointService{repo: repo, logger: logger}
}

func (s *EndpointService) Create(ctx context.Context, collectionID uuid.UUID, params map[string]any) (*entity.Endpoint, error) {
    ep := &entity.Endpoint{
        CollectionID: collectionID,
        Version:      "v1",
        Method:       "GET",
        IsActive:     true,
    }

    // Map params to endpoint fields
    if name, ok := params["name"].(string); ok {
        ep.Name = name
    }
    if desc, ok := params["description"].(string); ok {
        ep.Description = &desc
    }
    if path, ok := params["path"].(string); ok {
        ep.Path = path
    }
    if method, ok := params["method"].(string); ok {
        ep.Method = method
    }
    if dbType, ok := params["db_type"].(string); ok {
        ep.DBType = entity.EndpointType(dbType)
    }
    if schema, ok := params["schema"].(string); ok {
        ep.Schema = schema
    }
    if tableName, ok := params["table_name"].(string); ok {
        ep.TableName = tableName
    }
    if funcName, ok := params["func_name"].(string); ok {
        ep.FuncName = funcName
    }
    if ops, ok := params["operations"].([]any); ok {
        opsJSON, _ := json.Marshal(ops)
        ep.Operations = opsJSON
    }
    if security, ok := params["security"].(map[string]any); ok {
        secJSON, _ := json.Marshal(security)
        ep.SecurityPolicyJSON = secJSON
    }

    if err := ep.Validate(); err != nil {
        return nil, fmt.Errorf("validate endpoint: %w", err)
    }

    if err := s.repo.Create(ctx, ep); err != nil {
        return nil, fmt.Errorf("create endpoint: %w", err)
    }

    s.logger.Info(ctx, "endpoint created", "endpoint_id", ep.ID, "collection_id", collectionID)
    return ep, nil
}

func (s *EndpointService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Endpoint, error) {
    ep, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get endpoint %s: %w", id, err)
    }
    return ep, nil
}

func (s *EndpointService) Update(ctx context.Context, id uuid.UUID, updates map[string]any) (*entity.Endpoint, error) {
    ep, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }

    if name, ok := updates["name"].(string); ok {
        ep.Name = name
    }
    if path, ok := updates["path"].(string); ok {
        ep.Path = path
    }
    if method, ok := updates["method"].(string); ok {
        ep.Method = method
    }
    if tableName, ok := updates["table_name"].(string); ok {
        ep.TableName = tableName
    }
    if dbType, ok := updates["db_type"].(string); ok {
        ep.DBType = entity.EndpointType(dbType)
    }

    if err := ep.Validate(); err != nil {
        return nil, fmt.Errorf("validate endpoint: %w", err)
    }

    if err := s.repo.Update(ctx, ep); err != nil {
        return nil, fmt.Errorf("update endpoint %s: %w", id, err)
    }

    s.logger.Info(ctx, "endpoint updated", "endpoint_id", id)
    return ep, nil
}

func (s *EndpointService) Delete(ctx context.Context, id uuid.UUID) error {
    if err := s.repo.Delete(ctx, id); err != nil {
        return fmt.Errorf("delete endpoint %s: %w", id, err)
    }
    s.logger.Info(ctx, "endpoint deleted", "endpoint_id", id)
    return nil
}

func (s *EndpointService) List(ctx context.Context, collectionID uuid.UUID) ([]*entity.Endpoint, error) {
    return s.repo.ListByCollection(ctx, collectionID)
}

func (s *EndpointService) ToggleActive(ctx context.Context, id uuid.UUID, active bool) error {
    if err := s.repo.ToggleActive(ctx, id, active); err != nil {
        return fmt.Errorf("toggle endpoint %s: %w", id, err)
    }
    s.logger.Info(ctx, "endpoint toggled", "endpoint_id", id, "active", active)
    return nil
}
```

---

## Step 12.03 — Endpoint HTTP Handler

**Build:** Create `backend/internal/presentation/http/handler/endpoint_handler.go`:

```go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/application/service"
    "github.com/muhammadyunus/ForgeBase/internal/presentation/http/dto"
)

// EndpointHandler handles endpoint HTTP requests.
type EndpointHandler struct {
    endpointService service.EndpointService
}

// NewEndpointHandler creates a new endpoint handler.
func NewEndpointHandler(es service.EndpointService) *EndpointHandler {
    return &EndpointHandler{endpointService: es}
}

// Create handles POST /api/v1/collections/:col_id/endpoints
func (h *EndpointHandler) Create(c *gin.Context) {
    colID, err := uuid.Parse(c.Param("col_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid collection ID", ""))
        return
    }

    var params map[string]any
    if err := c.ShouldBindJSON(&params); err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid request body", err.Error()))
        return
    }
    params["collection_id"] = colID

    ep, err := h.endpointService.Create(c.Request.Context(), colID, params)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
        return
    }

    c.JSON(http.StatusCreated, toEndpointDTO(ep))
}

// List handles GET /api/v1/collections/:col_id/endpoints
func (h *EndpointHandler) List(c *gin.Context) {
    colID, err := uuid.Parse(c.Param("col_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid collection ID", ""))
        return
    }

    eps, err := h.endpointService.List(c.Request.Context(), colID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": toEndpointListDTO(eps)})
}

// GetByID handles GET /api/v1/endpoints/:id
func (h *EndpointHandler) GetByID(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid ID", ""))
        return
    }

    ep, err := h.endpointService.GetByID(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "Endpoint not found", err.Error()))
        return
    }

    c.JSON(http.StatusOK, toEndpointDTO(ep))
}

// Update handles PATCH /api/v1/endpoints/:id
func (h *EndpointHandler) Update(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid ID", ""))
        return
    }

    var updates map[string]any
    if err := c.ShouldBindJSON(&updates); err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid request body", err.Error()))
        return
    }

    ep, err := h.endpointService.Update(c.Request.Context(), id, updates)
    if err != nil {
        c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "Endpoint not found", err.Error()))
        return
    }

    c.JSON(http.StatusOK, toEndpointDTO(ep))
}

// Delete handles DELETE /api/v1/endpoints/:id
func (h *EndpointHandler) Delete(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid ID", ""))
        return
    }

    if err := h.endpointService.Delete(c.Request.Context(), id); err != nil {
        c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "Endpoint not found", err.Error()))
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "endpoint deleted"})
}

// Toggle handles POST /api/v1/endpoints/:id/toggle
func (h *EndpointHandler) Toggle(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid ID", ""))
        return
    }

    var req struct {
        Active bool `json:"active"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid request body", err.Error()))
        return
    }

    if err := h.endpointService.ToggleActive(c.Request.Context(), id, req.Active); err != nil {
        c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "endpoint toggled", "active": req.Active})
}

func toEndpointDTO(ep *entity.Endpoint) gin.H {
    return gin.H{
        "id":            ep.ID.String(),
        "collection_id": ep.CollectionID.String(),
        "name":          ep.Name,
        "description":   ep.Description,
        "path":          ep.Path,
        "method":        ep.Method,
        "version":       ep.Version,
        "is_active":     ep.IsActive,
        "db_type":       string(ep.DBType),
        "schema":        ep.Schema,
        "table_name":    ep.TableName,
        "func_name":     ep.FuncName,
        "params":        ep.Params,
        "operations":    ep.Operations,
        "created_at":    ep.CreatedAt,
        "updated_at":    ep.UpdatedAt,
    }
}

func toEndpointListDTO(eps []*entity.Endpoint) []gin.H {
    result := make([]gin.H, len(eps))
    for i, ep := range eps {
        result[i] = gin.H{
            "id":          ep.ID.String(),
            "name":        ep.Name,
            "path":        ep.Path,
            "method":      ep.Method,
            "version":     ep.Version,
            "is_active":   ep.IsActive,
            "db_type":     string(ep.DBType),
            "created_at":  ep.CreatedAt,
        }
    }
    return result
}
```

---

## Commit Instruction

```bash
git add .
git commit -m "feat: implement endpoint repository, service, and HTTP handlers"
```
