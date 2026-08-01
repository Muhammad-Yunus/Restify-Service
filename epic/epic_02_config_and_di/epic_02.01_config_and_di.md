# Epic 02 — Configuration & Dependency Injection

**Goal:** Set up configuration management with Koanf + .env, build manual DI bootstrap container.
**Dependencies:** Epic 01 (project structure)
**Commit:** `feat: add configuration system and DI bootstrap`

---

## Step 02.01 — Environment Configuration

**Build:**

1. Create `backend/configs/.env.example`:
```env
# Server
ForgeBase_HOST=0.0.0.0
ForgeBase_PORT=8080
ForgeBase_ENV=development
ForgeBase_JWT_SECRET=change-me-in-production
ForgeBase_JWT_EXPIRATION=24h

# Database
DATABASE_URL=postgres://user:pass@localhost:5432/ForgeBase?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379/0

# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@localhost:5672/

# SMTP
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-app-password

# EMQX MQTT
EMQX_BROKER_URL=tcp://localhost:1883

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# OpenTelemetry
OTEL_ENABLED=false
OTEL_ENDPOINT=http://localhost:4317
```

2. Create `backend/internal/config/config.go`:
```go
package config

import (
    "github.com/knadh/koanf/v2"
    "github.com/knadh/koanf/parsers/env"
    "github.com/knadh/koanf/parsers/dotenv"
    "github.com/knadh/koanf/providers/file"
    "github.com/knadh/koanf/providers/env"
)

// Config holds all application configuration.
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
    RabbitMQ RabbitMQConfig
    SMTP     SMTPConfig
    EMQX     EMQXConfig
    Logging  LoggingConfig
    OTEL     OTELConfig
    JWT      JWTConfig
}

type ServerConfig struct {
    Host     string
    Port     int
    Env      string
    JWTSecret string
    JWTEXP   string
}

type DatabaseConfig struct {
    URL string
}

type RedisConfig struct {
    URL string
}

type RabbitMQConfig struct {
    URL string
}

type SMTPConfig struct {
    Host string
    Port int
    User string
    Pass string
}

type EMQXConfig struct {
    URL string
}

type LoggingConfig struct {
    Level  string
    Format string
}

type OTELConfig struct {
    Enabled  bool
    Endpoint string
}

type JWTConfig struct {
    Secret string
    Exp    string
}

// Load reads configuration from .env file and environment variables.
func Load(configPath string) (*Config, error) {
    k := koanf.New(".")

    // Load .env file if exists
    if err := k.Load(dotenv.Provider(configPath, "."), nil); err != nil {
        // .env not required in production (env vars suffice)
    }

    // Load environment variables
    if err := k.Load(env.Provider(k, ".", "."), nil); err != nil {
        return nil, err
    }

    cfg := &Config{}
    if err := k.Unmarshal("", cfg); err != nil {
        return nil, err
    }

    return cfg, nil
}
```

**Test cases:**
- [ ] Unit: `Load()` reads `.env` file correctly
- [ ] Unit: `Load()` falls back to environment variables when `.env` missing
- [ ] Unit: All Config struct fields populated from environment
- [ ] Unit: Missing required config returns descriptive error

---

## Step 02.02 — Manual DI Bootstrap Container

**Build:** Create `backend/internal/di/bootstrap.go`:

