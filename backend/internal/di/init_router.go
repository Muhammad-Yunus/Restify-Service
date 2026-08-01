package di

import (
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/presentation/http/middleware"
	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/presentation/http/router"
)

// initRouter sets up the Gin HTTP router with routes, middleware, and dynamic BaaS routing.
func initRouter(env string, rateLimit *middleware.RateLimitMiddleware, deps *Container) repository.HTTPRouter {
	ginRouter := router.NewGinRouter(env)

	// Register all API routes (includes /health endpoint)
	handlerDeps := &router.HandlerDeps{
		AuthHandler:       deps.AuthHandler,
		UserHandler:       deps.UserHandler,
		WorkspaceHandler:  deps.WorkspaceHandler,
		TeamHandler:       deps.TeamHandler,
		CollectionHandler: deps.CollectionHandler,
		EndpointHandler:   deps.EndpointHandler,
		IntrospectHandler: deps.IntrospectHandler,
		LiveHandler:       deps.LiveHandler,
		AnalyticsHandler:  deps.AnalyticsHandler,
		AuthMW:            deps.AuthMiddleware,
	}

	router.RegisterAll(ginRouter.Engine(), handlerDeps, rateLimit, deps.LogRepo, deps.Logger)

	// Register WebSocket endpoint (public)
	ginRouter.Engine().GET("/ws", deps.WSHandler.Handle)

	// Set up dynamic BaaS routing
	if deps.BaasRouteRegistry != nil {
		dynRouter := router.NewDynamicRouter(deps.BaasRouteRegistry)
		dynRouter.RegisterDynamicRoutes(ginRouter.Engine())
	}

	return ginRouter
}
