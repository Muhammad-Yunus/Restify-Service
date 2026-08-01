package entity

import (
	"testing"

	"github.com/google/uuid"
)

func TestTeamValidate(t *testing.T) {
	valid := Team{Name: "Core", WorkspaceID: uuid.New()}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid team rejected: %v", err)
	}

	invalid := Team{}
	if err := invalid.Validate(); err == nil {
		t.Error("empty team accepted")
	}
}

func TestTeamMemberValidate(t *testing.T) {
	valid := TeamMember{TeamID: uuid.New(), UserID: uuid.New()}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid team member rejected: %v", err)
	}

	invalid := TeamMember{}
	if err := invalid.Validate(); err == nil {
		t.Error("empty team member accepted")
	}
}
