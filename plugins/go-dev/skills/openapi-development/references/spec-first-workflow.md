# Spec-First API Development Workflow

## Overview

Spec-first (or API-first) development means designing your API contract before writing implementation code. This approach ensures clear contracts, better documentation, and consistent interfaces.

## Benefits

1. **Contract-Driven**: API contract is agreed upon before implementation
2. **Better Documentation**: OpenAPI spec serves as living documentation
3. **Parallel Development**: Frontend/backend teams can work simultaneously
4. **Consistency**: Generated code ensures spec and implementation match
5. **Validation**: Automatic request/response validation from spec

## Workflow Steps

### 1. Design Phase

```
┌─────────────────────────────────────────────────────┐
│  1. Define resources and operations                 │
│  2. Create schemas for request/response bodies      │
│  3. Define error responses                          │
│  4. Review with stakeholders                        │
└─────────────────────────────────────────────────────┘
```

**Start with resources:**
- What entities does your API manage?
- What operations are needed (CRUD, custom actions)?
- What are the relationships between resources?

### 2. Spec Creation

Initialize modular structure:

```bash
/openapi-init v1
```

Creates:
```
api/v1/
├── openapi.yaml           # Main spec
├── paths/
│   └── _index.yaml        # Paths aggregator
├── components/
│   ├── schemas/_index.yaml
│   ├── parameters/_index.yaml
│   ├── responses/_index.yaml
│   └── requests/_index.yaml
└── .redocly.yaml          # Linter config
```

### 3. Add Resources

For each resource:

```bash
/openapi-add-path users --operations list,create,get,update,delete
```

This creates:
- `paths/users.yaml` - CRUD endpoints
- `components/schemas/user.yaml` - Data schemas
- Updates aggregator files

### 4. Validate Spec

```bash
make openapi-lint
```

Fix any issues before proceeding. Common issues:
- Missing operationId
- Missing summary/description
- Unused components
- Invalid $ref paths

### 5. Generate Code

```bash
make openapi-generate
```

Generates `internal/http/v1/api.gen.go`:
- Request/Response types
- ServerInterface with handler methods
- HandlerFromMux for Chi router

### 6. Implement Handlers

Create handler implementation:

```go
// internal/http/v1/handler_impl.go
type Handler struct {
    services *services.Registry
}

// Implement generated interface
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request, params ListUsersParams) {
    // Call service layer
    users, err := h.services.UserService().List(r.Context(), params.Limit, params.Offset)
    if err != nil {
        respondError(w, err)
        return
    }
    respondJSON(w, http.StatusOK, users)
}
```

### 7. Wire Router

```go
// cmd/server/main.go
handler := v1.NewHandler(services)
router := v1.HandlerFromMux(handler, chi.NewRouter())
```

## Keeping Spec and Code in Sync

### Rule: Spec is Source of Truth

```
OpenAPI Spec → Generate → Implement
     ↑                        │
     └────────────────────────┘
         (update spec first)
```

When requirements change:
1. Update OpenAPI spec first
2. Run `make openapi-generate`
3. Update handler implementations

### CI/CD Integration

```yaml
# .gitlab-ci.yml or .github/workflows/ci.yml
lint-openapi:
  script:
    - make openapi-lint

generate-check:
  script:
    - make openapi-generate
    - git diff --exit-code internal/http/v1/api.gen.go
```

The `git diff --exit-code` ensures generated code is committed after spec changes.

## Best Practices

### 1. Use Semantic Versioning for API

```yaml
info:
  version: "1.2.0"  # Major.Minor.Patch
```

- Major: Breaking changes
- Minor: New features (backward compatible)
- Patch: Bug fixes

### 2. Keep Specs Modular

Split large specs:
- One file per resource in `paths/`
- One file per schema in `components/schemas/`
- Use `$ref` for reusability

### 3. Document Everything

```yaml
/users:
  get:
    summary: List users                    # Short description
    description: |                         # Detailed description
      Returns a paginated list of users.
      Supports filtering by status and sorting by creation date.
    operationId: listUsers                 # Unique, URL-safe
```

### 4. Use Consistent Naming

| Element | Convention | Example |
|---------|------------|---------|
| Paths | kebab-case | `/user-accounts` |
| Parameters | camelCase | `userId`, `pageSize` |
| Schemas | PascalCase | `UserAccount`, `CreateUserRequest` |
| operationId | camelCase verb+noun | `listUsers`, `createUser` |

### 5. Define Error Responses

Create reusable error schemas:

```yaml
components:
  responses:
    BadRequest:
      description: Invalid request
      content:
        application/json:
          schema:
            $ref: "#/components/schemas/Error"
    NotFound:
      description: Resource not found
      content:
        application/json:
          schema:
            $ref: "#/components/schemas/Error"
```

## Common Patterns

### Pagination

```yaml
parameters:
  - name: limit
    in: query
    schema:
      type: integer
      default: 20
      maximum: 100
  - name: offset
    in: query
    schema:
      type: integer
      default: 0
```

### Filtering

```yaml
parameters:
  - name: status
    in: query
    schema:
      type: string
      enum: [active, inactive, pending]
  - name: search
    in: query
    schema:
      type: string
```

### Sorting

```yaml
parameters:
  - name: sortBy
    in: query
    schema:
      type: string
      enum: [createdAt, name, email]
  - name: sortOrder
    in: query
    schema:
      type: string
      enum: [asc, desc]
      default: asc
```
