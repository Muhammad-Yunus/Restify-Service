package service

import (
	"context"
	"errors"
	"testing"

	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// mockMQTTBroker implements repository.MQTTBroker for testing.
type mockMQTTBroker struct {
	publishFunc   func(ctx context.Context, topic string, payload []byte, qos byte, retained bool) error
	subscribeFunc func(ctx context.Context, topic string, qos byte, handler repository.MQTTHandler) error
}

func (m *mockMQTTBroker) Connect(ctx context.Context) error {
	return nil
}

func (m *mockMQTTBroker) Publish(ctx context.Context, topic string, payload []byte, qos byte, retained bool) error {
	if m.publishFunc != nil {
		return m.publishFunc(ctx, topic, payload, qos, retained)
	}
	return nil
}

func (m *mockMQTTBroker) Subscribe(ctx context.Context, topic string, qos byte, handler repository.MQTTHandler) error {
	if m.subscribeFunc != nil {
		return m.subscribeFunc(ctx, topic, qos, handler)
	}
	return nil
}

func (m *mockMQTTBroker) Unsubscribe(ctx context.Context, topic string) error {
	return nil
}

func (m *mockMQTTBroker) Close(ctx context.Context) error {
	return nil
}

func TestMQTTServicePublish(t *testing.T) {
	var receivedTopic string
	var receivedPayload []byte
	mb := &mockMQTTBroker{
		publishFunc: func(ctx context.Context, topic string, payload []byte, qos byte, retained bool) error {
			receivedTopic = topic
			receivedPayload = payload
			return nil
		},
	}
	ms := NewMQTTService(mb)

	payload := map[string]string{"event": "test"}
	err := ms.Publish(context.Background(), "test/topic", payload)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if receivedTopic != "test/topic" {
		t.Fatalf("expected topic 'test/topic', got %s", receivedTopic)
	}
	if len(receivedPayload) == 0 {
		t.Fatal("expected payload to be sent")
	}
}

func TestMQTTServicePublishFailure(t *testing.T) {
	expectedErr := errors.New("publish failed")
	mb := &mockMQTTBroker{
		publishFunc: func(ctx context.Context, topic string, payload []byte, qos byte, retained bool) error {
			return expectedErr
		},
	}
	ms := NewMQTTService(mb)

	err := ms.Publish(context.Background(), "test/topic", map[string]string{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error to wrap %v, got %v", expectedErr, err)
	}
}

func TestMQTTServicePublishMarshalFailure(t *testing.T) {
	mb := &mockMQTTBroker{}
	ms := NewMQTTService(mb)

	// channel cannot be marshaled
	type invalidPayload struct {
		Ch chan int
	}

	err := ms.Publish(context.Background(), "test/topic", invalidPayload{})
	if err == nil {
		t.Fatal("expected error for invalid payload, got nil")
	}
}

func TestMQTTServiceSubscribe(t *testing.T) {
	var subscribedTopic string
	mb := &mockMQTTBroker{
		subscribeFunc: func(ctx context.Context, topic string, qos byte, handler repository.MQTTHandler) error {
			subscribedTopic = topic
			return nil
		},
	}
	ms := NewMQTTService(mb)

	err := ms.Subscribe(context.Background(), "test/topic", 1, func(topic string, payload []byte) {})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if subscribedTopic != "test/topic" {
		t.Fatalf("expected subscribed topic to be 'test/topic', got %s", subscribedTopic)
	}
}

func TestMQTTServiceUnsubscribe(t *testing.T) {
	mb := &mockMQTTBroker{}
	ms := NewMQTTService(mb)

	err := ms.Unsubscribe(context.Background(), "test/topic")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestMQTTServiceClose(t *testing.T) {
	mb := &mockMQTTBroker{}
	ms := NewMQTTService(mb)

	err := ms.Close(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestMQTTServiceConnect(t *testing.T) {
	mb := &mockMQTTBroker{}
	ms := NewMQTTService(mb)

	err := ms.Connect(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
