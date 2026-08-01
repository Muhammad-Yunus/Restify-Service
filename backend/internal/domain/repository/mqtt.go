package repository

import "context"

// MQTTBroker is the MQTT broker repository interface.
type MQTTBroker interface {
	// Connect establishes connection to the broker.
	Connect(ctx context.Context) error

	// Publish publishes a message to a topic.
	Publish(ctx context.Context, topic string, payload []byte, qos byte, retained bool) error

	// Subscribe subscribes to a topic with a handler.
	Subscribe(ctx context.Context, topic string, qos byte, handler MQTTHandler) error

	// Unsubscribe unsubscribes from a topic.
	Unsubscribe(ctx context.Context, topic string) error

	// Close shuts down the MQTT connection.
	Close(ctx context.Context) error
}

// MQTTHandler processes a received MQTT message.
type MQTTHandler func(topic string, payload []byte)
