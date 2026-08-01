package repository

import "net/http"

// HTTPRouter defines the HTTP routing interface.
type HTTPRouter interface {
	// Group creates a route group with middleware.
	Group(basePath string, middleware ...Middleware) *RouterGroup

	// Handle registers a handler for a route.
	Handle(method, path string, handler http.HandlerFunc, middleware ...Middleware)

	// ServeHTTP implements http.Handler.
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// RouterGroup is a group of routes with shared middleware.
//
//nolint:unused // Fields are populated by the router implementation in Epic 13.
type RouterGroup struct {
	basePath   string
	middleware []Middleware
}

// Middleware is a function that wraps an HTTP handler.
type Middleware func(http.HandlerFunc) http.HandlerFunc
