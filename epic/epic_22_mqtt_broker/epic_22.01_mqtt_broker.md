# Epic 22 — MQTT Broker (EMQX)

**Goal:** Implement EMQX MQTT broker integration for real-time pub/sub messaging.
**Dependencies:** Epic 06 (EMQX adapter), Epic 05 (MQTT repository interface)
**Commit:** `feat: add EMQX MQTT broker integration`

---

## Step 22.01 — MQTT Service

**Build:** Create `backend/internal/application/service/mqtt_service.go`:

```go
package service

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
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
func (ms *MQTTService) Subscribe(ctx context.Context, topic string, handler repository.MQTTHandler) error {
    return ms.broker.Subscribe(ctx, topic, 1, handler)
}

// Unsubscribe unsubscribes from a topic.
func (ms *MQTTService) Unsubscribe(ctx context.Context, topic string) error {
    return ms.broker.Unsubscribe(ctx, topic)
}
```

---

## Step 22.02 — Real-time Event Bus

**Build:** Create `backend/internal/application/service/event_bus.go`:

```go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    "time"

    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
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
        // Compare handlers by pointer
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
```

---

## Step 22.03 — MQTT in DI Bootstrap

**Build:** Update `internal/di/bootstrap.go`:

```go
func initMQTT(ctx context.Context, cfg config.EMQXConfig) (repository.MQTTBroker, error) {
    return mqtt.NewEMQXBroker(ctx, cfg.URL)
}
```

---

## Commit Instruction

```bash
git add .
git commit -m "feat: add EMQX MQTT broker with real-time event bus"
```
