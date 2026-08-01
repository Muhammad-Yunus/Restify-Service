package entity

import "testing"

func TestRoleValidate(t *testing.T) {
	valid := Role{Name: RoleDeveloper}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid role rejected: %v", err)
	}

	invalid := Role{}
	if err := invalid.Validate(); err == nil {
		t.Error("empty role name accepted")
	}
}
