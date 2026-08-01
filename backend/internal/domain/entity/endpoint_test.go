package entity

import (
	"testing"

	"github.com/google/uuid"
)

func TestEndpointValidate(t *testing.T) {
	valid := Endpoint{
		CollectionID: uuid.New(),
		Name:         "list users",
		Path:         "/users",
		Method:       "GET",
		Version:      "v1",
		DBType:       EndpointTypeTable,
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid endpoint rejected: %v", err)
	}

	invalid := Endpoint{}
	if err := invalid.Validate(); err == nil {
		t.Error("empty endpoint accepted")
	}
}

func TestEndpointSecurityPolicyRoundtrip(t *testing.T) {
	e := &Endpoint{}

	limit := 60
	policy := SecurityPolicy{
		AuthRequired: true,
		AllowedRoles: []string{RoleDeveloper},
		RateLimit:    &limit,
	}

	if err := e.SetSecurityPolicy(policy); err != nil {
		t.Fatalf("SetSecurityPolicy: %v", err)
	}

	got := e.GetSecurityPolicy()

	if !got.AuthRequired {
		t.Error("auth_required not preserved")
	}

	if len(got.AllowedRoles) != 1 || got.AllowedRoles[0] != RoleDeveloper {
		t.Errorf("allowed_roles = %v, want [developer]", got.AllowedRoles)
	}

	if got.RateLimit == nil || *got.RateLimit != limit {
		t.Error("rate_limit not preserved")
	}
}
