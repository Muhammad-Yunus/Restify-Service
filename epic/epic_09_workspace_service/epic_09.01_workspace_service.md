# Epic 09 — Workspace Service & Repository

**Goal:** Implement workspace repository and service with full CRUD, plus slug generation.
**Dependencies:** Epic 04 (Workspace entity), Epic 05 (Repository interfaces), Epic 06 (DB adapter)
**Commit:** `feat: implement workspace repository and service`

---

## Step 09.01 — Workspace Repository Implementation

**Build:** Create `backend/internal/application/repository/workspace_repository.go`:

```go
package repository

import (
    "context"
    "errors"
    "fmt"
    "strings"
    "regexp"

    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
    "gorm.io/gorm"
)

// WorkspaceRepositoryImpl implements the repository.WorkspaceRepository interface.
type WorkspaceRepositoryImpl struct {
    db *gorm.DB
}

// NewWorkspaceRepository creates a new workspace repository.
func NewWorkspaceRepository(db repository.DB) repository.WorkspaceRepository {
    return &WorkspaceRepositoryImpl{db: gormDB(db)}
}

func (r *WorkspaceRepositoryImpl) Create(ctx context.Context, ws *entity.Workspace) error {
    if ws.ID == uuid.Nil {
        ws.ID = uuid.New()
    }
    ws.GenerateSlug()
    if err := r.db.WithContext(ctx).Create(ws).Error; err != nil {
        return fmt.Errorf("create workspace: %w", err)
    }
    return nil
}

func (r *WorkspaceRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entity.Workspace, error) {
    var ws entity.Workspace
    err := r.db.WithContext(ctx).Preload("Owner").First(&ws, "id = ?", id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, entity.ErrNotFound
        }
        return nil, fmt.Errorf("find workspace %s: %w", id, err)
    }
    return &ws, nil
}

func (r *WorkspaceRepositoryImpl) FindBySlug(ctx context.Context, slug string) (*entity.Workspace, error) {
    var ws entity.Workspace
    err := r.db.WithContext(ctx).Preload("Owner").First(&ws, "slug = ?", slug).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, entity.ErrNotFound
        }
        return nil, fmt.Errorf("find workspace by slug %s: %w", slug, err)
    }
    return &ws, nil
}

func (r *WorkspaceRepositoryImpl) Update(ctx context.Context, ws *entity.Workspace) error {
    if err := r.db.WithContext(ctx).Model(ws).Updates(map[string]any{
        "name":        ws.Name,
        "description": ws.Description,
        "is_public":   ws.IsPublic,
    }).Error; err != nil {
        return fmt.Errorf("update workspace %s: %w", ws.ID, err)
    }
    return nil
}

func (r *WorkspaceRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
    result := r.db.WithContext(ctx).Delete(&entity.Workspace{}, "id = ?", id)
    if result.Error != nil {
        return fmt.Errorf("delete workspace %s: %w", id, result.Error)
    }
    if result.RowsAffected == 0 {
        return entity.ErrNotFound
    }
    return nil
}

func (r *WorkspaceRepositoryImpl) List(ctx context.Context, ownerID uuid.UUID, page, pageSize int) ([]*entity.Workspace, int, error) {
    if page < 1 {
        page = 1
    }
    if pageSize < 1 || pageSize > 100 {
        pageSize = 20
    }
    offset := (page - 1) * pageSize

    var workspaces []*entity.Workspace
    var total int64

    if err := r.db.WithContext(ctx).Model(&entity.Workspace{}).
        Where("owner_id = ?", ownerID).Count(&total).Error; err != nil {
        return nil, 0, fmt.Errorf("count workspaces: %w", err)
    }

    if err := r.db.WithContext(ctx).
        Where("owner_id = ?", ownerID).
        Offset(offset).Limit(pageSize).
        Find(&workspaces).Error; err != nil {
        return nil, 0, fmt.Errorf("list workspaces: %w", err)
    }

    return workspaces, int(total), nil
}

func (r *WorkspaceRepositoryImpl) ListAll(ctx context.Context, page, pageSize int) ([]*entity.Workspace, int, error) {
    // Admin-only: list all workspaces
    if page < 1 {
        page = 1
    }
    if pageSize < 1 || pageSize > 100 {
        pageSize = 20
    }
    offset := (page - 1) * pageSize

    var workspaces []*entity.Workspace
    var total int64

    if err := r.db.WithContext(ctx).Model(&entity.Workspace{}).Count(&total).Error; err != nil {
        return nil, 0, fmt.Errorf("count all workspaces: %w", err)
    }

    if err := r.db.WithContext(ctx).
        Offset(offset).Limit(pageSize).
        Find(&workspaces).Error; err != nil {
        return nil, 0, fmt.Errorf("list all workspaces: %w", err)
    }

    return workspaces, int(total), nil
}

func (r *WorkspaceRepositoryImpl) CountByOwner(ctx context.Context, ownerID uuid.UUID) (int, error) {
    var count int64
    if err := r.db.WithContext(ctx).
        Model(&entity.Workspace{}).
        Where("owner_id = ?", ownerID).
        Count(&count).Error; err != nil {
        return 0, fmt.Errorf("count workspaces by owner: %w", err)
    }
    return int(count), nil
}
```

