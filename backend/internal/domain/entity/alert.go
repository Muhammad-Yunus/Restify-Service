package entity

import (
	"time"

	"github.com/google/uuid"
)

// AlertTrigger defines what condition triggers an alert.
type AlertTrigger string

const (
	TriggerErrorRate AlertTrigger = "error_rate"
	TriggerLatency   AlertTrigger = "latency"
	TriggerAuthFail  AlertTrigger = "auth_failure_burst"
	TriggerDBDown    AlertTrigger = "db_connection_loss"
	TriggerRateLimit AlertTrigger = "rate_limit_exceeded"
)

// AlertActionType defines how an alert is delivered.
type AlertActionType string

const (
	ActionWebhook AlertActionType = "webhook"
	ActionEmail   AlertActionType = "email"
	ActionMQTT    AlertActionType = "mqtt"
)

// AlertRule represents an alert configuration.
type AlertRule struct {
	ID            uuid.UUID    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name          string       `json:"name" gorm:"type:varchar(255);not null" validate:"required"`
	WorkspaceID   uuid.UUID    `json:"workspace_id" gorm:"type:uuid;not null;index" validate:"required"`
	EndpointID    *uuid.UUID   `json:"endpoint_id,omitempty" gorm:"type:uuid;index"`
	Trigger       AlertTrigger `json:"trigger" gorm:"type:varchar(50);not null" validate:"required"`
	Threshold     float64      `json:"threshold" gorm:"not null"`
	WindowMinutes int          `json:"window_minutes" gorm:"not null"`
	Actions       []byte       `json:"actions,omitempty" gorm:"type:jsonb"`
	IsEnabled     bool         `json:"is_enabled" gorm:"default:true"`
	CreatedAt     time.Time    `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt     time.Time    `json:"updated_at" gorm:"autoUpdateTime;not null"`
}

// Validate checks alert rule field constraints.
func (a *AlertRule) Validate() error {
	return validateStruct(a)
}

// AlertEvent represents a fired alert notification.
type AlertEvent struct {
	ID           uuid.UUID    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	RuleID       uuid.UUID    `json:"rule_id" gorm:"type:uuid;not null;index" validate:"required"`
	WorkspaceID  uuid.UUID    `json:"workspace_id" gorm:"type:uuid;not null;index" validate:"required"`
	Trigger      AlertTrigger `json:"trigger" gorm:"type:varchar(50);not null" validate:"required"`
	CurrentValue float64      `json:"current_value" gorm:"not null"`
	Threshold    float64      `json:"threshold" gorm:"not null"`
	Message      string       `json:"message" gorm:"type:text"`
	Notified     bool         `json:"notified" gorm:"default:false"`
	CreatedAt    time.Time    `json:"created_at" gorm:"autoCreateTime;not null"`
}

// Validate checks alert event field constraints.
func (e *AlertEvent) Validate() error {
	return validateStruct(e)
}
