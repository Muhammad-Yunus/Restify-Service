# Epic 13 — HTTP Router and Middleware

**Goal:** Set up Gin router with versioning, rate limiting, CORS, request logging, and route registration.
**Dependencies:** Epic 04 (APILog entity), Epic 05 (APILogRepository interface), Epic 06 (Gin router adapter), Epic 07 (Auth middleware)
**Commit:** `feat: add HTTP router with versioning, middleware, and route registration`

---

## Step 13.01 — Router Setup with Versioning

**Build:** Create `backend/internal/presentation/http/router/router.go`:

```go
package router

import (
    "net/http"
    "time"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// Router wraps Gin with additional middleware and versioning support.
type Router struct {
    engine *gin.Engine
}

// New creates a new router with default middleware.
func New() repository.HTTPRouter {
    engine := gin.New()
    engine.Use(gin.Recovery())
    engine.Use(gin.Logger())

    // CORS configuration
    engine.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-Key"},
        ExposeHeaders:    []string{"Link", "X-Total-Count"},
        AllowCredentials: false,
        MaxAge:           12 * time.Hour,
    }))

    return &Router{engine: engine}
}

func (r *Router) Group(basePath string, middleware ...repository.Middleware) *repository.RouterGroup {
    group := r.engine.Group(basePath)
    for _, m := range middleware {
        group.Use(middlewareToGin(m))
    }
    return &repository.RouterGroup{
        basePath:   basePath,
        middleware: append([]repository.Middleware{}, middleware...),
    }
}

func (r *Router) Handle(method, path string, handler http.HandlerFunc, middleware ...repository.Middleware) {
    ginHandler := func(c *gin.Context) {
        handler(c.Writer, c.Request)
    }
    ginMiddleware := make([]gin.HandlerFunc, len(middleware))
    for i, m := range middleware {
        ginMiddleware[i] = middlewareToGin(m)
    }

    switch method {
    case "GET":
        r.engine.GET(path, ginHandler, ginMiddleware...)
    case "POST":
        r.engine.POST(path, ginHandler, ginMiddleware...)
    case "PUT":
        r.engine.PUT(path, ginHandler, ginMiddleware...)
    case "DELETE":
        r.engine.DELETE(path, ginHandler, ginMiddleware...)
    case "PATCH":
        r.engine.PATCH(path, ginHandler, ginMiddleware...)
    case "HEAD":
        r.engine.HEAD(path, ginHandler, ginMiddleware...)
    case "OPTIONS":
        r.engine.OPTIONS(path, ginHandler, ginMiddleware...)
    default:
        panic(fmt.Sprintf("unsupported HTTP method: %s", method))
    }
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    r.engine.ServeHTTP(w, req)
}

// Engine returns the underlying Gin engine for advanced usage.
func (r *Router) Engine() *gin.Engine {
    return r.engine
}

func middlewareToGin(m repository.Middleware) gin.HandlerFunc {
    return func(c *gin.Context) {
        next := m(func(http.HandlerFunc) http.HandlerFunc {
            return func(w http.ResponseWriter, r *http.Request) {
                c.Next()
            }
        })
        next(c.Writer, c.Request)
    }
}
```

---

## Step 13.02 — Route Registration

**Build:** Create `backend/internal/presentation/http/router/routes.go`:

