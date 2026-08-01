package di

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/muhammadyunus/Restify-Service/internal/config"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
	"github.com/muhammadyunus/Restify-Service/internal/domain/service"
)

// Container holds all application dependencies.
type Container struct {
	Config *config.Config

	Logger repository.Logger
	DB     repository.DB
	Cache  repository.Cache
	Queue  repository.MessageQueue
	MQTT   repository.MQTTBroker

	Router *gin.Engine

	AuthService       service.AuthService
	UserService       service.UserService
	WorkspaceService  service.WorkspaceService
	TeamService       service.TeamService
	CollectionService service.CollectionService
	EndpointService   service.EndpointService

	LogRepo       repository.APILogRepository
	AnalyticsRepo repository.AnalyticsRepository
	AlertRepo     repository.AlertRepository

	APIIntrospector  service.APIIntrospector
	RESTGenerator    service.RESTGenerator
	AnalyticsService service.AnalyticsService
	AlertService     service.AlertService
	EmailService     service.EmailService
	SchedulerService service.SchedulerService

	closer []func(context.Context) error
}

// Bootstrap wires all application dependencies into a ready-to-run Container.
func Bootstrap(ctx context.Context, cfg *config.Config) (*Container, error) {
	c := &Container{Config: cfg}

	var err error

	c.Logger, err = initLogger(cfg.Logging)
	if err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}

	c.DB, err = initDatabase(cfg.Database)
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

	c.LogRepo, err = initLogRepo()
	if err != nil {
		return nil, fmt.Errorf("init log repo: %w", err)
	}

	c.AnalyticsRepo, err = initAnalyticsRepo()
	if err != nil {
		return nil, fmt.Errorf("init analytics repo: %w", err)
	}

	c.AlertRepo, err = initAlertRepo()
	if err != nil {
		return nil, fmt.Errorf("init alert repo: %w", err)
	}

	c.AuthService = service.NewAuthService(c.DB, c.Cache)
	c.UserService = service.NewUserService(c.DB, c.AuthService, c.Logger)
	c.WorkspaceService = service.NewWorkspaceService(c.DB, c.Logger)
	c.TeamService = service.NewTeamService(c.DB, c.Logger)
	c.CollectionService = service.NewCollectionService(c.DB, c.Logger)
	c.EndpointService = service.NewEndpointService(c.DB, c.Logger)

	c.APIIntrospector = service.NewPostgreSQLIntrospector(c.DB)
	c.RESTGenerator = service.NewRESTGenerator(c.APIIntrospector, c.Logger)
	c.EmailService = service.NewEmailService(c.Logger)
	c.AnalyticsService = service.NewAnalyticsService(c.LogRepo, c.AnalyticsRepo, c.Logger)
	c.AlertService = service.NewAlertService(c.AlertRepo, c.Queue, c.EmailService, c.Logger)
	c.SchedulerService = service.NewSchedulerService(c.Logger)

	c.Router = initRouter(cfg.Server.Env)

	return c, nil
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
