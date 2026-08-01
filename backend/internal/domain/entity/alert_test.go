package entity

import (
	"testing"

	"github.com/google/uuid"
)

func TestAlertRuleValidate(t *testing.T) {
	valid := AlertRule{
		Name:        "high error rate",
		WorkspaceID: uuid.New(),
		Trigger:     TriggerErrorRate,
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid alert rule rejected: %v", err)
	}

	invalid := AlertRule{}
	if err := invalid.Validate(); err == nil {
		t.Error("empty alert rule accepted")
	}
}

func TestAlertEventValidate(t *testing.T) {
	valid := AlertEvent{
		RuleID:      uuid.New(),
		WorkspaceID: uuid.New(),
		Trigger:     TriggerErrorRate,
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid alert event rejected: %v", err)
	}

	invalid := AlertEvent{}
	if err := invalid.Validate(); err == nil {
		t.Error("empty alert event accepted")
	}
}
