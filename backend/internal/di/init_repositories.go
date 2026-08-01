package di

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apprepo "github.com/muhammadyunus/Restify-Service/internal/application/repository"
	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

var errStubNotImplemented = errors.New("stub not implemented")

type logRepoStub struct{}

func (s *logRepoStub) Create(ctx context.Context, log *entity.APILog) error {
	return errStubNotImplemented
}

func (s *logRepoStub) CreateBatch(ctx context.Context, logs []*entity.APILog) error {
	return errStubNotImplemented
}

func (s *logRepoStub) FindByID(ctx context.Context, id uuid.UUID) (*entity.APILog, error) {
	return nil, errStubNotImplemented
}

func (s *logRepoStub) Search(ctx context.Context, workspaceID, endpointID uuid.UUID, level entity.LogLevel, method, path string, from, to time.Time, page, pageSize int) ([]*entity.APILog, int, error) {
	return nil, 0, errStubNotImplemented
}

func (s *logRepoStub) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	return 0, errStubNotImplemented
}

func (s *logRepoStub) CountByWorkspace(ctx context.Context, workspaceID uuid.UUID, from, to time.Time) (int64, error) {
	return 0, errStubNotImplemented
}

func (s *logRepoStub) Close(ctx context.Context) error {
	return nil
}

type analyticsRepoStub struct{}

func (s *analyticsRepoStub) RecordMetric(ctx context.Context, metric *entity.AnalyticsMetric) error {
	return errStubNotImplemented
}

func (s *analyticsRepoStub) RecordMetricsBatch(ctx context.Context, metrics []*entity.AnalyticsMetric) error {
	return errStubNotImplemented
}

func (s *analyticsRepoStub) GetMetrics(ctx context.Context, workspaceID uuid.UUID, metricName string, from, to time.Time, interval time.Duration) ([]*entity.AnalyticsMetric, error) {
	return nil, errStubNotImplemented
}

func (s *analyticsRepoStub) GetEndpointMetrics(ctx context.Context, endpointID uuid.UUID, from, to time.Time) ([]*entity.AnalyticsMetric, error) {
	return nil, errStubNotImplemented
}

func (s *analyticsRepoStub) GetOverview(ctx context.Context, workspaceID uuid.UUID, from, to time.Time) (repository.OverviewMetrics, error) {
	return repository.OverviewMetrics{}, errStubNotImplemented
}

func (s *analyticsRepoStub) AggregateOldMetrics(ctx context.Context, olderThan time.Time) error {
	return errStubNotImplemented
}

func (s *analyticsRepoStub) Close(ctx context.Context) error {
	return nil
}

type alertRepoStub struct{}

func (s *alertRepoStub) Create(ctx context.Context, rule *entity.AlertRule) error {
	return errStubNotImplemented
}

func (s *alertRepoStub) FindByID(ctx context.Context, id uuid.UUID) (*entity.AlertRule, error) {
	return nil, errStubNotImplemented
}

func (s *alertRepoStub) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entity.AlertRule, error) {
	return nil, errStubNotImplemented
}

func (s *alertRepoStub) Update(ctx context.Context, rule *entity.AlertRule) error {
	return errStubNotImplemented
}

func (s *alertRepoStub) Delete(ctx context.Context, id uuid.UUID) error {
	return errStubNotImplemented
}

func (s *alertRepoStub) ToggleEnabled(ctx context.Context, id uuid.UUID, enabled bool) error {
	return errStubNotImplemented
}

func (s *alertRepoStub) CreateEvent(ctx context.Context, event *entity.AlertEvent) error {
	return errStubNotImplemented
}

func (s *alertRepoStub) ListRecentEvents(ctx context.Context, workspaceID uuid.UUID, limit int) ([]*entity.AlertEvent, error) {
	return nil, errStubNotImplemented
}

func (s *alertRepoStub) MarkNotified(ctx context.Context, id uuid.UUID) error {
	return errStubNotImplemented
}

func (s *alertRepoStub) Close(ctx context.Context) error {
	return nil
}

func initLogRepo() (repository.APILogRepository, error) {
	return &logRepoStub{}, nil
}

func initAnalyticsRepo() (repository.AnalyticsRepository, error) {
	return &analyticsRepoStub{}, nil
}

func initAlertRepo() (repository.AlertRepository, error) {
	return &alertRepoStub{}, nil
}

func initUserRepository(pgDB repository.DB, gormDB *gorm.DB) repository.UserRepository {
	return apprepo.NewUserRepository(pgDB, gormDB)
}
