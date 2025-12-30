package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	AppName  string `env:"APP_NAME" envDefault:"myapp"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`

	HTTP HTTPConfig
	DB   DBConfig
}

type HTTPConfig struct {
	Host string `env:"HTTP_HOST" envDefault:"0.0.0.0"`
	Port int    `env:"HTTP_PORT" envDefault:"8080"`
}

type DBConfig struct {
	Host     string `env:"DB_HOST" envDefault:"localhost"`
	Port     int    `env:"DB_PORT" envDefault:"5432"`
	Name     string `env:"DB_NAME,notEmpty"`
	User     string `env:"DB_USER,notEmpty"`
	Password string `env:"DB_PASSWORD,notEmpty"`
	SSLMode  string `env:"DB_SSL_MODE" envDefault:"disable"`
	MaxConns int    `env:"DB_MAX_CONNS" envDefault:"10"`
	MinConns int    `env:"DB_MIN_CONNS" envDefault:"2"`
}

func New() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode,
	)
}

func (c *HTTPConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
