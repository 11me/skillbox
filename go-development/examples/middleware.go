// Package middleware provides HTTP middleware using stdlib-compatible signatures.
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

type ctxKey string

const (
	UserCtxKey    ctxKey = "user"
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

// RequestLogger logs requests with timing.
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqID := middleware.GetReqID(r.Context())

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
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

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
