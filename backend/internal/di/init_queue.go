package di

import (
	"context"
	"errors"

	"github.com/muhammadyunus/Restify-Service/internal/config"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

type queueStub struct{}

func (s *queueStub) Publish(ctx context.Context, queue string, message []byte) error {
	return errStubNotImplemented
}

func (s *queueStub) Consume(ctx context.Context, queue string, handler repository.MessageHandler) error {
	return errStubNotImplemented
}

func (s *queueStub) DeclareQueue(ctx context.Context, name string, opts repository.QueueOptions) error {
	return errStubNotImplemented
}

func (s *queueStub) Close(ctx context.Context) error {
	return nil
}

func initQueue(cfg config.RabbitMQConfig) (repository.MessageQueue, error) {
	if cfg.URL == "" {
		return nil, errors.New("rabbitmq url is required")
	}

	return &queueStub{}, nil
}
