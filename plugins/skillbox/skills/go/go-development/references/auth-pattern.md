# Authentication Patterns

Two authentication approaches for Go services.

## Pattern Selection

| Scenario | Pattern |
|----------|---------|
| Behind API Gateway (Kong, Istio, Envoy) | Gateway-based |
| Kubernetes with service mesh | Gateway-based |
| Internal microservices | Gateway-based |
| Public-facing API | Full JWT |
| No upstream auth proxy | Full JWT |
| Custom claims validation | Full JWT |

## Gateway-Based Auth

Trust upstream proxy/gateway that validates tokens and injects user info into headers.

### How It Works

```
Client → [API Gateway validates JWT] → [Adds X-User header] → Your Service
```

### Middleware

```go
type contextKey string

const userContextKey contextKey = "user"

// AuthMiddleware extracts user from X-User header set by upstream proxy.
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

// UserFromContext extracts user from context.
func UserFromContext(ctx context.Context) (string, bool) {
    user, ok := ctx.Value(userContextKey).(string)
    return user, ok
}
```

### Usage

```go
r := chi.NewRouter()
r.Use(AuthMiddleware)

r.Post("/orders", func(w http.ResponseWriter, r *http.Request) {
    user, ok := UserFromContext(r.Context())
    if !ok {
        // Should not happen if middleware is applied
        http.Error(w, "no user in context", http.StatusInternalServerError)
        return
    }

    // Use user for business logic
    order := createOrder(r.Context(), user, req)
})
```

### Advantages

- Simple implementation
- No crypto dependencies
- Token validation offloaded to gateway
- Easy to test (just set header)

### Considerations

- Requires trusted network (gateway → service)
- Must block direct access to service
- Header can be spoofed if exposed

---

## Full JWT Validation

Validate tokens in-app when no upstream auth proxy exists.

### Dependencies

```bash
go get gopkg.in/go-jose/go-jose.v2@latest
```

### Claims Structure

```go
// Claims represents JWT claims.
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
```

### Validator

```go
// JWTValidator validates JWT tokens.
type JWTValidator struct {
    issuers  map[string]struct{}
    keys     []any // RSA or ECDSA public keys
    audience string
}

// NewJWTValidator creates a validator with given issuers and keys.
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

// Validate parses and validates a JWT token.
func (v *JWTValidator) Validate(tokenString string) (*Claims, error) {
    // 1. Parse JWT
    tok, err := jwt.ParseSigned(tokenString)
    if err != nil {
        return nil, fmt.Errorf("parse token: %w", err)
    }

    // 2. Try each key (supports key rotation)
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

    // 3. Validate issuer
    if _, ok := v.issuers[claims.Issuer]; !ok {
        return nil, fmt.Errorf("invalid issuer: %s", claims.Issuer)
    }

    // 4. Validate expiration
    now := time.Now().Unix()
    if claims.ExpiresAt != 0 && now > claims.ExpiresAt {
        return nil, errors.New("token expired")
    }
    if claims.NotBefore != 0 && now < claims.NotBefore {
        return nil, errors.New("token not yet valid")
    }

    // 5. Validate audience
    if v.audience != "" && !containsAudience(claims.Audience, v.audience) {
        return nil, fmt.Errorf("invalid audience")
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
```

### Middleware

```go
// JWTMiddleware validates Bearer tokens.
func JWTMiddleware(validator *JWTValidator) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract token from Authorization header
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

            // Validate token
            claims, err := validator.Validate(token)
            if err != nil {
                http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
                return
            }

            // Add claims to context
            ctx := context.WithValue(r.Context(), claimsContextKey, claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// ClaimsFromContext extracts claims from context.
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
    claims, ok := ctx.Value(claimsContextKey).(*Claims)
    return claims, ok
}
```

### Key Loading

```go
// LoadRSAPublicKey loads an RSA public key from PEM.
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

// LoadECDSAPublicKey loads an ECDSA public key from PEM.
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
```

### Usage

```go
func main() {
    // Load public keys (supports rotation)
    key1, _ := LoadRSAPublicKey(key1PEM)
    key2, _ := LoadRSAPublicKey(key2PEM)

    validator := NewJWTValidator(
        []string{"https://auth.example.com"},
        []any{key1, key2},
        "my-api",
    )

    r := chi.NewRouter()
    r.Use(JWTMiddleware(validator))

    r.Get("/profile", func(w http.ResponseWriter, r *http.Request) {
        claims, _ := ClaimsFromContext(r.Context())
        // Use claims.UserID, claims.Email, claims.Roles
    })
}
```

---

## Role-Based Access Control

Add RBAC on top of auth middleware:

```go
// RequireRoles checks if user has required roles.
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
```

### Usage

```go
r.Route("/admin", func(r chi.Router) {
    r.Use(RequireRoles("admin", "superuser"))
    r.Get("/users", listUsers)
    r.Delete("/users/{id}", deleteUser)
})
```

---

## Testing

### Gateway-Based Auth

```go
func TestHandler_WithAuth(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/", nil)
    req.Header.Set("X-User", "test-user")

    rec := httptest.NewRecorder()
    handler := AuthMiddleware(myHandler)
    handler.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusOK, rec.Code)
}
```

### Full JWT

```go
func TestHandler_WithJWT(t *testing.T) {
    // Create test token
    token := createTestToken(t, claims)

    req := httptest.NewRequest(http.MethodGet, "/", nil)
    req.Header.Set("Authorization", "Bearer "+token)

    rec := httptest.NewRecorder()
    handler := JWTMiddleware(validator)(myHandler)
    handler.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusOK, rec.Code)
}
```
