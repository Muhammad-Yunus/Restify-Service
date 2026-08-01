package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/service"
	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/presentation/http/dto"
)

// WorkspaceHandler handles workspace HTTP requests.
type WorkspaceHandler struct {
	workspaceService service.WorkspaceService
}

// NewWorkspaceHandler creates a new workspace handler.
func NewWorkspaceHandler(ws service.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{workspaceService: ws}
}

// Create handles POST /api/v1/workspaces
func (h *WorkspaceHandler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required,max=255"`
		Description string `json:"description,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Validation Error", err.Error()))
		return
	}

	// In production, get ownerID from auth context
	// For now, generate a default owner ID (will be replaced in next epic)
	ownerID, _ := uuid.Parse(c.GetString("user_id"))

	ws, err := h.workspaceService.Create(c.Request.Context(), req.Name, req.Description, ownerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, toWorkspaceDTO(ws))
}

// List handles GET /api/v1/workspaces
func (h *WorkspaceHandler) List(c *gin.Context) {
	ownerID, _ := uuid.Parse(c.GetString("user_id"))

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	workspaces, total, err := h.workspaceService.List(c.Request.Context(), ownerID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       toWorkspaceListDTO(workspaces),
		"pagination": gin.H{"page": page, "page_size": pageSize, "total": total},
	})
}

// GetByID handles GET /api/v1/workspaces/:id
func (h *WorkspaceHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid ID", "workspace ID must be a UUID"))
		return
	}

	ws, err := h.workspaceService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "Workspace not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, toWorkspaceDTO(ws))
}

// Update handles PATCH /api/v1/workspaces/:id
func (h *WorkspaceHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid ID", ""))
		return
	}

	var updates map[string]any
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid request body", err.Error()))
		return
	}

	ws, err := h.workspaceService.Update(c.Request.Context(), id, updates)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "Workspace not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, toWorkspaceDTO(ws))
}

// Delete handles DELETE /api/v1/workspaces/:id
func (h *WorkspaceHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid ID", ""))
		return
	}

	if err := h.workspaceService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "Workspace not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "workspace deleted"})
}

func toWorkspaceDTO(ws *entity.Workspace) gin.H {
	return gin.H{
		"id":          ws.ID.String(),
		"name":        ws.Name,
		"description": ws.Description,
		"slug":        ws.Slug,
		"owner_id":    ws.OwnerID.String(),
		"is_public":   ws.IsPublic,
		"created_at":  ws.CreatedAt,
		"updated_at":  ws.UpdatedAt,
	}
}

func toWorkspaceListDTO(workspaces []*entity.Workspace) []gin.H {
	result := make([]gin.H, len(workspaces))
	for i, ws := range workspaces {
		result[i] = toWorkspaceDTO(ws)
	}
	return result
}