---

## Step 09.02 — Workspace Service

**Build:** Create `backend/internal/application/service/workspace_service.go`:

```go
package service

import (
    "context"
    "fmt"

    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// WorkspaceService implements the service.WorkspaceService interface.
type WorkspaceService struct {
    repo   repository.WorkspaceRepository
    logger repository.Logger
}

// NewWorkspaceService creates a new workspace service.
func NewWorkspaceService(repo repository.WorkspaceRepository, logger repository.Logger) repository.WorkspaceService {
    return &WorkspaceService{repo: repo, logger: logger}
}

func (s *WorkspaceService) Create(ctx context.Context, name, description string, ownerID uuid.UUID) (*entity.Workspace, error) {
    ws := &entity.Workspace{
        Name:        name,
        Description: &description,
        OwnerID:     ownerID,
        IsPublic:    false,
    }

    if err := ws.Validate(); err != nil {
        return nil, fmt.Errorf("validate workspace: %w", err)
    }

    if err := s.repo.Create(ctx, ws); err != nil {
        return nil, fmt.Errorf("create workspace: %w", err)
    }

    s.logger.Info(ctx, "workspace created", "workspace_id", ws.ID, "owner_id", ownerID)
    return ws, nil
}

func (s *WorkspaceService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Workspace, error) {
    ws, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get workspace %s: %w", id, err)
    }
    return ws, nil
}

func (s *WorkspaceService) Update(ctx context.Context, id uuid.UUID, updates map[string]any) (*entity.Workspace, error) {
    ws, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }

    if name, ok := updates["name"].(string); ok && name != "" {
        ws.Name = name
    }
    if desc, ok := updates["description"].(string); ok {
        ws.Description = &desc
    }
    if pub, ok := updates["is_public"].(bool); ok {
        ws.IsPublic = pub
    }

    if err := ws.Validate(); err != nil {
        return nil, fmt.Errorf("validate workspace: %w", err)
    }

    if err := s.repo.Update(ctx, ws); err != nil {
        return nil, fmt.Errorf("update workspace %s: %w", id, err)
    }

    s.logger.Info(ctx, "workspace updated", "workspace_id", id)
    return ws, nil
}

func (s *WorkspaceService) Delete(ctx context.Context, id uuid.UUID) error {
    // Check if workspace has collections or endpoints
    if err := s.repo.Delete(ctx, id); err != nil {
        return fmt.Errorf("delete workspace %s: %w", id, err)
    }
    s.logger.Info(ctx, "workspace deleted", "workspace_id", id)
    return nil
}

func (s *WorkspaceService) List(ctx context.Context, ownerID uuid.UUID, page, pageSize int) ([]*entity.Workspace, int, error) {
    return s.repo.List(ctx, ownerID, page, pageSize)
}
```

**Test cases:**
- [ ] Unit: `Create()` creates workspace with generated slug
- [ ] Unit: `Create()` validates name is not empty
- [ ] Unit: `GetByID()` returns workspace
- [ ] Unit: `Update()` updates name and description
- [ ] Unit: `Delete()` removes workspace
- [ ] Unit: `List()` returns paginated workspaces

---

## Step 09.03 — Workspace HTTP Handler

**Build:** Create `backend/internal/presentation/http/handler/workspace_handler.go`:

