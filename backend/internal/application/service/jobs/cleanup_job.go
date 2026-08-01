package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// CleanupJob handles periodic cleanup tasks.
type CleanupJob struct {
	logRepo      repository.APILogRepository
	cache        repository.Cache
	logger       repository.Logger
	retentionDays int
}

// NewCleanupJob creates a new cleanup job.
func NewCleanupJob(logRepo repository.APILogRepository, cache repository.Cache, logger repository.Logger) *CleanupJob {
	return &CleanupJob{
		logRepo:       logRepo,
		cache:         cache,
		logger:        logger,
		retentionDays: 90,
	}
}

// Run deletes logs older than retention period.
func (j *CleanupJob) Run(ctx context.Context) error {
	cutoff := time.Now().AddDate(0, 0, -j.retentionDays)

	deleted, err := j.logRepo.DeleteOlderThan(ctx, cutoff)
	if err != nil {
		return fmt.Errorf("cleanup old logs: %w", err)
	}

	j.logger.Info(ctx, "cleanup job completed", "logs_deleted", deleted)

	// Also clean up cache entries older than retention
	// Note: this is a best-effort cache cleanup
	if err := j.cache.Delete(ctx, "analytics:aggregation"); err != nil {
		j.logger.Warn(ctx, "cache cleanup partial", "key", "analytics:aggregation", "error", err)
	}

	return nil
}
