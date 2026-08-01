package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/baas"
)

// LiveHandler provides live introspection of registered BaaS endpoints.
type LiveHandler struct {
	routeReg *baas.RouteRegistry
}

// NewLiveHandler creates a new live handler.
func NewLiveHandler(routeReg *baas.RouteRegistry) *LiveHandler {
	return &LiveHandler{routeReg: routeReg}
}

// @Summary		List registered BaaS endpoints
// @Description	Get list of all registered BaaS endpoints
// @Tags			admin
// @Produce		json
// @Success		200	{object}	map[string]interface{}
// @Router			/api/v1/admin/endpoints [get]
// @Security		BearerAuth
func (h *LiveHandler) ListRegistered(c *gin.Context) {
	if h.routeReg == nil {
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"total": 0, "routes": []map[string]any{}}})
		return
	}

	endpoints := h.routeReg.Endpoints()

	routes := make([]map[string]any, 0, len(endpoints))
	for key, id := range endpoints {
		routes = append(routes, gin.H{
			"route_key":   key,
			"endpoint_id": id.String(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"total":  len(routes),
			"routes": routes,
		},
	})
}
