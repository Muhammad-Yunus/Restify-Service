package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const envFileContent = `FORGEBASE_SERVER_HOST=127.0.0.1
FORGEBASE_SERVER_PORT=9090
FORGEBASE_SERVER_ENV=test
FORGEBASE_DATABASE_URL=postgres://u:p@localhost:5432/db?sslmode=disable
FORGEBASE_REDIS_URL=redis://localhost:6379/0
FORGEBASE_RABBITMQ_URL=amqp://guest:guest@localhost:5672/
FORGEBASE_SMTP_HOST=smtp.test.com
FORGEBASE_SMTP_PORT=587
FORGEBASE_SMTP_USER=user@test.com
FORGEBASE_SMTP_PASS=secret
FORGEBASE_EMQX_URL=tcp://localhost:1883
FORGEBASE_LOGGING_LEVEL=debug
FORGEBASE_LOGGING_FORMAT=text
FORGEBASE_OTEL_ENABLED=true
FORGEBASE_OTEL_ENDPOINT=http://localhost:4317
FORGEBASE_JWT_SECRET=test-secret
FORGEBASE_JWT_EXPIRATION=1h
FORGEBASE_RATE_LIMIT_REQUESTS_PER_MINUTE=60
`

func writeEnvFile(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte(envFileContent), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	return path
}

func TestLoadFromEnvFile(t *testing.T) {
	cfg, err := Load(writeEnvFile(t))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Server.Host = %q, want 127.0.0.1", cfg.Server.Host)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %d, want 9090", cfg.Server.Port)
	}

	if cfg.Server.Env != "test" {
		t.Errorf("Server.Env = %q, want test", cfg.Server.Env)
	}

	if cfg.Database.URL != "postgres://u:p@localhost:5432/db?sslmode=disable" {
		t.Errorf("Database.URL = %q", cfg.Database.URL)
	}

	if cfg.Redis.URL != "redis://localhost:6379/0" {
		t.Errorf("Redis.URL = %q", cfg.Redis.URL)
	}

	if cfg.RabbitMQ.URL != "amqp://guest:guest@localhost:5672/" {
		t.Errorf("RabbitMQ.URL = %q", cfg.RabbitMQ.URL)
	}

	if cfg.SMTP.Host != "smtp.test.com" || cfg.SMTP.Port != 587 {
		t.Errorf("SMTP = %+v", cfg.SMTP)
	}

	if cfg.SMTP.User != "user@test.com" || cfg.SMTP.Pass != "secret" {
		t.Errorf("SMTP credentials = %+v", cfg.SMTP)
	}

	if cfg.EMQX.URL != "tcp://localhost:1883" {
		t.Errorf("EMQX.URL = %q", cfg.EMQX.URL)
	}

	if cfg.Logging.Level != "debug" || cfg.Logging.Format != "text" {
		t.Errorf("Logging = %+v", cfg.Logging)
	}

	if !cfg.OTEL.Enabled || cfg.OTEL.Endpoint != "http://localhost:4317" {
		t.Errorf("OTEL = %+v", cfg.OTEL)
	}

	if cfg.JWT.Secret != "test-secret" || cfg.JWT.Expiration != "1h" {
		t.Errorf("JWT = %+v", cfg.JWT)
	}

	if cfg.RateLimit.RequestsPerMinute != 60 {
		t.Errorf("RateLimit.RequestsPerMinute = %d, want 60", cfg.RateLimit.RequestsPerMinute)
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	t.Setenv("FORGEBASE_SERVER_HOST", "10.0.0.1")
	t.Setenv("FORGEBASE_SERVER_PORT", "8080")
	t.Setenv("FORGEBASE_SERVER_ENV", "production")
	t.Setenv("FORGEBASE_DATABASE_URL", "postgres://env:env@localhost:5432/db")
	t.Setenv("FORGEBASE_JWT_SECRET", "env-secret")
	t.Setenv("FORGEBASE_RATE_LIMIT_REQUESTS_PER_MINUTE", "120")

	cfg, err := Load(filepath.Join(t.TempDir(), ".env"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Server.Host != "10.0.0.1" {
		t.Errorf("Server.Host = %q, want 10.0.0.1", cfg.Server.Host)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want 8080", cfg.Server.Port)
	}

	if cfg.Database.URL != "postgres://env:env@localhost:5432/db" {
		t.Errorf("Database.URL = %q", cfg.Database.URL)
	}

	if cfg.JWT.Secret != "env-secret" {
		t.Errorf("JWT.Secret = %q, want env-secret", cfg.JWT.Secret)
	}

	if cfg.RateLimit.RequestsPerMinute != 120 {
		t.Errorf("RateLimit.RequestsPerMinute = %d, want 120", cfg.RateLimit.RequestsPerMinute)
	}
}

func TestLoadMissingRequiredConfig(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), ".env")); err == nil {
		t.Fatal("expected error for missing required config, got nil")
	}
}

func TestValidateReportsMissingFields(t *testing.T) {
	err := (&Config{}).Validate()
	if err == nil {
		t.Fatal("expected error for empty config, got nil")
	}

	for _, want := range []string{"server.host", "server.port", "server.env", "database.url", "jwt.secret"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error missing %q: %v", want, err)
		}
	}
}
