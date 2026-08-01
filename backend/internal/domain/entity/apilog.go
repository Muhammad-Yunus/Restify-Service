package entity

import (
	"time"

	"github.com/google/uuid"
)

// LogLevel represents log severity.
type LogLevel string

const (
	LevelDebug    LogLevel = "DEBUG"
	LevelInfo     LogLevel = "INFO"
	LevelWarn     LogLevel = "WARN"
	LevelError    LogLevel = "ERROR"
	LevelCritical LogLevel = "CRITICAL"
)

// APILog represents a request log entry.
type APILog struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	RequestID   string     `json:"request_id" gorm:"type:uuid;not null;index" validate:"required"`
	WorkspaceID *uuid.UUID `json:"workspace_id,omitempty" gorm:"type:uuid;index"`
	UserID      *uuid.UUID `json:"user_id,omitempty" gorm:"type:uuid;index"`
	Method      string     `json:"method" gorm:"type:varchar(10);not null;index" validate:"required"`
	Path        string     `json:"path" gorm:"type:varchar(1000);not null;index" validate:"required"`
	StatusCode  int        `json:"status_code" gorm:"not null;index" validate:"gte=100,lte=599"`
	LatencyMs   int64      `json:"latency_ms" gorm:"not null"`
	LogLevel    LogLevel   `json:"log_level" gorm:"type:varchar(20);not null" validate:"required"`
	Message     string     `json:"message" gorm:"type:text"`
	Meta        []byte     `json:"meta,omitempty" gorm:"type:jsonb"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime;not null"`
}

// Validate checks log field constraints.
func (l *APILog) Validate() error {
	return validateStruct(l)
}
