# Error Handling Pattern

Simple sentinel errors with wrapping.

## Core Rule

Use **few sentinel categories** (4-8) and wrap errors with context:

```go
// ❌ BAD: Many specific errors
var (
    ErrUserNotFound  = errors.New("user not found")
    ErrOrderNotFound = errors.New("order not found")
    // ... 50+ more
)

// ✅ GOOD: Few categories + wrap
var (
    ErrNotFound     = errors.New("not found")
    ErrConflict     = errors.New("conflict")
    ErrValidation   = errors.New("validation")
    ErrForbidden    = errors.New("forbidden")
    ErrUnauthorized = errors.New("unauthorized")
)
```

## Wrap Format

Pattern: `"<op>: <context>: %w"`, always `%w` at the end:

```go
fmt.Errorf("UserService.Get userID=%s: %w", id, err)
fmt.Errorf("OrderRepo.FindByID orderID=%s: %w", id, ErrNotFound)
```

## Don't Re-Wrap Sentinels

Repository creates sentinel, service wraps original err:

```go
// ❌ BAD: Re-creates sentinel, loses context
user, err := s.repo.FindByID(ctx, id)
if err != nil {
    if errors.Is(err, ErrNotFound) {
        return nil, fmt.Errorf("user %s: %w", id, ErrNotFound)  // WRONG
    }
}

// ✅ GOOD: Wrap original err
user, err := s.repo.FindByID(ctx, id)
if err != nil {
    return nil, fmt.Errorf("UserService.Get userID=%s: %w", id, err)
}
```

Repository guarantees the sentinel, service just adds context.

## Category Set (4-8)

| Category | HTTP | When |
|----------|------|------|
| `ErrNotFound` | 404 | Resource doesn't exist |
| `ErrConflict` | 409 | Duplicate, version conflict |
| `ErrValidation` | 400 | Invalid input |
| `ErrForbidden` | 403 | No permission |
| `ErrUnauthorized` | 401 | Not authenticated |
| `ErrTimeout` | 504 | External dependency timeout |
| `ErrUnavailable` | 503 | Service unavailable |

Don't bloat — add only what you actually map.

## Package: `internal/errs`

```go
package errs

import (
    "errors"
    "fmt"
    "net/http"
)

// Sentinel categories
var (
    ErrNotFound     = errors.New("not found")
    ErrConflict     = errors.New("conflict")
    ErrValidation   = errors.New("validation")
    ErrForbidden    = errors.New("forbidden")
    ErrUnauthorized = errors.New("unauthorized")
    ErrTimeout      = errors.New("timeout")
    ErrUnavailable  = errors.New("unavailable")
)

// Wrap adds operation context
func Wrap(op string, err error) error {
    if err == nil {
        return nil
    }
    return fmt.Errorf("%s: %w", op, err)
}

// NotFoundf creates not found error with context
func NotFoundf(op, format string, args ...any) error {
    return fmt.Errorf("%s: %w: %s", op, ErrNotFound, fmt.Sprintf(format, args...))
}

// Conflictf creates conflict error
func Conflictf(op, format string, args ...any) error {
    return fmt.Errorf("%s: %w: %s", op, ErrConflict, fmt.Sprintf(format, args...))
}

// Validationf creates validation error
func Validationf(op, format string, args ...any) error {
    return fmt.Errorf("%s: %w: %s", op, ErrValidation, fmt.Sprintf(format, args...))
}

// HTTPStatus maps error to HTTP status
func HTTPStatus(err error) int {
    switch {
    case errors.Is(err, ErrNotFound):
        return http.StatusNotFound
    case errors.Is(err, ErrConflict):
        return http.StatusConflict
    case errors.Is(err, ErrValidation):
        return http.StatusBadRequest
    case errors.Is(err, ErrForbidden):
        return http.StatusForbidden
    case errors.Is(err, ErrUnauthorized):
        return http.StatusUnauthorized
    case errors.Is(err, ErrTimeout):
        return http.StatusGatewayTimeout
    case errors.Is(err, ErrUnavailable):
        return http.StatusServiceUnavailable
    default:
        return http.StatusInternalServerError
    }
}

// Message returns client-safe message
func Message(err error) string {
    switch {
    case errors.Is(err, ErrNotFound):
        return "resource not found"
    case errors.Is(err, ErrConflict):
        return "resource conflict"
    case errors.Is(err, ErrValidation):
        return "validation failed"
    case errors.Is(err, ErrForbidden):
        return "forbidden"
    case errors.Is(err, ErrUnauthorized):
        return "unauthorized"
    case errors.Is(err, ErrTimeout):
        return "request timeout"
    case errors.Is(err, ErrUnavailable):
        return "service unavailable"
    default:
        return "internal error"
    }
}
```

## Usage: Repository

```go
func (r *UserRepo) FindByID(ctx context.Context, id string) (*User, error) {
    row := r.db.QueryRow(ctx, query, id)
    var user User
    if err := row.Scan(&user.ID, &user.Name); err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, errs.NotFoundf("UserRepo.FindByID", "userID=%s", id)
        }
        return nil, errs.Wrap("UserRepo.FindByID", err)
    }
    return &user, nil
}

func (r *UserRepo) Create(ctx context.Context, user *User) error {
    _, err := r.db.Exec(ctx, query, user.ID, user.Email)
    if err != nil {
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) && pgErr.Code == "23505" {
            return errs.Conflictf("UserRepo.Create", "email=%s", user.Email)
        }
        return errs.Wrap("UserRepo.Create", err)
    }
    return nil
}
```

## Usage: Service

```go
func (s *UserService) Get(ctx context.Context, id string) (*User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, errs.Wrap("UserService.Get", err)  // Just wrap, don't re-classify
    }
    return user, nil
}

func (s *UserService) Create(ctx context.Context, req CreateUserRequest) (*User, error) {
    if req.Email == "" {
        return nil, errs.Validationf("UserService.Create", "email is required")
    }

    user := &User{ID: uuid.NewString(), Email: req.Email}
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, errs.Wrap("UserService.Create", err)
    }
    return user, nil
}
```

## Usage: HTTP Handler

```go
func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
    status := errs.HTTPStatus(err)
    message := errs.Message(err)

    if status == http.StatusInternalServerError {
        h.logger.Error("internal error", slog.String("error", err.Error()))
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]string{"error": message})
}
```

## Exceptions: Typed Errors

Use typed error **only** when caller needs to extract data:

```go
// Rate limit with retry info
type RateLimitError struct {
    RetryAfter time.Duration
}

func (e RateLimitError) Error() string { return "rate limited" }
func (e RateLimitError) Unwrap() error { return errs.ErrUnavailable }

// Field-level validation
type FieldError struct {
    Field string
    Msg   string
}

func (e FieldError) Error() string { return e.Field + ": " + e.Msg }
func (e FieldError) Unwrap() error { return errs.ErrValidation }

// DB constraint violation
type ConstraintError struct {
    Constraint string
}

func (e ConstraintError) Error() string { return "constraint: " + e.Constraint }
func (e ConstraintError) Unwrap() error { return errs.ErrConflict }
```

## Link to Linting

- `err113` requires `%w` in `fmt.Errorf` — our helpers comply
- Raw `fmt.Errorf("message")` without `%w` will fail lint
- If you must bypass, use `//nolint:err113 // reason` with explanation

See [linting-pattern.md](linting-pattern.md) for nolintlint rules.

## Related

- [http-handler-pattern.md](http-handler-pattern.md) — HTTP handlers
- [logging-pattern.md](logging-pattern.md) — Logging with slog
