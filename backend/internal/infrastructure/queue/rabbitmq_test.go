package queue

import (
	"context"
	"strings"
	"testing"
)

func TestNewRabbitMQQueueInvalidURL(t *testing.T) {
	_, err := NewRabbitMQQueue(context.Background(), "not-a-url")
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}

	if !strings.Contains(err.Error(), "dial rabbitmq") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewRabbitMQQueueUnreachable(t *testing.T) {
	_, err := NewRabbitMQQueue(context.Background(), "amqp://guest:guest@127.0.0.1:1/")
	if err == nil {
		t.Fatal("expected error for unreachable broker, got nil")
	}
}
