# Epic 19 — API Analytics

**Goal:** Implement API analytics — request metrics, error rates, latency tracking, and dashboard data.
**Dependencies:** Epic 05 (Analytics repository interface), Epic 18 (Log repository)
**Commit:** `feat: add API analytics with metrics aggregation`

---

## Step 19.01 — Analytics Repository Implementation

**Build:** Create `backend/internal/application/repository/analytics_repository.go`:

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

// AnalyticsRepositoryImpl implements the repository.AnalyticsRepository interface.
type AnalyticsRepositoryImpl struct {
    db *gorm.DB
}

// NewAnalyticsRepository creates a new analytics repository.
func NewAnalyticsRepository(db repository.DB) repository.AnalyticsRepository {
    return &AnalyticsRepositoryImpl{db: gormDB(db)}
}

func (r *AnalyticsRepositoryImpl) RecordMetric(ctx context.Context, metric *entity.AnalyticsMetric) error {
    if metric.ID == uuid.Nil {
        metric.ID = uuid.New()
    }
    if err := r.db.WithContext(ctx).Create(metric).Error; err != nil {
        return fmt.Errorf("record analytics metric: %w", err)
    }
    return nil
}

func (r *AnalyticsRepositoryImpl) RecordMetricsBatch(ctx context.Context, metrics []*entity.AnalyticsMetric) error {
    if len(metrics) == 0 {
        return nil
    }
    if err := r.db.WithContext(ctx).Create(&metrics).Error; err != nil {
        return fmt.Errorf("batch record analytics metrics: %w", err)
    }
    return nil
}

func (r *AnalyticsRepositoryImpl) GetMetrics(ctx context.Context, workspaceID uuid.UUID, metricName string,
    from, to time.Time, interval time.Duration) ([]*entity.AnalyticsMetric, error) {

    var metrics []*entity.AnalyticsMetric
    query := r.db.WithContext(ctx).Model(&entity.AnalyticsMetric{}).
        Where("workspace_id = ?", workspaceID).
        Where("metric_name = ?", metricName).
        Where("period_start >= ?", from).
        Where("period_end <= ?", to).
        Order("period_start ASC")

    if err := query.Find(&metrics).Error; err != nil {
        return nil, fmt.Errorf("get analytics metrics: %w", err)
    }
    return metrics, nil
}

func (r *AnalyticsRepositoryImpl) GetEndpointMetrics(ctx context.Context, endpointID uuid.UUID, from, to time.Time) ([]*entity.AnalyticsMetric, error) {
    var metrics []*entity.AnalyticsMetric
    if err := r.db.WithContext(ctx).
        Where("endpoint_id = ?", endpointID).
        Where("period_start >= ?", from).
        Where("period_end <= ?", to).
        Order("period_start ASC").
        Find(&metrics).Error; err != nil {
        return nil, fmt.Errorf("get endpoint metrics: %w", err)
    }
    return metrics, nil
}

func (r *AnalyticsRepositoryImpl) GetOverview(ctx context.Context, workspaceID uuid.UUID, from, to time.Time) (repository.OverviewMetrics, error) {
    var metrics repository.OverviewMetrics

    // Total requests
    if err := r.db.WithContext(ctx).
        Model(&entity.AnalyticsMetric{}).
        Where("workspace_id = ?", workspaceID).
        Where("metric_name = 'requests'").
        Where("period_start >= ?", from).
        Where("period_end <= ?", to).
        Sum("metric_value", &metrics.TotalRequests).Error; err != nil {
        return metrics, fmt.Errorf("sum requests: %w", err)
    }

    // Average latency
    var avgLatency float64
    if err := r.db.WithContext(ctx).
        Model(&entity.AnalyticsMetric{}).
        Where("workspace_id = ?", workspaceID).
        Where("metric_name = 'avg_latency'").
        Where("period_start >= ?", from).
        Where("period_end <= ?", to).
        Avg("metric_value", &avgLatency).Error; err != nil {
        return metrics, fmt.Errorf("avg latency: %w", err)
    }
    metrics.AvgLatencyMs = avgLatency

    // Error rate
    var errorCount, totalCount float64
    r.db.WithContext(ctx).
        Model(&entity.AnalyticsMetric{}).
        Where("workspace_id = ?", workspaceID).
        Where("metric_name = 'errors'").
        Where("period_start >= ?", from).
        Where("period_end <= ?", to).
        Sum("metric_value", &errorCount).Error
    r.db.WithContext(ctx).
        Model(&entity.AnalyticsMetric{}).
        Where("workspace_id = ?", workspaceID).
        Where("metric_name = 'requests'").
        Where("period_start >= ?", from).
        Where("period_end <= ?", to).
        Sum("metric_value", &totalCount).Error

    if totalCount > 0 {
        metrics.ErrorRate = (errorCount / totalCount) * 100
    }

    return metrics, nil
}

