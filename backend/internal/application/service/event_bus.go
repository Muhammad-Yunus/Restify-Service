package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// EventBus provides pub/sub for real-time events.
type EventBus struct {
	broker repository.MQTTBroker
	mu     sync.RWMutex
	topics map[string][]repository.MQTTHandler
}

// NewEventBus creates a new event bus.
func NewEventBus(broker repository.MQTTBroker) *EventBus {
	return &EventBus{
		broker: broker,
		topics: make(map[string][]repository.MQTTHandler),
	}
}

// Publish publishes an event to a topic.
func (eb *EventBus) Publish(ctx context.Context, topic string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	return eb.broker.Publish(ctx, topic, data, 1, false)
}

// Subscribe registers a handler for a topic.
func (eb *EventBus) Subscribe(topic string, handler repository.MQTTHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.topics[topic] = append(eb.topics[topic], handler)
}

// Unsubscribe removes a handler from a topic.
func (eb *EventBus) Unsubscribe(topic string, handler repository.MQTTHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	handlers := eb.topics[topic]
	for i, h := range handlers {
		if &h == &handler {
			eb.topics[topic] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

// Start starts consuming all subscribed topics.
func (eb *EventBus) Start(ctx context.Context) error {
	eb.mu.RLock()
	topics := make([]string, 0, len(eb.topics))
	for t := range eb.topics {
		topics = append(topics, t)
	}
	eb.mu.RUnlock()

	for _, topic := range topics {
		if err := eb.broker.Subscribe(ctx, topic, 1, func(t string, payload []byte) {
			eb.mu.RLock()
			handlers := eb.topics[t]
			eb.mu.RUnlock()
			for _, h := range handlers {
				go h(t, payload)
			}
		}); err != nil {
			return fmt.Errorf("subscribe to %s: %w", topic, err)
		}
	}
	return nil
}
