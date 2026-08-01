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

// AlertHandler handles alert HTTP requests.
type AlertHandler struct {
	alertService service.AlertService
}

// NewAlertHandler creates a new alert handler.
func NewAlertHandler(alertService service.AlertService) *AlertHandler {
	return &AlertHandler{alertService: alertService}
}

// @Summary		List alert rules
// @Description	List all alert rules for a workspace
// @Tags			alerts
// @Produce		json
// @Param			workspace_id	path		string	true	"Workspace ID"	format(uuid)
// @Success		200				{object}	map[string]interface{}
// @Failure		400				{object}	map[string]interface{}
// @Failure		500				{object}	map[string]interface{}
// @Router			/api/v1/alerts/:workspace_id [get]
// @Security		BearerAuth
func (h *AlertHandler) List(c *gin.Context) {
	wsID, err := uuid.Parse(c.Param("workspace_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid workspace ID", err.Error()))
		return
	}
	rules, err := h.alertService.ListRules(c.Request.Context(), wsID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toAlertRuleListDTO(rules)})
}

// @Summary		Create alert rule
// @Description	Create a new alert rule for a workspace
// @Tags			alerts
// @Accept			json
// @Produce		json
// @Param			workspace_id	path		string				true	"Workspace ID"	format(uuid)
// @Param			body			body		entity.AlertRule	true	"Alert rule data"
// @Success		201				{object}	map[string]interface{}
// @Failure		400				{object}	map[string]interface{}
// @Failure		500				{object}	map[string]interface{}
// @Router			/api/v1/alerts/:workspace_id [post]
// @Security		BearerAuth
func (h *AlertHandler) Create(c *gin.Context) {
	wsID, err := uuid.Parse(c.Param("workspace_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid workspace ID", err.Error()))
		return
	}
	var rule entity.AlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Validation Error", err.Error()))
		return
	}
	rule.WorkspaceID = wsID

	if err := h.alertService.CreateRule(c.Request.Context(), &rule); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}
	c.JSON(http.StatusCreated, toAlertRuleDTO(&rule))
}

// @Summary		Update alert rule
// @Description	Update an alert rule
// @Tags			alerts
// @Accept			json
// @Produce		json
// @Param			workspace_id	path		string				true	"Workspace ID"	format(uuid)
// @Param			id				path		string				true	"Alert Rule ID"	format(uuid)
// @Param			body			body		entity.AlertRule	true	"Alert rule data"
// @Success		200				{object}	map[string]interface{}
// @Failure		400				{object}	map[string]interface{}
// @Failure		500				{object}	map[string]interface{}
// @Router			/api/v1/alerts/:workspace_id/:id [patch]
// @Security		BearerAuth
func (h *AlertHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid alert ID", err.Error()))
		return
	}
	var rule entity.AlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Validation Error", err.Error()))
		return
	}
	rule.ID = id

	if err := h.alertService.UpdateRule(c.Request.Context(), &rule); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, toAlertRuleDTO(&rule))
}

// @Summary		Delete alert rule
// @Description	Delete an alert rule
// @Tags			alerts
// @Produce		json
// @Param			workspace_id	path		string	true	"Workspace ID"	format(uuid)
// @Param			id				path		string	true	"Alert Rule ID"	format(uuid)
// @Success		200				{object}	map[string]string
// @Failure		400				{object}	map[string]interface{}
// @Failure		404				{object}	map[string]interface{}
// @Router			/api/v1/alerts/:workspace_id/:id [delete]
// @Security		BearerAuth
func (h *AlertHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid alert ID", err.Error()))
		return
	}
	if err := h.alertService.DeleteRule(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "Alert not found", err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "alert deleted"})
}

// @Summary		Toggle alert rule
// @Description	Toggle an alert rule's enabled/disabled status
// @Tags			alerts
// @Accept			json
// @Produce		json
// @Param			workspace_id	path		string	true	"Workspace ID"	format(uuid)
// @Param			id				path		string	true	"Alert Rule ID"	format(uuid)
// @Param			body			body		object	true	"Enabled status"
// @Success		200				{object}	map[string]interface{}
// @Failure		400				{object}	map[string]interface{}
// @Failure		500				{object}	map[string]interface{}
// @Router			/api/v1/alerts/:workspace_id/:id/toggle [put]
// @Security		BearerAuth
func (h *AlertHandler) Toggle(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid alert ID", err.Error()))
		return
	}
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Validation Error", err.Error()))
		return
	}
	if err := h.alertService.ToggleRule(c.Request.Context(), id, body.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "alert toggled", "enabled": body.Enabled})
}

// @Summary		List alert events
// @Description	List recent alert events for a workspace
// @Tags			alerts
// @Produce		json
// @Param			workspace_id	path		string	true	"Workspace ID"		format(uuid)
// @Param			limit			query		int		false	"Number of events"	default(20)
// @Success		200				{object}	map[string]interface{}
// @Failure		400				{object}	map[string]interface{}
// @Failure		500				{object}	map[string]interface{}
// @Router			/api/v1/alerts/:workspace_id/events [get]
// @Security		BearerAuth
func (h *AlertHandler) ListEvents(c *gin.Context) {
	wsID, err := uuid.Parse(c.Param("workspace_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid workspace ID", err.Error()))
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	events, err := h.alertService.ListRecentEvents(c.Request.Context(), wsID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toAlertEventListDTO(events)})
}

func toAlertRuleDTO(rule *entity.AlertRule) gin.H {
	return gin.H{
		"id":             rule.ID.String(),
		"name":           rule.Name,
		"workspace_id":   rule.WorkspaceID.String(),
		"endpoint_id":    rule.EndpointID,
		"trigger":        string(rule.Trigger),
		"threshold":      rule.Threshold,
		"window_minutes": rule.WindowMinutes,
		"is_enabled":     rule.IsEnabled,
		"created_at":     rule.CreatedAt,
		"updated_at":     rule.UpdatedAt,
	}
}

func toAlertRuleListDTO(rules []*entity.AlertRule) []gin.H {
	result := make([]gin.H, len(rules))
	for i, r := range rules {
		result[i] = toAlertRuleDTO(r)
	}
	return result
}

func toAlertEventListDTO(events []*entity.AlertEvent) []gin.H {
	result := make([]gin.H, len(events))
	for i, e := range events {
		result[i] = gin.H{
			"id":            e.ID.String(),
			"rule_id":       e.RuleID.String(),
			"trigger":       string(e.Trigger),
			"current_value": e.CurrentValue,
			"threshold":     e.Threshold,
			"message":       e.Message,
			"notified":      e.Notified,
			"created_at":    e.CreatedAt,
		}
	}
	return result
}
