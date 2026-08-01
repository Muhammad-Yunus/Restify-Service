package di

import (
	"context"
	"errors"

	"github.com/muhammadyunus/Restify-Service/internal/config"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

type mqttStub struct{}

func (s *mqttStub) Close(ctx context.Context) error {
	return nil
}

func initMQTT(cfg config.EMQXConfig) (repository.MQTTBroker, error) {
	if cfg.URL == "" {
		return nil, errors.New("emqx url is required")
	}

	return &mqttStub{}, nil
}
