package entity

import (
	"testing"

	"github.com/google/uuid"
)

func TestWorkspaceGenerateSlug(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "simple", in: "My Workspace", want: "my-workspace"},
		{name: "special characters", in: "Hello, World!!!", want: "hello-world"},
		{name: "collapsed dashes", in: "--Space--", want: "space"},
		{name: "already slug", in: "my-workspace", want: "my-workspace"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := &Workspace{Name: tc.in}
			w.GenerateSlug()

			if w.Slug != tc.want {
				t.Errorf("GenerateSlug(%q) = %q, want %q", tc.in, w.Slug, tc.want)
			}
		})
	}
}

func TestWorkspaceValidate(t *testing.T) {
	valid := Workspace{Name: "Acme", Slug: "acme", OwnerID: uuid.New()}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid workspace rejected: %v", err)
	}

	invalid := Workspace{Name: "Acme"}
	if err := invalid.Validate(); err == nil {
		t.Error("workspace without slug or owner accepted")
	}
}
