# Epic 28 — OpenTelemetry Tracing

**Goal:** Implement OpenTelemetry for distributed tracing, metrics export, and observability.
**Dependencies:** Epic 06 (Logger adapter), Epic 13 (HTTP Router)
**Commit:** `feat: add OpenTelemetry tracing and metrics`

---

## Step 28.01 — OTel Tracer Provider

**Build:** Create `backend/internal/infrastructure/tracing/otel.go`:

```go
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
)

// TracerProvider manages OpenTelemetry tracing.
type TracerProvider struct {
    provider *sdktrace.TracerProvider
    exporter *otlptracegrpc.Exporter
}

// NewTracerProvider creates a new OTel tracer provider.
func NewTracerProvider(ctx context.Context, endpoint string, serviceName string) (*TracerProvider, error) {
    if endpoint == "" {
        return nil, fmt.Errorf("tracing endpoint is required")
    }

    if serviceName == "" {
        return nil, fmt.Errorf("service name is required")
    }

    // Create OTLP exporter
    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint(endpoint),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        return nil, fmt.Errorf("create OTLP exporter: %w", err)
    }

    // Create tracer provider
    provider := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceName(serviceName),
        )),
    )

    // Set global tracer provider
    otel.SetTracerProvider(provider)

    // Configure propagator
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))

    return &TracerProvider{provider: provider, exporter: exporter}, nil
}

// Tracer returns a named otel.Tracer instance.
func (tp *TracerProvider) Tracer(name string) otel.Tracer {
    return otel.Tracer(name)
}

// Shutdown flushes all spans.
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
    if err := tp.provider.Shutdown(ctx); err != nil {
        return fmt.Errorf("shutdown tracer provider: %w", err)
    }
    if err := tp.exporter.Shutdown(ctx); err != nil {
        return fmt.Errorf("shutdown OTLP exporter: %w", err)
    }
    return nil
}
```

---

## Step 28.02 — OTel HTTP Middleware

**Build:** Create `backend/internal/infrastructure/tracing/middleware.go`:

```go
// Package tracing provides OpenTelemetry tracing integration.
package tracing

import (
    "fmt"

    "github.com/gin-gonic/gin"
    "go.opentelemetry.io/otel/attribute"
    oteltrace "go.opentelemetry.io/otel/trace"
    "go.opentelemetry.io/otel/codes"
)

// OTelMiddleware adds tracing to HTTP requests.
func OTelMiddleware(tracer oteltrace.Tracer) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, span := tracer.Start(c.Request.Context(), c.Request.Method+" "+c.Request.URL.Path)
        defer span.End()

        // Inject trace context into request
        c.Request = c.Request.WithContext(ctx)

        // Set span attributes
        span.SetAttributes(
            attribute.String("http.method", c.Request.Method),
            attribute.String("http.url", c.Request.URL.Path),
            attribute.String("http.client_ip", c.ClientIP()),
        )

        c.Next()

        // Set status code
        span.SetAttributes(
            attribute.Int("http.status_code", c.Writer.Status()),
        )

        if c.Writer.Status() >= 500 {
            span.RecordError(fmt.Errorf("http status: %d", c.Writer.Status()))
            span.SetStatus(codes.Error, "server error")
        }
    }
}
```

---

## Step 28.03 — OTel in DI Bootstrap

**Build:** Update `internal/di/bootstrap.go`:

```go
func initTracing(cfg config.OTELConfig, serviceName string) (*tracing.TracerProvider, error) {
    if !cfg.Enabled {
        return nil, nil // tracing disabled
    }
    if cfg.Endpoint == "" {
        return nil, fmt.Errorf("tracing endpoint is required when tracing is enabled")
    }
    return tracing.NewTracerProvider(context.Background(), cfg.Endpoint, serviceName)
}
```

**Test cases:**
- [ ] Unit: `NewTracerProvider()` rejects empty endpoint
- [ ] Unit: `NewTracerProvider()` rejects empty service name
- [ ] Unit: `NewTracerProvider()` creates provider with OTLP exporter
- [ ] Unit: `OTelMiddleware()` starts span for each request
- [ ] Unit: Span attributes include HTTP method and status
- [ ] Unit: `Tracer()` returns valid otel.Tracer
- [ ] Integration: Traces exported to OTLP collector
- [ ] Integration: 5xx responses set span status to Error

---

## Commit Instruction

```bash
git add .
git commit -m "feat: add OpenTelemetry tracing and metrics export"
```
