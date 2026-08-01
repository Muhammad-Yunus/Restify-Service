# Epic 16 — API Versioning and Dynamic Routing

**Goal:** Implement versioned URL routing that dispatches requests to dynamically generated handlers based on endpoint bindings.
**Dependencies:** Epic 13 (HTTP Router), Epic 15 (REST Generator)
**Commit:** `feat: add versioned dynamic routing for BaaS endpoints`

---

## Step 16.01 — Dynamic Route Registry

**Build:** Create `backend/internal/infrastructure/baas/route_registry.go`:

```go
package baas

import (
    "context"
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// RouteRegistry maps endpoint IDs to their generated HTTP handlers.
type RouteRegistry struct {
    registry map[string]http.HandlerFunc // key: "{version}/{path}"
    generator repository.RESTGenerator
    endpointRepo repository.EndpointRepository
    logger repository.Logger
}

// NewRouteRegistry creates a new route registry.
func NewRouteRegistry(gen repository.RESTGenerator, epRepo repository.EndpointRepository, logger repository.Logger) *RouteRegistry {
    return &RouteRegistry{
        registry:     make(map[string]http.HandlerFunc),
        generator:    gen,
        endpointRepo: epRepo,
        logger:       logger,
    }
}

// RegisterEndpoint registers a handler for an endpoint.
func (rr *RouteRegistry) RegisterEndpoint(ctx context.Context, endpoint *entity.Endpoint) error {
    handler, err := rr.generator.GenerateHandler(ctx, endpoint)
    if err != nil {
        return fmt.Errorf("generate handler for endpoint %s: %w", endpoint.ID, err)
    }

    key := fmt.Sprintf("%s%s", endpoint.Version, endpoint.Path)
    rr.registry[key] = handler
    rr.logger.Info(ctx, "endpoint registered", "key", key, "endpoint_id", endpoint.ID)
    return nil
}

// UnregisterEndpoint removes a handler registration.
func (rr *RouteRegistry) UnregisterEndpoint(endpointID string) {
    // Find and remove all keys associated with this endpoint
    // In practice, store endpoint ID as part of the key or maintain a separate map
    delete(rr.registry, endpointID)
}

// GetHandler returns the handler for a versioned path.
func (rr *RouteRegistry) GetHandler(version, path string) (http.HandlerFunc, bool) {
    key := fmt.Sprintf("%s%s", version, path)
    handler, ok := rr.registry[key]
    return handler, ok
}

// RefreshAll re-registers all active endpoints.
func (rr *RouteRegistry) RefreshAll(ctx context.Context) error {
    // Clear and re-register
    rr.registry = make(map[string]http.HandlerFunc)

    // Get all active endpoints (simplified — in production, query by workspace)
    // This would need a ListAll method on the repository
    return nil
}
```

---

## Step 16.02 — Dynamic Versioned Router

**Build:** Create `backend/internal/presentation/http/router/dynamic_router.go`:

```go
package router

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/muhammadyunus/ForgeBase/internal/infrastructure/baas"
)

// DynamicRouter wraps Gin to handle versioned BaaS routes dynamically.
type DynamicRouter struct {
    engine     *gin.Engine
    routeReg   *baas.RouteRegistry
}

// NewDynamicRouter creates a router with dynamic BaaS support.
func NewDynamicRouter(routeReg *baas.RouteRegistry) *DynamicRouter {
    engine := gin.New()
    engine.Use(gin.Recovery())
    return &DynamicRouter{engine: engine, routeReg: routeReg}
}

// RegisterDynamicRoutes sets up the version catch-all route.
func (dr *DynamicRouter) RegisterDynamicRoutes() {
    // Match /api/{version}/** dynamically
    dr.engine.Any("/api/:version/*path", func(c *gin.Context) {
        version := c.Param("version")
        restPath := c.Param("path")
        // Remove leading /
        restPath = strings.TrimPrefix(restPath, "/")

        handler, ok := dr.routeReg.GetHandler(version, "/"+restPath)
        if !ok {
            c.JSON(http.StatusNotFound, gin.H{
                "type":   "https://ForgeBase.api/errors/not-found",
                "title":  "Endpoint not found",
                "status": 404,
                "detail": fmt.Sprintf("no endpoint registered for %s%s", version, restPath),
            })
            return
        }

        // Create a new context with the endpoint info
        c.Set("baas_version", version)
        c.Set("baas_path", "/"+restPath)
        handler(c.Writer, c.Request)
    })
}

// Engine returns the underlying Gin engine.
func (dr *DynamicRouter) Engine() *gin.Engine {
    return dr.engine
}
```

---

## Step 16.03 — Endpoint Live Handler

**Build:** Create `backend/internal/presentation/http/handler/live_handler.go`:

```go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/muhammadyunus/ForgeBase/internal/infrastructure/baas"
)

// LiveHandler provides a live preview of generated endpoints.
type LiveHandler struct {
    routeReg *baas.RouteRegistry
}

// NewLiveHandler creates a new live handler.
func NewLiveHandler(routeReg *baas.RouteRegistry) *LiveHandler {
    return &LiveHandler{routeReg: routeReg}
}

// ListRegistered handles GET /api/v1/admin/endpoints
func (h *LiveHandler) ListRegistered(c *gin.Context) {
    // Return all registered routes
    c.JSON(http.StatusOK, gin.H{
        "data": gin.H{"routes": gin.H(h.routeReg.Registry())},
    })
}
```

---

## Commit Instruction

```bash
git add .
git commit -m "feat: add versioned dynamic routing for BaaS endpoints"
```
