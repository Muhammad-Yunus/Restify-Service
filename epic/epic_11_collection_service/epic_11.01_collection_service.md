# Epic 11 — Collection Service & Repository

**Goal:** Implement collection repository, service, and HTTP handlers.
**Dependencies:** Epic 04 (Collection entity), Epic 05 (Repository interfaces), Epic 06 (DB adapter), Epic 09 (Workspace service)
**Commit:** `feat: implement collection repository, service, and HTTP handlers`

---

## Step 11.01 — Collection Repository Implementation

**Build:** Create `backend/internal/application/repository/collection_repository.go`:

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

// CollectionRepositoryImpl implements the repository.CollectionRepository interface.
type CollectionRepositoryImpl struct {
    db *gorm.DB
}

// NewCollectionRepository creates a new collection repository.
func NewCollectionRepository(db repository.DB) repository.CollectionRepository {
    return &CollectionRepositoryImpl{db: gormDB(db)}
}

func (r *CollectionRepositoryImpl) Create(ctx context.Context, col *entity.Collection) error {
    if col.ID == uuid.Nil {
        col.ID = uuid.New()
    }
    col.GenerateSlug()
    if err := r.db.WithContext(ctx).Create(col).Error; err != nil {
        return fmt.Errorf("create collection: %w", err)
    }
    return nil
}

func (r *CollectionRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entity.Collection, error) {
    var col entity.Collection
    err := r.db.WithContext(ctx).Preload("Endpoints").First(&col, "id = ?", id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, entity.ErrNotFound
        }
        return nil, fmt.Errorf("find collection %s: %w", id, err)
    }
    return &col, nil
}

func (r *CollectionRepositoryImpl) FindBySlug(ctx context.Context, workspaceID uuid.UUID, slug string) (*entity.Collection, error) {
    var col entity.Collection
    err := r.db.WithContext(ctx).
        Preload("Endpoints").
        Where("workspace_id = ? AND slug = ?", workspaceID, slug).
        First(&col).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, entity.ErrNotFound
        }
        return nil, fmt.Errorf("find collection by slug: %w", err)
    }
    return &col, nil
}

func (r *CollectionRepositoryImpl) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entity.Collection, error) {
    var cols []*entity.Collection
    if err := r.db.WithContext(ctx).
        Preload("Endpoints").
        Where("workspace_id = ?", workspaceID).
        Find(&cols).Error; err != nil {
        return nil, fmt.Errorf("list collections for workspace %s: %w", workspaceID, err)
    }
    return cols, nil
}

func (r *CollectionRepositoryImpl) Update(ctx context.Context, col *entity.Collection) error {
    if err := r.db.WithContext(ctx).Model(col).Updates(map[string]any{
        "name":        col.Name,
        "description": col.Description,
    }).Error; err != nil {
        return fmt.Errorf("update collection %s: %w", col.ID, err)
    }
    return nil
}

func (r *CollectionRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
    // Delete all endpoints first (cascade)
    if err := r.db.WithContext(ctx).Delete(&entity.Endpoint{}, "collection_id = ?", id).Error; err != nil {
        return fmt.Errorf("delete endpoints: %w", err)
    }
    result := r.db.WithContext(ctx).Delete(&entity.Collection{}, "id = ?", id)
    if result.Error != nil {
        return fmt.Errorf("delete collection %s: %w", id, result.Error)
    }
    if result.RowsAffected == 0 {
        return entity.ErrNotFound
    }
    return nil
}

func (r *CollectionRepositoryImpl) CountByWorkspace(ctx context.Context, workspaceID uuid.UUID) (int, error) {
    var count int64
    if err := r.db.WithContext(ctx).
        Model(&entity.Collection{}).
        Where("workspace_id = ?", workspaceID).
        Count(&count).Error; err != nil {
        return 0, fmt.Errorf("count collections: %w", err)
    }
    return int(count), nil
}
```

**Test cases:**
- [ ] Unit: `Create()` creates collection with generated slug
- [ ] Unit: `FindByID()` returns collection with endpoints
- [ ] Unit: `FindBySlug()` finds by workspace + slug
- [ ] Unit: `Delete()` cascades to endpoints
- [ ] Integration: Full collection lifecycle

---

## Step 11.02 — Collection Service

**Build:** Create `backend/internal/application/service/collection_service.go`:

```go
package service

import (
    "context"
    "fmt"

    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// CollectionService implements the service.CollectionService interface.
type CollectionService struct {
    repo   repository.CollectionRepository
    logger repository.Logger
}

// NewCollectionService creates a new collection service.
func NewCollectionService(repo repository.CollectionRepository, logger repository.Logger) repository.CollectionService {
    return &CollectionService{repo: repo, logger: logger}
}

func (s *CollectionService) Create(ctx context.Context, name, description string, workspaceID uuid.UUID) (*entity.Collection, error) {
    col := &entity.Collection{
        Name:        name,
        Description: &description,
        WorkspaceID: workspaceID,
    }

    if err := col.Validate(); err != nil {
        return nil, fmt.Errorf("validate collection: %w", err)
    }

    if err := s.repo.Create(ctx, col); err != nil {
        return nil, fmt.Errorf("create collection: %w", err)
    }

    s.logger.Info(ctx, "collection created", "collection_id", col.ID, "workspace_id", workspaceID)
    return col, nil
}

func (s *CollectionService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Collection, error) {
    col, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get collection %s: %w", id, err)
    }
    return col, nil
}

