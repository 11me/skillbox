# HTTP Handler Pattern

Production HTTP handler patterns using chi router with handler-per-entity organization.

## Why chi?

| Feature | chi | gorilla/mux | gin |
|---------|-----|-------------|-----|
| stdlib compatible | Yes | Yes | No |
| Middleware signature | `http.Handler` | `http.Handler` | custom |
| Dependencies | 0 | 0 | many |
| Performance | excellent | good | excellent |
| Learning curve | low | low | medium |

**Recommendation:** chi for new projects — stdlib-compatible, minimal, fast.

## File Organization

**Handler-per-entity** — each entity gets its own handler file:

```
internal/http/v1/
├── router.go              # Router setup + path constants
├── user_handler.go        # User handlers
├── order_handler.go       # Order handlers
├── dto.go                 # Request/Response types
└── helpers.go             # decode*, encode* helpers
```

## Router Setup (router.go)

```go
package v1

import (
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

// Path constants — single source of truth for URLs.
const (
    PathPrefix = "/api/v1"

    // Users
    UsersPath    = "/users"
    UserByIDPath = "/users/{userID}"

    // Orders
    OrdersPath    = "/orders"
    OrderByIDPath = "/orders/{orderID}"
)

// NewRouter creates the HTTP router with all handlers.
func NewRouter(
    userHandler *UserHandler,
    orderHandler *OrderHandler,
) http.Handler {
    r := chi.NewRouter()

    // Global middleware
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.Timeout(60 * time.Second))

    // Health (no auth)
    r.Get("/health", healthHandler)
    r.Get("/ready", readyHandler)

    // API v1
    r.Route(PathPrefix, func(r chi.Router) {
        // Users
        r.Post(UsersPath, userHandler.Create)
        r.Get(UsersPath, userHandler.List)
        r.Get(UserByIDPath, userHandler.GetByID)
        r.Put(UserByIDPath, userHandler.Update)
        r.Delete(UserByIDPath, userHandler.Delete)

        // Orders
        r.Post(OrdersPath, orderHandler.Create)
        r.Get(OrdersPath, orderHandler.List)
        r.Get(OrderByIDPath, orderHandler.GetByID)
    })

    return r
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ok"))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ready"))
}
```

## Handler Structure (user_handler.go)

Each handler receives only its required dependencies:

```go
package v1

import (
    "context"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/go-playground/validator/v10"
)

// UserService defines required service methods.
type UserService interface {
    Create(ctx context.Context, name, email string) (*User, error)
    GetByID(ctx context.Context, id string) (*User, error)
    List(ctx context.Context, limit, offset int) ([]*User, int64, error)
    Update(ctx context.Context, id, name, email string) (*User, error)
    Delete(ctx context.Context, id string) error
}

// UserHandler handles user HTTP endpoints.
type UserHandler struct {
    userService UserService
    validate    *validator.Validate
}

// NewUserHandler creates a new user handler.
func NewUserHandler(svc UserService) *UserHandler {
    return &UserHandler{
        userService: svc,
        validate:    validator.New(),
    }
}

// Create handles POST /users.
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    req, err := decodeCreateUserRequest(r, h.validate)
    if err != nil {
        encodeErrorResponse(w, err)
        return
    }

    user, err := h.userService.Create(ctx, req.Name, req.Email)
    if err != nil {
        encodeErrorResponse(w, err)
        return
    }

    encodeJSONResponse(w, http.StatusCreated, toUserResponse(user))
}

// GetByID handles GET /users/{userID}.
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    userID := chi.URLParam(r, "userID")
    if userID == "" {
        encodeErrorResponse(w, NewBadRequestError("user ID is required"))
        return
    }

    user, err := h.userService.GetByID(ctx, userID)
    if err != nil {
        encodeErrorResponse(w, err)
        return
    }

    encodeJSONResponse(w, http.StatusOK, toUserResponse(user))
}

// List handles GET /users.
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    limit := getIntQuery(r, "limit", 20)
    offset := getIntQuery(r, "offset", 0)

    users, total, err := h.userService.List(ctx, limit, offset)
    if err != nil {
        encodeErrorResponse(w, err)
        return
    }

    encodeJSONResponse(w, http.StatusOK, ListResponse[UserResponse]{
        Items:      toUserResponses(users),
        TotalCount: total,
        Limit:      limit,
        Offset:     offset,
    })
}

// Update handles PUT /users/{userID}.
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    userID := chi.URLParam(r, "userID")
    if userID == "" {
        encodeErrorResponse(w, NewBadRequestError("user ID is required"))
        return
    }

    req, err := decodeUpdateUserRequest(r, h.validate)
    if err != nil {
        encodeErrorResponse(w, err)
        return
    }

    user, err := h.userService.Update(ctx, userID, deref(req.Name), deref(req.Email))
    if err != nil {
        encodeErrorResponse(w, err)
        return
    }

    encodeJSONResponse(w, http.StatusOK, toUserResponse(user))
}

// Delete handles DELETE /users/{userID}.
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    userID := chi.URLParam(r, "userID")
    if userID == "" {
        encodeErrorResponse(w, NewBadRequestError("user ID is required"))
        return
    }

    if err := h.userService.Delete(ctx, userID); err != nil {
        encodeErrorResponse(w, err)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}
```

