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

// @Summary		Create workspace
// @Description	Create a new workspace
// @Tags			workspaces
// @Accept			json
// @Produce		json
// @Param			body	body		object	true	"Workspace data"
// @Success		201		{object}	map[string]interface{}
// @Failure		400		{object}	map[string]interface{}
// @Failure		500		{object}	map[string]interface{}
// @Router			/api/v1/workspaces [post]
// @Security		BearerAuth
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

// @Summary		List workspaces
// @Description	List workspaces with pagination
// @Tags			workspaces
// @Produce		json
// @Param			page		query		int	false	"Page number"	default(1)
// @Param			page_size	query		int	false	"Page size"		default(20)
// @Success		200			{object}	map[string]interface{}
// @Failure		500			{object}	map[string]interface{}
// @Router			/api/v1/workspaces [get]
// @Security		BearerAuth
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

// @Summary		Get workspace by ID
// @Description	Get a workspace by ID
// @Tags			workspaces
// @Produce		json
// @Param			id	path		string	true	"Workspace ID"	format(uuid)
// @Success		200	{object}	map[string]interface{}
// @Failure		400	{object}	map[string]interface{}
// @Failure		404	{object}	map[string]interface{}
// @Router			/api/v1/workspaces/:id [get]
// @Security		BearerAuth
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

// @Summary		Update workspace
// @Description	Update workspace information
// @Tags			workspaces
// @Accept			json
// @Produce		json
// @Param			id		path		string	true	"Workspace ID"	format(uuid)
// @Param			body	body		object	true	"Workspace updates"
// @Success		200		{object}	map[string]interface{}
// @Failure		400		{object}	map[string]interface{}
// @Failure		404		{object}	map[string]interface{}
// @Router			/api/v1/workspaces/:id [patch]
// @Security		BearerAuth
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

// @Summary		Delete workspace
// @Description	Delete a workspace by ID (admin only)
// @Tags			workspaces
// @Produce		json
// @Param			id	path		string	true	"Workspace ID"	format(uuid)
// @Success		200	{object}	map[string]string
// @Failure		400	{object}	map[string]interface{}
// @Failure		404	{object}	map[string]interface{}
// @Router			/api/v1/workspaces/:id [delete]
// @Security		BearerAuth
// @Security		AdminRole
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
