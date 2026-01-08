---
name: go-development
description: This skill should be used when the user is working with Go projects (go.mod, *.go files), asking about "Go patterns", "Go architecture", "services", "repositories", or mentions "Go development", "Go project", "Go handlers", "Go testing".
---

# Go Development Guide

Production-ready patterns extracted from real projects.

## Table of Contents

- [Core Principle: LESS CODE = FEWER BUGS](#core-principle-less-code--fewer-bugs)
- [Architecture Decisions](#architecture-decisions)
- [Linter Enforcement](#linter-enforcement)
- [Enforced Project Structure](#enforced-project-structure)
- [References](#references)
  - [Core Patterns](#core-patterns)
  - [HTTP Layer](#http-layer)
  - [Production Patterns](#production-patterns)
  - [Best Practices](#best-practices)
- [Examples](#examples)
- [Templates](#templates)
- [Commands](#commands)
- [Dependencies](#dependencies)
- [Version](#version)

## Core Principle: LESS CODE = FEWER BUGS

**This applies to ALL code — not just project scaffolding.**

### What this means:

- ❌ Don't write error types that won't be used
- ❌ Don't create custom error type for every error case
- ❌ Don't add methods "for completeness"
- ❌ Don't create helper functions that have one caller
- ❌ Don't add validation for impossible scenarios
- ❌ Don't implement interfaces "just in case"
- ❌ Don't add logging/metrics that nobody reads
- ❌ Don't create `common`, `helpers`, `utils`, `shared`, `misc` packages

- ✅ Write only what the feature requires
- ✅ Use few sentinel categories (ErrNotFound, ErrConflict) + wrap with context
- ✅ Typed errors only when caller needs to extract data (RetryAfter, Field)
- ✅ Add code when there's actual need
- ✅ Delete unused code immediately
- ✅ Prefer inline code over tiny functions
- ✅ Trust the type system, don't over-validate
- ✅ Name packages by what they provide, not what's in them
- ✅ Keep helper functions close to usage (unexported)
- ✅ If truly shared, create small purpose-named packages (`internal/optional/`)

- ❌ Don't modify `.golangci.yml` — linting config is protected

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
| **Errors** | Sentinel categories + wrap (errs package) |
| **Logging** | slog (small) / zap (large) — ask user |
| **Tracing** | OpenTelemetry (optional) — ask user |
| **Migrations** | `goose/v3` |

## Linter Enforcement

All rules are enforced via `golangci-lint` (revive rules) in `.golangci.yml`:

| Rule | Linter | Description |
|------|--------|-------------|
| `userID` not `userId` | `var-naming` | Go idiom: acronyms in caps |
| `any` not `interface{}` | `use-any` | Go 1.18+ alias |
| No `common/helpers/utils` packages | `var-naming` | `extraBadPackageNames` |
| Error wrapping | `err113`, `errorlint` | Sentinel errors + wrap |
| Test helpers | `thelper` | Must use `t.Helper()` |
| Test parallelism | `tparallel` | Suggests `t.Parallel()` |
| Test env vars | `tenv` | Detects `os.Setenv` in tests |

### Enforcement Layers

1. **SessionStart** — Context injection with linter rules reminder
2. **Agents/Commands** — Self-enforce by running `golangci-lint` after code generation
3. **Stop hook** — Blocks session end if Go files modified but lint not run

### Required Validation

After generating/modifying Go code:

```bash
golangci-lint run ./...
```

**Do not modify `.golangci.yml`** — linting config is protected by `golangci-guard` hook.

## Enforced Project Structure

The skill MUST enforce this structure for all Go projects:

```
project/
├── cmd/
│   └── app/
│       └── main.go           # Entry point only
├── internal/
│   ├── config/
│   │   └── config.go         # Config with envPrefix
│   ├── errs/
│   │   └── errors.go         # Sentinel errors + helpers
│   ├── optional/
│   │   └── optional.go       # Pointer conversion helpers
│   ├── models/
│   │   └── {entity}.go       # Domain models + mappers
│   ├── services/
│   │   ├── registry.go       # Service registry
│   │   └── {entity}.go       # Business logic
│   ├── storage/
│   │   ├── storage.go        # Storage interface
│   │   ├── {entity}.go       # Repository impl
│   │   ├── main_test.go      # TestMain with testcontainers
│   │   └── testmigration/    # Test data fixtures (SQL)
│   │       ├── 100001_users.up.sql
│   │       └── 100001_users.down.sql
│   └── http/
│       └── v1/
│           ├── router.go     # Router + path constants
│           ├── {entity}_handler.go
│           ├── dto.go        # Request/Response types
│           └── json.go       # encode/decode
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

### Structure Rules

| Rule | Correct | Wrong |
|------|---------|-------|
| Handler files | `user_handler.go`, `order_handler.go` | `handlers.go` (all in one) |
| Mappers | In `models/{entity}.go` with model | In separate `mappers/` package |
| DTOs | In `http/v1/dto.go` | Mixed with domain models |
| Path constants | In `router.go` | Hardcoded strings in handlers |
| IDs | `string` type | `uuid.UUID` type |
| ID generation | `uuid.NewString()` in service | In handler or repository |
| Config nesting | Use `envPrefix` tag | Full env var names in nested structs |
| Doc comments | `// User represents...` | `// This struct...` |
| Section organization | Separate files | `// ----- Section -----` |

## References

### Core Patterns

| Pattern | File |
|---------|------|
| Entry Point | [entrypoint-pattern.md](references/entrypoint-pattern.md) |
| Configuration | [config-pattern.md](references/config-pattern.md) |
| Package Structure | [package-structure-decision.md](references/package-structure-decision.md) |
| Database & Transactions | [database-pattern.md](references/database-pattern.md) |
| Advisory Locks | [advisory-lock-pattern.md](references/advisory-lock-pattern.md) |
| Service Layer | [service-pattern.md](references/service-pattern.md) |
| Repository | [repository-pattern.md](references/repository-pattern.md) |
| Filter Pattern | [filter-pattern.md](references/filter-pattern.md) |
| Mapper | [mapper-pattern.md](references/mapper-pattern.md) |
| JSONB Types | [jsonb-pattern.md](references/jsonb-pattern.md) |
| Optional Helper | [optional-pattern.md](references/optional-pattern.md) |
| Error Handling | [error-handling.md](references/error-handling.md) |
| Logging | [logging-pattern.md](references/logging-pattern.md) |
| Testing | [testing-pattern.md](references/testing-pattern.md) |
| Test Fixtures | [test-fixtures-pattern.md](references/test-fixtures-pattern.md) |
| Money | [money-pattern.md](references/money-pattern.md) |
| Build & Deploy | [build-deploy.md](references/build-deploy.md) |

### HTTP Layer

| Pattern | File |
|---------|------|
| HTTP Handlers | [http-handler-pattern.md](references/http-handler-pattern.md) |
| Middleware | [middleware-pattern.md](references/middleware-pattern.md) |
| Validation | [validation-pattern.md](references/validation-pattern.md) |
| Authentication | [auth-pattern.md](references/auth-pattern.md) |

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
| Package Naming | [package-naming.md](references/package-naming.md) |
| Control Structures | [control-structures.md](references/control-structures.md) |
| Interface Design | [interface-design.md](references/interface-design.md) |
| Allocation (new/make) | [allocation-patterns.md](references/allocation-patterns.md) |
| Defer | [defer-patterns.md](references/defer-patterns.md) |
| Embedding | [embedding-patterns.md](references/embedding-patterns.md) |
| Blank Identifier | [blank-identifier.md](references/blank-identifier.md) |
| Concurrency | [concurrency-pattern.md](references/concurrency-pattern.md) |
| Channel Axioms | [channel-axioms.md](references/channel-axioms.md) |
| Linting | [linting-pattern.md](references/linting-pattern.md) |
| Common Pitfalls | [common-pitfalls.md](references/common-pitfalls.md) |
| Performance | [performance-tips.md](references/performance-tips.md) |
| Code Quality | [code-quality.md](references/code-quality.md) |

## Examples

### Core

| Component | File |
|-----------|------|
| Main | [main.go](examples/main.go) |
| Backend | [backend.go](examples/backend.go) |
| Config | [config.go](examples/config.go) |
| Filter | [filter.go](examples/filter.go) |
| Common Models | [common_models.go](examples/common_models.go) |
| Common Storage | [common_storage.go](examples/common_storage.go) |
| Database Client | [pg-client.go](examples/pg-client.go) |
| Advisory Lock | [advisory_lock.go](examples/advisory_lock.go) |
| Repository | [repository.go](examples/repository.go) |
| Service | [service.go](examples/service.go) |
| Mapper | [mapper.go](examples/mapper.go) |
| JSONB Types | [jsonb.go](examples/jsonb.go) |
| Optional Helper | [optional.go](examples/optional.go) |
| Errors | [errors.go](examples/errors.go) |
| Logger (slog) | [logger_slog.go](examples/logger_slog.go) |
| Logger (zap) | [logger_zap.go](examples/logger_zap.go) |
| Test Setup | [main_test.go](examples/main_test.go) |
| Test Fixtures (SQL) | [testmigration_example.sql](examples/testmigration_example.sql) |
| Repository Tests | [repository_test.go](examples/repository_test.go) |
| Service Tests | [service_test.go](examples/service_test.go) |
| Money | [money.go](examples/money.go) |
| Money Tests | [money_test.go](examples/money_test.go) |

### HTTP Layer

| Component | File |
|-----------|------|
| HTTP Handler | [handler.go](examples/handler.go) |
| Middleware | [middleware.go](examples/middleware.go) |
| HTTP Errors | [http_errors.go](examples/http_errors.go) |
| Authentication | [auth.go](examples/auth.go) |

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

**ALWAYS use `@latest`** — enforced by PreToolUse hook.

| Pattern | Action |
|---------|--------|
| `go get pkg` (no version) | Hook suggests adding `@latest` |
| `go get pkg@latest` | Approved |
| `go get pkg@v1.2.3` | Hook reminds to verify via Context7 |

Recommended libraries:

```bash
# Core
go get golang.org/x/sync/errgroup@latest
go get github.com/caarlos0/env/v10@latest
go get github.com/jackc/pgx/v5@latest
go get github.com/Masterminds/squirrel@latest
go get github.com/pressly/goose/v3@latest
go get github.com/avast/retry-go@latest
go get github.com/google/uuid@latest
go get github.com/shopspring/decimal@latest

# HTTP Layer
go get github.com/go-chi/chi/v5@latest
go get github.com/go-playground/validator/v10@latest
go get gopkg.in/go-jose/go-jose.v2@latest

# Testing
go get github.com/stretchr/testify@latest
go get github.com/testcontainers/testcontainers-go@latest

# Tracing (OpenTelemetry)
go get go.opentelemetry.io/otel@latest
go get go.opentelemetry.io/otel/sdk@latest
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc@latest
go get go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp@latest
go get github.com/exaring/otelpgx@latest
```

## Version

- 1.20.0 — Advisory lock pattern for serializable transactions: Serialize() method, enforcement hook, /go-add-repository includes Serialize() by default
- 1.19.0 — Filter pattern for database queries: XxxFilter, getXxxCondition(), /go-add-repository generates filter scaffold
- 1.18.0 — Dependency version enforcement: PreToolUse hook for `go get @latest`
- 1.17.0 — Testcontainers integration: typed closer, testmigration/ fixtures, goose with sql.Open
- 1.16.1 — Fix golangci-lint v2 schema: imports-blocklist, remove deprecated options
- 1.16.0 — Protected .golangci.yml (hook + documentation)
- 1.15.0 — Package naming: no common/helpers/utils, purpose-named packages (optional/, json.go)
- 1.14.0 — Error handling: sentinel categories + wrap (internal/errs), helpers, no re-wrap
- 1.13.1 — golangci-lint v2 fix: typecheck is built-in, wsl (not wsl_v5), troubleshooting
- 1.13.0 — golangci-lint v2 migration (formatters, nolintlint anti-cheat, err113)
- 1.12.0 — Comprehensive linting configuration (revive, depguard, exclusion patterns)
- 1.11.1 — Channel axioms (nil/closed behavior, WaitMany, broadcast signaling)
- 1.11.0 — Effective Go patterns (control structures, interfaces, allocation, defer, embedding, blank identifier)
- 1.10.3 — Package structure decision guide (pkg/ vs internal/)
- 1.10.2 — Config: embedded structs only (no named fields in main Config)
- 1.10.1 — Fix: UserColumns() in models package only (not repository)
- 1.10.0 — Repository pattern: Save as upsert, UserColumns() function, Values() method
- 1.9.0 — Authentication patterns (gateway-based, full JWT validation)
- 1.8.0 — Code quality tools (deadcode analysis, make deadcode target)
- 1.7.1 — Comment conventions (stdlib-style doc comments, no decorative separators)
- 1.7.0 — Handler-per-entity, Optional[T], JSONB types, Mapper pattern, Config envPrefix, IDs as string
- 1.6.1 — Money pattern: NUMERIC database storage (was TEXT)
- 1.6.0 — Money pattern (decimal precision, currency conversion, exchange rates)
- 1.5.0 — Enhanced testing (testcontainers, testify mock, CI/Local detection)
- 1.4.0 — Entry point pattern (backend, errgroup, graceful shutdown)
- 1.3.0 — Best practices (naming, concurrency, pitfalls, performance)
- 1.2.0 — Simple errors, OpenTelemetry tracing
- 1.1.0 — HTTP Layer, Enhanced Errors, Production Patterns (pagination, health, workers)
- 1.0.0 — Initial release
