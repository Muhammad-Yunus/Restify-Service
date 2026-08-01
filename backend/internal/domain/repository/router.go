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
type RouterGroup struct {
	basePath   string
	middleware []Middleware
}

// NewRouterGroup creates a route group with shared middleware.
func NewRouterGroup(basePath string, middleware ...Middleware) *RouterGroup {
	return &RouterGroup{
		basePath:   basePath,
		middleware: middleware,
	}
}

// Middleware is a function that wraps an HTTP handler.
type Middleware func(http.HandlerFunc) http.HandlerFunc