## DTOs (dto.go)

Request and response types in one place:

```go
package v1

import "time"

// -----------------------------------------------------------------------------
// User DTOs
// -----------------------------------------------------------------------------

type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=2,max=100"`
    Email string `json:"email" validate:"required,email"`
}

type UpdateUserRequest struct {
    Name  *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
    Email *string `json:"email,omitempty" validate:"omitempty,email"`
}

type UserResponse struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

// toUserResponse converts domain model to response DTO.
func toUserResponse(u *User) UserResponse {
    return UserResponse{
        ID:        u.ID,
        Name:      u.Name,
        Email:     u.Email,
        CreatedAt: u.CreatedAt,
    }
}

// toUserResponses converts slice of domain models.
func toUserResponses(users []*User) []UserResponse {
    result := make([]UserResponse, len(users))
    for i, u := range users {
        result[i] = toUserResponse(u)
    }
    return result
}

// -----------------------------------------------------------------------------
// Generic DTOs
// -----------------------------------------------------------------------------

type ListResponse[T any] struct {
    Items      []T   `json:"items"`
    TotalCount int64 `json:"total_count"`
    Limit      int   `json:"limit"`
    Offset     int   `json:"offset"`
}

type ErrorResponse struct {
    Error   string `json:"error"`
    Code    string `json:"code,omitempty"`
    Details any    `json:"details,omitempty"`
}
```

## Helpers (helpers.go)

Decode and encode functions:

```go
package v1

import (
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/go-playground/validator/v10"
)

// -----------------------------------------------------------------------------
// Decode Functions
// -----------------------------------------------------------------------------

func decodeCreateUserRequest(r *http.Request, v *validator.Validate) (*CreateUserRequest, error) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        return nil, NewBadRequestError("invalid JSON")
    }
    if err := v.StructCtx(r.Context(), &req); err != nil {
        return nil, NewValidationError(err)
    }
    return &req, nil
}

func decodeUpdateUserRequest(r *http.Request, v *validator.Validate) (*UpdateUserRequest, error) {
    var req UpdateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        return nil, NewBadRequestError("invalid JSON")
    }
    if err := v.StructCtx(r.Context(), &req); err != nil {
        return nil, NewValidationError(err)
    }
    return &req, nil
}

// -----------------------------------------------------------------------------
// Encode Functions
// -----------------------------------------------------------------------------

func encodeJSONResponse(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if data != nil {
        json.NewEncoder(w).Encode(data)
    }
}

func encodeErrorResponse(w http.ResponseWriter, err error) {
    status := HTTPStatusCode(err)
    message := ErrorMessage(err)
    code := GetErrorCode(err)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(ErrorResponse{
        Error: message,
        Code:  string(code),
    })
}

