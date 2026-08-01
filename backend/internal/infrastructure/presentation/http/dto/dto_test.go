package dto

import (
	"net/http"
	"testing"
)

func TestRegisterRequestValidateAcceptsValidRequest(t *testing.T) {
	req := RegisterRequest{
		Email:    "user@example.com",
		Password: "S3curePass!",
		FullName: "John Doe",
	}

	if err := req.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestRegisterRequestValidateRejectsInvalidEmail(t *testing.T) {
	req := RegisterRequest{
		Email:    "not-an-email",
		Password: "S3curePass!",
	}

	if err := req.Validate(); err == nil {
		t.Fatal("Validate() expected error for invalid email")
	}
}

func TestRegisterRequestValidateRejectsShortPassword(t *testing.T) {
	req := RegisterRequest{
		Email:    "user@example.com",
		Password: "short",
	}

	if err := req.Validate(); err == nil {
		t.Fatal("Validate() expected error for short password")
	}
}

func TestLoginRequestValidateRejectsMissingPassword(t *testing.T) {
	req := LoginRequest{Email: "user@example.com"}

	if err := req.Validate(); err == nil {
		t.Fatal("Validate() expected error for missing password")
	}
}

func TestRefreshRequestValidateRejectsEmptyToken(t *testing.T) {
	req := RefreshRequest{}

	if err := req.Validate(); err == nil {
		t.Fatal("Validate() expected error for empty refresh token")
	}
}

func TestProblemDetailReturnsCorrectStructure(t *testing.T) {
	resp := ProblemDetail(http.StatusBadRequest, "Validation Error", "email is required")

	if resp["status"] != http.StatusBadRequest {
		t.Errorf("status = %v, want %d", resp["status"], http.StatusBadRequest)
	}

	if resp["title"] != "Validation Error" {
		t.Errorf("title = %v, want Validation Error", resp["title"])
	}

	if resp["detail"] != "email is required" {
		t.Errorf("detail = %v, want email is required", resp["detail"])
	}

	if resp["instance"] != "Bad Request" {
		t.Errorf("instance = %v, want Bad Request", resp["instance"])
	}
}

func TestProblemDetailTypeURLFollowsRFC7807Format(t *testing.T) {
	resp := ProblemDetail(http.StatusBadRequest, "Email already registered", "duplicate email")

	typ, ok := resp["type"].(string)
	if !ok {
		t.Fatalf("type = %v, want string", resp["type"])
	}

	if want := "https://ForgeBase.api/errors/email-already-registered"; typ != want {
		t.Errorf("type = %q, want %q", typ, want)
	}
}

func TestProblemDetailWithPathSetsInstance(t *testing.T) {
	resp := ProblemDetailWithPath(http.StatusNotFound, "Not Found", "record not found", "/api/v1/users/42")

	if resp["instance"] != "/api/v1/users/42" {
		t.Errorf("instance = %v, want /api/v1/users/42", resp["instance"])
	}
}

func TestProblemDetailWithErrorsIncludesErrorsArray(t *testing.T) {
	fieldErrors := []FieldError{
		{Field: "email", Message: "required"},
	}

	resp := ProblemDetailWithErrors(http.StatusBadRequest, "Validation Error", "invalid input", fieldErrors)

	errors, ok := resp["errors"].([]FieldError)
	if !ok {
		t.Fatalf("errors = %v, want []FieldError", resp["errors"])
	}

	if len(errors) != 1 {
		t.Fatalf("len(errors) = %d, want 1", len(errors))
	}

	if errors[0].Field != "email" || errors[0].Message != "required" {
		t.Errorf("errors[0] = %+v, want email/required", errors[0])
	}
}
