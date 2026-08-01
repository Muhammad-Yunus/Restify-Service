# Epic 18 — API Logging

**Goal:** Implement request/response logging with structured output, log search, and retention policy.
**Dependencies:** Epic 06 (Logger adapter), Epic 05 (APILog repository interface)
**Commit:** `feat: add structured API request/response logging`

---

## Step 18.01 — APILog Repository Implementation

**Build:** Create `backend/internal/application/repository/apilog_repository.go`:

```go
package repository

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
    "gorm.io/gorm"
)

// APILogRepositoryImpl implements the repository.APILogRepository interface.
type APILogRepositoryImpl struct {
    db *gorm.DB
}

// NewAPILogRepository creates a new API log repository.
func NewAPILogRepository(db repository.DB) repository.APILogRepository {
    return &APILogRepositoryImpl{db: gormDB(db)}
}

func (r *APILogRepositoryImpl) Create(ctx context.Context, log *entity.APILog) error {
    if log.ID == uuid.Nil {
        log.ID = uuid.New()
    }
    if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
        return fmt.Errorf("create API log: %w", err)
    }
    return nil
}

func (r *APILogRepositoryImpl) CreateBatch(ctx context.Context, logs []*entity.APILog) error {
    if len(logs) == 0 {
        return nil
    }
    if err := r.db.WithContext(ctx).Create(&logs).Error; err != nil {
        return fmt.Errorf("batch create API logs: %w", err)
    }
    return nil
}

func (r *APILogRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entity.APILog, error) {
    var log entity.APILog
    err := r.db.WithContext(ctx).First(&log, "id = ?", id).Error
    if err != nil {
        return nil, fmt.Errorf("find API log %s: %w", id, err)
    }
    return &log, nil
}

func (r *APILogRepositoryImpl) Search(ctx context.Context, workspaceID, endpointID uuid.UUID,
    level entity.LogLevel, method, path string, from, to time.Time, page, pageSize int) ([]*entity.APILog, int, error) {

    if page < 1 {
        page = 1
    }
    if pageSize < 1 || pageSize > 100 {
        pageSize = 50
    }
    offset := (page - 1) * pageSize

    query := r.db.WithContext(ctx).Model(&entity.APILog{})

    if !workspaceID.IsNil() {
        query = query.Where("workspace_id = ?", workspaceID)
    }
    if !endpointID.IsNil() {
        query = query.Where("endpoint_id = ?", endpointID)
    }
    if level != "" {
        query = query.Where("log_level = ?", level)
    }
    if method != "" {
        query = query.Where("method = ?", method)
    }
    if path != "" {
        query = query.Where("path LIKE ?", "%"+path+"%")
    }
    if !from.IsZero() {
        query = query.Where("created_at >= ?", from)
    }
    if !to.IsZero() {
        query = query.Where("created_at <= ?", to)
    }

    var total int64
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, fmt.Errorf("count API logs: %w", err)
    }

    var logs []*entity.APILog
    if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&logs).Error; err != nil {
        return nil, 0, fmt.Errorf("search API logs: %w", err)
    }

    return logs, int(total), nil
}

func (r *APILogRepositoryImpl) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
    result := r.db.WithContext(ctx).Delete(&entity.APILog{}, "created_at < ?", before)
    if result.Error != nil {
        return 0, fmt.Errorf("delete old logs: %w", result.Error)
    }
    return result.RowsAffected, nil
}

func (r *APILogRepositoryImpl) CountByWorkspace(ctx context.Context, workspaceID uuid.UUID, from, to time.Time) (int64, error) {
    var count int64
    query := r.db.WithContext(ctx).Model(&entity.APILog{}).Where("workspace_id = ?", workspaceID)
    if !from.IsZero() {
        query = query.Where("created_at >= ?", from)
    }
    if !to.IsZero() {
        query = query.Where("created_at <= ?", to)
    }
    if err := query.Count(&count).Error; err != nil {
        return 0, fmt.Errorf("count logs by workspace: %w", err)
    }
    return count, nil
}
```

