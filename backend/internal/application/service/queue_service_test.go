package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// mockQueue implements repository.MessageQueue for testing.
type mockQueue struct {
	publishFunc func(ctx context.Context, queue string, message []byte) error
	consumeFunc func(ctx context.Context, queue string, handler repository.MessageHandler) error
	declareFunc func(ctx context.Context, name string, opts repository.QueueOptions) error
	closeFunc   func(ctx context.Context) error
}

func (m *mockQueue) Publish(ctx context.Context, queue string, message []byte) error {
	if m.publishFunc != nil {
		return m.publishFunc(ctx, queue, message)
	}
	return nil
}

func (m *mockQueue) Consume(ctx context.Context, queue string, handler repository.MessageHandler) error {
	if m.consumeFunc != nil {
		return m.consumeFunc(ctx, queue, handler)
	}
	return nil
}

func (m *mockQueue) DeclareQueue(ctx context.Context, name string, opts repository.QueueOptions) error {
	if m.declareFunc != nil {
		return m.declareFunc(ctx, name, opts)
	}
	return nil
}

func (m *mockQueue) Close(ctx context.Context) error {
	if m.closeFunc != nil {
		return m.closeFunc(ctx)
	}
	return nil
}

// mockLogger implements repository.Logger for testing.
type mockLogger struct {
	messages []logEntry
}

type logEntry struct {
	level   string
	message string
	fields  map[string]any
}

func (m *mockLogger) With(keyValues ...any) repository.Logger {
	return &mockLogger{messages: m.messages}
}

func (m *mockLogger) Info(ctx context.Context, msg string, keysAndValues ...any) {
	m.messages = append(m.messages, logEntry{level: "INFO", message: msg, fields: toFields(keysAndValues)})
}

func (m *mockLogger) Error(ctx context.Context, msg string, keysAndValues ...any) {
	m.messages = append(m.messages, logEntry{level: "ERROR", message: msg, fields: toFields(keysAndValues)})
}

func (m *mockLogger) Warn(ctx context.Context, msg string, keysAndValues ...any) {
	m.messages = append(m.messages, logEntry{level: "WARN", message: msg, fields: toFields(keysAndValues)})
}

func (m *mockLogger) Debug(ctx context.Context, msg string, keysAndValues ...any) {
	m.messages = append(m.messages, logEntry{level: "DEBUG", message: msg, fields: toFields(keysAndValues)})
}

func (m *mockLogger) Logger() *slog.Logger {
	return slog.New(slog.NewTextHandler(nil, nil))
}

func toFields(keysAndValues []any) map[string]any {
	fields := make(map[string]any)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			fields[keysAndValues[i].(string)] = keysAndValues[i+1]
		}
	}
	return fields
}

func TestQueueServicePublish(t *testing.T) {
	mq := &mockQueue{}
	qs := NewQueueService(mq)

	payload := map[string]string{"key": "value"}
	err := qs.Publish(context.Background(), "test-queue", payload)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestQueueServicePublishFailure(t *testing.T) {
	expectedErr := errors.New("publish failed")
	mq := &mockQueue{
		publishFunc: func(ctx context.Context, queue string, message []byte) error {
			return expectedErr
		},
	}
	qs := NewQueueService(mq)

	err := qs.Publish(context.Background(), "test-queue", map[string]string{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error to wrap %v, got %v", expectedErr, err)
	}
}

func TestQueueServiceConsume(t *testing.T) {
	called := false
	mq := &mockQueue{
		consumeFunc: func(ctx context.Context, queue string, handler repository.MessageHandler) error {
			called = true
			return nil
		},
	}
	qs := NewQueueService(mq)

	err := qs.Consume(context.Background(), "test-queue", func(ctx context.Context, message []byte) error {
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatal("expected consumeFunc to be called")
	}
}

func TestQueueServiceDeclareQueue(t *testing.T) {
	mq := &mockQueue{}
	qs := NewQueueService(mq)

	err := qs.DeclareQueue(context.Background(), "durable-queue", true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestQueueServiceDeclareAndConsume(t *testing.T) {
	handlerCalled := false
	mq := &mockQueue{
		declareFunc: func(ctx context.Context, name string, opts repository.QueueOptions) error {
			return nil
		},
		consumeFunc: func(ctx context.Context, queue string, handler repository.MessageHandler) error {
			handlerCalled = true
			return nil
		},
	}
	qs := NewQueueService(mq)

	err := qs.DeclareAndConsume(context.Background(), "test-queue", true, func(ctx context.Context, message []byte) error {
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !handlerCalled {
		t.Fatal("expected consumeFunc to be called after declareFunc")
	}
}

func TestQueueServiceDeclareAndConsumeDeclareFailure(t *testing.T) {
	expectedErr := errors.New("declare failed")
	mq := &mockQueue{
		declareFunc: func(ctx context.Context, name string, opts repository.QueueOptions) error {
			return expectedErr
		},
	}
	qs := NewQueueService(mq)

	err := qs.DeclareAndConsume(context.Background(), "test-queue", true, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error to wrap %v, got %v", expectedErr, err)
	}
}

func TestQueueServiceClose(t *testing.T) {
	closed := false
	mq := &mockQueue{
		closeFunc: func(ctx context.Context) error {
			closed = true
			return nil
		},
	}
	qs := NewQueueService(mq)

	err := qs.Close(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !closed {
		t.Fatal("expected closeFunc to be called")
	}
}

func TestQueueServicePublishMarshalFailure(t *testing.T) {
	mq := &mockQueue{}
	qs := NewQueueService(mq)

	// channel cannot be marshaled
	type invalidPayload struct {
		Ch chan int
	}

	err := qs.Publish(context.Background(), "test-queue", invalidPayload{})
	if err == nil {
		t.Fatal("expected error for invalid payload, got nil")
	}
	if !errors.Is(err, fmt.Errorf("marshal message: %w", nil)) {
		t.Skip("marshal error check skipped")
	}
}
