# oapi-codegen Configuration Reference

## Overview

oapi-codegen generates Go code from OpenAPI 3.x specifications. Configuration can be via CLI flags or YAML config file (recommended).

## Installation

```bash
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
```

## Configuration File

### Basic Structure

```yaml
# oapi-codegen.yaml
package: api                           # Go package name
output: internal/http/v1/api.gen.go    # Output file path

generate:
  # Choose ONE server type:
  chi-server: true      # Chi router (recommended)
  # echo-server: true   # Echo framework
  # gin-server: true    # Gin framework
  # fiber-server: true  # Fiber framework
  # std-http-server: true  # Standard net/http
  # gorilla-server: true   # Gorilla mux

  models: true          # Generate types
  # client: true        # Generate client SDK
  # embedded-spec: true # Embed spec in binary

output-options:
  skip-prune: false     # Remove unused types
```

## Generation Modes

### Server Only (types + server)

```yaml
generate:
  chi-server: true
  models: true
output: internal/http/v1/server.gen.go
```

### Client Only (types + client)

```yaml
generate:
  client: true
  models: true
output: pkg/client/client.gen.go
```

### Types Only

```yaml
generate:
  models: true
output: internal/models/types.gen.go
```

### Full Generation

```yaml
generate:
  chi-server: true
  client: true
  models: true
  embedded-spec: true
output: internal/http/v1/api.gen.go
```

## Router Support

| Router | Config Key | Framework |
|--------|------------|-----------|
| Chi | `chi-server: true` | `github.com/go-chi/chi/v5` |
| Echo | `echo-server: true` | `github.com/labstack/echo/v4` |
| Gin | `gin-server: true` | `github.com/gin-gonic/gin` |
| Fiber | `fiber-server: true` | `github.com/gofiber/fiber/v2` |
| Gorilla | `gorilla-server: true` | `github.com/gorilla/mux` |
| Iris | `iris-server: true` | `github.com/kataras/iris/v12` |
| net/http | `std-http-server: true` | Standard library |

### Chi Server Example

Generated interface:

```go
type ServerInterface interface {
    ListUsers(w http.ResponseWriter, r *http.Request, params ListUsersParams)
    CreateUser(w http.ResponseWriter, r *http.Request)
    GetUser(w http.ResponseWriter, r *http.Request, userID string)
    UpdateUser(w http.ResponseWriter, r *http.Request, userID string)
    DeleteUser(w http.ResponseWriter, r *http.Request, userID string)
}
```

Mounting:

```go
handler := &MyHandler{}
router := chi.NewRouter()
HandlerFromMux(handler, router)
```

## Output Options

```yaml
output-options:
  skip-prune: false              # Remove unused types
  skip-fmt: false                # Skip gofmt
  include-tags: [users, orders]  # Only generate for these tags
  exclude-tags: [internal]       # Skip these tags
  user-templates:                # Custom templates
    type-definitions: ./templates/types.tmpl
```

### Tag Filtering

Generate only endpoints with specific tags:

```yaml
output-options:
  include-tags:
    - users
    - orders
```

Exclude specific tags:

```yaml
output-options:
  exclude-tags:
    - internal
    - deprecated
```

## Import Mapping

For multi-file specs, map external schemas to existing Go packages:

```yaml
import-mapping:
  ./common/schemas.yaml: github.com/yourorg/common/types
  ./external/payment.yaml: github.com/yourorg/payment
```

### Self-Mapping (v2.4.0+)

Split large specs into same package:

```yaml
# types.yaml config
output: internal/api/types.gen.go
generate:
  models: true
import-mapping:
  types.yaml: "-"  # Self-reference

# server.yaml config
output: internal/api/server.gen.go
generate:
  chi-server: true
import-mapping:
  types.yaml: "-"  # Use same package
```

## Additional Initialisms

Define custom initialisms for proper Go naming:

```yaml
additional-initialisms:
  - UUID
  - SSO
  - OAuth
  - JWT
```

Result: `UserUUID`, `SSOToken`, `OAuthConfig`

## OpenAPI Extensions

### x-go-type-name

Override generated type name:

```yaml
UserID:
  type: string
  format: uuid
  x-go-type-name: UserID
```

### x-go-type

Map to existing Go type:

```yaml
Timestamp:
  type: string
  format: date-time
  x-go-type: time.Time

CustomID:
  type: string
  x-go-type: github.com/yourorg/types.CustomID
```

### x-go-type-import

Import with alias:

```yaml
Money:
  type: object
  x-go-type: money.Money
  x-go-type-import:
    path: github.com/shopspring/decimal
    name: money
```

### x-oapi-codegen-extra-tags

Add struct tags:

```yaml
User:
  properties:
    email:
      type: string
      x-oapi-codegen-extra-tags:
        validate: "required,email"
        db: "email"
        json: "email,omitempty"
```

Generated:

```go
type User struct {
    Email string `json:"email,omitempty" validate:"required,email" db:"email"`
}
```

### x-go-json-ignore

Exclude field from JSON:

```yaml
User:
  properties:
    password:
      type: string
      x-go-json-ignore: true
```

Generated:

```go
type User struct {
    Password string `json:"-"`
}
```

## Optional Pointer Control (v2.5.0+)

```yaml
output-options:
  # Don't generate pointers for optional fields
  prefer-skip-optional-pointer: true

  # Use omitzero tag (Go 1.24+)
  prefer-skip-optional-pointer-with-omitzero: true
```

## CLI Usage

```bash
# Using config file
oapi-codegen --config oapi-codegen.yaml api/v1/bundle.yaml

# CLI flags (override config)
oapi-codegen \
  -package api \
  -generate types,chi-server \
  -o internal/http/v1/api.gen.go \
  api/v1/bundle.yaml
```

## Common Configurations

### Microservice API

```yaml
package: api
output: internal/http/v1/api.gen.go
generate:
  chi-server: true
  models: true
output-options:
  skip-prune: true
```

### SDK Package

```yaml
package: client
output: pkg/sdk/client.gen.go
generate:
  client: true
  models: true
```

### Shared Types Library

```yaml
package: types
output: pkg/types/types.gen.go
generate:
  models: true
output-options:
  skip-prune: false
```
