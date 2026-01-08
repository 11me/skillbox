// Package examples demonstrates authentication patterns for Go services.
package examples

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gopkg.in/go-jose/go-jose.v2/jwt"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const (
	userContextKey   contextKey = "user"
	claimsContextKey contextKey = "claims"
)

// AuthMiddleware extracts user from X-User header set by upstream proxy.
// Use this when auth is handled by API Gateway (Kong, Istio, Envoy).
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Header.Get("X-User")
		if user == "" {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserFromContext extracts user string from context.
func UserFromContext(ctx context.Context) (string, bool) {
	user, ok := ctx.Value(userContextKey).(string)
	return user, ok
}

// Claims represents JWT claims with custom fields.
type Claims struct {
	// Standard claims
	Issuer    string   `json:"iss"`
	Subject   string   `json:"sub"`
	Audience  []string `json:"aud"`
	ExpiresAt int64    `json:"exp"`
	IssuedAt  int64    `json:"iat"`
	NotBefore int64    `json:"nbf,omitempty"`
	ID        string   `json:"jti,omitempty"`

	// Custom claims
	UserID string   `json:"user_id,omitempty"`
	Email  string   `json:"email,omitempty"`
	Roles  []string `json:"roles,omitempty"`
}

// JWTValidator validates JWT tokens with support for key rotation.
type JWTValidator struct {
	issuers  map[string]struct{}
	keys     []any
	audience string
}

// NewJWTValidator creates a validator with given issuers and public keys.
func NewJWTValidator(issuers []string, keys []any, audience string) *JWTValidator {
	issuerSet := make(map[string]struct{}, len(issuers))
	for _, iss := range issuers {
		issuerSet[iss] = struct{}{}
	}
	return &JWTValidator{
		issuers:  issuerSet,
		keys:     keys,
		audience: audience,
	}
}

// Validate parses and validates a JWT token string.
func (v *JWTValidator) Validate(tokenString string) (*Claims, error) {
	tok, err := jwt.ParseSigned(tokenString)
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	var claims Claims
	var verified bool
	for _, key := range v.keys {
		if err := tok.Claims(key, &claims); err == nil {
			verified = true
			break
		}
	}
	if !verified {
		return nil, errors.New("invalid signature")
	}

	if _, ok := v.issuers[claims.Issuer]; !ok {
		return nil, fmt.Errorf("invalid issuer: %s", claims.Issuer)
	}

	now := time.Now().Unix()
	if claims.ExpiresAt != 0 && now > claims.ExpiresAt {
		return nil, errors.New("token expired")
	}
	if claims.NotBefore != 0 && now < claims.NotBefore {
		return nil, errors.New("token not yet valid")
	}

	if v.audience != "" && !containsAudience(claims.Audience, v.audience) {
		return nil, errors.New("invalid audience")
	}

	return &claims, nil
}

func containsAudience(audiences []string, target string) bool {
	for _, aud := range audiences {
		if aud == target {
			return true
		}
	}
	return false
}

// JWTMiddleware validates Bearer tokens from Authorization header.
func JWTMiddleware(validator *JWTValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				http.Error(w, `{"error":"missing authorization"}`, http.StatusUnauthorized)
				return
			}

			const prefix = "Bearer "
			if !strings.HasPrefix(auth, prefix) {
				http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
				return
			}
			token := strings.TrimPrefix(auth, prefix)

			claims, err := validator.Validate(token)
			if err != nil {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), claimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClaimsFromContext extracts JWT claims from context.
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*Claims)
	return claims, ok
}

// RequireRoles checks if user has any of the required roles.
func RequireRoles(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			if !hasAnyRole(claims.Roles, roles) {
				http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func hasAnyRole(userRoles, requiredRoles []string) bool {
	roleSet := make(map[string]struct{}, len(userRoles))
	for _, r := range userRoles {
		roleSet[r] = struct{}{}
	}
	for _, r := range requiredRoles {
		if _, ok := roleSet[r]; ok {
			return true
		}
	}
	return false
}

// LoadRSAPublicKey loads an RSA public key from PEM-encoded data.
func LoadRSAPublicKey(pemData []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return rsaPub, nil
}

// LoadECDSAPublicKey loads an ECDSA public key from PEM-encoded data.
func LoadECDSAPublicKey(pemData []byte) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	ecPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("not an ECDSA public key")
	}

	return ecPub, nil
}

// Usage:
//
// Gateway-based auth (behind API Gateway):
//
//	r := chi.NewRouter()
//	r.Use(AuthMiddleware)
//	r.Get("/profile", func(w http.ResponseWriter, r *http.Request) {
//	    user, _ := UserFromContext(r.Context())
//	    // use user
//	})
//
// Full JWT validation:
//
//	key, _ := LoadRSAPublicKey(keyPEM)
//	validator := NewJWTValidator(
//	    []string{"https://auth.example.com"},
//	    []any{key},
//	    "my-api",
//	)
//
//	r := chi.NewRouter()
//	r.Use(JWTMiddleware(validator))
//	r.Get("/profile", func(w http.ResponseWriter, r *http.Request) {
//	    claims, _ := ClaimsFromContext(r.Context())
//	    // use claims.UserID, claims.Email, claims.Roles
//	})
//
// Role-based access control:
//
//	r.Route("/admin", func(r chi.Router) {
//	    r.Use(RequireRoles("admin", "superuser"))
//	    r.Get("/users", listUsers)
//	})
