# Package Structure Decision: pkg/ vs internal/

Decision guide for placing packages in `pkg/` or `internal/`.

## Decision Tree

```
Is it used by other pkg/* packages?
├── YES → MUST be in pkg/
└── NO → Continue...

Does it import from internal/*?
├── YES → MUST stay in internal/
└── NO → Continue...

Is it reusable by other projects/modules?
├── YES → pkg/
└── NO → internal/
```

## Package Placement Examples

| Package | Location | Reason |
|---------|----------|--------|
| observability/tracing | `pkg/observability/` | Reusable OTel wrapper, no internal deps |
| postgres pool | `pkg/pg/` | Generic DB client |
| resilience (retry, circuit breaker) | `pkg/resilience/` | Generic patterns |
| logger | `pkg/logger/` | Reusable logging wrapper |
| config | `internal/config/` | App-specific env vars |
| models | `internal/models/` | Domain-specific entities |
| services | `internal/services/` | Business logic |
| storage | `internal/storage/` | App-specific repositories |
| http handlers | `internal/http/` | App-specific API layer |

## Red Flags (Refactoring Needed)

1. **`pkg/*` imports `internal/*`** — architectural violation
   ```go
   // pkg/arp/client.go
   import "myapp/internal/observability"  // ❌ WRONG
   ```
   Fix: Move `observability` to `pkg/`

2. **`internal/*` without internal dependencies** — candidate for `pkg/`
   ```
   internal/observability/  ← has no imports from internal/*
   ```
   Question: Can this be reused? If yes → move to `pkg/`

3. **Code duplication between projects** — extract to `pkg/`
   ```
   project-a/internal/retry/
   project-b/internal/retry/  ← same code
   ```
   Fix: Create shared `pkg/resilience/` module

## Observability Placement

```
pkg/observability/               # Infrastructure setup
├── observability.go             # NewTracerProvider(), NewMeterProvider()
└── observability_test.go

internal/services/               # Usage in business logic
└── order.go                     # tracer.Start(ctx, "CreateOrder")
```

**Rule:**
- Observability **infrastructure** (provider setup, helpers) → `pkg/`
- Observability **usage** (spans in business logic) → stays in `internal/services/`

## Quick Reference

| Question | YES | NO |
|----------|-----|-----|
| Used by `pkg/*`? | → `pkg/` | Continue |
| Imports `internal/*`? | → `internal/` | Continue |
| Reusable across projects? | → `pkg/` | → `internal/` |

## Common pkg/ Candidates

```
pkg/
├── logger/          # Logging wrapper (slog/zap)
├── postgres/        # Database client
├── observability/   # OTel tracing/metrics setup
├── resilience/      # Retry, circuit breaker
├── httputil/        # HTTP client helpers
└── validation/      # Generic validation utilities
```

## Common internal/ Packages

```
internal/
├── config/          # App-specific configuration
├── models/          # Domain entities
├── services/        # Business logic
├── storage/         # Repositories
├── http/            # HTTP handlers, middleware
└── workers/         # Background job processors
```
