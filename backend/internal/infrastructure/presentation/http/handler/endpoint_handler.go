package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/service"
	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/presentation/http/dto"
)

// EndpointHandler handles endpoint HTTP requests.
type EndpointHandler struct {
	endpointService service.EndpointService
}

// NewEndpointHandler creates a new endpoint handler.
func NewEndpointHandler(es service.EndpointService) *EndpointHandler {
	return &EndpointHandler{endpointService: es}
}

// Create handles POST /api/v1/collections/:col_id/endpoints
func (h *EndpointHandler) Create(c *gin.Context) {
	colID, err := uuid.Parse(c.Param("col_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid collection ID", "collection ID must be a UUID"))
		return
	}

	var params map[string]any
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid request body", err.Error()))
		return
	}

	if _, ok := params["name"].(string); !ok {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Missing required field", "name is required"))
		return
	}

	params["collection_id"] = colID

	ep, err := h.endpointService.Create(c.Request.Context(), colID, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, toEndpointDTO(ep))
}

// List handles GET /api/v1/collections/:col_id/endpoints
func (h *EndpointHandler) List(c *gin.Context) {
	colID, err := uuid.Parse(c.Param("col_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid collection ID", "collection ID must be a UUID"))
		return
	}

	eps, err := h.endpointService.List(c.Request.Context(), colID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": toEndpointListDTO(eps)})
}

// GetByID handles GET /api/v1/endpoints/:id
func (h *EndpointHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid ID", "endpoint ID must be a UUID"))
		return
	}

	ep, err := h.endpointService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "Endpoint not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, toEndpointDTO(ep))
}

// Update handles PATCH /api/v1/endpoints/:id
func (h *EndpointHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid ID", "endpoint ID must be a UUID"))
		return
	}

	var updates map[string]any
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid request body", err.Error()))
		return
	}

	ep, err := h.endpointService.Update(c.Request.Context(), id, updates)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "Endpoint not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, toEndpointDTO(ep))
}

// Delete handles DELETE /api/v1/endpoints/:id
func (h *EndpointHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid ID", "endpoint ID must be a UUID"))
		return
	}

	if err := h.endpointService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "Endpoint not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "endpoint deleted"})
}

// Toggle handles POST /api/v1/endpoints/:id/toggle
func (h *EndpointHandler) Toggle(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid ID", "endpoint ID must be a UUID"))
		return
	}

	var req struct {
		Active bool `json:"active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid request body", err.Error()))
		return
	}

	if err := h.endpointService.ToggleActive(c.Request.Context(), id, req.Active); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "endpoint toggled", "active": req.Active})
}

func toEndpointDTO(ep *entity.Endpoint) gin.H {
	return gin.H{
		"id":            ep.ID.String(),
		"collection_id": ep.CollectionID.String(),
		"name":          ep.Name,
		"description":   ep.Description,
		"path":          ep.Path,
		"method":        ep.Method,
		"version":       ep.Version,
		"is_active":     ep.IsActive,
		"db_type":       string(ep.DBType),
		"schema":        ep.Schema,
		"table_name":    ep.TableName,
		"func_name":     ep.FuncName,
		"params":        string(ep.Params),
		"operations":    string(ep.Operations),
		"created_at":    ep.CreatedAt,
		"updated_at":    ep.UpdatedAt,
	}
}

func toEndpointListDTO(eps []*entity.Endpoint) []gin.H {
	result := make([]gin.H, len(eps))
	for i, ep := range eps {
		result[i] = gin.H{
			"id":          ep.ID.String(),
			"name":        ep.Name,
			"path":        ep.Path,
			"method":      ep.Method,
			"version":     ep.Version,
			"is_active":   ep.IsActive,
			"db_type":     string(ep.DBType),
			"created_at":  ep.CreatedAt,
		}
	}
	return result
}
