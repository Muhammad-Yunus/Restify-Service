package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitMiddleware provides simple in-memory rate limiting.
type RateLimitMiddleware struct {
	mu      sync.Mutex
	clients map[string]*clientRate
	limit   int
	window  time.Duration
}

type clientRate struct {
	requests int
	resetAt  time.Time
}

// NewRateLimitMiddleware creates a rate limiter.
func NewRateLimitMiddleware(requestsPerMinute int) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		clients: make(map[string]*clientRate),
		limit:   requestsPerMinute,
		window:  time.Minute,
	}
}

// Limit is the Gin middleware.
func (rl *RateLimitMiddleware) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()

		rl.mu.Lock()
		client, exists := rl.clients[key]
		now := time.Now()

		if !exists || now.After(client.resetAt) {
			client = &clientRate{requests: 0, resetAt: now.Add(rl.window)}
			rl.clients[key] = client
		}

		client.requests++
		count := client.requests
		rl.mu.Unlock()

		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(rl.limit-count))

		if count > rl.limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": int(time.Until(client.resetAt).Seconds()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
