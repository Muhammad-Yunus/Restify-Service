# Epic 06 — Infrastructure Adapters

**Goal:** Implement all repository interfaces with concrete infrastructure adapters.
**Dependencies:** Epic 05 (Repository interfaces defined)
**Commit:** `feat: implement all infrastructure adapters for repository interfaces`

---

## Step 06.01 — Logger Adapter (slog + tint)

**Build:** Create `backend/internal/infrastructure/logging/logger.go`:

```go
package logging

import (
    "context"
    "io"
    "log/slog"
    "os"

    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
    "github.com/lmittmann/tint"
)

// SLogger wraps slog.Logger with the repository.Logger interface.
type SLogger struct {
    logger *slog.Logger
}

// New creates a new structured logger.
func New(level string, format string, output io.Writer) repository.Logger {
    opts := &slog.HandlerOptions{
        Level: parseLevel(level),
    }

    var handler slog.Handler
    if format == "text" {
        handler = slog.NewTextHandler(output, opts)
    } else {
        handler = tint.NewHandler(output, &tint.Options{
            Level: opts.Level,
            TimeFormat: "2006-01-02T15:04:05.000Z",
        })
    }

    return &SLogger{logger: slog.New(handler)}
}

// NewDefault creates a logger writing to stderr.
func NewDefault(level string) repository.Logger {
    return New(level, "json", os.Stderr)
}

func (l *SLogger) With(keyValues ...any) repository.Logger {
    return &SLogger{logger: l.logger.With(keyValues...)}
}

func (l *SLogger) Debug(ctx context.Context, msg string, keyValues ...any) {
    l.logger.Log(ctx, slog.LevelDebug, msg, keyValues...)
}

func (l *SLogger) Info(ctx context.Context, msg string, keyValues ...any) {
    l.logger.Log(ctx, slog.LevelInfo, msg, keyValues...)
}

func (l *SLogger) Warn(ctx context.Context, msg string, keyValues ...any) {
    l.logger.Log(ctx, slog.LevelWarn, msg, keyValues...)
}

func (l *SLogger) Error(ctx context.Context, msg string, keyValues ...any) {
    l.logger.Log(ctx, slog.LevelError, msg, keyValues...)
}

func (l *SLogger) Logger() *slog.Logger {
    return l.logger
}

func parseLevel(level string) slog.Level {
    switch level {
    case "debug":
        return slog.LevelDebug
    case "warn":
        return slog.LevelWarn
    case "error":
        return slog.LevelError
    default:
        return slog.LevelInfo
    }
}

// Compile-time check.
var _ repository.Logger = (*SLogger)(nil)
```

**Test cases:**
- [ ] Unit: `New()` creates logger with correct level
- [ ] Unit: `With()` adds key-value pairs to context
- [ ] Unit: `Debug()` writes at debug level
- [ ] Unit: `Error()` writes at error level
- [ ] Unit: Output format is correct (JSON vs text)

---

## Step 06.02 — Cache Adapter (Redis)

**Build:** Create `backend/internal/infrastructure/cache/redis.go`:

```go
package cache

import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// RedisCache implements the repository.Cache interface.
type RedisCache struct {
    client *redis.Client
}

// NewRedisCache creates a new Redis cache connection.
func NewRedisCache(ctx context.Context, url string) (*RedisCache, error) {
    opts, err := redis.ParseURL(url)
    if err != nil {
        return nil, fmt.Errorf("parse redis URL: %w", err)
    }

    client := redis.NewClient(opts)

    if err := client.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("ping redis: %w", err)
    }

    return &RedisCache{client: client}, nil
}

func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
    return r.client.Get(ctx, key).Result()
}

func (r *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
    return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
    return r.client.Del(ctx, key).Err()
}

func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
    count, err := r.client.Exists(ctx, key).Result()
    if err != nil {
        return false, err
    }
    return count > 0, nil
}

func (r *RedisCache) Close(ctx context.Context) error {
    return r.client.Close()
}

// Compile-time check.
var _ repository.Cache = (*RedisCache)(nil)
```

**Test cases:**
- [ ] Unit: `NewRedisCache()` with valid URL creates client
- [ ] Unit: `Get()` returns value for existing key
- [ ] Unit: `Get()` returns error for missing key
- [ ] Unit: `Set()` stores value with TTL
- [ ] Unit: `Delete()` removes key
- [ ] Unit: `Exists()` returns true for existing key
- [ ] Integration: Full get/set/delete cycle

---

## Step 06.03 — Message Queue Adapter (RabbitMQ)

**Build:** Create `backend/internal/infrastructure/queue/rabbitmq.go`:

