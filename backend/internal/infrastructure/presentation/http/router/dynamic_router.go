package router

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/baas"
)

// DynamicRouter wraps Gin to handle versioned BaaS routes dynamically.
type DynamicRouter struct {
	registry *baas.RouteRegistry
}

// NewDynamicRouter creates a router with dynamic BaaS support.
func NewDynamicRouter(routeReg *baas.RouteRegistry) *DynamicRouter {
	return &DynamicRouter{registry: routeReg}
}

// RegisterDynamicRoutes sets up the version catch-all route.
func (dr *DynamicRouter) RegisterDynamicRoutes(engine *gin.Engine) {
	// Match /api/{version}/** dynamically
	engine.Any("/api/:version/*path", func(c *gin.Context) {
		version := c.Param("version")
		restPath := c.Param("path")
		// Remove leading /
		restPath = strings.TrimPrefix(restPath, "/")

		key := fmt.Sprintf("%s/%s", version, restPath)
		handler, ok := dr.registry.GetHandler(version, "/"+restPath)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{
				"type":   "https://Restify.api/errors/not-found",
				"title":  "Endpoint not found",
				"status": 404,
				"detail": fmt.Sprintf("no endpoint registered for %s", key),
			})
			return
		}

		// Store endpoint info in context for logging/analytics
		if endpointID, exists := dr.registry.GetEndpointID(version, "/"+restPath); exists {
			c.Set("baas_endpoint_id", endpointID)
			c.Set("baas_version", version)
			c.Set("baas_path", "/"+restPath)
		}

		handler(c.Writer, c.Request)
	})
}
