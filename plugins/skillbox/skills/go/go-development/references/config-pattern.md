# Configuration Pattern

Using `caarlos0/env` for environment-only configuration.

## Pattern

```go
package config

import (
    "github.com/caarlos0/env/v10"
)

type Config struct {
    AppName  string `env:"APP_NAME" envDefault:"myapp"`
    LogLevel string `env:"LOG_LEVEL" envDefault:"info"`

    HTTP     HTTPConfig
    DB       DBConfig
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
```

## Key Features

- **env only** — no YAML/JSON files
- **notEmpty** — required fields
- **envDefault** — sensible defaults
- **Nested structs** — logical grouping

## Usage

```go
func main() {
    cfg, err := config.New()
    if err != nil {
        log.Fatalf("config: %v", err)
    }

    // Use cfg.DB.DSN(), cfg.HTTP.Port, etc.
}
```

## Environment File

```bash
# .env.example
APP_NAME=myapp
LOG_LEVEL=info

HTTP_HOST=0.0.0.0
HTTP_PORT=8080

DB_HOST=localhost
DB_PORT=5432
DB_NAME=mydb
DB_USER=myuser
DB_PASSWORD=mypassword
DB_SSL_MODE=disable
DB_MAX_CONNS=10
```
