package repository

import "context"

// MQTTBroker is the MQTT broker repository interface.
type MQTTBroker interface {
	// Close shuts down the MQTT connection.
	Close(ctx context.Context) error
}