```go
package di

import (
    "context"
    "fmt"

    "github.com/muhammadyunus/ForgeBase/internal/config"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
    "github.com/muhammadyunus/ForgeBase/internal/domain/service"
)

// Container holds all application dependencies.
type Container struct {
    Config *config.Config

    // Infrastructure
    DB         repository.DB
    Cache      repository.Cache
    Queue      repository.MessageQueue
    MQTT       repository.MQTTBroker
    Logger     repository.Logger

    // Domain Services
    AuthService    service.AuthService
    UserService    service.UserService
    WorkspaceService service.WorkspaceService
    TeamService    service.TeamService
    CollectionService service.CollectionService
    EndpointService  service.EndpointService

    // Repositories (wired during bootstrap)
    LogRepo       repository.APILogRepository
    AnalyticsRepo repository.AnalyticsRepository
    AlertRepo     repository.AlertRepository

    // Application Services
    APIIntrospector service.APIIntrospector
    RESTGenerator   service.RESTGenerator
    AnalyticsService service.AnalyticsService
    AlertService    service.AlertService
    EmailService    service.EmailService
    SchedulerService service.SchedulerService

    // HTTP
    Router repository.HTTPRouter

    // Lifecycle
    closer []func(context.Context) error
}

// Bootstrap creates and wires all application dependencies.
func Bootstrap(ctx context.Context, cfg *config.Config) (*Container, error) {
    c := &Container{Config: cfg}

    // Infrastructure (order matters: later services depend on earlier ones)
    var err error

    c.Logger, err = initLogger(cfg.Logging)
    if err != nil {
        return nil, fmt.Errorf("init logger: %w", err)
    }

    c.DB, err = initDatabase(ctx, cfg.Database)
    if err != nil {
        return nil, fmt.Errorf("init database: %w", err)
    }
    c.registerClose(func(ctx context.Context) error { return c.DB.Close(ctx) })

    c.Cache, err = initCache(ctx, cfg.Redis)
    if err != nil {
        return nil, fmt.Errorf("init cache: %w", err)
    }
    c.registerClose(func(ctx context.Context) error { return c.Cache.Close(ctx) })

    c.Queue, err = initQueue(ctx, cfg.RabbitMQ)
    if err != nil {
        return nil, fmt.Errorf("init queue: %w", err)
    }
    c.registerClose(func(ctx context.Context) error { return c.Queue.Close(ctx) })

    c.MQTT, err = initMQTT(ctx, cfg.EMQX)
    if err != nil {
        return nil, fmt.Errorf("init mqtt: %w", err)
    }
    c.registerClose(func(ctx context.Context) error { return c.MQTT.Close(ctx) })

    // Domain Services (depend on infrastructure)
    c.AuthService = service.NewAuthService(c.DB, c.Cache, cfg.JWT)
    c.UserService = service.NewUserService(c.DB, c.AuthService, c.Logger)
    c.WorkspaceService = service.NewWorkspaceService(c.DB, c.Logger)
    c.TeamService = service.NewTeamService(c.DB, c.Logger)
    c.CollectionService = service.NewCollectionService(c.DB, c.Logger)
    c.EndpointService = service.NewEndpointService(c.DB, c.Logger)

    // Application Services
    c.APIIntrospector = service.NewPostgreSQLIntrospector(c.DB)
    c.RESTGenerator = service.NewRESTGenerator(c.APIIntrospector, c.Logger)
    c.AnalyticsService = service.NewAnalyticsService(c.LogRepo, c.AnalyticsRepo, c.Logger)
    c.AlertService = service.NewAlertService(c.AlertRepo, c.Queue, c.EmailService, c.Logger)
    c.EmailService = service.NewEmailService(cfg.SMTP, c.Logger)
    c.SchedulerService = service.NewSchedulerService(c.Logger)

    // HTTP Router (depends on all services)
    c.Router, err = initRouter(ctx, c)
    if err != nil {
        return nil, fmt.Errorf("init router: %w", err)
    }

    return c, nil
}

func (c *Container) registerClose(fn func(context.Context) error) {
    c.closer = append(c.closer, fn)
}

// Close gracefully shuts down all dependencies.
func (c *Container) Close(ctx context.Context) error {
    for i := len(c.closer) - 1; i >= 0; i-- {
        if err := c.closer[i](ctx); err != nil {
            return err
        }
    }
    return nil
}
```

**Build stub functions** in `internal/di/init_*.go` — each returns a stub implementation for now, to be filled in subsequent epics:
- `initLogger()` → returns Logger repo (filled in Epic 06)
- `initDatabase()` → returns DB repo (filled in Epic 03)
- `initCache()` → returns Cache repo (filled in Epic 24)
- `initQueue()` → returns Queue repo (filled in Epic 06)
- `initMQTT()` → returns MQTT repo (filled in Epic 22)
- `initRouter()` → returns HTTP router (filled in Epic 13)
- `initLogRepo()` → returns APILog repository (filled in Epic 18)
- `initAnalyticsRepo()` → returns Analytics repository (filled in Epic 19)
- `initAlertRepo()` → returns Alert repository (filled in Epic 20)

**Test cases:**
- [ ] Unit: `Bootstrap()` creates container without panicking (stubs OK)
- [ ] Unit: `Close()` calls all registered closers in reverse order
- [ ] Unit: Missing database config returns clear error
- [ ] E2E: Full bootstrap cycle completes successfully with real infra (integration test)

---

## Step 02.03 — App Interface and Run Method

**Build:** Create `backend/internal/di/app.go`:

```go
package di

import (
    "context"
    "net/http"

    "github.com/gin-gonic/gin"
)

// App is the top-level application object.
type App struct {
    container *Container
    server    *http.Server
    engine    *gin.Engine
}

// Run starts the HTTP server and blocks until context is cancelled.
func (a *App) Run(ctx context.Context) error {
    addr := fmt.Sprintf("%s:%d", a.container.Config.Server.Host, a.container.Config.Server.Port)

    a.server = &http.Server{
        Addr:    addr,
        Handler: a.engine,
    }

    go func() {
        <-ctx.Done()
        a.server.Shutdown(context.Background())
    }()

    a.container.Logger.Info("starting server", "addr", addr)
    if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        return err
    }
    return nil
}

// NewApp creates an App from a Container.
func NewApp(c *Container) (*App, error) {
    return &App{container: c, engine: c.Router}, nil
}
```

Update `main.go` to wire everything together.

**Test cases:**
- [ ] Unit: `App.Run()` starts HTTP server
- [ ] Unit: `App.Run()` returns when context cancelled
- [ ] E2E: Server responds to health check

---

## Step 02.04 — Health Check Endpoint

**Build:** Add a simple health check route in the router stub:

```go
// In router initialization
router.GET("/health", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": "ok"})
})
```

**Test cases:**
- [ ] E2E: `GET /health` returns `200 OK` with JSON body
- [ ] E2E: Response includes `status: "ok"`

---

## Commit Instruction

```bash
git add .
git commit -m "feat: add Koanf configuration system and manual DI bootstrap container"
```
