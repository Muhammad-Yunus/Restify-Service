package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	httpHandler "github.com/muhammadyunus/Restify-Service/internal/infrastructure/presentation/http/handler"
	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/presentation/http/middleware"
)

// HandlerDeps holds all HTTP handlers.
type HandlerDeps struct {
	AuthHandler       *httpHandler.AuthHandler
	UserHandler       *httpHandler.UserHandler
	WorkspaceHandler  *httpHandler.WorkspaceHandler
	TeamHandler       *httpHandler.TeamHandler
	CollectionHandler *httpHandler.CollectionHandler
	EndpointHandler   *httpHandler.EndpointHandler
	IntrospectHandler *httpHandler.IntrospectHandler
	AuthMW            *middleware.AuthMiddleware
}

// routerDeps holds dependencies needed for route registration.
type routerDeps struct {
	logRepo interface{ Create(context interface{}, entry interface{}) error }
	logger  interface{ Error(ctx interface{}, msg string, v ...interface{}) }
}

// RegisterAll registers all API routes with versioning.
func RegisterAll(r *gin.Engine, deps *HandlerDeps, rateLimiter *middleware.RateLimitMiddleware, logRepo interface{}, logger interface{}) {

	// Health check (public)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "ForgeBase"})
	})

	// Rate limiting (default to 100 requests per minute)
	rateLimitMW := rateLimiter.Limit()

	// Auth routes (public)
	authGroup := r.Group("/api/v1/auth")
	{
		authGroup.POST("/register", deps.AuthHandler.Register)
		authGroup.POST("/login", deps.AuthHandler.Login)
		authGroup.POST("/refresh", deps.AuthHandler.Refresh)
		authGroup.POST("/logout", deps.AuthHandler.Logout, deps.AuthMW.RequireAuth())
	}

	// User routes (authenticated)
	userGroup := r.Group("/api/v1/users", deps.AuthMW.RequireAuth(), rateLimitMW)
	{
		userGroup.GET("", deps.UserHandler.List)
		userGroup.GET("/:id", deps.UserHandler.GetByID)
		userGroup.PATCH("/:id", deps.UserHandler.Update)
		userGroup.DELETE("/:id", deps.AuthMW.RequireRole("administrator"), deps.UserHandler.Delete)
	}

	// Workspace routes (authenticated)
	wsGroup := r.Group("/api/v1/workspaces", deps.AuthMW.RequireAuth(), rateLimitMW)
	{
		wsGroup.GET("", deps.WorkspaceHandler.List)
		wsGroup.POST("", deps.WorkspaceHandler.Create)
		wsGroup.GET("/:id", deps.WorkspaceHandler.GetByID)
		wsGroup.PATCH("/:id", deps.WorkspaceHandler.Update)
		wsGroup.DELETE("/:id", deps.AuthMW.RequireRole("administrator"), deps.WorkspaceHandler.Delete)
	}

	// Team routes (authenticated)
	teamGroup := r.Group("/api/v1/teams", deps.AuthMW.RequireAuth(), rateLimitMW)
	{
		teamGroup.GET("/:id", deps.TeamHandler.GetByID)
		teamGroup.POST("/:id/members", deps.TeamHandler.AddMember)
		teamGroup.GET("/:id/members", deps.TeamHandler.ListMembers)
		teamGroup.DELETE("/:id/members/:user_id", deps.TeamHandler.RemoveMember)
	}

	// Collection routes (authenticated)
	colGroup := r.Group("/api/v1/collections", deps.AuthMW.RequireAuth(), rateLimitMW)
	{
		colGroup.GET("", deps.CollectionHandler.List)
		colGroup.POST("", deps.CollectionHandler.Create)
		colGroup.GET("/:id", deps.CollectionHandler.GetByID)
		colGroup.PATCH("/:id", deps.CollectionHandler.Update)
		colGroup.DELETE("/:id", deps.AuthMW.RequireRole("administrator"), deps.CollectionHandler.Delete)
	}

	// Endpoint routes (authenticated)
	epGroup := r.Group("/api/v1/endpoints", deps.AuthMW.RequireAuth(), rateLimitMW)
	{
		epGroup.GET("", deps.EndpointHandler.List)
		epGroup.POST("", deps.EndpointHandler.Create)
		epGroup.GET("/:id", deps.EndpointHandler.GetByID)
		epGroup.PATCH("/:id", deps.EndpointHandler.Update)
		epGroup.DELETE("/:id", deps.AuthMW.RequireRole("administrator"), deps.EndpointHandler.Delete)
		epGroup.POST("/:id/toggle", deps.EndpointHandler.Toggle)
	}

	// Introspection routes (authenticated)
	introGroup := r.Group("/api/v1/introspect", deps.AuthMW.RequireAuth(), rateLimitMW)
	{
		introGroup.GET("/schemas/:schema/tables", deps.IntrospectHandler.DiscoverTables)
		introGroup.GET("/schemas/:schema/tables/:table", deps.IntrospectHandler.GetTableSchema)
		introGroup.GET("/schemas/:schema/functions", deps.IntrospectHandler.DiscoverFunctions)
		introGroup.GET("/schemas/:schema/functions/:name", deps.IntrospectHandler.GetFunctionSignature)
		introGroup.GET("/schemas/:schema/procedures", deps.IntrospectHandler.DiscoverProcedures)
	}
}
