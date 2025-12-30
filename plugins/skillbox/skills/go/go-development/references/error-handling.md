# Error Handling Pattern

Simple typed errors with HTTP status mapping.

## Core Principle

**Simple is better.** Use a single `Error` struct with string-based error codes.

- ❌ No chainable errors with `CausedBy()`
- ❌ No stack trace capture
- ❌ No complex error chains
- ✅ Simple struct with Code, Message, Err
- ✅ Standard `fmt.Errorf("%w")` for wrapping
- ✅ Direct HTTP status mapping

## Error Structure

```go
package errors

import (
    "errors"
    "net/http"
)

// ErrorCode classifies errors for HTTP mapping.
type ErrorCode string

const (
    CodeOK           ErrorCode = "ok"
    CodeInvalid      ErrorCode = "invalid"
    CodeNotFound     ErrorCode = "not_found"
    CodeConflict     ErrorCode = "conflict"
    CodeUnauthorized ErrorCode = "unauthorized"
    CodeForbidden    ErrorCode = "forbidden"
    CodeInternal     ErrorCode = "internal"
    CodeUnavailable  ErrorCode = "unavailable"
)

// Error is the application error type.
type Error struct {
    Code    ErrorCode
    Message string
    Err     error // wrapped error (optional)
}

func (e *Error) Error() string {
    if e.Err != nil {
        return e.Message + ": " + e.Err.Error()
    }
    return e.Message
}

func (e *Error) Unwrap() error {
    return e.Err
}
```

## HTTP Status Mapping

```go
// HTTPStatusCode maps error code to HTTP status.
func HTTPStatusCode(err error) int {
    var e *Error
    if !errors.As(err, &e) {
        return http.StatusInternalServerError
    }
    switch e.Code {
    case CodeOK:
        return http.StatusOK
    case CodeInvalid:
        return http.StatusBadRequest
    case CodeNotFound:
        return http.StatusNotFound
    case CodeConflict:
        return http.StatusConflict
    case CodeUnauthorized:
        return http.StatusUnauthorized
    case CodeForbidden:
        return http.StatusForbidden
    case CodeUnavailable:
        return http.StatusServiceUnavailable
    default:
        return http.StatusInternalServerError
    }
}

// ErrorCode extracts error code from error.
func GetErrorCode(err error) ErrorCode {
    var e *Error
    if errors.As(err, &e) {
        return e.Code
    }
    return CodeInternal
}
```

## Client-Safe Messages

Internal errors should not expose details to clients:

```go
// ErrorMessage returns client-safe message.
// Internal errors return generic message for security.
func ErrorMessage(err error) string {
    var e *Error
    if !errors.As(err, &e) || e.Code == CodeInternal {
        return "an internal error has occurred"
    }
    return e.Message
}
```

## Creating Errors

### Pre-defined Package Errors

```go
// Package-level errors for common cases
var (
    ErrUserNotFound  = &Error{Code: CodeNotFound, Message: "user not found"}
    ErrEmailTaken    = &Error{Code: CodeConflict, Message: "email already registered"}
    ErrUnauthorized  = &Error{Code: CodeUnauthorized, Message: "unauthorized"}
    ErrForbidden     = &Error{Code: CodeForbidden, Message: "forbidden"}
)
```

### Inline Creation

```go
// Validation errors with specific messages
return &Error{Code: CodeInvalid, Message: "email format is invalid"}

// Not found with context
return &Error{Code: CodeNotFound, Message: fmt.Sprintf("user %s not found", id)}
```

### Wrapping with Context

```go
// Wrap database error
user, err := s.repo.FindByID(ctx, id)
if err != nil {
    return nil, &Error{
        Code:    CodeInternal,
        Message: "failed to find user",
        Err:     err,
    }
}

// Or use standard wrapping (simpler)
if err != nil {
    return nil, fmt.Errorf("find user: %w", err)
}
```

## Usage in Repository

```go
func (r *userRepo) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
    row := r.db.QueryRow(ctx, query, id)

    var user User
    if err := row.Scan(&user.ID, &user.Name, &user.Email); err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, ErrUserNotFound
        }
        return nil, &Error{Code: CodeInternal, Message: "failed to query user", Err: err}
    }
    return &user, nil
}

func (r *userRepo) Create(ctx context.Context, user *User) error {
    _, err := r.db.Exec(ctx, insertQuery, user.ID, user.Name, user.Email)
    if err != nil {
        // Check for unique constraint violation
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) && pgErr.Code == "23505" {
            return ErrEmailTaken
        }
        return &Error{Code: CodeInternal, Message: "failed to create user", Err: err}
    }
    return nil
}
```

## Usage in Service

```go
func (s *UserService) Create(ctx context.Context, req CreateUserRequest) (*User, error) {
    // Validation
    if req.Email == "" {
        return nil, &Error{Code: CodeInvalid, Message: "email is required"}
    }

    // Check existing
    existing, err := s.repo.FindByEmail(ctx, req.Email)
    if err != nil && GetErrorCode(err) != CodeNotFound {
        return nil, err
    }
    if existing != nil {
        return nil, ErrEmailTaken
    }

    // Create
    user := &User{
        ID:    uuid.New(),
        Name:  req.Name,
        Email: req.Email,
    }

    if err := s.repo.Create(ctx, user); err != nil {
        return nil, err
    }

    return user, nil
}
```

## Usage in HTTP Handler

```go
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.writeError(w, r, &Error{Code: CodeInvalid, Message: "invalid request body"})
        return
    }

    user, err := h.services.Users().Create(r.Context(), req)
    if err != nil {
        h.writeError(w, r, err)
        return
    }

    h.writeJSON(w, http.StatusCreated, user)
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
    status := HTTPStatusCode(err)
    message := ErrorMessage(err)

    // Log internal errors
    if status == http.StatusInternalServerError {
        h.logger.Error("internal error",
            slog.String("error", err.Error()),
            slog.String("path", r.URL.Path),
        )
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]string{
        "error": message,
    })
}
```

## Type Checking

```go
// Using errors.Is for pre-defined errors
if errors.Is(err, ErrUserNotFound) {
    // Handle not found
}

// Using errors.As for any Error
var e *Error
if errors.As(err, &e) {
    switch e.Code {
    case CodeNotFound:
        // Handle not found
    case CodeInvalid:
        // Handle validation
    }
}

// Using helper function
if GetErrorCode(err) == CodeNotFound {
    // Handle not found
}
```

## Best Practices

### DO:
- ✅ Use pre-defined errors for common cases
- ✅ Return client-safe messages for internal errors
- ✅ Log full error details server-side
- ✅ Use `errors.Is/As` for type checking
- ✅ Keep error messages user-friendly

### DON'T:
- ❌ Expose internal error details to clients
- ❌ Create errors without proper codes
- ❌ Use generic "error occurred" messages
- ❌ Forget to log internal errors

## Related

- [http-handler-pattern.md](http-handler-pattern.md) — HTTP handlers
- [logging-pattern.md](logging-pattern.md) — Logging with slog
