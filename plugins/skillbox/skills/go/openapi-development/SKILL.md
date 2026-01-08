---
name: openapi-development
description: This skill should be used when the user asks about "OpenAPI", "spec-first API", "oapi-codegen", "REST API generation", or mentions "OpenAPI spec", "Go API code generation", "swagger".
---

# OpenAPI Development

Spec-first API development workflow combining OpenAPI 3.x specifications with Go code generation via oapi-codegen.

## Core Principles

1. **Spec-First**: Design API contract before implementation
2. **Modular Specs**: Separate files for paths, schemas, parameters, responses
3. **Generated Code**: Never manually edit `*.gen.go` files
4. **Validation**: Lint specs before generation, lint code after

## Workflow

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Design API     │ ──▶ │  Generate Code   │ ──▶ │  Implement      │
│  (OpenAPI spec) │     │  (oapi-codegen)  │     │  (handlers)     │
└─────────────────┘     └──────────────────┘     └─────────────────┘
        │                        │                        │
        ▼                        ▼                        ▼
   redocly lint           Chi + types +            Wire to services
                          server interface
```

## Commands

| Command | Description |
|---------|-------------|
| `/openapi-init` | Initialize modular OpenAPI spec structure |
| `/openapi-add-path` | Add new resource path with CRUD operations |
| `/openapi-generate` | Generate Go code from OpenAPI spec |

## Project Structure

After initialization:

```
project/
├── api/v1/
│   ├── openapi.yaml              # Main spec with $ref
│   ├── paths/
│   │   ├── _index.yaml           # Paths aggregator
│   │   └── users.yaml            # Resource paths
│   ├── components/
│   │   ├── schemas/_index.yaml   # Schemas aggregator
│   │   ├── parameters/_index.yaml
│   │   ├── responses/_index.yaml
│   │   └── requests/_index.yaml
│   └── .redocly.yaml             # Linter config
├── internal/http/v1/
│   ├── api.gen.go                # Generated (DO NOT EDIT)
│   ├── router.go                 # Chi router setup
│   └── handler_impl.go           # Interface implementation
├── oapi-codegen.yaml             # Generation config
└── Makefile                      # Includes openapi targets
```

## Makefile Targets

```makefile
openapi-lint:      # Lint OpenAPI spec with Redocly
openapi-bundle:    # Bundle spec (resolve $refs)
openapi-generate:  # Generate Go code from spec
openapi-preview:   # Preview API docs locally
```

## Default Configuration

### oapi-codegen.yaml

```yaml
package: api
generate:
  chi-server: true      # Chi router (recommended)
  models: true          # Request/Response types
  embedded-spec: false  # Don't embed spec in binary
output: internal/http/v1/api.gen.go
```

### Supported Routers

| Router | Config Key | Notes |
|--------|------------|-------|
| Chi | `chi-server: true` | Recommended, lightweight |
| Echo | `echo-server: true` | Full-featured framework |
| Gin | `gin-server: true` | High performance |
| Fiber | `fiber-server: true` | Express-like |
| net/http | `std-http-server: true` | No dependencies |

## Integration with go-development

Generated code integrates with existing service/repository patterns:

### Handler Implementation Pattern

```go
// internal/http/v1/handler_impl.go
package v1

import (
    "context"
    "net/http"

    "{{module}}/internal/services"
)

type Handler struct {
    services *services.Registry
}

func NewHandler(svc *services.Registry) *Handler {
    return &Handler{services: svc}
}

// Implement generated ServerInterface
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request, params ListUsersParams) {
    ctx := r.Context()

    users, err := h.services.UserService().List(ctx, params.Limit, params.Offset)
    if err != nil {
        // Map domain error to HTTP response
        respondError(w, err)
        return
    }

    respondJSON(w, http.StatusOK, users)
}
```

### Router Setup

```go
// internal/http/v1/router.go
package v1

import (
    "github.com/go-chi/chi/v5"

    "{{module}}/internal/services"
)

func NewRouter(svc *services.Registry) *chi.Mux {
    r := chi.NewRouter()

    handler := NewHandler(svc)

    // Use generated HandlerFromMux
    return HandlerFromMux(handler, r)
}
```

## OpenAPI Extensions

### x-go-type-name

Override generated Go type name:

```yaml
UserID:
  type: string
  format: uuid
  x-go-type-name: UserID
```

### x-oapi-codegen-extra-tags

Add struct tags for validation or ORM:

```yaml
User:
  type: object
  properties:
    email:
      type: string
      format: email
      x-oapi-codegen-extra-tags:
        validate: "required,email"
        db: "email"
```

### x-go-type

Map to existing Go type:

```yaml
Timestamp:
  type: string
  format: date-time
  x-go-type: time.Time
```

## Anti-Patterns

### DO NOT

- Manually edit `*.gen.go` files (changes will be overwritten)
- Put business logic in handlers (use services)
- Skip spec validation before generation
- Use `interface{}` in generated types (configure proper types)

### DO

- Keep specs modular with `$ref`
- Use `_index.yaml` aggregators
- Run `openapi-lint` before `openapi-generate`
- Run `golangci-lint` after generation
- Implement handlers as thin wrappers around services

## Dependencies

| Tool | Purpose | Installation |
|------|---------|--------------|
| Redocly CLI | Lint & bundle specs | `npm install -g @redocly/cli@latest` |
| oapi-codegen | Generate Go code | `go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest` |

## Version History

- **v0.48.0**: Initial release with spec-first workflow