// -----------------------------------------------------------------------------
// Query Helpers
// -----------------------------------------------------------------------------

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

func getStringQuery(r *http.Request, key string) string {
    return r.URL.Query().Get(key)
}

func getBoolQuery(r *http.Request, key string, defaultVal bool) bool {
    val := r.URL.Query().Get(key)
    if val == "" {
        return defaultVal
    }
    b, err := strconv.ParseBool(val)
    if err != nil {
        return defaultVal
    }
    return b
}

// -----------------------------------------------------------------------------
// Utility
// -----------------------------------------------------------------------------

// deref returns the value of a pointer or empty string if nil.
func deref(s *string) string {
    if s == nil {
        return ""
    }
    return *s
}
```

## Error Helpers

Add to helpers.go or separate errors.go:

```go
// HandlerError represents HTTP errors.
type HandlerError struct {
    Status  int
    Code    string
    Message string
}

func (e *HandlerError) Error() string {
    return e.Message
}

func NewBadRequestError(msg string) error {
    return &HandlerError{
        Status:  http.StatusBadRequest,
        Code:    "bad_request",
        Message: msg,
    }
}

func NewNotFoundError(msg string) error {
    return &HandlerError{
        Status:  http.StatusNotFound,
        Code:    "not_found",
        Message: msg,
    }
}

func NewValidationError(err error) error {
    return &HandlerError{
        Status:  http.StatusBadRequest,
        Code:    "validation_error",
        Message: formatValidationError(err),
    }
}

// HTTPStatusCode extracts HTTP status from error.
func HTTPStatusCode(err error) int {
    if he, ok := err.(*HandlerError); ok {
        return he.Status
    }
    // Check for domain errors (see error-handling.md)
    var domainErr *Error
    if errors.As(err, &domainErr) {
        return domainErr.HTTPStatus()
    }
    return http.StatusInternalServerError
}

// ErrorMessage returns client-safe error message.
func ErrorMessage(err error) string {
    if he, ok := err.(*HandlerError); ok {
        return he.Message
    }
    var domainErr *Error
    if errors.As(err, &domainErr) {
        return domainErr.Message
    }
    return "internal error"
}

// GetErrorCode returns error code string.
func GetErrorCode(err error) string {
    if he, ok := err.(*HandlerError); ok {
        return he.Code
    }
    return "internal"
}
```

## Adding Protected Routes

```go
func NewRouter(
    userHandler *UserHandler,
    orderHandler *OrderHandler,
    authMiddleware func(http.Handler) http.Handler,
) http.Handler {
    r := chi.NewRouter()

    // ... global middleware

    r.Route(PathPrefix, func(r chi.Router) {
        // Public routes
        r.Post("/auth/login", authHandler.Login)
        r.Post("/auth/register", authHandler.Register)

        // Protected routes
        r.Group(func(r chi.Router) {
            r.Use(authMiddleware)

            r.Post(UsersPath, userHandler.Create)
            r.Get(UsersPath, userHandler.List)
            r.Get(UserByIDPath, userHandler.GetByID)
            r.Put(UserByIDPath, userHandler.Update)
            r.Delete(UserByIDPath, userHandler.Delete)
        })
    })

    return r
}
```

## Best Practices

### DO:
- One handler struct per entity
- Separate decode functions for each request type
- Keep DTOs separate from domain models
- Use path constants in router.go
- Use stdlib `http.Handler` signature

### DON'T:
- Don't put all handlers in one struct
- Don't put business logic in handlers
- Don't access database directly
- Don't return domain models (use DTOs)

## Dependencies

```bash
go get github.com/go-chi/chi/v5@latest
go get github.com/go-playground/validator/v10@latest
```

## Related

- [middleware-pattern.md](middleware-pattern.md) — Middleware chain
- [validation-pattern.md](validation-pattern.md) — Input validation
- [error-handling.md](error-handling.md) — Error handling
