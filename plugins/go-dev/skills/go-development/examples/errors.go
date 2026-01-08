// Package errs provides sentinel errors with helpers.
package errs

import (
	"errors"
	"fmt"
	"net/http"
)

// Sentinel categories (4-8 typical).
var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrValidation   = errors.New("validation")
	ErrForbidden    = errors.New("forbidden")
	ErrUnauthorized = errors.New("unauthorized")
	ErrTimeout      = errors.New("timeout")
	ErrUnavailable  = errors.New("unavailable")
)

// Wrap adds operation context to error.
func Wrap(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", op, err)
}

// NotFoundf creates not found error with context.
func NotFoundf(op, format string, args ...any) error {
	return fmt.Errorf("%s: %w: %s", op, ErrNotFound, fmt.Sprintf(format, args...))
}

// Conflictf creates conflict error with context.
func Conflictf(op, format string, args ...any) error {
	return fmt.Errorf("%s: %w: %s", op, ErrConflict, fmt.Sprintf(format, args...))
}

// Validationf creates validation error with context.
func Validationf(op, format string, args ...any) error {
	return fmt.Errorf("%s: %w: %s", op, ErrValidation, fmt.Sprintf(format, args...))
}

// Forbiddenf creates forbidden error with context.
func Forbiddenf(op, format string, args ...any) error {
	return fmt.Errorf("%s: %w: %s", op, ErrForbidden, fmt.Sprintf(format, args...))
}

// Unauthorizedf creates unauthorized error with context.
func Unauthorizedf(op, format string, args ...any) error {
	return fmt.Errorf("%s: %w: %s", op, ErrUnauthorized, fmt.Sprintf(format, args...))
}

// HTTPStatus maps error to HTTP status code.
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

// Message returns client-safe error message.
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

// Usage:
//
//	// Repository: create sentinel
//	func (r *UserRepo) FindByID(ctx context.Context, id string) (*User, error) {
//	    row := r.db.QueryRow(ctx, query, id)
//	    var user User
//	    if err := row.Scan(&user.ID, &user.Name); err != nil {
//	        if errors.Is(err, pgx.ErrNoRows) {
//	            return nil, errs.NotFoundf("UserRepo.FindByID", "userID=%s", id)
//	        }
//	        return nil, errs.Wrap("UserRepo.FindByID", err)
//	    }
//	    return &user, nil
//	}
//
//	// Service: just wrap, don't re-classify
//	func (s *UserService) Get(ctx context.Context, id string) (*User, error) {
//	    user, err := s.repo.FindByID(ctx, id)
//	    if err != nil {
//	        return nil, errs.Wrap("UserService.Get", err)
//	    }
//	    return user, nil
//	}
//
//	// Handler: map to HTTP
//	func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
//	    status := errs.HTTPStatus(err)
//	    message := errs.Message(err)
//	    if status == http.StatusInternalServerError {
//	        h.logger.Error("internal error", slog.String("error", err.Error()))
//	    }
//	    w.WriteHeader(status)
//	    json.NewEncoder(w).Encode(map[string]string{"error": message})
//	}
