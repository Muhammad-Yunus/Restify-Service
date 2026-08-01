package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/presentation/http/handler"
	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/presentation/http/middleware"
)

// RegisterUserRoutes registers user-related routes.
func RegisterUserRoutes(r *gin.Engine, userHandler *handler.UserHandler, authMW *middleware.AuthMiddleware) {
	users := r.Group("/api/v1/users")
	{
		users.GET("", userHandler.List)
		users.GET("/:id", authMW.RequireAuth(), userHandler.GetByID)
		users.PATCH("/:id", authMW.RequireAuth(), userHandler.Update)
		users.DELETE("/:id", authMW.RequireRole("administrator"), userHandler.Delete)
	}
}

// RegisterWorkspaceRoutes registers workspace-related routes.
func RegisterWorkspaceRoutes(r *gin.Engine, wsHandler *handler.WorkspaceHandler, authMW *middleware.AuthMiddleware) {
	workspaces := r.Group("/api/v1/workspaces")
	{
		workspaces.GET("", authMW.RequireAuth(), wsHandler.List)
		workspaces.POST("", authMW.RequireAuth(), wsHandler.Create)
		workspaces.GET("/:id", authMW.RequireAuth(), wsHandler.GetByID)
		workspaces.PATCH("/:id", authMW.RequireAuth(), wsHandler.Update)
		workspaces.DELETE("/:id", authMW.RequireRole("administrator"), wsHandler.Delete)
	}
}

// RegisterTeamRoutes registers team-related routes.
func RegisterTeamRoutes(r *gin.Engine, teamHandler *handler.TeamHandler, authMW *middleware.AuthMiddleware) {
	teams := r.Group("/api/v1/teams")
	{
		teams.GET("/:id", authMW.RequireAuth(), teamHandler.GetByID)
		teams.POST("/:id/members", authMW.RequireAuth(), teamHandler.AddMember)
		teams.GET("/:id/members", authMW.RequireAuth(), teamHandler.ListMembers)
		teams.DELETE("/:id/members/:user_id", authMW.RequireAuth(), teamHandler.RemoveMember)
	}
}

// RegisterAuthRoutes registers authentication-related routes.
func RegisterAuthRoutes(r *gin.Engine, authHandler *handler.AuthHandler) {
	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/logout", authHandler.Logout)
	}
}

// RegisterHealthRoute registers the health check route.
func RegisterHealthRoute(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}
