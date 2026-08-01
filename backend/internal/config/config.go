package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
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
	RateLimit RateLimitConfig
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host string
	Port int
	Env  string
}

// DatabaseConfig holds the primary database connection settings.
type DatabaseConfig struct {
	URL string
}

// RedisConfig holds the cache connection settings.
type RedisConfig struct {
	URL string
}

// RabbitMQConfig holds the message queue connection settings.
type RabbitMQConfig struct {
	URL string
}

// SMTPConfig holds the email delivery settings.
type SMTPConfig struct {
	Host string
	Port int
	User string
	Pass string
}

// EMQXConfig holds the MQTT broker connection settings.
type EMQXConfig struct {
	URL string
}

// LoggingConfig holds structured logging settings.
type LoggingConfig struct {
	Level  string
	Format string
}

// OTELConfig holds OpenTelemetry tracing settings.
type OTELConfig struct {
	Enabled  bool
	Endpoint string
}

// JWTConfig holds authentication token settings.
type JWTConfig struct {
	Secret     string
	Expiration string
}

// RateLimitConfig holds rate limiting settings.
type RateLimitConfig struct {
	RequestsPerMinute int `koanf:"requests_per_minute"`
}

func (r *RateLimitConfig) Unmarshal(k *koanf.Koanf) error {
	v := k.Int("rate.limit.requests_per_minute")
	if v > 0 {
		r.RequestsPerMinute = v
	}
	return nil
}

// Load reads configuration from a .env file (optional) and environment
// variables, with environment variables taking precedence.
func Load(configPath string) (*Config, error) {
	k := koanf.New(".")

	transform := func(key, value string) (string, interface{}) {
		return normalizeKey(key), value
	}

	if err := k.Load(file.Provider(configPath), dotenv.ParserEnvWithValue("FORGEBASE_", ".", transform)); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("load %s: %w", configPath, err)
		}
	}

	if err := k.Load(env.ProviderWithValue("FORGEBASE_", ".", transform), nil); err != nil {
		return nil, fmt.Errorf("load environment: %w", err)
	}

	cfg := &Config{}
	if err := k.Unmarshal("", cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Rate limit: manually parse from koanf to handle nested key mapping
	if cfg.RateLimit.RequestsPerMinute == 0 {
		if v := k.Int("rate.limit.requests.per.minute"); v > 0 {
			cfg.RateLimit.RequestsPerMinute = v
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate ensures all required settings are present.
func (c *Config) Validate() error {
	var errs []error

	if c.Server.Host == "" {
		errs = append(errs, errors.New("server.host is required"))
	}

	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		errs = append(errs, errors.New("server.port must be between 1 and 65535"))
	}

	if c.Server.Env == "" {
		errs = append(errs, errors.New("server.env is required"))
	}

	if c.Database.URL == "" {
		errs = append(errs, errors.New("database.url is required"))
	}

	if c.JWT.Secret == "" {
		errs = append(errs, errors.New("jwt.secret is required"))
	}

	if c.RateLimit.RequestsPerMinute <= 0 {
		errs = append(errs, errors.New("rate_limit.requests_per_minute must be greater than 0"))
	}

	if len(errs) == 0 {
		return nil
	}

	return fmt.Errorf("invalid configuration: %w", errors.Join(errs...))
}

func normalizeKey(key string) string {
	key = strings.TrimPrefix(key, "FORGEBASE_")

	return strings.ToLower(strings.ReplaceAll(key, "_", "."))
}
