// Package handler provides HTTP handlers using chi router.
//
// This example shows:
// - Handler structure with dependency injection
// - Request parsing (path params, query params, JSON body)
// - Response helpers
// - Validation integration
// - Error handling
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"project/internal/common"
	"project/internal/models"
	"project/internal/services"
)

// Handler handles HTTP requests.
type Handler struct {
	services *services.Registry
	validate *validator.Validate
}

// New creates a new Handler.
func New(svc *services.Registry) *Handler {
	v := validator.New(validator.WithRequiredStructEnabled())
	return &Handler{
		services: svc,
		validate: v,
	}
}

// NewRouter creates a new chi router with all routes.
func NewRouter(h *Handler) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health endpoints (no auth)
	r.Get("/health", h.Health)
	r.Get("/ready", h.Ready)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
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

// ---------- Request/Response DTOs ----------

// CreateUserRequest is the request body for creating a user.
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email"`
}

// UpdateUserRequest is the request body for updating a user.
type UpdateUserRequest struct {
	Name  *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Email *string `json:"email,omitempty" validate:"omitempty,email"`
}

// UserResponse is the response body for a user.
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListResponse is a generic paginated response.
type ListResponse[T any] struct {
	Items      []T   `json:"items"`
	TotalCount int64 `json:"total_count"`
	Limit      int   `json:"limit"`
	Offset     int   `json:"offset"`
}

// ---------- Handlers ----------

// Health returns 200 OK if the service is alive.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	h.json(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Ready returns 200 OK if the service is ready to serve traffic.
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	// Check dependencies (DB, cache, etc.)
	if err := h.services.Health().CheckReady(r.Context()); err != nil {
		h.error(w, http.StatusServiceUnavailable, "service not ready")
		return
	}
	h.json(w, http.StatusOK, map[string]string{"status": "ready"})
}

// ListUsers returns a paginated list of users.
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit := getIntQuery(r, "limit", 20)
	offset := getIntQuery(r, "offset", 0)

	users, total, err := h.services.Users().List(r.Context(), limit, offset)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.json(w, http.StatusOK, ListResponse[UserResponse]{
		Items:      mapUsers(users),
		TotalCount: total,
		Limit:      limit,
		Offset:     offset,
	})
}

// GetUser returns a single user by ID.
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		h.error(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := h.services.Users().GetByID(r.Context(), userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.json(w, http.StatusOK, toUserResponse(user))
}

// CreateUser creates a new user.
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.validationError(w, err)
		return
	}

	user, err := h.services.Users().Create(r.Context(), req.Name, req.Email)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.json(w, http.StatusCreated, toUserResponse(user))
}

// UpdateUser updates an existing user.
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		h.error(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.validationError(w, err)
		return
	}

	user, err := h.services.Users().Update(r.Context(), userID, req.Name, req.Email)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.json(w, http.StatusOK, toUserResponse(user))
}

// DeleteUser deletes a user by ID.
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		h.error(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	if err := h.services.Users().Delete(r.Context(), userID); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ---------- Response Helpers ----------

func (h *Handler) json(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (h *Handler) error(w http.ResponseWriter, status int, message string) {
	h.json(w, status, map[string]string{"error": message})
}

func (h *Handler) validationError(w http.ResponseWriter, err error) {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		h.error(w, http.StatusBadRequest, "validation failed")
		return
	}

	details := make(map[string]string)
	for _, e := range validationErrors {
		details[e.Field()] = formatValidationError(e)
	}

	h.json(w, http.StatusBadRequest, map[string]any{
		"error":   "validation failed",
		"details": details,
	})
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch {
	case common.IsNotFound(err):
		h.error(w, http.StatusNotFound, err.Error())
	case common.IsValidationFailed(err):
		h.error(w, http.StatusBadRequest, err.Error())
	case common.IsStateConflict(err):
		h.error(w, http.StatusConflict, err.Error())
	case common.IsUnauthorized(err):
		h.error(w, http.StatusUnauthorized, err.Error())
	default:
		h.error(w, http.StatusInternalServerError, "internal server error")
	}
}

// ---------- Helpers ----------

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

func formatValidationError(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "field is required"
	case "email":
		return "invalid email format"
	case "min":
		return "value too small"
	case "max":
		return "value too large"
	default:
		return "invalid value"
	}
}

func toUserResponse(u *models.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func mapUsers(users []*models.User) []UserResponse {
	result := make([]UserResponse, len(users))
	for i, u := range users {
		result[i] = toUserResponse(u)
	}
	return result
}
