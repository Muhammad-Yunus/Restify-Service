package tracing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTracerProviderEmptyEndpoint(t *testing.T) {
	_, err := NewTracerProvider(context.Background(), "", "test-service")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tracing endpoint is required")
}

func TestNewTracerProviderEmptyServiceName(t *testing.T) {
	_, err := NewTracerProvider(context.Background(), "http://localhost:4317", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service name is required")
}

func TestTracer(t *testing.T) {
	// This test only validates that Tracer() returns a non-nil tracer
	// Full integration requires a running OTLP collector
	tp, err := NewTracerProvider(context.Background(), "http://localhost:4317", "test-service")
	if err != nil {
		t.Skipf("Skipping test: %v", err)
	}
	defer tp.Shutdown(context.Background())

	tracer := tp.Tracer("test")
	assert.NotNil(t, tracer)
}
