---
description: Generate a new service with factory method for Go projects
---

# /go-add-service

Generate a new service with business logic and register it in the Service Registry.

## Usage

```
/go-add-service <service-name>
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `service-name` | Yes | Name in PascalCase (e.g., `User`, `Order`, `Payment`) |

## Prerequisites

- Go project with Service Registry pattern
- `internal/services/registry.go` exists
- Associated repository (optional, will prompt)

## Generated Files

```
internal/services/
├── registry.go       # Updated with factory method
└── <name>.go         # New service file
```

## Steps

1. **Validate service name**:
   - Must be PascalCase: `[A-Z][a-zA-Z0-9]*`
   - Must not already exist in `internal/services/`

2. **Check for associated repository**:
   - Ask: "Does this service need a repository?"
   - If yes, verify repository exists or suggest creating it first

3. **Generate service file** (`internal/services/<name>.go`):

```go
package services

import (
    "context"

    "{{MODULE}}/internal/config"
    "{{MODULE}}/internal/models"
    "{{MODULE}}/internal/storage"
)

type {{Name}}Service struct {
    storage storage.Storage
    conf    *config.Config
}

func New{{Name}}Service(storage storage.Storage, conf *config.Config) *{{Name}}Service {
    return &{{Name}}Service{
        storage: storage,
        conf:    conf,
    }
}

// TODO: Add service methods
// Example:
// func (s *{{Name}}Service) Create(ctx context.Context, req *models.Create{{Name}}Request) (*models.{{Name}}, error) {
//     // Validate
//     // Business logic
//     // Call repository
//     return nil, nil
// }
```

4. **Update Service Registry** (`internal/services/registry.go`):
   - Add factory method:

```go
func (r *Registry) {{Name}}Service() *{{Name}}Service {
    return New{{Name}}Service(r.storage, r.conf)
}
```

5. **Report created files** and suggest next steps.

## Example

```
/go-add-service Order
```

Creates `internal/services/order.go`:
```go
package services

type OrderService struct {
    storage storage.Storage
    conf    *config.Config
}

func NewOrderService(storage storage.Storage, conf *config.Config) *OrderService {
    return &OrderService{
        storage: storage,
        conf:    conf,
    }
}
```

Updates `registry.go`:
```go
func (r *Registry) OrderService() *OrderService {
    return NewOrderService(r.storage, r.conf)
}
```

## Next Steps After Generation

1. Implement service methods with business logic
2. Add validation using typed errors
3. Create tests in `internal/services/<name>_test.go`
4. Wire up in handlers if needed
