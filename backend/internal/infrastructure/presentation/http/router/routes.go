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