```go
package queue

import (
    "context"
    "fmt"

    "github.com/google/uuid"
    "github.com/streadway/amqp"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// RabbitMQQueue implements the repository.MessageQueue interface.
type RabbitMQQueue struct {
    conn    *amqp.Connection
    channel *amqp.Channel
}

// NewRabbitMQQueue creates a new RabbitMQ connection.
func NewRabbitMQQueue(ctx context.Context, url string) (*RabbitMQQueue, error) {
    conn, err := amqp.Dial(url)
    if err != nil {
        return nil, fmt.Errorf("dial rabbitmq: %w", err)
    }

    ch, err := conn.Channel()
    if err != nil {
        conn.Close()
        return nil, fmt.Errorf("create channel: %w", err)
    }

    return &RabbitMQQueue{conn: conn, channel: ch}, nil
}

func (q *RabbitMQQueue) Publish(ctx context.Context, queue string, message []byte) error {
    if err := q.channel.PublishWithContext(
        ctx,
        "",           // exchange
        queue,        // routing key = queue name
        false,        // mandatory
        false,        // immediate
        amqp.Publishing{
            ContentType: "application/json",
            Body:        message,
            MessageId:   uuid.New().String(),
        },
    ); err != nil {
        return fmt.Errorf("publish to %s: %w", queue, err)
    }
    return nil
}

func (q *RabbitMQQueue) Consume(ctx context.Context, queue string, handler repository.MessageHandler) error {
    msgs, err := q.channel.ConsumeWithContext(
        ctx,
        queue,
        "",     // consumer
        false,  // auto-ack
        false,  // exclusive
        false,  // no-local
        false,  // no-wait
        nil,    // args
    )
    if err != nil {
        return fmt.Errorf("consume %s: %w", queue, err)
    }

    go func() {
        for msg := range msgs {
            if err := handler(ctx, msg.Body); err != nil {
                msg.Nack(false, true) // requeue on error
                continue
            }
            msg.Ack(false)
        }
    }()

    return nil
}

func (q *RabbitMQQueue) DeclareQueue(ctx context.Context, name string, opts repository.QueueOptions) error {
    _, err := q.channel.QueueDeclareWithContext(
        ctx,
        name,
        opts.Durable,
        opts.AutoDelete,
        false,
        opts.Arguments,
        nil,
    )
    if err != nil {
        return fmt.Errorf("declare queue %s: %w", name, err)
    }
    return nil
}

func (q *RabbitMQQueue) Close(ctx context.Context) error {
    if err := q.channel.Close(); err != nil {
        return err
    }
    return q.conn.Close()
}

// Compile-time check.
var _ repository.MessageQueue = (*RabbitMQQueue)(nil)
```

**Test cases:**
- [ ] Unit: `NewRabbitMQQueue()` with valid URL creates connection
- [ ] Unit: `Publish()` sends message to queue
- [ ] Unit: `Consume()` receives and processes messages
- [ ] Unit: Failed messages are nacked and requeued
- [ ] Integration: Full publish/subscribe cycle

---

## Step 06.04 — MQTT Broker Adapter (EMQX + Paho)

**Build:** Create `backend/internal/infrastructure/mqtt/emqx.go`:

```go
package mqtt

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
    mqtt "github.com/eclipse/paho.mqtt.golang"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// EMQXBroker implements the repository.MQTTBroker interface.
type EMQXBroker struct {
    client mqtt.Client
}

// NewEMQXBroker creates a new MQTT broker connection.
func NewEMQXBroker(ctx context.Context, brokerURL string) (*EMQXBroker, error) {
    opts := mqtt.NewClientOptions()
    opts.AddBroker(brokerURL)
    opts.SetClientID(fmt.Sprintf("ForgeBase-%s", uuid.New().String()[:8]))
    opts.SetAutoReconnect(true)
    opts.SetConnectionHandler(func(c mqtt.Client) {
        // Re-subscribe on reconnect
    })

    client := mqtt.NewClient(opts)
    token := client.Connect()
    if timeout := 10 * time.Second; !token.WaitTimeout(timeout) {
        return nil, fmt.Errorf("connect to MQTT broker: timeout")
    }
    if err := token.Error(); err != nil {
        return nil, fmt.Errorf("connect to MQTT broker: %w", err)
    }

    return &EMQXBroker{client: client}, nil
}

func (m *EMQXBroker) Connect(ctx context.Context) error {
    token := m.client.Connect()
    if !token.WaitTimeout(10 * time.Second) {
        return fmt.Errorf("MQTT connect timeout")
    }
    return token.Error()
}

func (m *EMQXBroker) Publish(ctx context.Context, topic string, payload []byte, qos byte, retained bool) error {
    token := m.client.Publish(topic, qos, retained, payload)
    if !token.WaitTimeout(5 * time.Second) {
        return fmt.Errorf("MQTT publish timeout")
    }
    return token.Error()
}

func (m *EMQXBroker) Subscribe(ctx context.Context, topic string, qos byte, handler repository.MQTTHandler) error {
    token := m.client.Subscribe(topic, qos, func(client mqtt.Client, msg mqtt.Message) {
        handler(msg.Topic(), msg.Payload())
    })
    if !token.WaitTimeout(5 * time.Second) {
        return fmt.Errorf("MQTT subscribe timeout")
    }
    return token.Error()
}

func (m *EMQXBroker) Unsubscribe(ctx context.Context, topic string) error {
    token := m.client.Unsubscribe(topic)
    if !token.WaitTimeout(5 * time.Second) {
        return fmt.Errorf("MQTT unsubscribe timeout")
    }
    return token.Error()
}

func (m *EMQXBroker) Close(ctx context.Context) error {
    m.client.Disconnect(250)
    return nil
}

// Compile-time check.
var _ repository.MQTTBroker = (*EMQXBroker)(nil)
```

