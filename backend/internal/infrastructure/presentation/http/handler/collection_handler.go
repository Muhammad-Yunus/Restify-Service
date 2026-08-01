package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/service"
	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/presentation/http/dto"
)

// CollectionHandler handles collection HTTP requests.
type CollectionHandler struct {
	collectionService service.CollectionService
}

// NewCollectionHandler creates a new collection handler.
func NewCollectionHandler(cs service.CollectionService) *CollectionHandler {
	return &CollectionHandler{collectionService: cs}
}

// Create handles POST /api/v1/workspaces/:ws_id/collections
func (h *CollectionHandler) Create(c *gin.Context) {
	wsID, err := uuid.Parse(c.Param("ws_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid workspace ID", "workspace ID must be a UUID"))
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required,max=255"`
		Description string `json:"description,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Validation Error", err.Error()))
		return
	}

	col, err := h.collectionService.Create(c.Request.Context(), req.Name, req.Description, wsID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, toCollectionDTO(col))
}

// List handles GET /api/v1/workspaces/:ws_id/collections
func (h *CollectionHandler) List(c *gin.Context) {
	wsID, err := uuid.Parse(c.Param("ws_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid workspace ID", "workspace ID must be a UUID"))
		return
	}

	cols, err := h.collectionService.List(c.Request.Context(), wsID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": toCollectionListDTO(cols)})
}

// GetByID handles GET /api/v1/collections/:id
func (h *CollectionHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid ID", "collection ID must be a UUID"))
		return
	}

	col, err := h.collectionService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "Collection not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, toCollectionDTO(col))
}

// Update handles PATCH /api/v1/collections/:id
func (h *CollectionHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid ID", "collection ID must be a UUID"))
		return
	}

	var updates map[string]any
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid request body", err.Error()))
		return
	}

	col, err := h.collectionService.Update(c.Request.Context(), id, updates)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "Collection not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, toCollectionDTO(col))
}

// Delete handles DELETE /api/v1/collections/:id
func (h *CollectionHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid ID", "collection ID must be a UUID"))
		return
	}

	if err := h.collectionService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "Collection not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "collection deleted"})
}

func toCollectionDTO(col *entity.Collection) gin.H {
	return gin.H{
		"id":           col.ID.String(),
		"name":         col.Name,
		"description":  col.Description,
		"slug":         col.Slug,
		"workspace_id": col.WorkspaceID.String(),
		"endpoint_count": len(col.Endpoints),
		"created_at":   col.CreatedAt,
		"updated_at":   col.UpdatedAt,
	}
}

func toCollectionListDTO(cols []*entity.Collection) []gin.H {
	result := make([]gin.H, len(cols))
	for i, col := range cols {
		result[i] = gin.H{
			"id":           col.ID.String(),
			"name":         col.Name,
			"slug":         col.Slug,
			"workspace_id": col.WorkspaceID.String(),
			"endpoint_count": len(col.Endpoints),
			"created_at":   col.CreatedAt,
		}
	}
	return result
}
