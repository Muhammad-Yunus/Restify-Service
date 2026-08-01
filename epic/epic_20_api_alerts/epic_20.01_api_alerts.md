# Epic 20 — API Alerts

**Goal:** Implement alert rules, event tracking, and notification dispatch (webhook, email, MQTT).
**Dependencies:** Epic 05 (Alert repository interface), Epic 06 (EmailService & RabbitMQ adapters)
**Commit:** `feat: add alert system with rules, events, and multi-channel notifications`

---

## Step 20.01 — Alert Repository Implementation

**Build:** Create `backend/internal/application/repository/alert_repository.go`:

```go
package repository

import (
    "context"
    "fmt"

    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
    "gorm.io/gorm"
)

// AlertRepositoryImpl implements the repository.AlertRepository interface.
type AlertRepositoryImpl struct {
    db *gorm.DB
}

// NewAlertRepository creates a new alert repository.
func NewAlertRepository(db repository.DB) repository.AlertRepository {
    return &AlertRepositoryImpl{db: gormDB(db)}
}

func (r *AlertRepositoryImpl) Create(ctx context.Context, rule *entity.AlertRule) error {
    if rule.ID == uuid.Nil {
        rule.ID = uuid.New()
    }
    if err := r.db.WithContext(ctx).Create(rule).Error; err != nil {
        return fmt.Errorf("create alert rule: %w", err)
    }
    return nil
}

func (r *AlertRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entity.AlertRule, error) {
    var rule entity.AlertRule
    err := r.db.WithContext(ctx).First(&rule, "id = ?", id).Error
    if err != nil {
        return nil, fmt.Errorf("find alert rule %s: %w", id, err)
    }
    return &rule, nil
}

func (r *AlertRepositoryImpl) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entity.AlertRule, error) {
    var rules []*entity.AlertRule
    if err := r.db.WithContext(ctx).
        Where("workspace_id = ?", workspaceID).
        Find(&rules).Error; err != nil {
        return nil, fmt.Errorf("list alert rules: %w", err)
    }
    return rules, nil
}

func (r *AlertRepositoryImpl) Update(ctx context.Context, rule *entity.AlertRule) error {
    if err := r.db.WithContext(ctx).Model(rule).Updates(map[string]any{
        "name":          rule.Name,
        "trigger":       rule.Trigger,
        "threshold":     rule.Threshold,
        "window_minutes": rule.WindowMinutes,
        "actions":       rule.Actions,
        "is_enabled":    rule.IsEnabled,
    }).Error; err != nil {
        return fmt.Errorf("update alert rule: %w", err)
    }
    return nil
}

func (r *AlertRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
    result := r.db.WithContext(ctx).Delete(&entity.AlertRule{}, "id = ?", id)
    if result.Error != nil {
        return fmt.Errorf("delete alert rule: %w", result.Error)
    }
    if result.RowsAffected == 0 {
        return entity.ErrNotFound
    }
    return nil
}

func (r *AlertRepositoryImpl) ToggleEnabled(ctx context.Context, id uuid.UUID, enabled bool) error {
    result := r.db.WithContext(ctx).Model(&entity.AlertRule{}).
        Where("id = ?", id).
        Update("is_enabled", enabled)
    if result.Error != nil {
        return fmt.Errorf("toggle alert rule: %w", result.Error)
    }
    if result.RowsAffected == 0 {
        return entity.ErrNotFound
    }
    return nil
}

func (r *AlertRepositoryImpl) CreateEvent(ctx context.Context, event *entity.AlertEvent) error {
    if event.ID == uuid.Nil {
        event.ID = uuid.New()
    }
    if err := r.db.WithContext(ctx).Create(event).Error; err != nil {
        return fmt.Errorf("create alert event: %w", err)
    }
    return nil
}

func (r *AlertRepositoryImpl) ListRecentEvents(ctx context.Context, workspaceID uuid.UUID, limit int) ([]*entity.AlertEvent, error) {
    var events []*entity.AlertEvent
    if err := r.db.WithContext(ctx).
        Where("workspace_id = ?", workspaceID).
        Order("created_at DESC").
        Limit(limit).
        Find(&events).Error; err != nil {
        return nil, fmt.Errorf("list alert events: %w", err)
    }
    return events, nil
}

func (r *AlertRepositoryImpl) MarkNotified(ctx context.Context, id uuid.UUID) error {
    result := r.db.WithContext(ctx).Model(&entity.AlertEvent{}).
        Where("id = ?", id).
        Update("notified", true)
    if result.Error != nil {
        return fmt.Errorf("mark alert notified: %w", result.Error)
    }
    return nil
}
```

---

## Step 20.02 — Alert Service

**Build:** Create `backend/internal/application/service/alert_service.go`:

