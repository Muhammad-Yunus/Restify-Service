package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"

	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// SchedulerService manages scheduled background jobs using gocron v2.
type SchedulerService struct {
	scheduler gocron.Scheduler
	logger    repository.Logger
	jobs      map[string]gocron.Job
}

// NewSchedulerService creates a new scheduler service.
func NewSchedulerService(logger repository.Logger) *SchedulerService {
	s, err := gocron.NewScheduler(gocron.WithLogger(slog.Default()))
	if err != nil {
		panic(fmt.Errorf("create scheduler: %w", err))
	}
	return &SchedulerService{
		scheduler: s,
		logger:    logger,
		jobs:      make(map[string]gocron.Job),
	}
}

// Start begins the scheduler.
func (s *SchedulerService) Start(ctx context.Context) error {
	s.scheduler.Start()
	s.logger.Info(ctx, "scheduler started")
	return nil
}

// Stop gracefully stops the scheduler.
func (s *SchedulerService) Stop(ctx context.Context) error {
	if err := s.scheduler.Shutdown(); err != nil {
		return fmt.Errorf("stop scheduler: %w", err)
	}
	s.logger.Info(ctx, "scheduler stopped")
	return nil
}

// RegisterCron registers a job with a cron expression.
func (s *SchedulerService) RegisterCron(name string, cronExpr string, job func(context.Context) error) error {
	jobFunc := func() {
		ctx := context.Background()
		if err := job(ctx); err != nil {
			s.logger.Error(ctx, "scheduled job failed", "name", name, "error", err)
		}
	}

	jobRef, err := s.scheduler.NewJob(
		gocron.CronJob(cronExpr, false),
		gocron.NewTask(jobFunc),
		gocron.WithName(name),
		gocron.WithEventListeners(
			gocron.AfterJobRuns(func(_ uuid.UUID, jobName string) {
				s.logger.Info(context.Background(), "job completed", "name", jobName)
			}),
			gocron.AfterJobRunsWithError(func(_ uuid.UUID, jobName string, err error) {
				s.logger.Error(context.Background(), "job failed", "name", jobName, "error", err)
			}),
		),
	)
	if err != nil {
		return fmt.Errorf("register cron job %s: %w", name, err)
	}
	s.jobs[name] = jobRef
	return nil
}

// Register registers a job with a fixed duration interval.
func (s *SchedulerService) Register(name string, interval time.Duration, job func(context.Context) error) error {
	jobFunc := func() {
		ctx := context.Background()
		if err := job(ctx); err != nil {
			s.logger.Error(ctx, "scheduled job failed", "name", name, "error", err)
		}
	}

	jobRef, err := s.scheduler.NewJob(
		gocron.DurationJob(interval),
		gocron.NewTask(jobFunc),
		gocron.WithName(name),
		gocron.WithEventListeners(
			gocron.AfterJobRuns(func(_ uuid.UUID, jobName string) {
				s.logger.Info(context.Background(), "job completed", "name", jobName)
			}),
			gocron.AfterJobRunsWithError(func(_ uuid.UUID, jobName string, err error) {
				s.logger.Error(context.Background(), "job failed", "name", jobName, "error", err)
			}),
		),
	)
	if err != nil {
		return fmt.Errorf("register job %s: %w", name, err)
	}
	s.jobs[name] = jobRef
	return nil
}

// GetJob returns a registered job by name (for testing).
func (s *SchedulerService) GetJob(name string) (gocron.Job, bool) {
	j, ok := s.jobs[name]
	return j, ok
}

// Jobs returns all registered jobs (for testing).
func (s *SchedulerService) Jobs() []gocron.Job {
	return s.scheduler.Jobs()
}
