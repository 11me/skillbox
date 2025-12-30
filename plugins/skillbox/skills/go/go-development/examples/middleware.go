// Package middleware provides HTTP middleware using stdlib-compatible signatures.
//
// This example shows:
// - Recovery middleware with stack trace logging
// - Request logging with structured output
// - JWT authentication middleware
// - Context key pattern for user injection
package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// ---------- Context Keys ----------

type ctxKey string

const (
	// UserCtxKey is the context key for authenticated user.
	UserCtxKey ctxKey = "user"
	// RequestCtxKey is the context key for request metadata.
	RequestCtxKey ctxKey = "request_context"
)

// User represents an authenticated user.
type User struct {
	ID    string
	Email string
	Role  string
}

// RequestContext contains request metadata.
type RequestContext struct {
	RequestID string
	IP        string
	UserAgent string
}

// ---------- Recovery Middleware ----------

// Recovery recovers from panics and logs the stack trace.
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
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

// ---------- Request Logging Middleware ----------

// RequestLogger logs requests with timing.
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

// ---------- Authentication Middleware ----------

// AuthService defines the interface for authentication.
type AuthService interface {
	ValidateToken(ctx context.Context, token string) (*User, error)
}

// Auth validates JWT tokens and injects user into context.
func Auth(authSvc AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				unauthorized(w)
				return
			}

			user, err := authSvc.ValidateToken(r.Context(), token)
			if err != nil {
				unauthorized(w)
				return
			}

			ctx := context.WithValue(r.Context(), UserCtxKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request) string {
	// Check Authorization header
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// Check cookie
	if cookie, err := r.Cookie("token"); err == nil {
		return cookie.Value
	}

	return ""
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
}

// ---------- Role-Based Access ----------

// RequireRole ensures the user has one of the allowed roles.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	roleSet := make(map[string]bool)
	for _, r := range roles {
		roleSet[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok {
				unauthorized(w)
				return
			}

			if !roleSet[user.Role] {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]string{"error": "forbidden"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ---------- Context Enrichment ----------

// ContextEnrichment adds request metadata to context.
func ContextEnrichment(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCtx := RequestContext{
			RequestID: middleware.GetReqID(r.Context()),
			IP:        r.RemoteAddr,
			UserAgent: r.UserAgent(),
		}

		ctx := context.WithValue(r.Context(), RequestCtxKey, reqCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ---------- Context Helpers ----------

// UserFromContext returns the authenticated user from context.
func UserFromContext(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(UserCtxKey).(*User)
	return user, ok
}

// RequestFromContext returns the request context.
func RequestFromContext(ctx context.Context) (RequestContext, bool) {
	reqCtx, ok := ctx.Value(RequestCtxKey).(RequestContext)
	return reqCtx, ok
}
