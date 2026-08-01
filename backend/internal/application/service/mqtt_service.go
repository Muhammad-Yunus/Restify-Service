package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// MQTTService manages MQTT broker operations.
type MQTTService struct {
	broker repository.MQTTBroker
}

// NewMQTTService creates a new MQTT service.
func NewMQTTService(broker repository.MQTTBroker) *MQTTService {
	return &MQTTService{broker: broker}
}

// Publish publishes a message to an MQTT topic.
func (ms *MQTTService) Publish(ctx context.Context, topic string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal MQTT payload: %w", err)
	}
	return ms.broker.Publish(ctx, topic, data, 1, false)
}

// Subscribe subscribes to an MQTT topic.
func (ms *MQTTService) Subscribe(ctx context.Context, topic string, qos byte, handler repository.MQTTHandler) error {
	return ms.broker.Subscribe(ctx, topic, qos, handler)
}

// Unsubscribe unsubscribes from a topic.
func (ms *MQTTService) Unsubscribe(ctx context.Context, topic string) error {
	return ms.broker.Unsubscribe(ctx, topic)
}

// Connect establishes connection to the MQTT broker.
func (ms *MQTTService) Connect(ctx context.Context) error {
	return ms.broker.Connect(ctx)
}

// Close shuts down the MQTT connection.
func (ms *MQTTService) Close(ctx context.Context) error {
	return ms.broker.Close(ctx)
}
