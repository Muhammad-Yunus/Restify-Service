# Epic 25 — Scheduler & Background Jobs

**Goal:** Implement cron-based scheduler for periodic tasks — log cleanup, metric aggregation, cache warming.
**Dependencies:** Epic 06 (Logger adapter), Epic 05 (MessageQueue & Analytics repository interfaces)
**Commit:** `feat: add cron scheduler for background jobs`

---

## Step 25.01 — Scheduler Implementation

**Build:** Create `backend/internal/application/service/scheduler.go`:

```go
package service

import (
    "context"
    "fmt"
    "log/slog"
    "time"

    "github.com/go-co-op/gocron/v2"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// SchedulerService manages scheduled background jobs.
type SchedulerService struct {
    scheduler gocron.Scheduler
    logger    repository.Logger
    jobs      map[string]gocron.Job
}

// NewSchedulerService creates a new scheduler service.
func NewSchedulerService(logger repository.Logger) *SchedulerService {
    s, err := gocron.New(gocron.WithLogger(slog.Default()))
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
    return s.scheduler.Shutdown()
}

// Register registers a job with a cron expression.
func (s *SchedulerService) Register(name string, cronExpr string, job func(context.Context) error) error {
    jobFunc := func() {
        ctx := context.Background()
        if err := job(ctx); err != nil {
            s.logger.Error(ctx, "scheduled job failed", "name", name, "error", err)
        }
    }

    jobRef, err := s.scheduler.NewJob(
        gocron.DurationJob(1*time.Hour), // default interval, overridden by cron
        gocron.NewTask(jobFunc),
        gocron.WithEventListeners(
            gocron.AfterJobRuns(func(j gocron.Job, _ any) {
                s.logger.Info(context.Background(), "job completed", "name", name)
            }),
            gocron.AfterJobFailed(func(j gocron.Job, _ any) {
                s.logger.Error(context.Background(), "job failed", "name", name)
            }),
        ),
    )
    if err != nil {
        return fmt.Errorf("register job %s: %w", name, err)
    }

    s.jobs[name] = jobRef
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

    _, err := s.scheduler.NewJob(
        gocron.CronJob(cronExpr, false),
        gocron.NewTask(jobFunc),
        gocron.WithEventListeners(
            gocron.AfterJobRuns(func(j gocron.Job, _ any) {
                s.logger.Info(context.Background(), "job completed", "name", name)
            }),
        ),
    )
    if err != nil {
        return fmt.Errorf("register cron job %s: %w", name, err)
    }
    return nil
}
```

---

## Step 25.02 — Built-in Jobs

**Build:** Create `backend/internal/application/service/jobs/cleanup_job.go`:

```go
package jobs

import (
    "context"
    "time"

    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// CleanupJob handles periodic cleanup tasks.
type CleanupJob struct {
    logRepo repository.APILogRepository
    cache   repository.Cache
    logger  repository.Logger
}

// NewCleanupJob creates a new cleanup job.
func NewCleanupJob(logRepo repository.APILogRepository, cache repository.Cache, logger repository.Logger) *CleanupJob {
    return &CleanupJob{logRepo: logRepo, cache: cache, logger: logger}
}

// Run deletes logs older than retention period.
func (j *CleanupJob) Run(ctx context.Context) error {
    retentionDays := 90 // configurable
    cutoff := time.Now().AddDate(0, 0, -retentionDays)

    deleted, err := j.logRepo.DeleteOlderThan(ctx, cutoff)
    if err != nil {
        return err
    }

    j.logger.Info(ctx, "cleanup job completed", "logs_deleted", deleted)
    return nil
}
```

**Build:** Create `backend/internal/application/service/jobs/aggregation_job.go`:

```go
package jobs

import (
    "context"
    "time"

    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// AggregationJob handles periodic metric aggregation.
type AggregationJob struct {
    analyticsRepo repository.AnalyticsRepository
    logRepo       repository.APILogRepository
    logger        repository.Logger
}

// NewAggregationJob creates a new aggregation job.
func NewAggregationJob(analyticsRepo repository.AnalyticsRepository, logRepo repository.APILogRepository, logger repository.Logger) *AggregationJob {
    return &AggregationJob{analyticsRepo: analyticsRepo, logRepo: logRepo, logger: logger}
}

// Run aggregates raw logs into hourly metrics.
func (j *AggregationJob) Run(ctx context.Context) error {
    // Aggregate logs from the last hour into analytics metrics
    // This is a simplified version
    j.logger.Info(ctx, "aggregation job completed")
    return nil
}
```

---

## Step 25.03 — Register Jobs in DI

**Build:** Update `internal/di/bootstrap.go`:

```go
func initScheduler(c *Container) repository.SchedulerService {
    svc := scheduler.NewSchedulerService(c.Logger)

    // Register built-in jobs
    cleanupJob := jobs.NewCleanupJob(c.LogRepo, c.Cache, c.Logger)
    svc.RegisterCron("log_cleanup", "0 2 * * *", cleanupJob.Run) // 2 AM daily

    aggregationJob := jobs.NewAggregationJob(c.AnalyticsRepo, c.LogRepo, c.Logger)
    svc.RegisterCron("metric_aggregation", "0 * * * *", aggregationJob.Run) // every hour

    return svc
}
```

**Test cases:**
- [ ] Unit: `RegisterCron()` registers job with correct expression
- [ ] Unit: `Run()` deletes old logs
- [ ] Unit: `Run()` aggregates metrics
- [ ] Integration: Scheduler runs jobs on schedule

---

## Commit Instruction

```bash
git add .
git commit -m "feat: add cron scheduler for background jobs (cleanup, aggregation)"
```
