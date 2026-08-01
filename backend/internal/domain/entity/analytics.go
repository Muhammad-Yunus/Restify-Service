package entity

import (
	"time"

	"github.com/google/uuid"
)

// AnalyticsMetric represents an aggregated API metric.
type AnalyticsMetric struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	WorkspaceID uuid.UUID  `json:"workspace_id" gorm:"type:uuid;not null;index" validate:"required"`
	EndpointID  *uuid.UUID `json:"endpoint_id,omitempty" gorm:"type:uuid;index"`
	MetricName  string     `json:"metric_name" gorm:"type:varchar(100);not null" validate:"required"`
	MetricValue float64    `json:"metric_value" gorm:"not null"`
	PeriodStart time.Time  `json:"period_start" gorm:"not null;index" validate:"required"`
	PeriodEnd   time.Time  `json:"period_end" gorm:"not null" validate:"required"`
	Labels      []byte     `json:"labels,omitempty" gorm:"type:jsonb"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime;not null"`
}

// Validate checks analytics field constraints.
func (a *AnalyticsMetric) Validate() error {
	return validateStruct(a)
}
