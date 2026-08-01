package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
	httpHandler "github.com/muhammadyunus/Restify-Service/internal/infrastructure/presentation/http/handler"
	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/presentation/http/middleware"
	"github.com/muhammadyunus/Restify-Service/internal/version"
)

//	@title			ForgeBase API
//	@version		1.0
//	@description	ForgeBase is a Backend-as-a-Service (BaaS) platform that allows developers to create REST APIs on top of PostgreSQL databases.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.email	support@forgebase.io

//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT

//	@host		localhost:8080
//	@BasePath	/api/v1
//	@schemes	http https

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Example: Bearer {token}
//
//	@securityDefinitions.apikey	AdminRole
//	@in							header
//	@name						Authorization
//	@description				Required for admin-only endpoints. Example: Bearer {token}
//	@securityDefinitions.apikey	WorkspaceAdminRole
//	@in							header
//	@name						Authorization
//	@description				Required for workspace admin endpoints. Example: Bearer {token}

// HandlerDeps holds all HTTP handlers.
type HandlerDeps struct {
	AuthHandler       *httpHandler.AuthHandler
	UserHandler       *httpHandler.UserHandler
	WorkspaceHandler  *httpHandler.WorkspaceHandler
	TeamHandler       *httpHandler.TeamHandler
	CollectionHandler *httpHandler.CollectionHandler
	EndpointHandler   *httpHandler.EndpointHandler
	IntrospectHandler *httpHandler.IntrospectHandler
	LiveHandler       *httpHandler.LiveHandler
	AnalyticsHandler  *httpHandler.AnalyticsHandler
	AlertHandler      *httpHandler.AlertHandler
	AuthMW            *middleware.AuthMiddleware
}

// routerDeps holds dependencies needed for route registration.
type routerDeps struct {
	logRepo repository.APILogRepository
	logger  repository.Logger
	db      repository.DB
	cache   repository.Cache
}

// RegisterAll registers all API routes with versioning.
func RegisterAll(r *gin.Engine, deps *HandlerDeps, rateLimiter *middleware.RateLimitMiddleware, logRepo repository.APILogRepository, logger repository.Logger, db repository.DB, cache repository.Cache) {
	rd := &routerDeps{
		logRepo: logRepo,
		logger:  logger,
		db:      db,
		cache:   cache,
	}

	// Health check (public)
	r.GET("/health", func(c *gin.Context) {
		status := "ok"
		checks := gin.H{}

		if rd.db != nil {
			// Try a lightweight query to verify database connectivity
			var healthy bool
			if err := rd.db.Raw(c.Request.Context(), "SELECT true", &healthy); err != nil {
				status = "degraded"
				checks["database"] = "error"
			} else {
				checks["database"] = "ok"
			}
		}

		if rd.cache != nil {
			if _, err := rd.cache.Get(c.Request.Context(), "health:ping"); err != nil {
				status = "degraded"
				checks["cache"] = "error"
			} else {
				checks["cache"] = "ok"
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"status":     status,
			"service":    "ForgeBase",
			"version":    version.Version,
			"built_at":   version.BuiltAt,
			"git_commit": version.GitCommit,
			"checks":     checks,
		})
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

	// Live BaaS endpoint introspection (authenticated)
	r.GET("/api/v1/admin/endpoints", deps.LiveHandler.ListRegistered, deps.AuthMW.RequireAuth())

	// Analytics routes (authenticated)
	analyticsGroup := r.Group("/api/v1/analytics", deps.AuthMW.RequireAuth(), rateLimitMW)
	{
		analyticsGroup.GET("/overview/:workspace_id", deps.AnalyticsHandler.Overview)
		analyticsGroup.GET("/endpoints/:endpoint_id/metrics", deps.AnalyticsHandler.EndpointMetrics)
		analyticsGroup.GET("/logs", deps.AnalyticsHandler.SearchLogs)
	}

	// Alert routes (authenticated)
	alertGroup := r.Group("/api/v1/alerts", deps.AuthMW.RequireAuth(), rateLimitMW)
	{
		alertGroup.GET("/:workspace_id", deps.AlertHandler.List)
		alertGroup.POST("/:workspace_id", deps.AlertHandler.Create)
		alertGroup.PUT("/:workspace_id/:id/toggle", deps.AlertHandler.Toggle)
		alertGroup.PATCH("/:workspace_id/:id", deps.AlertHandler.Update)
		alertGroup.DELETE("/:workspace_id/:id", deps.AlertHandler.Delete)
		alertGroup.GET("/:workspace_id/events", deps.AlertHandler.ListEvents)
	}
}
