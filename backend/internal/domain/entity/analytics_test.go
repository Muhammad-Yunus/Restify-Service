package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAnalyticsMetricValidate(t *testing.T) {
	valid := AnalyticsMetric{
		WorkspaceID: uuid.New(),
		MetricName:  "requests_total",
		PeriodStart: time.Now().Add(-time.Hour),
		PeriodEnd:   time.Now(),
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid analytics metric rejected: %v", err)
	}

	invalid := AnalyticsMetric{}
	if err := invalid.Validate(); err == nil {
		t.Error("empty analytics metric accepted")
	}
}
