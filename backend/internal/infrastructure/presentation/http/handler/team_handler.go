package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/service"
	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/presentation/http/dto"
)

// TeamHandler handles team HTTP requests.
type TeamHandler struct {
	teamService service.TeamService
}

// NewTeamHandler creates a new team handler.
func NewTeamHandler(ts service.TeamService) *TeamHandler {
	return &TeamHandler{teamService: ts}
}

// Create handles POST /api/v1/workspaces/:ws_id/teams
func (h *TeamHandler) Create(c *gin.Context) {
	wsID, err := uuid.Parse(c.Param("ws_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid workspace ID", "workspace ID must be a UUID"))
		return
	}

	var req struct {
		Name string `json:"name" binding:"required,max=255"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Validation Error", err.Error()))
		return
	}

	team, err := h.teamService.Create(c.Request.Context(), req.Name, wsID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, toTeamDTO(team))
}

// GetByID handles GET /api/v1/teams/:id
func (h *TeamHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid ID", "team ID must be a UUID"))
		return
	}

	team, err := h.teamService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "Team not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, toTeamDTO(team))
}

// AddMember handles POST /api/v1/teams/:id/members
func (h *TeamHandler) AddMember(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid team ID", "team ID must be a UUID"))
		return
	}

	var req struct {
		UserID string `json:"user_id" binding:"required,uuid"`
		Role   string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Validation Error", err.Error()))
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid user ID", "user ID must be a UUID"))
		return
	}

	if err := h.teamService.AddMember(c.Request.Context(), teamID, userID, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "member added"})
}

// ListMembers handles GET /api/v1/teams/:id/members
func (h *TeamHandler) ListMembers(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid team ID", "team ID must be a UUID"))
		return
	}

	members, err := h.teamService.ListMembers(c.Request.Context(), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": toMemberListDTO(members)})
}

// RemoveMember handles DELETE /api/v1/teams/:id/members/:user_id
func (h *TeamHandler) RemoveMember(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid team ID", "team ID must be a UUID"))
		return
	}

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid user ID", "user ID must be a UUID"))
		return
	}

	if err := h.teamService.RemoveMember(c.Request.Context(), teamID, userID); err != nil {
		c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "Member not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member removed"})
}

func toTeamDTO(team *entity.Team) gin.H {
	return gin.H{
		"id":           team.ID.String(),
		"name":         team.Name,
		"workspace_id": team.WorkspaceID.String(),
		"created_at":   team.CreatedAt,
	}
}

func toMemberListDTO(members []*entity.TeamMember) []gin.H {
	result := make([]gin.H, len(members))
	for i, m := range members {
		result[i] = gin.H{
			"id":        m.ID.String(),
			"user_id":   m.UserID.String(),
			"role":      m.Role,
			"joined_at": m.JoinedAt,
		}
		if m.User != nil {
			result[i]["email"] = m.User.Email
			if m.User.FullName != nil {
				result[i]["full_name"] = *m.User.FullName
			}
		}
	}
	return result
}
