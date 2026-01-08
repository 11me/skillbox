# Middleware Pattern

Production middleware patterns using stdlib-compatible signatures.

## Middleware Signature

```go
// Standard middleware signature (chi, gorilla, stdlib)
func MyMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Before handler
        // ...

        next.ServeHTTP(w, r)

        // After handler
        // ...
    })
}
```

## Common Middleware Stack

Recommended order (top to bottom):

```go
r := chi.NewRouter()

// 1. Request ID (always first)
r.Use(middleware.RequestID)

// 2. Real IP (before logging)
r.Use(middleware.RealIP)

// 3. Structured logging
r.Use(RequestLogger(logger))

// 4. Panic recovery (catch panics from handlers)
r.Use(RecoveryMiddleware(logger))

// 5. CORS (if needed)
r.Use(cors.Handler(corsOptions))

// 6. Timeout
r.Use(middleware.Timeout(60 * time.Second))

// 7. Auth (on protected routes only)
r.Group(func(r chi.Router) {
    r.Use(AuthMiddleware(authService))
    // protected routes...
})
```

## Recovery Middleware

**Critical:** Prevents service crash on panics, logs stack trace.

```go
func RecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if rec := recover(); rec != nil {
                    reqID := middleware.GetReqID(r.Context())

                    logger.Error("panic recovered",
                        slog.String("request_id", reqID),
                        slog.Any("panic", rec),
                        slog.String("stack", string(debug.Stack())),
                    )

                    w.Header().Set("Content-Type", "application/json")
                    w.WriteHeader(http.StatusInternalServerError)
                    json.NewEncoder(w).Encode(map[string]string{
                        "error":      "internal server error",
                        "request_id": reqID,
                    })
                }
            }()

            next.ServeHTTP(w, r)
        })
    }
}
```

## Request Logging Middleware

Structured logging with timing:

```go
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            reqID := middleware.GetReqID(r.Context())

            // Wrap response writer to capture status
            ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}

            next.ServeHTTP(ww, r)

            logger.Info("request",
                slog.String("request_id", reqID),
                slog.String("method", r.Method),
                slog.String("path", r.URL.Path),
                slog.Int("status", ww.status),
                slog.Duration("duration", time.Since(start)),
                slog.String("ip", r.RemoteAddr),
            )
        })
    }
}

type responseWriter struct {
    http.ResponseWriter
    status int
}

func (w *responseWriter) WriteHeader(status int) {
    w.status = status
    w.ResponseWriter.WriteHeader(status)
}
```

## Auth Middleware

JWT-based authentication:

```go
type ctxKey string

const UserCtxKey ctxKey = "user"

func AuthMiddleware(authSvc AuthService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract token
            token := extractToken(r)
            if token == "" {
                http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
                return
            }

            // Validate token
            user, err := authSvc.ValidateToken(r.Context(), token)
            if err != nil {
                http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
                return
            }

            // Add user to context
            ctx := context.WithValue(r.Context(), UserCtxKey, user)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func extractToken(r *http.Request) string {
    // 1. Check Authorization header
    auth := r.Header.Get("Authorization")
    if strings.HasPrefix(auth, "Bearer ") {
        return strings.TrimPrefix(auth, "Bearer ")
    }

    // 2. Check cookie (optional)
    if cookie, err := r.Cookie("token"); err == nil {
        return cookie.Value
    }

    return ""
}

// Helper to get user from context
func UserFromContext(ctx context.Context) (*User, bool) {
    user, ok := ctx.Value(UserCtxKey).(*User)
    return user, ok
}
```

## Role-Based Middleware

```go
func RequireRole(roles ...string) func(http.Handler) http.Handler {
    roleSet := make(map[string]bool)
    for _, r := range roles {
        roleSet[r] = true
    }

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user, ok := UserFromContext(r.Context())
            if !ok {
                http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
                return
            }

            if !roleSet[user.Role] {
                http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// Usage
r.Group(func(r chi.Router) {
    r.Use(AuthMiddleware(authSvc))
    r.Use(RequireRole("admin", "moderator"))
    r.Get("/admin/users", h.ListAllUsers)
})
```

## CORS Middleware

```go
import "github.com/go-chi/cors"

func CORSMiddleware() func(http.Handler) http.Handler {
    return cors.Handler(cors.Options{
        AllowedOrigins:   []string{"https://example.com"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
        ExposedHeaders:   []string{"Link"},
        AllowCredentials: true,
        MaxAge:           300,
    })
}
```

## Rate Limiting Middleware

Simple in-memory rate limiter:

```go
import "golang.org/x/time/rate"

func RateLimitMiddleware(rps int) func(http.Handler) http.Handler {
    limiter := rate.NewLimiter(rate.Limit(rps), rps*2)

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

## Context Enrichment

Add request metadata to context for logging/tracing:

```go
type requestContext struct {
    RequestID string
    UserAgent string
    IP        string
}

const RequestCtxKey ctxKey = "request_context"

func ContextEnrichment(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        reqCtx := requestContext{
            RequestID: middleware.GetReqID(r.Context()),
            UserAgent: r.UserAgent(),
            IP:        r.RemoteAddr,
        }

        ctx := context.WithValue(r.Context(), RequestCtxKey, reqCtx)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## Multi-Auth Pattern

Support multiple auth methods with fallback:

```go
func WithAuthOptions(handlers ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
    return func(final http.Handler) http.Handler {
        // Chain handlers: if one fails, try next
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            for _, handler := range handlers {
                // Try each auth method
                attempted := false
                handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                    attempted = true
                    if _, ok := UserFromContext(r.Context()); ok {
                        final.ServeHTTP(w, r)
                    }
                })).ServeHTTP(w, r)

                if attempted {
                    return
                }
            }

            // All methods failed
            http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
        })
    }
}

// Usage
r.Use(WithAuthOptions(
    JWTAuth(authSvc),
    APIKeyAuth(apiKeySvc),
    SessionAuth(sessionSvc),
))
```

## Best Practices

### DO:
- ✅ Use stdlib `http.Handler` signature
- ✅ Always recover from panics in production
- ✅ Log structured data (request ID, duration, status)
- ✅ Put auth middleware only on protected routes
- ✅ Use context for passing request-scoped data

### DON'T:
- ❌ Put business logic in middleware
- ❌ Modify response after calling `next.ServeHTTP()`
- ❌ Forget to call `next.ServeHTTP()` (breaks chain)
- ❌ Use middleware for single-route logic (use handler instead)

## Dependencies

```bash
go get github.com/go-chi/chi/v5@latest
go get github.com/go-chi/cors@latest         # if CORS needed
go get golang.org/x/time/rate@latest         # if rate limiting needed
```

## Related

- [http-handler-pattern.md](http-handler-pattern.md) — HTTP handlers
- [error-handling.md](error-handling.md) — Error handling
