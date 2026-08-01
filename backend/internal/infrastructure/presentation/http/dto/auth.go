package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// validateStruct runs struct validation and wraps any error with context.
func validateStruct(v any) error {
	if err := validate.Struct(v); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

// RegisterRequest is the request body for user registration.
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=128"`
	FullName string `json:"full_name,omitempty" validate:"max=255"`
}

// Validate checks the register request fields.
func (r *RegisterRequest) Validate() error {
	return validateStruct(r)
}

// LoginRequest is the request body for user login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Validate checks the login request fields.
func (r *LoginRequest) Validate() error {
	return validateStruct(r)
}

// RefreshRequest is the request body for token refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Validate checks the refresh request fields.
func (r *RefreshRequest) Validate() error {
	return validateStruct(r)
}

// AuthResponse is the response body for auth operations.
type AuthResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresIn    int           `json:"expires_in"`
}

// UserResponse is a sanitized user representation.
type UserResponse struct {
	ID       string   `json:"id"`
	Email    string   `json:"email"`
	FullName *string  `json:"full_name,omitempty"`
	Roles    []string `json:"roles"`
}
