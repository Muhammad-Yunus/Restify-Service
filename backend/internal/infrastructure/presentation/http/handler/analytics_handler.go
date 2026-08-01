package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
	"github.com/muhammadyunus/Restify-Service/internal/domain/service"
	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/presentation/http/dto"
)

// AnalyticsHandler handles analytics HTTP requests.
type AnalyticsHandler struct {
	analyticsService service.AnalyticsService
	logRepo          repository.APILogRepository
}

// NewAnalyticsHandler creates a new analytics handler.
func NewAnalyticsHandler(
	analyticsService service.AnalyticsService,
	logRepo repository.APILogRepository,
) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		logRepo:          logRepo,
	}
}

// Overview handles GET /api/v1/analytics/overview
func (h *AnalyticsHandler) Overview(c *gin.Context) {
	workspaceID, _ := uuid.Parse(c.Param("workspace_id"))

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

	// Default time range to last 7 days if not specified
	if from.IsZero() {
		from = time.Now().AddDate(0, 0, -7)
	}
	if to.IsZero() {
		to = time.Now()
	}

	overview, err := h.analyticsService.GetOverview(c.Request.Context(), workspaceID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, toOverviewDTO(overview))
}

// EndpointMetrics handles GET /api/v1/analytics/endpoints/:endpoint_id/metrics
func (h *AnalyticsHandler) EndpointMetrics(c *gin.Context) {
	endpointID, err := uuid.Parse(c.Param("endpoint_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid endpoint ID", "endpoint ID must be a UUID"))
		return
	}

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

	metrics, err := h.analyticsService.GetEndpointMetrics(c.Request.Context(), endpointID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, toMetricListDTO(metrics))
}

// SearchLogs handles GET /api/v1/analytics/logs
func (h *AnalyticsHandler) SearchLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

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

	workspaceID, _ := uuid.Parse(c.Query("workspace_id"))
	endpointID, _ := uuid.Parse(c.Query("endpoint_id"))

	level := entity.LogLevel(c.Query("level"))
	method := c.Query("method")
	path := c.Query("path")

	logs, total, err := h.logRepo.Search(c.Request.Context(), workspaceID, endpointID, level, method, path, from, to, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       toLogListDTO(logs),
		"pagination": gin.H{"page": page, "page_size": pageSize, "total": total},
	})
}

func toOverviewDTO(overview repository.OverviewMetrics) gin.H {
	return gin.H{
		"total_requests": overview.TotalRequests,
		"avg_latency_ms": overview.AvgLatencyMs,
		"error_rate":     overview.ErrorRate,
	}
}

func toMetricListDTO(metrics []*entity.AnalyticsMetric) gin.H {
	var result []gin.H
	for _, m := range metrics {
		result = append(result, gin.H{
			"id":           m.ID.String(),
			"metric_name":  m.MetricName,
			"metric_value": m.MetricValue,
			"period_start": m.PeriodStart,
			"period_end":   m.PeriodEnd,
		})
	}
	return gin.H{"data": result}
}

func toLogListDTO(logs []*entity.APILog) []gin.H {
	result := make([]gin.H, len(logs))
	for i, log := range logs {
		result[i] = gin.H{
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
	return result
}
