package router

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// GinRouter implements the repository.HTTPRouter interface.
type GinRouter struct {
	engine *gin.Engine
}

// NewGinRouter creates a new Gin router with default middleware (CORS, logging, recovery).
func NewGinRouter(env string) *GinRouter {
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(gin.Logger())
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-Key"},
		ExposeHeaders:    []string{"Link", "X-Total-Count"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	return &GinRouter{engine: engine}
}

// Group creates a route group with middleware.
func (r *GinRouter) Group(basePath string, middleware ...repository.Middleware) *repository.RouterGroup {
	group := r.engine.Group(basePath)

	for _, m := range middleware {
		group.Use(middlewareToGin(m))
	}

	return repository.NewRouterGroup(basePath, middleware...)
}

// Handle registers a handler for a route.
func (r *GinRouter) Handle(method, path string, handler http.HandlerFunc, middleware ...repository.Middleware) {
	if !supportedMethod(method) {
		panic(fmt.Sprintf("unsupported HTTP method: %s", method))
	}

	handlers := make([]gin.HandlerFunc, 0, len(middleware)+1)

	for _, m := range middleware {
		handlers = append(handlers, middlewareToGin(m))
	}

	handlers = append(handlers, gin.WrapF(handler))

	r.engine.Handle(method, path, handlers...)
}

// ServeHTTP implements http.Handler.
func (r *GinRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.engine.ServeHTTP(w, req)
}

// Engine returns the underlying Gin engine for advanced usage.
func (r *GinRouter) Engine() *gin.Engine {
	return r.engine
}

// middlewareToGin adapts a domain middleware to a Gin handler. The domain
// middleware wraps an http.HandlerFunc; the inner handler advances the Gin
// chain via c.Next.
func middlewareToGin(m repository.Middleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Next()
		})

		m(next)(c.Writer, c.Request)
	}
}

func supportedMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodDelete, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

// Compile-time check.
var _ repository.HTTPRouter = (*GinRouter)(nil)
