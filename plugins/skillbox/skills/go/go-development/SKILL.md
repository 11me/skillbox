---
name: go-development
description: |
  Primary Go development guide. Trigger when:
  - Working with Go projects (go.mod detected)
  - User asks about Go patterns, architecture
  - `/go` command invoked
  - Creating/scaffolding Go services

  Covers: Config, Database, Services, Repositories, Errors, Logging, Testing,
  HTTP Handlers, Middleware, Validation, Pagination, Health Checks, Workers, Tracing,
  Naming, Concurrency, Pitfalls, Performance.
triggers:
  - go.mod
  - "*.go"
  - /go
---

# Go Development Guide

Production-ready patterns extracted from real projects.

## Core Principle: LESS CODE = FEWER BUGS

**This applies to ALL code — not just project scaffolding.**

### What this means:

- ❌ Don't write error types that won't be used
- ❌ Don't add methods "for completeness"
- ❌ Don't create helper functions that have one caller
- ❌ Don't add validation for impossible scenarios
- ❌ Don't implement interfaces "just in case"
- ❌ Don't add logging/metrics that nobody reads

- ✅ Write only what the feature requires
- ✅ Add code when there's actual need
- ✅ Delete unused code immediately
- ✅ Prefer inline code over tiny functions
- ✅ Trust the type system, don't over-validate

### Examples:

| Bad | Good |
|-----|------|
| Define 10 error types, use 2 | Define errors as you need them |
| Create `GetByID`, `GetByEmail`, `GetByName` upfront | Create only the method you're using now |
| Add `Update`, `Delete` when only `Create` is needed | Add `Update` when the feature requires it |
| Write helper `formatUserName()` called once | Inline the logic |
| Validate internal struct fields | Trust internal code, validate at boundaries |

## Architecture Decisions

| Aspect | Choice |
|--------|--------|
| **Config** | `caarlos0/env` (env only) |
| **Structure** | `cmd/`, `internal/`, `pkg/` |
| **Database** | `pgx/v5` + `squirrel` |
| **Transactions** | Context injection + retry |
| **DI** | Service Registry |
| **Errors** | Simple typed errors with HTTP mapping |
| **Logging** | slog (small) / zap (large) — ask user |
| **Tracing** | OpenTelemetry (optional) — ask user |
| **Migrations** | `goose/v3` |

## Project Structure

```
project/
├── cmd/
│   └── app/
│       └── main.go
├── internal/
│   ├── config/
│   ├── models/
│   ├── services/
│   │   └── registry.go
│   ├── storage/
│   └── common/
│       └── errors.go
├── pkg/
│   ├── logger/
│   └── postgres/
├── migrations/
├── go.mod
├── .env.example
├── Makefile
├── Dockerfile
└── docker-compose.yml
```

## References

### Core Patterns

| Pattern | File |
|---------|------|
| Configuration | [config-pattern.md](references/config-pattern.md) |
| Database & Transactions | [database-pattern.md](references/database-pattern.md) |
| Service Layer | [service-pattern.md](references/service-pattern.md) |
| Repository | [repository-pattern.md](references/repository-pattern.md) |
| Error Handling | [error-handling.md](references/error-handling.md) |
| Logging | [logging-pattern.md](references/logging-pattern.md) |
| Testing | [testing-pattern.md](references/testing-pattern.md) |
| Build & Deploy | [build-deploy.md](references/build-deploy.md) |

### HTTP Layer

| Pattern | File |
|---------|------|
| HTTP Handlers | [http-handler-pattern.md](references/http-handler-pattern.md) |
| Middleware | [middleware-pattern.md](references/middleware-pattern.md) |
| Validation | [validation-pattern.md](references/validation-pattern.md) |

### Production Patterns

| Pattern | File |
|---------|------|
| Pagination | [pagination-pattern.md](references/pagination-pattern.md) |
| Health Checks | [health-check-pattern.md](references/health-check-pattern.md) |
| Background Workers | [worker-pattern.md](references/worker-pattern.md) |
| Tracing | [tracing-pattern.md](references/tracing-pattern.md) |

### Best Practices

| Topic | File |
|-------|------|
| Naming Conventions | [naming-conventions.md](references/naming-conventions.md) |
| Concurrency | [concurrency-pattern.md](references/concurrency-pattern.md) |
| Common Pitfalls | [common-pitfalls.md](references/common-pitfalls.md) |
| Performance | [performance-tips.md](references/performance-tips.md) |

## Examples

### Core

| Component | File |
|-----------|------|
| Config | [config.go](examples/config.go) |
| Database Client | [pg-client.go](examples/pg-client.go) |
| Repository | [repository.go](examples/repository.go) |
| Service | [service.go](examples/service.go) |
| Errors | [errors.go](examples/errors.go) |
| Logger (slog) | [logger_slog.go](examples/logger_slog.go) |
| Logger (zap) | [logger_zap.go](examples/logger_zap.go) |
| Tests | [service_test.go](examples/service_test.go) |
| Main | [main.go](examples/main.go) |

### HTTP Layer

| Component | File |
|-----------|------|
| HTTP Handler | [handler.go](examples/handler.go) |
| Middleware | [middleware.go](examples/middleware.go) |
| HTTP Errors | [http_errors.go](examples/http_errors.go) |

### Production

| Component | File |
|-----------|------|
| Pagination | [pagination.go](examples/pagination.go) |
| Health Check | [health.go](examples/health.go) |
| Worker | [worker.go](examples/worker.go) |
| Tracing | [tracing.go](examples/tracing.go) |

## Templates

| File | Purpose |
|------|---------|
| [Dockerfile](templates/Dockerfile) | Multi-stage build |
| [Makefile](templates/Makefile) | Build automation |
| [docker-compose.yml](templates/docker-compose.yml) | Local development |
| [.env.example](templates/.env.example) | Environment template |

## Commands

| Command | Description |
|---------|-------------|
| `/go-add-service` | Generate service + interface |
| `/go-add-repository` | Generate repository + interface |
| `/go-add-model` | Generate model + mapper |

## Dependencies

Always use latest versions:

```bash
# Core
go get github.com/caarlos0/env/v10@latest
go get github.com/jackc/pgx/v5@latest
go get github.com/Masterminds/squirrel@latest
go get github.com/pressly/goose/v3@latest
go get github.com/avast/retry-go@latest
go get github.com/google/uuid@latest

# HTTP Layer
go get github.com/go-chi/chi/v5@latest
go get github.com/go-playground/validator/v10@latest

# Testing
go get github.com/stretchr/testify@latest

# Tracing (OpenTelemetry)
go get go.opentelemetry.io/otel@latest
go get go.opentelemetry.io/otel/sdk@latest
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc@latest
go get go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp@latest
go get github.com/exaring/otelpgx@latest
```

## Version

- 1.3.0 — Best practices (naming, concurrency, pitfalls, performance)
- 1.2.0 — Simple errors, OpenTelemetry tracing
- 1.1.0 — HTTP Layer, Enhanced Errors, Production Patterns (pagination, health, workers)
- 1.0.0 — Initial release