func (s *CollectionService) Update(ctx context.Context, id uuid.UUID, updates map[string]any) (*entity.Collection, error) {
    col, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }

    if name, ok := updates["name"].(string); ok && name != "" {
        col.Name = name
    }
    if desc, ok := updates["description"].(string); ok {
        col.Description = &desc
    }

    if err := col.Validate(); err != nil {
        return nil, fmt.Errorf("validate collection: %w", err)
    }

    if err := s.repo.Update(ctx, col); err != nil {
        return nil, fmt.Errorf("update collection %s: %w", id, err)
    }

    s.logger.Info(ctx, "collection updated", "collection_id", id)
    return col, nil
}

func (s *CollectionService) Delete(ctx context.Context, id uuid.UUID) error {
    if err := s.repo.Delete(ctx, id); err != nil {
        return fmt.Errorf("delete collection %s: %w", id, err)
    }
    s.logger.Info(ctx, "collection deleted", "collection_id", id)
    return nil
}

func (s *CollectionService) List(ctx context.Context, workspaceID uuid.UUID) ([]*entity.Collection, error) {
    return s.repo.ListByWorkspace(ctx, workspaceID)
}
```

---

## Step 11.03 — Collection HTTP Handler

**Build:** Create `backend/internal/presentation/http/handler/collection_handler.go`:

```go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/application/service"
    "github.com/muhammadyunus/ForgeBase/internal/presentation/http/dto"
)

// CollectionHandler handles collection HTTP requests.
type CollectionHandler struct {
    collectionService service.CollectionService
}

// NewCollectionHandler creates a new collection handler.
func NewCollectionHandler(cs service.CollectionService) *CollectionHandler {
    return &CollectionHandler{collectionService: cs}
}

// Create handles POST /api/v1/workspaces/:ws_id/collections
func (h *CollectionHandler) Create(c *gin.Context) {
    wsID, err := uuid.Parse(c.Param("ws_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid workspace ID", ""))
        return
    }

    var req struct {
        Name        string `json:"name" validate:"required,max=255"`
        Description string `json:"description,omitempty"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Validation Error", err.Error()))
        return
    }

    col, err := h.collectionService.Create(c.Request.Context(), req.Name, req.Description, wsID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
        return
    }

    c.JSON(http.StatusCreated, toCollectionDTO(col))
}

// List handles GET /api/v1/workspaces/:ws_id/collections
func (h *CollectionHandler) List(c *gin.Context) {
    wsID, err := uuid.Parse(c.Param("ws_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid workspace ID", ""))
        return
    }

    cols, err := h.collectionService.List(c.Request.Context(), wsID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": toCollectionListDTO(cols)})
}

// GetByID handles GET /api/v1/collections/:id
func (h *CollectionHandler) GetByID(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid ID", ""))
        return
    }

    col, err := h.collectionService.GetByID(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "Collection not found", err.Error()))
        return
    }

    c.JSON(http.StatusOK, toCollectionDTO(col))
}

// Update handles PATCH /api/v1/collections/:id
func (h *CollectionHandler) Update(c *gin.Context) {
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

    col, err := h.collectionService.Update(c.Request.Context(), id, updates)
    if err != nil {
        c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "Collection not found", err.Error()))
        return
    }

    c.JSON(http.StatusOK, toCollectionDTO(col))
}

// Delete handles DELETE /api/v1/collections/:id
func (h *CollectionHandler) Delete(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid ID", ""))
        return
    }

    if err := h.collectionService.Delete(c.Request.Context(), id); err != nil {
        c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "Collection not found", err.Error()))
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "collection deleted"})
}

func toCollectionDTO(col *entity.Collection) gin.H {
    endpoints := make([]gin.H, len(col.Endpoints))
    for i, ep := range col.Endpoints {
        endpoints[i] = gin.H{
            "id":       ep.ID.String(),
            "name":     ep.Name,
            "path":     ep.Path,
            "method":   ep.Method,
            "version":  ep.Version,
            "is_active": ep.IsActive,
        }
    }
    return gin.H{
        "id":           col.ID.String(),
        "name":         col.Name,
        "description":  col.Description,
        "slug":         col.Slug,
        "workspace_id": col.WorkspaceID.String(),
        "endpoints":    endpoints,
        "created_at":   col.CreatedAt,
        "updated_at":   col.UpdatedAt,
    }
}

func toCollectionListDTO(cols []*entity.Collection) []gin.H {
    result := make([]gin.H, len(cols))
    for i, col := range cols {
        result[i] = gin.H{
            "id":           col.ID.String(),
            "name":         col.Name,
            "slug":         col.Slug,
            "workspace_id": col.WorkspaceID.String(),
            "endpoint_count": len(col.Endpoints),
            "created_at":   col.CreatedAt,
        }
    }
    return result
}
```

---

## Commit Instruction

```bash
git add .
git commit -m "feat: implement collection repository, service, and HTTP handlers"
```
