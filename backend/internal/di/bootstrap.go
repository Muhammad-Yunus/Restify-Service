package di

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/muhammadyunus/Restify-Service/internal/application/service"
	"github.com/muhammadyunus/Restify-Service/internal/config"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
	domservice "github.com/muhammadyunus/Restify-Service/internal/domain/service"
	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/auth"
	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/baas"
	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/presentation/http/handler"
	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/presentation/http/middleware"
)

// Container holds all application dependencies.
type Container struct {
	Config *config.Config

	Logger repository.Logger
	DB     repository.DB
	GORM   *gorm.DB
	Cache  repository.Cache
	Queue  repository.MessageQueue
	MQTT   repository.MQTTBroker

	Router          repository.HTTPRouter
	RateLimiter     *middleware.RateLimitMiddleware
	AuthMiddleware  *middleware.AuthMiddleware

	AuthHandler       *handler.AuthHandler
	UserHandler       *handler.UserHandler
	WorkspaceHandler  *handler.WorkspaceHandler
	TeamHandler       *handler.TeamHandler
	CollectionHandler *handler.CollectionHandler
	EndpointHandler   *handler.EndpointHandler
	IntrospectHandler *handler.IntrospectHandler
	LiveHandler       *handler.LiveHandler
	AnalyticsHandler  *handler.AnalyticsHandler
	AlertHandler      *handler.AlertHandler
	BaasRouteRegistry *baas.RouteRegistry

	AuthService       domservice.AuthService
	UserService       domservice.UserService
	WorkspaceService  domservice.WorkspaceService
	TeamService       domservice.TeamService
	CollectionService domservice.CollectionService
	EndpointService   domservice.EndpointService

	LogRepo       repository.APILogRepository
	AnalyticsRepo repository.AnalyticsRepository
	AlertRepo     repository.AlertRepository
	EndpointRepo  *service.ServiceEndpointRepo

	APIIntrospector  domservice.APIIntrospector
	RESTGenerator    domservice.RESTGenerator
	AnalyticsService domservice.AnalyticsService
	AlertService     domservice.AlertService
	EmailService     domservice.EmailService
	SchedulerService domservice.SchedulerService

	IntrospectService *service.IntrospectorService
	QueueService      *service.QueueService
	WorkerPool        *service.WorkerPool

	closer []func(context.Context) error
}

// RegisterWorker registers a worker to be started by the pool.
func (c *Container) RegisterWorker(worker *service.Worker) {
	c.WorkerPool.Add(worker)
}

// RegisterQueueWorker is a convenience method to register a worker with a queue handler.
func (c *Container) RegisterQueueWorker(name string, queue string, handler repository.MessageHandler) {
	worker := service.NewWorker(
		name,
		queue,
		handler,
		c.Logger,
		c.QueueService.Consume,
	)
	c.WorkerPool.Add(worker)
}

// Bootstrap wires all application dependencies into a ready-to-run Container.
func Bootstrap(ctx context.Context, cfg *config.Config) (*Container, error) {
	c := &Container{Config: cfg}

	var err error

	c.Logger = initLogger(cfg.Logging)

	c.DB, c.GORM, err = initDatabase(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("init database: %w", err)
	}

	c.registerClose(c.DB.Close)

	c.Cache, err = initCache(cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("init cache: %w", err)
	}

	c.registerClose(c.Cache.Close)

	c.Queue, err = initQueue(cfg.RabbitMQ)
	if err != nil {
		return nil, fmt.Errorf("init queue: %w", err)
	}

	c.registerClose(c.Queue.Close)

	c.MQTT, err = initMQTT(cfg.EMQX)
	if err != nil {
		return nil, fmt.Errorf("init mqtt: %w", err)
	}
	c.registerClose(c.MQTT.Close)

	if err := c.wireServices(); err != nil {
		return nil, err
	}

	return c, nil
}

