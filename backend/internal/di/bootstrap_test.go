package di

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/muhammadyunus/Restify-Service/internal/config"
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

func TestBootstrap(t *testing.T) {
	container, err := Bootstrap(context.Background(), testConfig())
	if err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}

	if container == nil {
		t.Fatal("container is nil")
	}

	if container.Router == nil {
		t.Error("router is nil")
	}

	if container.Logger == nil {
		t.Error("logger is nil")
	}

	if container.DB == nil || container.Cache == nil || container.Queue == nil || container.MQTT == nil {
		t.Error("infrastructure dependency is nil")
	}

	if len(container.closer) == 0 {
		t.Error("no closers registered")
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
	container, err := Bootstrap(context.Background(), testConfig())
	if err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}

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
