package di

import (
	"context"
	"errors"

	"github.com/muhammadyunus/Restify-Service/internal/config"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

type mqttStub struct{}

func (s *mqttStub) Connect(ctx context.Context) error {
	return errStubNotImplemented
}

func (s *mqttStub) Publish(ctx context.Context, topic string, payload []byte, qos byte, retained bool) error {
	return errStubNotImplemented
}

func (s *mqttStub) Subscribe(ctx context.Context, topic string, qos byte, handler repository.MQTTHandler) error {
	return errStubNotImplemented
}

func (s *mqttStub) Unsubscribe(ctx context.Context, topic string) error {
	return errStubNotImplemented
}

func (s *mqttStub) Close(ctx context.Context) error {
	return nil
}

func initMQTT(cfg config.EMQXConfig) (repository.MQTTBroker, error) {
	if cfg.URL == "" {
		return nil, errors.New("emqx url is required")
	}

	return &mqttStub{}, nil
}