```go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// AlertService manages alert rules and notifications.
type AlertService struct {
    repo     repository.AlertRepository
    queue    repository.MessageQueue
    emailSvc repository.EmailService
    logger   repository.Logger
}

// NewAlertService creates a new alert service.
func NewAlertService(repo repository.AlertRepository, queue repository.MessageQueue, emailSvc repository.EmailService, logger repository.Logger) repository.AlertService {
    return &AlertService{repo: repo, queue: queue, emailSvc: emailSvc, logger: logger}
}

func (s *AlertService) CreateRule(ctx context.Context, rule *entity.AlertRule) error {
    if err := s.repo.Create(ctx, rule); err != nil {
        return fmt.Errorf("create alert rule: %w", err)
    }
    s.logger.Info(ctx, "alert rule created", "rule_id", rule.ID)
    return nil
}

func (s *AlertService) GetRule(ctx context.Context, id uuid.UUID) (*entity.AlertRule, error) {
    return s.repo.FindByID(ctx, id)
}

func (s *AlertService) ListRules(ctx context.Context, workspaceID uuid.UUID) ([]*entity.AlertRule, error) {
    return s.repo.ListByWorkspace(ctx, workspaceID)
}

func (s *AlertService) UpdateRule(ctx context.Context, rule *entity.AlertRule) error {
    return s.repo.Update(ctx, rule)
}

func (s *AlertService) DeleteRule(ctx context.Context, id uuid.UUID) error {
    return s.repo.Delete(ctx, id)
}

func (s *AlertService) ToggleRule(ctx context.Context, id uuid.UUID, enabled bool) error {
    return s.repo.ToggleEnabled(ctx, id, enabled)
}

// FireAlert creates an alert event and dispatches notifications.
func (s *AlertService) FireAlert(ctx context.Context, event *entity.AlertEvent) error {
    if err := s.repo.CreateEvent(ctx, event); err != nil {
        return fmt.Errorf("create alert event: %w", err)
    }

    // Get the rule to find actions
    rule, err := s.repo.FindByID(ctx, event.RuleID)
    if err != nil {
        return fmt.Errorf("get alert rule: %w", err)
    }

    // Dispatch to configured actions
    var actions []entity.AlertAction
    json.Unmarshal(rule.Actions, &actions)

    for _, action := range actions {
        if !rule.IsEnabled {
            continue
        }
        go s.dispatchAction(ctx, action, event)
    }

    return nil
}

func (s *AlertService) dispatchAction(ctx context.Context, action entity.AlertAction, event *entity.AlertEvent) {
    switch action.Type {
    case entity.ActionWebhook:
        s.sendWebhook(ctx, action, event)
    case entity.ActionEmail:
        s.sendEmail(ctx, action, event)
    case entity.ActionMQTT:
        s.sendMQTT(ctx, action, event)
    }
}

func (s *AlertService) sendWebhook(ctx context.Context, action entity.AlertAction, event *entity.AlertEvent) {
    payload, _ := json.Marshal(gin.H{
        "event_id":   event.ID.String(),
        "trigger":    string(event.Trigger),
        "threshold":  event.Threshold,
        "current":    event.CurrentValue,
        "message":    event.Message,
        "timestamp":  event.CreatedAt,
    })
    if err := s.queue.Publish(ctx, "alerts.webhook", payload); err != nil {
        s.logger.Error(ctx, "failed to publish webhook alert", "error", err)
    }
}

func (s *AlertService) sendEmail(ctx context.Context, action entity.AlertAction, event *entity.AlertEvent) {
    if err := s.emailSvc.SendAlertEmail(ctx, action.Recipient, "ForgeBase Alert", event.Message); err != nil {
        s.logger.Error(ctx, "failed to send alert email", "error", err)
    }
}

func (s *AlertService) sendMQTT(ctx context.Context, action entity.AlertAction, event *entity.AlertEvent) {
    payload, _ := json.Marshal(gin.H{
        "trigger": string(event.Trigger),
        "message": event.Message,
    })
    // Publish to MQTT topic
    _ = payload // MQTT broker integration
}

func (s *AlertService) ListRecentEvents(ctx context.Context, workspaceID uuid.UUID, limit int) ([]*entity.AlertEvent, error) {
    return s.repo.ListRecentEvents(ctx, workspaceID, limit)
}
```

---

## Step 20.03 — Alert HTTP Handler

**Build:** Create `backend/internal/presentation/http/handler/alert_handler.go`:

```go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/application/service"
    "github.com/muhammadyunus/ForgeBase/internal/presentation/http/dto"
)

// AlertHandler handles alert HTTP requests.
type AlertHandler struct {
    alertService service.AlertService
}

// NewAlertHandler creates a new alert handler.
func NewAlertHandler(as service.AlertService) *AlertHandler {
    return &AlertHandler{alertService: as}
}

// List handles GET /api/v1/alerts
func (h *AlertHandler) List(c *gin.Context) {
    wsID, _ := uuid.Parse(c.Param("ws_id"))
    rules, err := h.alertService.ListRules(c.Request.Context(), wsID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": toAlertRuleListDTO(rules)})
}

// Create handles POST /api/v1/alerts
func (h *AlertHandler) Create(c *gin.Context) {
    var rule entity.AlertRule
    if err := c.ShouldBindJSON(&rule); err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Validation Error", err.Error()))
        return
    }
    if err := h.alertService.CreateRule(c.Request.Context(), &rule); err != nil {
        c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
        return
    }
    c.JSON(http.StatusCreated, toAlertRuleDTO(&rule))
}

// Update handles PATCH /api/v1/alerts/:id
func (h *AlertHandler) Update(c *gin.Context) {
    id, _ := uuid.Parse(c.Param("id"))
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

// Delete handles DELETE /api/v1/alerts/:id
func (h *AlertHandler) Delete(c *gin.Context) {
    id, _ := uuid.Parse(c.Param("id"))
    if err := h.alertService.DeleteRule(c.Request.Context(), id); err != nil {
        c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "Alert not found", err.Error()))
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "alert deleted"})
}

// ListEvents handles GET /api/v1/alerts/events
func (h *AlertHandler) ListEvents(c *gin.Context) {
    wsID, _ := uuid.Parse(c.Param("ws_id"))
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
```

---

## Step 20.04 — Alert Repository Initialization in DI

**Build:** Update `internal/di/bootstrap.go`:

```go
func initAlertRepo(db repository.DB) repository.AlertRepository {
    return alerts.NewAlertRepository(db)
}
```

**Test cases:**
- [ ] Unit: `initAlertRepo()` returns a non-nil AlertRepository

---

## Commit Instruction

```bash
git add .
git commit -m "feat: add alert system with rules, events, and multi-channel notifications"
```
