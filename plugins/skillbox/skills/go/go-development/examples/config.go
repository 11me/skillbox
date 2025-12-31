package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
)

// Config is the main application configuration.
// Uses envPrefix to cleanly scope nested struct env vars.
type Config struct {
	AppName  string `env:"APP_NAME" envDefault:"myapp"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`

	Server ServerConfig `envPrefix:"SERVER_"`
	DB     DBConfig     `envPrefix:"DB_"`
	Redis  RedisConfig  `envPrefix:"REDIS_"`
	Queue  QueueConfig  `envPrefix:"QUEUE_"`
}

// ServerConfig holds HTTP server settings.
// Env vars: SERVER_USE_TLS, SERVER_LISTEN_ADDR
type ServerConfig struct {
	UseTLS     bool   `env:"USE_TLS" envDefault:"false"`
	ListenAddr string `env:"LISTEN_ADDR" envDefault:"0.0.0.0:8080"`
}

// DBConfig holds database connection settings.
// Env vars: DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD, etc.
type DBConfig struct {
	Host             string `env:"HOST" envDefault:"localhost"`
	Port             int    `env:"PORT" envDefault:"5432"`
	Name             string `env:"NAME,notEmpty"`
	User             string `env:"USER,notEmpty"`
	Password         string `env:"PASSWORD,notEmpty"`
	SSLMode          string `env:"SSL_MODE" envDefault:"disable"`
	MaxConns         int    `env:"MAX_CONNS" envDefault:"10"`
	MinConns         int    `env:"MIN_CONNS" envDefault:"2"`
	MigrationEnabled bool   `env:"MIGRATION_ENABLED" envDefault:"true"`
}

// RedisConfig holds Redis connection settings.
// Env vars: REDIS_ADDR, REDIS_PASSWORD, REDIS_DB
type RedisConfig struct {
	Addr     string `env:"ADDR" envDefault:"localhost:6379"`
	Password string `env:"PASSWORD"`
	DB       int    `env:"DB" envDefault:"0"`
}

// QueueConfig holds queue worker settings.
// Env vars: QUEUE_WORKERS
type QueueConfig struct {
	Workers int `env:"WORKERS" envDefault:"5"`
}

// New parses environment variables into Config struct.
func New() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// DSN returns PostgreSQL connection string.
func (c *DBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode,
	)
}
