package dto

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ProblemDetail creates an RFC 7807 Problem Details response.
func ProblemDetail(status int, title string, detail string) gin.H {
	return gin.H{
		"type":     fmt.Sprintf("https://ForgeBase.api/errors/%s", strings.ToLower(strings.ReplaceAll(title, " ", "-"))),
		"title":    title,
		"status":   status,
		"detail":   detail,
		"instance": http.StatusText(status),
	}
}

// ProblemDetailWithPath creates an RFC 7807 Problem Details response with a specific path.
func ProblemDetailWithPath(status int, title string, detail string, path string) gin.H {
	resp := ProblemDetail(status, title, detail)
	resp["instance"] = path

	return resp
}

// ProblemDetailWithErrors creates a Problem Details response with validation errors.
func ProblemDetailWithErrors(status int, title string, detail string, errors []FieldError) gin.H {
	resp := ProblemDetail(status, title, detail)
	resp["errors"] = errors

	return resp
}

// FieldError represents a single validation error.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
