package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// AggregationJob handles periodic metric aggregation.
type AggregationJob struct {
	analyticsRepo repository.AnalyticsRepository
	logRepo       repository.APILogRepository
	logger        repository.Logger
}

// NewAggregationJob creates a new aggregation job.
func NewAggregationJob(analyticsRepo repository.AnalyticsRepository, logRepo repository.APILogRepository, logger repository.Logger) *AggregationJob {
	return &AggregationJob{
		analyticsRepo: analyticsRepo,
		logRepo:       logRepo,
		logger:        logger,
	}
}

// Run aggregates raw logs into hourly metrics.
func (j *AggregationJob) Run(ctx context.Context) error {
	// Roll up hourly metrics older than one day into daily metrics
	cutoff := time.Now().AddDate(0, 0, -1)

	if err := j.analyticsRepo.AggregateOldMetrics(ctx, cutoff); err != nil {
		return fmt.Errorf("aggregate old metrics: %w", err)
	}

	j.logger.Info(ctx, "aggregation job completed", "cutoff", cutoff.Format(time.DateOnly))
	return nil
}
