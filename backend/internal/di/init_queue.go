package di

import (
	"context"
	"errors"

	"github.com/muhammadyunus/Restify-Service/internal/config"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

type queueStub struct{}

func (s *queueStub) Close(ctx context.Context) error {
	return nil
}

func initQueue(cfg config.RabbitMQConfig) (repository.MessageQueue, error) {
	if cfg.URL == "" {
		return nil, errors.New("rabbitmq url is required")
	}

	return &queueStub{}, nil
}
