package entity

import (
	"testing"

	"github.com/google/uuid"
)

func TestCollectionGenerateSlug(t *testing.T) {
	valid := Collection{Name: "User API"}
	valid.GenerateSlug()

	if valid.Slug != "user-api" {
		t.Errorf("GenerateSlug() = %q, want %q", valid.Slug, "user-api")
	}
}

func TestCollectionValidate(t *testing.T) {
	valid := Collection{Name: "User API", Slug: "user-api", WorkspaceID: uuid.New()}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid collection rejected: %v", err)
	}

	invalid := Collection{Name: "User API"}
	if err := invalid.Validate(); err == nil {
		t.Error("collection without slug or workspace accepted")
	}
}
