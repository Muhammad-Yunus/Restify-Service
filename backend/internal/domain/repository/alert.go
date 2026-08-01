package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
)

// AlertRepository defines the data access contract for alert rules and events.
type AlertRepository interface {
	// Create inserts a new alert rule.
	Create(ctx context.Context, rule *entity.AlertRule) error

	// FindByID returns an alert rule by UUID.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.AlertRule, error)

	// ListByWorkspace returns all alert rules for a workspace.
	ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entity.AlertRule, error)

	// Update partially updates an alert rule.
	Update(ctx context.Context, rule *entity.AlertRule) error

	// Delete removes an alert rule.
	Delete(ctx context.Context, id uuid.UUID) error

	// ToggleEnabled enables or disables a rule.
	ToggleEnabled(ctx context.Context, id uuid.UUID, enabled bool) error

	// CreateEvent records a fired alert event.
	CreateEvent(ctx context.Context, event *entity.AlertEvent) error

	// ListRecentEvents returns recent alert events for a workspace.
	ListRecentEvents(ctx context.Context, workspaceID uuid.UUID, limit int) ([]*entity.AlertEvent, error)

	// MarkNotified marks an alert event as notified.
	MarkNotified(ctx context.Context, id uuid.UUID) error

	// Close shuts down any resources owned by the repository.
	Close(ctx context.Context) error
}
