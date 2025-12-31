# Validation Pattern

Input validation using go-playground/validator.

## Core Principle

> **Validate at boundaries, trust internal code.**

- ✅ Validate HTTP request bodies
- ✅ Validate external API responses
- ❌ Don't validate internal function arguments
- ❌ Don't validate between services

## Setup

```go
package handler

import (
    "github.com/go-playground/validator/v10"
)

type Handler struct {
    services *services.Registry
    validate *validator.Validate
}

func New(svc *services.Registry) *Handler {
    v := validator.New(validator.WithRequiredStructEnabled())

    // Register custom validators if needed
    v.RegisterValidation("uuid", validateUUID)

    return &Handler{
        services: svc,
        validate: v,
    }
}
```

## Request DTOs with Validation Tags

```go
// request.go
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"omitempty,min=18,max=120"`
    Role     string `json:"role" validate:"required,oneof=user admin moderator"`
    Password string `json:"password" validate:"required,min=8,max=72"`
}

type UpdateUserRequest struct {
    Name  *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
    Email *string `json:"email,omitempty" validate:"omitempty,email"`
}

type ListUsersRequest struct {
    Limit  int    `validate:"min=1,max=100"`
    Offset int    `validate:"min=0"`
    Status string `validate:"omitempty,oneof=active inactive"`
}
```

## Common Validation Tags

| Tag | Description | Example |
|-----|-------------|---------|
| `required` | Field must be present and non-zero | `validate:"required"` |
| `omitempty` | Skip validation if empty | `validate:"omitempty,email"` |
| `min` | Minimum length/value | `validate:"min=2"` |
| `max` | Maximum length/value | `validate:"max=100"` |
| `len` | Exact length | `validate:"len=10"` |
| `email` | Valid email format | `validate:"email"` |
| `url` | Valid URL | `validate:"url"` |
| `uuid` | Valid UUID (v4) | `validate:"uuid"` |
| `oneof` | One of allowed values | `validate:"oneof=a b c"` |
| `gt` | Greater than | `validate:"gt=0"` |
| `gte` | Greater than or equal | `validate:"gte=18"` |
| `lt` | Less than | `validate:"lt=100"` |
| `lte` | Less than or equal | `validate:"lte=120"` |
| `eqfield` | Equal to another field | `validate:"eqfield=Password"` |

## Validation in Handler

```go
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    // 1. Decode JSON
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.error(w, r, http.StatusBadRequest, "invalid JSON body")
        return
    }

    // 2. Validate
    if err := h.validate.Struct(req); err != nil {
        h.validationError(w, r, err)
        return
    }

    // 3. Call service (already validated, no need to re-validate)
    user, err := h.services.Users().Create(r.Context(), req.Name, req.Email)
    if err != nil {
        h.handleError(w, r, err)
        return
    }

    h.json(w, r, http.StatusCreated, toUserResponse(user))
}
```

## Validation Error Formatting

Convert validator errors to user-friendly messages:

```go
func (h *Handler) validationError(w http.ResponseWriter, r *http.Request, err error) {
    var validationErrors validator.ValidationErrors
    if !errors.As(err, &validationErrors) {
        h.error(w, r, http.StatusBadRequest, "validation failed")
        return
    }

    errors := make(map[string]string)
    for _, e := range validationErrors {
        errors[toSnakeCase(e.Field())] = formatValidationError(e)
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(map[string]any{
        "error":   "validation failed",
        "details": errors,
    })
}

func formatValidationError(e validator.FieldError) string {
    switch e.Tag() {
    case "required":
        return "field is required"
    case "email":
        return "invalid email format"
    case "min":
        if e.Type().Kind() == reflect.String {
            return fmt.Sprintf("must be at least %s characters", e.Param())
        }
        return fmt.Sprintf("must be at least %s", e.Param())
    case "max":
        if e.Type().Kind() == reflect.String {
            return fmt.Sprintf("must be at most %s characters", e.Param())
        }
        return fmt.Sprintf("must be at most %s", e.Param())
    case "oneof":
        return fmt.Sprintf("must be one of: %s", e.Param())
    default:
        return fmt.Sprintf("failed on '%s' validation", e.Tag())
    }
}

func toSnakeCase(s string) string {
    var result []rune
    for i, r := range s {
        if i > 0 && unicode.IsUpper(r) {
            result = append(result, '_')
        }
        result = append(result, unicode.ToLower(r))
    }
    return string(result)
}
```

Example error response:

```json
{
    "error": "validation failed",
    "details": {
        "name": "must be at least 2 characters",
        "email": "invalid email format",
        "role": "must be one of: user admin moderator"
    }
}
```

## Custom Validators

```go
func validateUUID(fl validator.FieldLevel) bool {
    _, err := uuid.Parse(fl.Field().String())
    return err == nil
}

func validatePhoneNumber(fl validator.FieldLevel) bool {
    phone := fl.Field().String()
    // Simple E.164 format check
    matched, _ := regexp.MatchString(`^\+[1-9]\d{10,14}$`, phone)
    return matched
}

// Register
v.RegisterValidation("uuid", validateUUID)
v.RegisterValidation("phone", validatePhoneNumber)

// Usage
type Request struct {
    UserID string `json:"user_id" validate:"required,uuid"`
    Phone  string `json:"phone" validate:"required,phone"`
}
```

## Cross-Field Validation

```go
type RegisterRequest struct {
    Password        string `json:"password" validate:"required,min=8"`
    ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

type DateRangeRequest struct {
    StartDate time.Time `json:"start_date" validate:"required"`
    EndDate   time.Time `json:"end_date" validate:"required,gtfield=StartDate"`
}
```

## Query Parameter Validation

```go
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
    req := ListUsersRequest{
        Limit:  getIntQuery(r, "limit", 20),
        Offset: getIntQuery(r, "offset", 0),
        Status: r.URL.Query().Get("status"),
    }

    if err := h.validate.Struct(req); err != nil {
        h.validationError(w, r, err)
        return
    }

    users, total, err := h.services.Users().List(r.Context(), req.Limit, req.Offset, req.Status)
    if err != nil {
        h.handleError(w, r, err)
        return
    }

    h.json(w, r, http.StatusOK, ListResponse[UserResponse]{
        Items:      mapSlice(users, toUserResponse),
        TotalCount: total,
        Limit:      req.Limit,
        Offset:     req.Offset,
    })
}
```

## Nested Struct Validation

```go
type CreateOrderRequest struct {
    CustomerID string             `json:"customer_id" validate:"required,uuid"`
    Items      []OrderItemRequest `json:"items" validate:"required,min=1,dive"`
}

type OrderItemRequest struct {
    ProductID string `json:"product_id" validate:"required,uuid"`
    Quantity  int    `json:"quantity" validate:"required,min=1,max=100"`
}
```

The `dive` tag validates each element in the slice.

## Optional Fields (Partial Updates)

Use pointers for optional fields in PATCH requests:

```go
type UpdateUserRequest struct {
    Name  *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
    Email *string `json:"email,omitempty" validate:"omitempty,email"`
    Age   *int    `json:"age,omitempty" validate:"omitempty,min=18,max=120"`
}

// In service
func (s *UserService) Update(ctx context.Context, id string, req UpdateUserRequest) (*User, error) {
    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    if req.Name != nil {
        user.Name = *req.Name
    }
    if req.Email != nil {
        user.Email = *req.Email
    }
    if req.Age != nil {
        user.Age = *req.Age
    }

    return s.repo.Update(ctx, user)
}
```

## Best Practices

### DO:
- ✅ Validate all external input (HTTP, gRPC, message queues)
- ✅ Use struct tags for declarative validation
- ✅ Return user-friendly error messages
- ✅ Validate early, fail fast
- ✅ Use `omitempty` for optional fields

### DON'T:
- ❌ Validate internal function arguments
- ❌ Validate between service calls
- ❌ Put validation logic in domain models
- ❌ Return raw validator errors to users

## Dependencies

```bash
go get github.com/go-playground/validator/v10@latest
```

## Related

- [http-handler-pattern.md](http-handler-pattern.md) — HTTP handlers
- [error-handling.md](error-handling.md) — Error handling
