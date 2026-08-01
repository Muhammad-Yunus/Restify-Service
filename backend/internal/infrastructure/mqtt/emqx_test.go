package mqtt

import (
	"context"
	"strings"
	"testing"
)

func TestNewEMQXBrokerUnreachable(t *testing.T) {
	_, err := NewEMQXBroker(context.Background(), "tcp://127.0.0.1:1")
	if err == nil {
		t.Fatal("expected error for unreachable broker, got nil")
	}

	if !strings.Contains(err.Error(), "connect to MQTT broker") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewEMQXBrokerInvalidURL(t *testing.T) {
	_, err := NewEMQXBroker(context.Background(), "://invalid")
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}
}