func (r *AnalyticsRepositoryImpl) AggregateOldMetrics(ctx context.Context, olderThan time.Time) error {
    // This would roll up hourly metrics into daily for storage efficiency
    // Simplified for now
    return nil
}
```

---

## Step 19.02 — Analytics Service

**Build:** Create `backend/internal/application/service/analytics_service.go`:

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

// AnalyticsService manages API analytics.
type AnalyticsService struct {
    logRepo   repository.APILogRepository
    metricRepo repository.AnalyticsRepository
    logger    repository.Logger
}

// NewAnalyticsService creates a new analytics service.
func NewAnalyticsService(logRepo repository.APILogRepository, metricRepo repository.AnalyticsRepository, logger repository.Logger) repository.AnalyticsService {
    return &AnalyticsService{
        logRepo:    logRepo,
        metricRepo: metricRepo,
        logger:     logger,
    }
}

func (s *AnalyticsService) RecordRequest(ctx context.Context, log *entity.APILog) error {
    if err := s.logRepo.Create(ctx, log); err != nil {
        return fmt.Errorf("record log: %w", err)
    }

    // Aggregate metrics asynchronously
    go s.aggregateMetrics(ctx, log)
    return nil
}

func (s *AnalyticsService) aggregateMetrics(ctx context.Context, log *entity.APILog) {
    // Handle nil WorkspaceID safely
    var workspaceID uuid.UUID
    if log.WorkspaceID != nil {
        workspaceID = *log.WorkspaceID
    }

    periodStart := log.CreatedAt.Truncate(time.Hour)
    periodEnd := log.CreatedAt.Add(time.Hour).Truncate(time.Hour)

    // Aggregate request count
    metric := &entity.AnalyticsMetric{
        WorkspaceID: workspaceID,
        MetricName:  "requests",
        MetricValue: 1,
        PeriodStart: periodStart,
        PeriodEnd:   periodEnd,
    }
    if err := s.metricRepo.RecordMetric(ctx, metric); err != nil {
        s.logger.Error(ctx, "failed to record metric", "error", err)
    }

    // Aggregate latency
    latencyMetric := &entity.AnalyticsMetric{
        WorkspaceID: workspaceID,
        MetricName:  "latency",
        MetricValue: float64(log.LatencyMs),
        PeriodStart: periodStart,
        PeriodEnd:   periodEnd,
    }
    if err := s.metricRepo.RecordMetric(ctx, latencyMetric); err != nil {
        s.logger.Error(ctx, "failed to record latency metric", "error", err)
    }

    // Aggregate errors
    if log.StatusCode >= 400 {
        errorMetric := &entity.AnalyticsMetric{
            WorkspaceID: workspaceID,
            MetricName:  "errors",
            MetricValue: 1,
            PeriodStart: periodStart,
            PeriodEnd:   periodEnd,
        }
        if err := s.metricRepo.RecordMetric(ctx, errorMetric); err != nil {
            s.logger.Error(ctx, "failed to record error metric", "error", err)
        }
    }
}

func (s *AnalyticsService) GetOverview(ctx context.Context, workspaceID uuid.UUID, from, to time.Time) (repository.OverviewMetrics, error) {
    return s.metricRepo.GetOverview(ctx, workspaceID, from, to)
}

func (s *AnalyticsService) GetEndpointMetrics(ctx context.Context, endpointID uuid.UUID, from, to time.Time) ([]*entity.AnalyticsMetric, error) {
    return s.metricRepo.GetEndpointMetrics(ctx, endpointID, from, to)
}
```

---

## Step 19.03 — Analytics HTTP Handler

**Build:** Create `backend/internal/presentation/http/handler/analytics_handler.go`:

```go
package handler

import (
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/application/service"
    "github.com/muhammadyunus/ForgeBase/internal/presentation/http/dto"
)

// AnalyticsHandler handles analytics HTTP requests.
type AnalyticsHandler struct {
    analyticsService service.AnalyticsService
}

// NewAnalyticsHandler creates a new analytics handler.
func NewAnalyticsHandler(as service.AnalyticsService) *AnalyticsHandler {
    return &AnalyticsHandler{analyticsService: as}
}

// GetOverview handles GET /api/v1/analytics/overview
func (h *AnalyticsHandler) GetOverview(c *gin.Context) {
    wsIDStr := c.Param("ws_id")
    wsID, _ := uuid.Parse(wsIDStr)

    days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
    to := time.Now()
    from := to.AddDate(0, 0, -days)

    metrics, err := h.analyticsService.GetOverview(c.Request.Context(), wsID, from, to)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": metrics})
}

// GetEndpointMetrics handles GET /api/v1/analytics/endpoints/:id
func (h *AnalyticsHandler) GetEndpointMetrics(c *gin.Context) {
    epID, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid endpoint ID", ""))
        return
    }

    days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
    to := time.Now()
    from := to.AddDate(0, 0, -days)

    metrics, err := h.analyticsService.GetEndpointMetrics(c.Request.Context(), epID, from, to)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": metrics})
}
```

---

## Step 19.04 — Analytics Repository Initialization in DI

**Build:** Update `internal/di/bootstrap.go`:

```go
func initAnalyticsRepo(db repository.DB) repository.AnalyticsRepository {
    return analytics.NewAnalyticsRepository(db)
}
```

**Test cases:**
- [ ] Unit: `initAnalyticsRepo()` returns a non-nil AnalyticsRepository

---

## Commit Instruction

```bash
git add .
git commit -m "feat: add API analytics with metrics aggregation and overview"
```
