# Configuration Pattern

Using `caarlos0/env` for environment-only configuration with `envPrefix` for clean nested structs.

## Pattern

```go
package config

import (
    "fmt"

    "github.com/caarlos0/env/v10"
)

type Config struct {
    AppName  string `env:"APP_NAME" envDefault:"myapp"`
    LogLevel string `env:"LOG_LEVEL" envDefault:"info"`

    Server ServerConfig `envPrefix:"SERVER_"`
    DB     DBConfig     `envPrefix:"DB_"`
    Redis  RedisConfig  `envPrefix:"REDIS_"`
    Queue  QueueConfig  `envPrefix:"QUEUE_"`
}

type ServerConfig struct {
    UseTLS     bool   `env:"USE_TLS" envDefault:"false"`
    ListenAddr string `env:"LISTEN_ADDR" envDefault:"0.0.0.0:8080"`
}

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

type RedisConfig struct {
    Addr     string `env:"ADDR" envDefault:"localhost:6379"`
    Password string `env:"PASSWORD"`
    DB       int    `env:"DB" envDefault:"0"`
}

type QueueConfig struct {
    Workers int `env:"WORKERS" envDefault:"5"`
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

## Key Feature: envPrefix

The `envPrefix` tag automatically prefixes all child struct env vars:

```go
type Config struct {
    Server ServerConfig `envPrefix:"SERVER_"`  // ← prefix defined here
}

type ServerConfig struct {
    ListenAddr string `env:"LISTEN_ADDR"`  // → SERVER_LISTEN_ADDR
    UseTLS     bool   `env:"USE_TLS"`      // → SERVER_USE_TLS
}
```

### Without envPrefix (old way)

```go
// ❌ Verbose, error-prone
type ServerConfig struct {
    ListenAddr string `env:"SERVER_LISTEN_ADDR"`
    UseTLS     bool   `env:"SERVER_USE_TLS"`
}
```

### With envPrefix (new way)

```go
// ✅ Clean, prefix defined once
type Config struct {
    Server ServerConfig `envPrefix:"SERVER_"`
}

type ServerConfig struct {
    ListenAddr string `env:"LISTEN_ADDR"`
    UseTLS     bool   `env:"USE_TLS"`
}
```

## Environment Variables

The config above reads these env vars:

| Struct | Field | Env Var |
|--------|-------|---------|
| Config | AppName | `APP_NAME` |
| Config | LogLevel | `LOG_LEVEL` |
| ServerConfig | UseTLS | `SERVER_USE_TLS` |
| ServerConfig | ListenAddr | `SERVER_LISTEN_ADDR` |
| DBConfig | Host | `DB_HOST` |
| DBConfig | Port | `DB_PORT` |
| DBConfig | Name | `DB_NAME` |
| DBConfig | User | `DB_USER` |
| DBConfig | Password | `DB_PASSWORD` |
| DBConfig | SSLMode | `DB_SSL_MODE` |
| DBConfig | MaxConns | `DB_MAX_CONNS` |
| DBConfig | MigrationEnabled | `DB_MIGRATION_ENABLED` |
| RedisConfig | Addr | `REDIS_ADDR` |
| RedisConfig | Password | `REDIS_PASSWORD` |
| QueueConfig | Workers | `QUEUE_WORKERS` |

## Tags Reference

| Tag | Purpose | Example |
|-----|---------|---------|
| `env:"NAME"` | Env var name | `env:"DB_HOST"` |
| `envDefault:"val"` | Default value | `envDefault:"localhost"` |
| `envPrefix:"PFX_"` | Prefix for nested struct | `envPrefix:"DB_"` |
| `notEmpty` | Required, fail if empty | `env:"PASSWORD,notEmpty"` |
| `expand` | Expand $VAR in value | `env:"URL,expand"` |
| `file` | Read from file path | `env:"SECRET,file"` |

## Usage

```go
func main() {
    cfg, err := config.New()
    if err != nil {
        log.Fatalf("config: %v", err)
    }

    // Access nested configs
    if cfg.DB.MigrationEnabled {
        runMigrations(cfg.DB)
    }

    // Use DSN helper
    db, err := pgx.Connect(ctx, cfg.DB.DSN())
}
```

## Environment File

```bash
# .env.example
APP_NAME=myapp
LOG_LEVEL=info

# Server
SERVER_USE_TLS=false
SERVER_LISTEN_ADDR=0.0.0.0:8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=mydb
DB_USER=myuser
DB_PASSWORD=mypassword
DB_SSL_MODE=disable
DB_MAX_CONNS=10
DB_MIN_CONNS=2
DB_MIGRATION_ENABLED=true

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Queue
QUEUE_WORKERS=5
```

## Best Practices

| DO | DON'T |
|----|-------|
| Use `envPrefix` for nested structs | Repeat prefix in every field |
| Use `notEmpty` for required fields | Panic on missing config |
| Provide sensible defaults | Require all values |
| Keep related settings together | Scatter settings across structs |
| Use typed fields (int, bool) | Parse everything as string |

## Dependencies

```bash
go get github.com/caarlos0/env/v10@latest
```
