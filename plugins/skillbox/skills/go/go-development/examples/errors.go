// Package errors provides typed errors with HTTP status mapping.
//
// This example shows:
// - Simple Error struct with Code, Message, Err
// - ErrorCode string constants for classification
// - HTTP status code mapping
// - Client-safe error messages
package errors

import (
	"errors"
	"net/http"
)

// ---------- Error Codes ----------

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

// ---------- Error Type ----------

// Error is the application error type.
type Error struct {
	Code    ErrorCode
	Message string
	Err     error // wrapped error (optional)
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the wrapped error for errors.Is/errors.As.
func (e *Error) Unwrap() error {
	return e.Err
}

// ---------- Pre-defined Errors ----------

// Common errors for reuse across packages.
var (
	ErrNotFound     = &Error{Code: CodeNotFound, Message: "resource not found"}
	ErrUnauthorized = &Error{Code: CodeUnauthorized, Message: "unauthorized"}
	ErrForbidden    = &Error{Code: CodeForbidden, Message: "forbidden"}
	ErrUnavailable  = &Error{Code: CodeUnavailable, Message: "service unavailable"}
)

// ---------- HTTP Status Mapping ----------

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

// ---------- Error Code Extraction ----------

// GetErrorCode extracts the error code from an error.
// Returns CodeInternal if the error is not an *Error.
func GetErrorCode(err error) ErrorCode {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return CodeInternal
}

// ---------- Client-Safe Messages ----------

// ErrorMessage returns client-safe message.
// Internal errors return a generic message for security.
func ErrorMessage(err error) string {
	var e *Error
	if !errors.As(err, &e) || e.Code == CodeInternal {
		return "an internal error has occurred"
	}
	return e.Message
}

// ---------- Type Checking Helpers ----------

// IsNotFound checks if the error is a not found error.
func IsNotFound(err error) bool {
	return GetErrorCode(err) == CodeNotFound
}

// IsInvalid checks if the error is a validation error.
func IsInvalid(err error) bool {
	return GetErrorCode(err) == CodeInvalid
}

// IsConflict checks if the error is a conflict error.
func IsConflict(err error) bool {
	return GetErrorCode(err) == CodeConflict
}

// IsUnauthorized checks if the error is an unauthorized error.
func IsUnauthorized(err error) bool {
	return GetErrorCode(err) == CodeUnauthorized
}

// IsForbidden checks if the error is a forbidden error.
func IsForbidden(err error) bool {
	return GetErrorCode(err) == CodeForbidden
}

// IsInternal checks if the error is an internal error.
func IsInternal(err error) bool {
	return GetErrorCode(err) == CodeInternal
}

// ---------- Usage Examples ----------

// Example usage:
//
//	// Pre-defined errors
//	var ErrUserNotFound = &errors.Error{Code: errors.CodeNotFound, Message: "user not found"}
//	var ErrEmailTaken = &errors.Error{Code: errors.CodeConflict, Message: "email already registered"}
//
//	// In repository
//	func (r *userRepo) FindByID(ctx context.Context, id string) (*User, error) {
//	    row := r.db.QueryRow(ctx, query, id)
//	    var user User
//	    if err := row.Scan(&user.ID, &user.Name); err != nil {
//	        if errors.Is(err, pgx.ErrNoRows) {
//	            return nil, ErrUserNotFound
//	        }
//	        return nil, &errors.Error{Code: errors.CodeInternal, Message: "failed to query user", Err: err}
//	    }
//	    return &user, nil
//	}
//
//	// In service
//	func (s *UserService) Create(ctx context.Context, email string) (*User, error) {
//	    if email == "" {
//	        return nil, &errors.Error{Code: errors.CodeInvalid, Message: "email is required"}
//	    }
//	    // ...
//	}
//
//	// In HTTP handler
//	func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
//	    user, err := h.services.Users().GetByID(ctx, id)
//	    if err != nil {
//	        status := errors.HTTPStatusCode(err)
//	        message := errors.ErrorMessage(err)
//	        // Log internal errors
//	        if status == http.StatusInternalServerError {
//	            h.logger.Error("internal error", slog.String("error", err.Error()))
//	        }
//	        w.WriteHeader(status)
//	        json.NewEncoder(w).Encode(map[string]string{"error": message})
//	        return
//	    }
//	    // ...
//	}
