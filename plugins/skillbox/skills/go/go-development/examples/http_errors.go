// Package handler provides HTTP error handling utilities.
//
// This example shows:
// - Error to HTTP status mapping using errors package
// - API error response format
// - Request ID for tracing
// - Logging integration with internal error protection
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"project/internal/errors"
)

// ---------- API Error Response ----------

// ErrorResponse is the standard API error response format.
type ErrorResponse struct {
	Error     string            `json:"error"`
	Code      string            `json:"code,omitempty"`
	Details   map[string]string `json:"details,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}

// ---------- Error Handler ----------

// ErrorHandler handles errors and writes appropriate HTTP responses.
type ErrorHandler struct {
	logger *slog.Logger
}

// NewErrorHandler creates a new ErrorHandler.
func NewErrorHandler(logger *slog.Logger) *ErrorHandler {
	return &ErrorHandler{logger: logger}
}

// Handle converts an error to an HTTP response.
func (h *ErrorHandler) Handle(w http.ResponseWriter, r *http.Request, err error) {
	reqID := middleware.GetReqID(r.Context())
	status := errors.HTTPStatusCode(err)
	message := errors.ErrorMessage(err)

	// Log internal errors with full details
	if status == http.StatusInternalServerError {
		h.logger.Error("internal error",
			slog.String("request_id", reqID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("error", err.Error()),
		)
	}

	h.writeError(w, status, message, string(errors.GetErrorCode(err)), reqID)
}

// HandleWithCode handles error with a custom error code.
func (h *ErrorHandler) HandleWithCode(w http.ResponseWriter, r *http.Request, err error, code string) {
	reqID := middleware.GetReqID(r.Context())
	status := errors.HTTPStatusCode(err)
	message := errors.ErrorMessage(err)

	if status == http.StatusInternalServerError {
		h.logger.Error("internal error",
			slog.String("request_id", reqID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("error", err.Error()),
		)
	}

	h.writeError(w, status, message, code, reqID)
}

// ---------- Response Writers ----------

func (h *ErrorHandler) writeError(w http.ResponseWriter, status int, message, code, requestID string) {
	resp := ErrorResponse{
		Error:     message,
		Code:      code,
		RequestID: requestID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// ---------- Convenience Functions ----------

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// WriteError writes an error response.
func WriteError(w http.ResponseWriter, err error, requestID string) {
	status := errors.HTTPStatusCode(err)
	message := errors.ErrorMessage(err)
	code := errors.GetErrorCode(err)

	resp := ErrorResponse{
		Error:     message,
		Code:      string(code),
		RequestID: requestID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// WriteValidationErrors writes validation errors with field details.
func WriteValidationErrors(w http.ResponseWriter, requestID string, details map[string]string) {
	resp := ErrorResponse{
		Error:     "validation failed",
		Code:      string(errors.CodeInvalid),
		Details:   details,
		RequestID: requestID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(resp)
}

// WriteNotFound writes a 404 response.
func WriteNotFound(w http.ResponseWriter, requestID, resource string) {
	resp := ErrorResponse{
		Error:     resource + " not found",
		Code:      string(errors.CodeNotFound),
		RequestID: requestID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(resp)
}

// WriteUnauthorized writes a 401 response.
func WriteUnauthorized(w http.ResponseWriter, requestID string) {
	resp := ErrorResponse{
		Error:     "unauthorized",
		Code:      string(errors.CodeUnauthorized),
		RequestID: requestID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(resp)
}

// WriteInternalError writes a 500 response.
// Note: Always use generic message for security.
func WriteInternalError(w http.ResponseWriter, requestID string) {
	resp := ErrorResponse{
		Error:     "an internal error has occurred",
		Code:      string(errors.CodeInternal),
		RequestID: requestID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(resp)
}

// ---------- Usage Example ----------

// Example usage in handler:
//
//	func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
//	    reqID := middleware.GetReqID(r.Context())
//
//	    var req CreateUserRequest
//	    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//	        WriteError(w, &errors.Error{Code: errors.CodeInvalid, Message: "invalid request body"}, reqID)
//	        return
//	    }
//
//	    user, err := h.services.Users().Create(r.Context(), req)
//	    if err != nil {
//	        // Log internal errors
//	        if errors.IsInternal(err) {
//	            h.logger.Error("create user failed", slog.String("error", err.Error()))
//	        }
//	        WriteError(w, err, reqID)
//	        return
//	    }
//
//	    WriteJSON(w, http.StatusCreated, user)
//	}
