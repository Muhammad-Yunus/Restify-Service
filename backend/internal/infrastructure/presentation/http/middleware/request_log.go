package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// RequestLoggingMiddleware logs every HTTP request to the APILog repository.
func RequestLoggingMiddleware(logRepo repository.APILogRepository, logger repository.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start).Milliseconds()
		statusCode := c.Writer.Status()

		logLevel := determineLogLevel(statusCode)
		requestID, _ := c.Get("request_id")
		userID, _ := c.Get("user_id")

		logEntry := &entity.APILog{
			RequestID:  requestID.(string),
			Method:     c.Request.Method,
			Path:       path,
			StatusCode: statusCode,
			LatencyMs:  latency,
			LogLevel:   logLevel,
			CreatedAt:  start,
		}

		if userID != nil {
			if s, ok := userID.(string); ok && s != "" {
				if uid, err := uuid.Parse(s); err == nil {
					logEntry.UserID = &uid
				}
			}
		}

		// Async write to avoid blocking response
		go func() {
			if err := logRepo.Create(c.Request.Context(), logEntry); err != nil {
				logger.Error(c.Request.Context(), "failed to write log", "error", err)
			}
		}()
	}
}

func determineLogLevel(statusCode int) entity.LogLevel {
	switch {
	case statusCode >= 500:
		return entity.LevelError
	case statusCode >= 400:
		return entity.LevelWarn
	default:
		return entity.LevelInfo
	}
}
