package di

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/muhammadyunus/Restify-Service/internal/config"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

func testConfig() *config.Config {
	return &config.Config{
		Server:   config.ServerConfig{Host: "127.0.0.1", Port: 8080, Env: "test"},
		Database: config.DatabaseConfig{URL: "postgres://u:p@localhost:5432/db?sslmode=disable"},
		Redis:    config.RedisConfig{URL: "redis://localhost:6379/0"},
		RabbitMQ: config.RabbitMQConfig{URL: "amqp://guest:guest@localhost:5672/"},
		EMQX:     config.EMQXConfig{URL: "tcp://localhost:1883"},
		Logging:  config.LoggingConfig{Level: "info", Format: "json"},
		JWT:      config.JWTConfig{Secret: "test-secret"},
	}
}

// dbStub satisfies repository.DB without a live database.
type dbStub struct{}

func (s *dbStub) BeginTransaction(ctx context.Context) (repository.Transaction, error) {
	return nil, nil
}

func (s *dbStub) WithTransaction(ctx context.Context, fn func(tx repository.Transaction) error) error {
	return nil
}

func (s *dbStub) Raw(ctx context.Context, query string, dest any, args ...any) error {
	return nil
}

func (s *dbStub) Query(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	return nil, nil
}

func (s *dbStub) Close(ctx context.Context) error {
	return nil
}

// testContainer builds a fully wired Container with stubbed infrastructure,
// avoiding the need for a live database.
func testContainer(t *testing.T) *Container {
	t.Helper()

	cfg := testConfig()
	c := &Container{Config: cfg}

	logger, err := initLogger(cfg.Logging)
	if err != nil {
		t.Fatalf("init logger: %v", err)
	}

	c.Logger = logger

	c.DB = &dbStub{}
	c.Cache = &cacheStub{}
	c.Queue = &queueStub{}
	c.MQTT = &mqttStub{}

	if err := c.wireServices(); err != nil {
		t.Fatalf("wire services: %v", err)
	}

	return c
}

func TestBootstrap(t *testing.T) {
	container := testContainer(t)

	if container.Router == nil {
		t.Error("router is nil")
	}

	if container.Logger == nil {
		t.Error("logger is nil")
	}

	if container.DB == nil || container.Cache == nil || container.Queue == nil || container.MQTT == nil {
		t.Error("infrastructure dependency is nil")
	}
}

func TestBootstrapMissingDatabase(t *testing.T) {
	cfg := testConfig()
	cfg.Database = config.DatabaseConfig{}

	_, err := Bootstrap(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for missing database config, got nil")
	}

	if !strings.Contains(err.Error(), "database url is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCloseReverseOrder(t *testing.T) {
	c := &Container{}

	var order []string

	c.registerClose(func(context.Context) error {
		order = append(order, "first")
		return nil
	})
	c.registerClose(func(context.Context) error {
		order = append(order, "second")
		return nil
	})
	c.registerClose(func(context.Context) error {
		order = append(order, "third")
		return nil
	})

	if err := c.Close(context.Background()); err != nil {
		t.Fatalf("Close: %v", err)
	}

	want := []string{"third", "second", "first"}
	if !slices.Equal(order, want) {
		t.Errorf("close order = %v, want %v", order, want)
	}
}

func TestHealthEndpoint(t *testing.T) {
	container := testContainer(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	container.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if !strings.Contains(rec.Body.String(), `"ok"`) {
		t.Errorf("body = %q, want status ok", rec.Body.String())
	}
}
