package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
)

// AnalyticsRepository defines the data access contract for API analytics.
type AnalyticsRepository interface {
	// RecordMetric stores an aggregated metric.
	RecordMetric(ctx context.Context, metric *entity.AnalyticsMetric) error

	// RecordMetricsBatch stores multiple metrics.
	RecordMetricsBatch(ctx context.Context, metrics []*entity.AnalyticsMetric) error

	// GetMetrics returns metrics for a workspace in a time range.
	GetMetrics(ctx context.Context, workspaceID uuid.UUID, metricName string,
		from, to time.Time, interval time.Duration) ([]*entity.AnalyticsMetric, error)

	// GetEndpointMetrics returns metrics for a specific endpoint.
	GetEndpointMetrics(ctx context.Context, endpointID uuid.UUID,
		from, to time.Time) ([]*entity.AnalyticsMetric, error)

	// GetOverview returns summary metrics for dashboard.
	GetOverview(ctx context.Context, workspaceID uuid.UUID, from, to time.Time) (OverviewMetrics, error)

	// AggregateOldMetrics rolls up hourly metrics into daily for storage efficiency.
	AggregateOldMetrics(ctx context.Context, olderThan time.Time) error

	// Close shuts down any resources owned by the repository.
	Close(ctx context.Context) error
}

// OverviewMetrics holds dashboard summary data.
type OverviewMetrics struct {
	TotalRequests  int64          `json:"total_requests"`
	AvgLatencyMs   float64        `json:"avg_latency_ms"`
	ErrorRate      float64        `json:"error_rate"` // percentage
	TopEndpoints   []TopEndpoint  `json:"top_endpoints"`
	RequestsByHour []HourlyMetric `json:"requests_by_hour"`
}

// TopEndpoint summarizes traffic for a single endpoint.
type TopEndpoint struct {
	Path      string  `json:"path"`
	Requests  int64   `json:"requests"`
	ErrorRate float64 `json:"error_rate"`
}

// HourlyMetric counts requests and errors for one hour.
type HourlyMetric struct {
	Hour     time.Time `json:"hour"`
	Requests int64     `json:"requests"`
	Errors   int64     `json:"errors"`
}
