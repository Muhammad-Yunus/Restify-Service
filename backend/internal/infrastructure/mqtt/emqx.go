package mqtt

import (
	"context"
	"errors"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"

	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// EMQXBroker implements the repository.MQTTBroker interface.
type EMQXBroker struct {
	client mqtt.Client
}

// NewEMQXBroker creates a new MQTT broker connection.
func NewEMQXBroker(ctx context.Context, brokerURL string) (*EMQXBroker, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(fmt.Sprintf("ForgeBase-%s", uuid.New().String()[:8]))
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(false)
	opts.SetConnectTimeout(5 * time.Second)

	client := mqtt.NewClient(opts)
	token := client.Connect()

	if !token.WaitTimeout(10 * time.Second) {
		return nil, errors.New("connect to MQTT broker: timeout")
	}

	if err := token.Error(); err != nil {
		return nil, fmt.Errorf("connect to MQTT broker: %w", err)
	}

	return &EMQXBroker{client: client}, nil
}

// Connect establishes connection to the broker.
func (m *EMQXBroker) Connect(ctx context.Context) error {
	token := m.client.Connect()

	if !token.WaitTimeout(10 * time.Second) {
		return errors.New("mqtt connect timeout")
	}

	if err := token.Error(); err != nil {
		return fmt.Errorf("connect to mqtt broker: %w", err)
	}

	return nil
}

// Publish publishes a message to a topic.
func (m *EMQXBroker) Publish(ctx context.Context, topic string, payload []byte, qos byte, retained bool) error {
	token := m.client.Publish(topic, qos, retained, payload)

	if !token.WaitTimeout(5 * time.Second) {
		return errors.New("mqtt publish timeout")
	}

	if err := token.Error(); err != nil {
		return fmt.Errorf("publish to %s: %w", topic, err)
	}

	return nil
}

// Subscribe subscribes to a topic with a handler.
func (m *EMQXBroker) Subscribe(ctx context.Context, topic string, qos byte, handler repository.MQTTHandler) error {
	token := m.client.Subscribe(topic, qos, func(client mqtt.Client, msg mqtt.Message) {
		handler(msg.Topic(), msg.Payload())
	})

	if !token.WaitTimeout(5 * time.Second) {
		return errors.New("mqtt subscribe timeout")
	}

	if err := token.Error(); err != nil {
		return fmt.Errorf("subscribe to %s: %w", topic, err)
	}

	return nil
}

// Unsubscribe unsubscribes from a topic.
func (m *EMQXBroker) Unsubscribe(ctx context.Context, topic string) error {
	token := m.client.Unsubscribe(topic)

	if !token.WaitTimeout(5 * time.Second) {
		return errors.New("mqtt unsubscribe timeout")
	}

	if err := token.Error(); err != nil {
		return fmt.Errorf("unsubscribe from %s: %w", topic, err)
	}

	return nil
}

// Close shuts down the MQTT connection.
func (m *EMQXBroker) Close(ctx context.Context) error {
	m.client.Disconnect(250)

	return nil
}

// Compile-time check.
var _ repository.MQTTBroker = (*EMQXBroker)(nil)
