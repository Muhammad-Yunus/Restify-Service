package di

import (
	"context"

	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

type logRepoStub struct{}

func (s *logRepoStub) Close(ctx context.Context) error {
	return nil
}

type analyticsRepoStub struct{}

func (s *analyticsRepoStub) Close(ctx context.Context) error {
	return nil
}

type alertRepoStub struct{}

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
