# HTTP Handler Pattern

Production HTTP handler patterns using chi router.

## Why chi?

| Feature | chi | gorilla/mux | gin |
|---------|-----|-------------|-----|
| stdlib compatible | ✅ | ✅ | ❌ |
| Middleware signature | `http.Handler` | `http.Handler` | custom |
| Dependencies | 0 | 0 | many |
| Performance | excellent | good | excellent |
| Learning curve | low | low | medium |

**Recommendation:** chi for new projects — stdlib-compatible, minimal, fast.

## Router Setup

```go
package main

import (
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *Handler) http.Handler {
    r := chi.NewRouter()

    // Global middleware
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.Timeout(60 * time.Second))

    // Health check (no auth)
    r.Get("/health", h.Health)
    r.Get("/ready", h.Ready)

    // API routes
    r.Route("/api/v1", func(r chi.Router) {
        // Users
        r.Route("/users", func(r chi.Router) {
            r.Get("/", h.ListUsers)
            r.Post("/", h.CreateUser)
            r.Route("/{userID}", func(r chi.Router) {
                r.Get("/", h.GetUser)
                r.Put("/", h.UpdateUser)
                r.Delete("/", h.DeleteUser)
            })
        })
    })

    return r
}
```

## Handler Structure

```go
package handler

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"

    "project/internal/services"
)

type Handler struct {
    services *services.Registry
}

func New(svc *services.Registry) *Handler {
    return &Handler{services: svc}
}
```

## Handler Function Signature

Always use stdlib signature for compatibility:

```go
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    // 1. Parse path parameters
    userID, err := uuid.Parse(chi.URLParam(r, "userID"))
    if err != nil {
        h.error(w, r, http.StatusBadRequest, "invalid user ID")
        return
    }

    // 2. Call service
    user, err := h.services.Users().GetByID(r.Context(), userID)
    if err != nil {
        h.handleError(w, r, err)
        return
    }

    // 3. Return response
    h.json(w, r, http.StatusOK, user)
}
```

## Response Helpers

```go
// json writes JSON response
func (h *Handler) json(w http.ResponseWriter, r *http.Request, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if data != nil {
        json.NewEncoder(w).Encode(data)
    }
}

// error writes error response
func (h *Handler) error(w http.ResponseWriter, r *http.Request, status int, message string) {
    h.json(w, r, status, map[string]string{
        "error": message,
    })
}

// noContent writes 204 No Content
func (h *Handler) noContent(w http.ResponseWriter) {
    w.WriteHeader(http.StatusNoContent)
}
```

## Request Parsing

### Path Parameters

```go
// GET /users/{userID}
userID := chi.URLParam(r, "userID")

// Parse UUID
id, err := uuid.Parse(chi.URLParam(r, "userID"))
if err != nil {
    h.error(w, r, http.StatusBadRequest, "invalid user ID")
    return
}
```

### Query Parameters

```go
// GET /users?limit=10&offset=0&status=active
limit := r.URL.Query().Get("limit")
status := r.URL.Query().Get("status")

// With defaults
func getIntQuery(r *http.Request, key string, defaultVal int) int {
    val := r.URL.Query().Get(key)
    if val == "" {
        return defaultVal
    }
    i, err := strconv.Atoi(val)
    if err != nil {
        return defaultVal
    }
    return i
}

limit := getIntQuery(r, "limit", 20)
offset := getIntQuery(r, "offset", 0)
```

### JSON Body

```go
// POST /users
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.error(w, r, http.StatusBadRequest, "invalid JSON")
        return
    }

    // Validate (see validation-pattern.md)
    if err := h.validate.Struct(req); err != nil {
        h.error(w, r, http.StatusBadRequest, formatValidationError(err))
        return
    }

    user, err := h.services.Users().Create(r.Context(), req.Name, req.Email)
    if err != nil {
        h.handleError(w, r, err)
        return
    }

    h.json(w, r, http.StatusCreated, user)
}
```

## Request/Response DTOs

Keep DTOs in handler package, separate from domain models:

```go
// request.go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=2,max=100"`
    Email string `json:"email" validate:"required,email"`
}

type UpdateUserRequest struct {
    Name  *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
    Email *string `json:"email,omitempty" validate:"omitempty,email"`
}

// response.go
type UserResponse struct {
    ID        uuid.UUID `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

type ListResponse[T any] struct {
    Items      []T   `json:"items"`
    TotalCount int64 `json:"total_count"`
    Limit      int   `json:"limit"`
    Offset     int   `json:"offset"`
}

// Mapper
func toUserResponse(u *models.User) UserResponse {
    return UserResponse{
        ID:        u.ID,
        Name:      u.Name,
        Email:     u.Email,
        CreatedAt: u.CreatedAt,
    }
}
```

## Route Groups with Middleware

```go
r.Route("/api/v1", func(r chi.Router) {
    // Public routes
    r.Post("/auth/login", h.Login)
    r.Post("/auth/register", h.Register)

    // Protected routes
    r.Group(func(r chi.Router) {
        r.Use(h.AuthMiddleware)

        r.Route("/users", func(r chi.Router) {
            r.Get("/me", h.GetCurrentUser)
            r.Put("/me", h.UpdateCurrentUser)
        })

        // Admin only
        r.Group(func(r chi.Router) {
            r.Use(h.AdminOnly)
            r.Get("/admin/users", h.ListAllUsers)
        })
    })
})
```

## Best Practices

### DO:
- ✅ Use stdlib `http.Handler` signature
- ✅ Parse and validate input early
- ✅ Return early on errors
- ✅ Use DTOs for request/response (not domain models)
- ✅ Keep handlers thin — delegate to services

### DON'T:
- ❌ Put business logic in handlers
- ❌ Access database directly from handlers
- ❌ Return domain models directly (use DTOs)
- ❌ Ignore request body close (`defer r.Body.Close()` not needed — stdlib does it)

## Dependencies

```bash
go get github.com/go-chi/chi/v5@latest
```

## Related

- [middleware-pattern.md](middleware-pattern.md) — Middleware chain
- [validation-pattern.md](validation-pattern.md) — Input validation
- [error-handling.md](error-handling.md) — Error handling