```go
package handler

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/application/service"
    "github.com/muhammadyunus/ForgeBase/internal/presentation/http/dto"
)

// WorkspaceHandler handles workspace HTTP requests.
type WorkspaceHandler struct {
    workspaceService service.WorkspaceService
}

// NewWorkspaceHandler creates a new workspace handler.
func NewWorkspaceHandler(ws service.WorkspaceService) *WorkspaceHandler {
    return &WorkspaceHandler{workspaceService: ws}
}

// Create handles POST /api/v1/workspaces
func (h *WorkspaceHandler) Create(c *gin.Context) {
    var req struct {
        Name        string `json:"name" validate:"required,max=255"`
        Description string `json:"description,omitempty"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Validation Error", err.Error()))
        return
    }

    ownerIDStr, ok := c.Get("user_id")
    if !ok {
        c.JSON(http.StatusUnauthorized, dto.ProblemDetail(http.StatusUnauthorized, "Unauthorized", "user not found in context"))
        return
    }
    ownerID, err := uuid.Parse(ownerIDStr.(string))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid user ID", err.Error()))
        return
    }

    ws, err := h.workspaceService.Create(c.Request.Context(), req.Name, req.Description, ownerID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
        return
    }

    c.JSON(http.StatusCreated, toWorkspaceDTO(ws))
}

// List handles GET /api/v1/workspaces
func (h *WorkspaceHandler) List(c *gin.Context) {
    ownerIDStr, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, dto.ProblemDetail(http.StatusUnauthorized, "Unauthorized", ""))
        return
    }
    ownerID, _ := uuid.Parse(ownerIDStr.(string))

    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

    workspaces, total, err := h.workspaceService.List(c.Request.Context(), ownerID, page, pageSize)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "data":       toWorkspaceListDTO(workspaces),
        "pagination": gin.H{"page": page, "page_size": pageSize, "total": total},
    })
}

// GetByID handles GET /api/v1/workspaces/:id
func (h *WorkspaceHandler) GetByID(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid ID", "workspace ID must be a UUID"))
        return
    }

    ws, err := h.workspaceService.GetByID(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "Workspace not found", err.Error()))
        return
    }

    c.JSON(http.StatusOK, toWorkspaceDTO(ws))
}

// Update handles PATCH /api/v1/workspaces/:id
func (h *WorkspaceHandler) Update(c *gin.Context) {
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

    ws, err := h.workspaceService.Update(c.Request.Context(), id, updates)
    if err != nil {
        c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "Workspace not found", err.Error()))
        return
    }

    c.JSON(http.StatusOK, toWorkspaceDTO(ws))
}

// Delete handles DELETE /api/v1/workspaces/:id
func (h *WorkspaceHandler) Delete(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid ID", ""))
        return
    }

    if err := h.workspaceService.Delete(c.Request.Context(), id); err != nil {
        c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "Workspace not found", err.Error()))
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "workspace deleted"})
}

func toWorkspaceDTO(ws *entity.Workspace) gin.H {
    return gin.H{
        "id":          ws.ID.String(),
        "name":        ws.Name,
        "description": ws.Description,
        "slug":        ws.Slug,
        "owner_id":    ws.OwnerID.String(),
        "is_public":   ws.IsPublic,
        "created_at":  ws.CreatedAt,
        "updated_at":  ws.UpdatedAt,
    }
}

func toWorkspaceListDTO(workspaces []*entity.Workspace) []gin.H {
    result := make([]gin.H, len(workspaces))
    for i, ws := range workspaces {
        result[i] = toWorkspaceDTO(ws)
    }
    return result
}
```

---

## Step 09.04 — Register Workspace Routes

**Build:** Update router to include workspace routes:

```go
func RegisterWorkspaceRoutes(rg *gin.RouterGroup, wsHandler *handler.WorkspaceHandler, authMiddleware *middleware.AuthMiddleware) {
    workspaces := rg.Group("/workspaces")
    {
        workspaces.GET("", authMiddleware.RequireAuth(), wsHandler.List)
        workspaces.POST("", authMiddleware.RequireAuth(), wsHandler.Create)
        workspaces.GET("/:id", authMiddleware.RequireAuth(), wsHandler.GetByID)
        workspaces.PATCH("/:id", authMiddleware.RequireAuth(), wsHandler.Update)
        workspaces.DELETE("/:id", authMiddleware.RequireRole("administrator"), wsHandler.Delete)
    }
}
```

---

## Commit Instruction

```bash
git add .
git commit -m "feat: implement workspace repository, service, and HTTP handlers"
```
