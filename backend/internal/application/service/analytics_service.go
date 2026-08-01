package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
	"github.com/muhammadyunus/Restify-Service/internal/domain/service"
)

// AnalyticsService manages API analytics.
type AnalyticsService struct {
	logRepo    repository.APILogRepository
	metricRepo repository.AnalyticsRepository
	logger     repository.Logger
}

// NewAnalyticsService creates a new analytics service.
func NewAnalyticsService(logRepo repository.APILogRepository, metricRepo repository.AnalyticsRepository, logger repository.Logger) service.AnalyticsService {
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

	go s.aggregateMetrics(ctx, log)
	return nil
}

func (s *AnalyticsService) aggregateMetrics(ctx context.Context, log *entity.APILog) {
	var workspaceID uuid.UUID
	if log.WorkspaceID != nil {
		workspaceID = *log.WorkspaceID
	}

	periodStart := log.CreatedAt.Truncate(time.Hour)
	periodEnd := periodStart.Add(time.Hour)

	// Request count
	requestMetric := &entity.AnalyticsMetric{
		WorkspaceID: workspaceID,
		MetricName:  "requests",
		MetricValue: 1,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	}
	if err := s.metricRepo.RecordMetric(ctx, requestMetric); err != nil {
		s.logger.Error(ctx, "failed to record metric", "error", err)
	}

	// Latency
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

	// Errors (status >= 400)
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
