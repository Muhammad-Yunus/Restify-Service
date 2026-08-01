package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// AlertRepositoryImpl implements the repository.AlertRepository interface.
type AlertRepositoryImpl struct {
	db *gorm.DB
}

// NewAlertRepository creates a new alert repository.
func NewAlertRepository(db repository.DB, gormDB *gorm.DB) repository.AlertRepository {
	return &AlertRepositoryImpl{db: gormDB}
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
		"name":           rule.Name,
		"trigger":        rule.Trigger,
		"threshold":      rule.Threshold,
		"window_minutes": rule.WindowMinutes,
		"actions":        rule.Actions,
		"is_enabled":     rule.IsEnabled,
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

func (r *AlertRepositoryImpl) Close(ctx context.Context) error {
	return nil
}
