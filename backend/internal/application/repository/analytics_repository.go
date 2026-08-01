package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// AnalyticsRepositoryImpl implements the repository.AnalyticsRepository interface.
type AnalyticsRepositoryImpl struct {
	db *gorm.DB
}

// NewAnalyticsRepository creates a new analytics repository.
func NewAnalyticsRepository(db repository.DB, gormDB *gorm.DB) repository.AnalyticsRepository {
	return &AnalyticsRepositoryImpl{db: gormDB}
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

func (r *AnalyticsRepositoryImpl) GetMetrics(ctx context.Context, workspaceID uuid.UUID,
	metricName string, from, to time.Time, interval time.Duration) ([]*entity.AnalyticsMetric, error) {

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
	var totalRequests struct{ Total float64 }
	if err := r.db.WithContext(ctx).
		Model(&entity.AnalyticsMetric{}).
		Where("workspace_id = ?", workspaceID).
		Where("metric_name = ?", "requests").
		Where("period_start >= ?", from).
		Where("period_end <= ?", to).
		Select("COALESCE(SUM(metric_value), 0) as total").
		Scan(&totalRequests).Error; err != nil {
		return metrics, fmt.Errorf("sum requests: %w", err)
	}
	metrics.TotalRequests = int64(totalRequests.Total)

	// Average latency
	var avgLatency struct{ Latency float64 }
	if err := r.db.WithContext(ctx).
		Model(&entity.AnalyticsMetric{}).
		Where("workspace_id = ?", workspaceID).
		Where("metric_name = ?", "latency").
		Where("period_start >= ?", from).
		Where("period_end <= ?", to).
		Select("COALESCE(AVG(metric_value), 0) as latency").
		Scan(&avgLatency).Error; err != nil {
		return metrics, fmt.Errorf("avg latency: %w", err)
	}
	metrics.AvgLatencyMs = avgLatency.Latency

	// Error rate
	var errorCount struct{ Cnt float64 }
	if err := r.db.WithContext(ctx).
		Model(&entity.AnalyticsMetric{}).
		Where("workspace_id = ?", workspaceID).
		Where("metric_name = ?", "errors").
		Where("period_start >= ?", from).
		Where("period_end <= ?", to).
		Select("COALESCE(SUM(metric_value), 0) as cnt").
		Scan(&errorCount).Error; err != nil {
		return metrics, fmt.Errorf("sum errors: %w", err)
	}

	var totalCount struct{ Cnt float64 }
	if err := r.db.WithContext(ctx).
		Model(&entity.AnalyticsMetric{}).
		Where("workspace_id = ?", workspaceID).
		Where("metric_name = ?", "requests").
		Where("period_start >= ?", from).
		Where("period_end <= ?", to).
		Select("COALESCE(SUM(metric_value), 0) as cnt").
		Scan(&totalCount).Error; err != nil {
		return metrics, fmt.Errorf("sum requests for error rate: %w", err)
	}

	if totalCount.Cnt > 0 {
		metrics.ErrorRate = (errorCount.Cnt / totalCount.Cnt) * 100
	}

	return metrics, nil
}

func (r *AnalyticsRepositoryImpl) AggregateOldMetrics(ctx context.Context, olderThan time.Time) error {
	// This would roll up hourly metrics into daily for storage efficiency — simplified for now
	return nil
}

func (r *AnalyticsRepositoryImpl) Close(ctx context.Context) error {
	return nil
}
