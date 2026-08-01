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

		c.Request = c.Request.WithContext(ctx)

		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.Path),
			attribute.String("http.client_ip", c.ClientIP()),
		)

		c.Next()

		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
		)

		if c.Writer.Status() >= 500 {
			span.RecordError(fmt.Errorf("http status: %d", c.Writer.Status()))
			span.SetStatus(codes.Error, "server error")
		}
	}
}
