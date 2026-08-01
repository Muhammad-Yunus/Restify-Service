// Package tracing provides OpenTelemetry tracing integration.
package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// TracerProvider manages OpenTelemetry tracing.
type TracerProvider struct {
	provider *sdktrace.TracerProvider
}

// NewTracerProvider creates a new OTel tracer provider.
func NewTracerProvider(ctx context.Context, endpoint string, serviceName string) (*TracerProvider, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("tracing endpoint is required")
	}

	if serviceName == "" {
		return nil, fmt.Errorf("service name is required")
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create OTLP exporter: %w", err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
	)

	otel.SetTracerProvider(provider)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &TracerProvider{provider: provider}, nil
}

// Tracer returns a named otel.Tracer instance.
func (tp *TracerProvider) Tracer(name string) oteltrace.Tracer {
	return otel.Tracer(name)
}

// Shutdown flushes all spans.
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	if err := tp.provider.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown tracer provider: %w", err)
	}
	return nil
}