// wireServices wires repositories, services, and the HTTP router.
func (c *Container) wireServices() error {
	var err error

	c.LogRepo, err = initLogRepo()
	if err != nil {
		return fmt.Errorf("init log repo: %w", err)
	}

	c.AnalyticsRepo, err = initAnalyticsRepo(c.DB, c.GORM)
	if err != nil {
		return fmt.Errorf("init analytics repo: %w", err)
	}

	c.AlertRepo, err = initAlertRepo(c.DB, c.GORM)
	if err != nil {
		return fmt.Errorf("init alert repo: %w", err)
	}

	c.AuthService = domservice.NewAuthService(c.DB, c.Cache)
	c.UserService = domservice.NewUserService(nil, c.Logger)
	c.WorkspaceService = domservice.NewWorkspaceService(nil, c.Logger)
	c.TeamService = domservice.NewTeamService(c.DB, c.Logger)
	c.CollectionService = domservice.NewCollectionService(c.DB, c.Logger)

	c.EndpointRepo = &service.ServiceEndpointRepo{DB: c.GORM}
	c.EndpointService = service.NewEndpointService(c.GORM, c.Logger)
	c.APIIntrospector = baas.NewPostgreSQLIntrospector(c.DB)
	c.RESTGenerator = baas.NewRESTGenerator(c.APIIntrospector, c.Logger)
	c.EmailService = domservice.NewEmailService(c.Logger)
	c.AnalyticsService = domservice.NewAnalyticsService(c.LogRepo, c.AnalyticsRepo, c.Logger)
	c.AlertService = domservice.NewAlertService(c.AlertRepo, c.Queue, c.EmailService, c.Logger, c.MQTT)
	c.SchedulerService = domservice.NewSchedulerService(c.Logger)
	c.IntrospectService = service.NewIntrospectorService(c.APIIntrospector)

	// Initialize message queue service and worker pool
	c.QueueService = service.NewQueueService(c.Queue)
	c.WorkerPool = service.NewWorkerPool(c.Logger)

	// Initialize auth middleware and rate limiter
	jwtExpiration, _ := time.ParseDuration(c.Config.JWT.Expiration)
	jwtSvc := auth.NewJWTService(c.Config.JWT.Secret, jwtExpiration)
	blacklist := auth.NewTokenBlacklist(c.Cache)
	c.RateLimiter = middleware.NewRateLimitMiddleware(c.Config.RateLimit.RequestsPerMinute)
	c.AuthMiddleware = middleware.NewAuthMiddleware(jwtSvc, blacklist)

	// Initialize HTTP handlers
	c.AuthHandler = handler.NewAuthHandler(c.AuthService, jwtSvc)
	c.UserHandler = handler.NewUserHandler(c.UserService)
	c.WorkspaceHandler = handler.NewWorkspaceHandler(c.WorkspaceService)
	c.TeamHandler = handler.NewTeamHandler(c.TeamService)
	c.CollectionHandler = handler.NewCollectionHandler(c.CollectionService)
	c.EndpointHandler = handler.NewEndpointHandler(c.EndpointService)
	c.IntrospectHandler = handler.NewIntrospectHandler(c.IntrospectService)
	c.AnalyticsHandler = handler.NewAnalyticsHandler(c.AnalyticsService, c.LogRepo)

	c.AlertHandler = handler.NewAlertHandler(c.AlertService)
	c.BaasRouteRegistry = baas.NewRouteRegistry(c.RESTGenerator, c.EndpointRepo, c.Logger)
	c.LiveHandler = handler.NewLiveHandler(c.BaasRouteRegistry)

	c.Router = initRouter(c.Config.Server.Env, c.RateLimiter, c)

	return nil
}

func (c *Container) registerClose(fn func(context.Context) error) {
	c.closer = append(c.closer, fn)
}

// Close gracefully shuts down all dependencies in reverse order.
func (c *Container) Close(ctx context.Context) error {
	for i := len(c.closer) - 1; i >= 0; i-- {
		if err := c.closer[i](ctx); err != nil {
			return err
		}
	}

	return nil
}