```go
package router

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/muhammadyunus/ForgeBase/internal/presentation/http/handler"
    "github.com/muhammadyunus/ForgeBase/internal/presentation/http/middleware"
)

// RegisterAll registers all API routes with versioning.
func RegisterAll(r *gin.Engine, deps *HandlerDeps) {
    // Health check (public)
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "ForgeBase"})
    })

    // WebSocket route (authenticated)
    if deps.WebSocketHandler != nil {
        r.GET("/api/v1/ws", deps.AuthMW.RequireAuth(), deps.WebSocketHandler.Handle)
    }

    // API v1 routes
    v1 := r.Group("/api/v1")
    {
        // Auth routes (public)
        authGroup := v1.Group("/auth")
        {
            authGroup.POST("/register", deps.AuthHandler.Register)
            authGroup.POST("/login", deps.AuthHandler.Login)
            authGroup.POST("/refresh", deps.AuthHandler.Refresh)
            authGroup.POST("/logout", deps.AuthHandler.Logout, deps.AuthMW.RequireAuth())
        }

        // User routes (authenticated)
        userGroup := v1.Group("/users", deps.AuthMW.RequireAuth())
        {
            userGroup.GET("", deps.UserHandler.List)
            userGroup.GET("/:id", deps.UserHandler.GetByID)
            userGroup.PATCH("/:id", deps.UserHandler.Update)
            userGroup.DELETE("/:id", deps.AuthMW.RequireRole("administrator"), deps.UserHandler.Delete)
        }

        // Workspace routes (authenticated)
        wsGroup := v1.Group("/workspaces", deps.AuthMW.RequireAuth())
        {
            wsGroup.GET("", deps.WorkspaceHandler.List)
            wsGroup.POST("", deps.WorkspaceHandler.Create)
            wsGroup.GET("/:id", deps.WorkspaceHandler.GetByID)
            wsGroup.PATCH("/:id", deps.WorkspaceHandler.Update)
            wsGroup.DELETE("/:id", deps.AuthMW.RequireRole("administrator"), deps.WorkspaceHandler.Delete)
        }

        // Team routes (authenticated)
        teamGroup := v1.Group("/teams", deps.AuthMW.RequireAuth())
        {
            teamGroup.GET("", deps.TeamHandler.List)
            teamGroup.GET("/:id", deps.TeamHandler.GetByID)
            teamGroup.POST("/:id/members", deps.TeamHandler.AddMember)
            teamGroup.DELETE("/:id/members/:user_id", deps.TeamHandler.RemoveMember)
            teamGroup.GET("/:id/members", deps.TeamHandler.ListMembers)
        }

        // Collection routes (authenticated)
        colGroup := v1.Group("/collections", deps.AuthMW.RequireAuth())
        {
            colGroup.GET("", deps.CollectionHandler.List)
            colGroup.GET("/:id", deps.CollectionHandler.GetByID)
            colGroup.POST("/:id/endpoints", deps.EndpointHandler.Create)
        }

        // Endpoint routes (authenticated)
        epGroup := v1.Group("/endpoints", deps.AuthMW.RequireAuth())
        {
            epGroup.GET("", deps.EndpointHandler.List)
            epGroup.GET("/:id", deps.EndpointHandler.GetByID)
            epGroup.PATCH("/:id", deps.EndpointHandler.Update)
            epGroup.DELETE("/:id", deps.EndpointHandler.Delete)
            epGroup.POST("/:id/toggle", deps.EndpointHandler.Toggle)
        }

        // Analytics routes (authenticated)
        analyticsGroup := v1.Group("/analytics", deps.AuthMW.RequireAuth())
        {
            analyticsGroup.GET("/overview", deps.AnalyticsHandler.GetOverview)
            analyticsGroup.GET("/endpoints/:id", deps.AnalyticsHandler.GetEndpointMetrics)
        }

        // Alert routes (authenticated)
        alertGroup := v1.Group("/alerts", deps.AuthMW.RequireAuth())
        {
            alertGroup.GET("", deps.AlertHandler.List)
            alertGroup.POST("", deps.AlertHandler.Create)
            alertGroup.PATCH("/:id", deps.AlertHandler.Update)
            alertGroup.DELETE("/:id", deps.AlertHandler.Delete)
            alertGroup.GET("/events", deps.AlertHandler.ListEvents)
        }

        // Log routes (authenticated)
        logGroup := v1.Group("/logs", deps.AuthMW.RequireAuth())
        {
            logGroup.GET("", deps.LogHandler.Search)
            logGroup.GET("/:id", deps.LogHandler.GetByID)
        }

        // WebSocket endpoint
        r.GET("/ws", deps.WebSocketHandler.Handle)
    }
}

// HandlerDeps holds all HTTP handlers.
type HandlerDeps struct {
    AuthHandler       *handler.AuthHandler
    UserHandler       *handler.UserHandler
    WorkspaceHandler  *handler.WorkspaceHandler
    TeamHandler       *handler.TeamHandler
    CollectionHandler *handler.CollectionHandler
    EndpointHandler   *handler.EndpointHandler
    AnalyticsHandler  *handler.AnalyticsHandler
    AlertHandler      *handler.AlertHandler
    LogHandler        *handler.LogHandler
    WebSocketHandler  *handler.WebSocketHandler
    AuthMW            *middleware.AuthMiddleware
}
```

---

## Step 13.03 — Rate Limiting Middleware

**Build:** Create `backend/internal/presentation/http/middleware/rate_limit.go`:

```go
package middleware

import (
    "net/http"
    "sync"
    "time"

    "github.com/gin-gonic/gin"
)

// RateLimitMiddleware provides simple in-memory rate limiting.
type RateLimitMiddleware struct {
    mu       sync.Mutex
    clients  map[string]*clientRate
    limit    int
    window   time.Duration
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
        c.Header("X-RateLimit-Remaining", strconv.Itoa(rl.limit - count))

        if count > rl.limit {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "rate limit exceeded",
                "retry_after": int(time.Until(now.Add(rl.window)).Seconds()),
            })
            c.Abort()
            return
        }

        c.Next()
    }
}
```

---

## Step 13.04 — Request ID Middleware

**Build:** Create `backend/internal/presentation/http/middleware/request_id.go`:

```go
package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

// RequestIDMiddleware adds a unique request ID to each request.
func RequestIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := uuid.New().String()
        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)
        c.Next()
    }
}
```

---

## Step 13.05 — Request Logging Middleware

**Build:** Create `backend/internal/presentation/http/middleware/request_log.go`:

```go
package middleware

import (
    "time"

    "github.com/gin-gonic/gin"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// RequestLoggingMiddleware logs every HTTP request.
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
            RequestID:   requestID.(string),
            Method:      c.Request.Method,
            Path:        path,
            StatusCode:  statusCode,
            LatencyMs:   latency,
            LogLevel:    logLevel,
            Message:     c.Writer.String(),
            CreatedAt:   start,
        }

        if userID != nil {
            // Parse userID if needed
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
```

---

## Step 13.06 — Router Initialization in DI

**Build:** Update `internal/di/bootstrap.go`:

```go
func initRouter(_ context.Context, c *di.Container) (repository.HTTPRouter, error) {
    return router.NewGinRouter(), nil
}
```

**Test cases:**
- [ ] Unit: `initRouter()` returns a non-nil HTTPRouter

---

## Commit Instruction

```bash
git add .
git commit -m "feat: add HTTP router with versioning, rate limiting, CORS, request ID, and logging middleware"
```