**Test cases:**
- [ ] Unit: `NewEMQXBroker()` creates client
- [ ] Unit: `Publish()` sends message to topic
- [ ] Unit: `Subscribe()` receives messages via handler
- [ ] Integration: Full publish/subscribe cycle with EMQX

---

## Step 06.05 — HTTP Router Adapter (Gin)

**Build:** Create `backend/internal/infrastructure/presentation/http/router/gin.go`:

```go
package router

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// GinRouter implements the repository.HTTPRouter interface.
type GinRouter struct {
    engine *gin.Engine
}

// NewGinRouter creates a new Gin router.
func NewGinRouter() repository.HTTPRouter {
    gin.SetMode(gin.ReleaseMode)
    engine := gin.New()
    engine.Use(gin.Recovery())
    engine.Use(gin.Logger())
    return &GinRouter{engine: engine}
}

func (r *GinRouter) Group(basePath string, middleware ...repository.Middleware) *repository.RouterGroup {
    group := r.engine.Group(basePath)
    for _, m := range middleware {
        group.Use(middlewareToGin(m))
    }
    return &repository.RouterGroup{
        basePath: basePath,
        middleware: append([]repository.Middleware{}, middleware...),
    }
}

func (r *GinRouter) Handle(method, path string, handler http.HandlerFunc, middleware ...repository.Middleware) {
    ginHandler := func(c *gin.Context) {
        handler(c.Writer, c.Request)
    }
    routes := map[string]func(string, http.HandlerFunc, ...gin.HandlerFunc){
        "GET":     r.engine.GET,
        "POST":    r.engine.POST,
        "PUT":     r.engine.PUT,
        "DELETE":  r.engine.DELETE,
        "PATCH":   r.engine.PATCH,
        "HEAD":    r.engine.HEAD,
        "OPTIONS": r.engine.OPTIONS,
    }

    fn, ok := routes[method]
    if !ok {
        panic(fmt.Sprintf("unsupported HTTP method: %s", method))
    }

    ginMiddleware := make([]gin.HandlerFunc, len(middleware))
    for i, m := range middleware {
        ginMiddleware[i] = middlewareToGin(m)
    }
    fn(path, ginHandler, ginMiddleware...)
}

func (r *GinRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    r.engine.ServeHTTP(w, req)
}

// middlewareToGin converts a domain middleware to a Gin handler.
// NOTE: This is a simplified adapter. In production, use the
// middleware directly as gin.HandlerFunc or implement proper
// middleware-to-Gin conversion in the DI layer.
func middlewareToGin(m repository.Middleware) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Call the middleware and let it decide whether to abort
        m(c)
    }
}

// Compile-time check.
var _ repository.HTTPRouter = (*GinRouter)(nil)
```

**Test cases:**
- [ ] Unit: `NewGinRouter()` creates router without panic
- [ ] Unit: `Handle()` registers routes correctly
- [ ] Unit: `Group()` creates route groups with middleware
- [ ] Unit: `ServeHTTP()` routes requests correctly
- [ ] Integration: GET/POST/PUT/DELETE all work

---

## Step 06.06 — Update DI Bootstrap with Real Adapters

**Build:** Update `internal/di/bootstrap.go` to use real infrastructure adapters:

```go
func initLogger(cfg config.LoggingConfig) (repository.Logger, error) {
    return logging.NewDefault(cfg.Level)
}

func initDatabase(ctx context.Context, cfg config.DatabaseConfig) (repository.DB, error) {
    pg, err := database.NewPostgresDB(ctx, cfg)
    if err != nil {
        return nil, err
    }
    _, err = database.NewGORMDB(pg)
    if err != nil {
        pg.Close(ctx)
        return nil, err
    }
    return pg, nil
}

func initCache(ctx context.Context, cfg config.RedisConfig) (repository.Cache, error) {
    return cache.NewRedisCache(ctx, cfg.URL)
}

func initQueue(ctx context.Context, cfg config.RabbitMQConfig) (repository.MessageQueue, error) {
    return queue.NewRabbitMQQueue(ctx, cfg.URL)
}

func initMQTT(ctx context.Context, cfg config.EMQXConfig) (repository.MQTTBroker, error) {
    return mqtt.NewEMQXBroker(ctx, cfg.URL)
}

func initRouter(_ context.Context, c *di.Container) (repository.HTTPRouter, error) {
    return router.NewGinRouter(), nil
}
```

**Test cases:**
- [ ] Integration: Full bootstrap with all real adapters (requires all infra running)
- [ ] E2E: Application starts with all dependencies connected

---

## Commit Instruction

```bash
git add .
git commit -m "feat: implement all infrastructure adapters for repository interfaces"
```