---

## Step 18.02 — Log Service

**Build:** Create `backend/internal/application/service/log_service.go`:

```go
package service

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// LogService manages API log operations.
type LogService struct {
    repo   repository.APILogRepository
    logger repository.Logger
}

// NewLogService creates a new log service.
func NewLogService(repo repository.APILogRepository, logger repository.Logger) repository.LogService {
    return &LogService{repo: repo, logger: logger}
}

func (s *LogService) RecordRequest(ctx context.Context, log *entity.APILog) error {
    return s.repo.Create(ctx, log)
}

func (s *LogService) Search(ctx context.Context, workspaceID, endpointID uuid.UUID,
    level entity.LogLevel, method, path string, from, to time.Time, page, pageSize int) ([]*entity.APILog, int, error) {

    return s.repo.Search(ctx, workspaceID, endpointID, level, method, path, from, to, page, pageSize)
}

func (s *LogService) GetByID(ctx context.Context, id uuid.UUID) (*entity.APILog, error) {
    return s.repo.FindByID(ctx, id)
}

func (s *LogService) RetainDays(ctx context.Context, days int) (int64, error) {
    cutoff := time.Now().AddDate(0, 0, -days)
    return s.repo.DeleteOlderThan(ctx, cutoff)
}
```

---

## Step 18.03 — Log HTTP Handler

**Build:** Create `backend/internal/presentation/http/handler/log_handler.go`:

```go
package handler

import (
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/application/service"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
    "github.com/muhammadyunus/ForgeBase/internal/presentation/http/dto"
)

// LogHandler handles API log HTTP requests.
type LogHandler struct {
    logService service.AnalyticsService
}

// NewLogHandler creates a new log handler.
func NewLogHandler(ls service.AnalyticsService) *LogHandler {
    return &LogHandler{logService: ls}
}

// Search handles GET /api/v1/logs
func (h *LogHandler) Search(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

    // Parse optional filters
    var from, to time.Time
    if f := c.Query("from"); f != "" {
        if t, err := time.Parse(time.RFC3339, f); err == nil {
            from = t
        }
    }
    if t := c.Query("to"); t != "" {
        if date, err := time.Parse(time.RFC3339, t); err == nil {
            to = date
        }
    }

    logs, total, err := h.logService.Search(c.Request.Context(), uuid.Nil, uuid.Nil, "", "", "", from, to, page, pageSize)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "data":       toLogListDTO(logs),
        "pagination": gin.H{"page": page, "page_size": pageSize, "total": total},
    })
}

// GetByID handles GET /api/v1/logs/:id
func (h *LogHandler) GetByID(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid ID", ""))
        return
    }

    log, err := h.logService.GetByID(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "Log not found", err.Error()))
        return
    }

    c.JSON(http.StatusOK, toLogDTO(log))
}

func toLogDTO(log *entity.APILog) gin.H {
    return gin.H{
        "id":          log.ID.String(),
        "request_id":  log.RequestID,
        "method":      log.Method,
        "path":        log.Path,
        "status_code": log.StatusCode,
        "latency_ms":  log.LatencyMs,
        "log_level":   string(log.LogLevel),
        "message":     log.Message,
        "created_at":  log.CreatedAt,
    }
}

func toLogListDTO(logs []*entity.APILog) []gin.H {
    result := make([]gin.H, len(logs))
    for i, log := range logs {
        result[i] = toLogDTO(log)
    }
    return result
}
```

---

## Step 18.04 — Log Repository Initialization in DI

**Build:** Update `internal/di/bootstrap.go`:

```go
func initLogRepo(db repository.DB) repository.APILogRepository {
    return logging.NewAPILogRepository(db)
}
```

**Test cases:**
- [ ] Unit: `initLogRepo()` returns a non-nil APILogRepository

---

## Commit Instruction

```bash
git add .
git commit -m "feat: add API request/response logging with search and retention"
```
