package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
	domservice "github.com/muhammadyunus/Restify-Service/internal/domain/service"
)

// AlertService manages alert rules and notifications.
type AlertService struct {
	repo     repository.AlertRepository
	queue    repository.MessageQueue
	emailSvc domservice.EmailService
	logger   repository.Logger
	mqtt     repository.MQTTBroker
}

// NewAlertService creates a new alert service.
func NewAlertService(repo repository.AlertRepository, queue repository.MessageQueue, emailSvc domservice.EmailService, logger repository.Logger, mqtt repository.MQTTBroker) domservice.AlertService {
	return &AlertService{repo: repo, queue: queue, emailSvc: emailSvc, logger: logger, mqtt: mqtt}
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
	if len(rule.Actions) > 0 {
		if err := json.Unmarshal(rule.Actions, &actions); err != nil {
			return fmt.Errorf("unmarshal alert actions: %w", err)
		}
	}

	for _, action := range actions {
		if !rule.IsEnabled {
			continue
		}
		go s.dispatchAction(ctx, action, event, rule.WorkspaceID)
	}

	return nil
}

func (s *AlertService) dispatchAction(ctx context.Context, action entity.AlertAction, event *entity.AlertEvent, workspaceID uuid.UUID) {
	switch action.Type {
	case entity.ActionWebhook:
		s.sendWebhook(ctx, action, event)
	case entity.ActionEmail:
		s.sendEmail(ctx, action, event)
	case entity.ActionMQTT:
		s.sendMQTT(ctx, action, event, workspaceID)
	}
}

func (s *AlertService) sendWebhook(ctx context.Context, action entity.AlertAction, event *entity.AlertEvent) {
	payload, _ := json.Marshal(map[string]any{
		"event_id":  event.ID.String(),
		"trigger":   string(event.Trigger),
		"threshold": event.Threshold,
		"current":   event.CurrentValue,
		"message":   event.Message,
		"timestamp": event.CreatedAt,
	})
	if err := s.queue.Publish(ctx, "alerts.webhook", payload); err != nil {
		s.logger.Error(ctx, "failed to publish webhook alert", "error", err)
	}
}

func (s *AlertService) sendEmail(ctx context.Context, action entity.AlertAction, event *entity.AlertEvent) {
	if err := s.emailSvc.SendAlertEmail(ctx, action.Recipient, "Alert: "+string(event.Trigger), event.Message); err != nil {
		s.logger.Error(ctx, "failed to send alert email", "error", err)
	}
}

func (s *AlertService) sendMQTT(ctx context.Context, action entity.AlertAction, event *entity.AlertEvent, workspaceID uuid.UUID) {
	payload, _ := json.Marshal(map[string]any{
		"trigger":   string(event.Trigger),
		"message":   event.Message,
		"workspace": workspaceID.String(),
	})
	if err := s.mqtt.Publish(ctx, action.Topic, payload, 0, false); err != nil {
		s.logger.Error(ctx, "failed to publish MQTT alert", "error", err)
	}
}

func (s *AlertService) ListRecentEvents(ctx context.Context, workspaceID uuid.UUID, limit int) ([]*entity.AlertEvent, error) {
	return s.repo.ListRecentEvents(ctx, workspaceID, limit)
}
