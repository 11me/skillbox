# Optional Pattern

Generic helper for converting values to nullable pointers with smart nil handling.

## Problem

Go doesn't have a built-in `Optional<T>` type. Common approaches have issues:

```go
// Problem 1: Empty string becomes pointer to empty string
name := ""
user.LastName = &name  // *string points to "", not nil

// Problem 2: Zero time becomes pointer to zero time
var t time.Time
user.DeletedAt = &t  // *time.Time points to 0001-01-01, not nil

// Problem 3: Verbose nil checks everywhere
if name != "" {
    user.LastName = &name
}
```

## Solution

Generic `optional.Of[T]` function that returns nil for "empty" values:

```go
func Of[T any](val T) *T {
    anyVal := any(val)
    switch anyVal.(type) {
    case string:
        if any(val).(string) == "" {
            return nil
        }
    case time.Time:
        if anyVal.(time.Time).IsZero() {
            return nil
        }
    }
    return &val
}
```

## How It Works

| Input | Output |
|-------|--------|
| `optional.Of("")` | `nil` |
| `optional.Of("John")` | `*string → "John"` |
| `optional.Of(time.Time{})` | `nil` |
| `optional.Of(time.Now())` | `*time.Time → now` |
| `optional.Of(0)` | `*int → 0` |
| `optional.Of(42)` | `*int → 42` |
| `optional.Of(false)` | `*bool → false` |

**Note:** Only empty strings and zero `time.Time` are converted to nil. Other zero values (0, false) become valid pointers.

## Use Cases

### 1. Model Builders

```go
type User struct {
    ID        string
    FirstName string
    LastName  *string    // optional
    Username  *string    // optional
    DeletedAt *time.Time // optional
}

func NewUserFromTelegram(tgUser *TgUser) *User {
    return &User{
        ID:        uuid.NewString(),
        FirstName: tgUser.FirstName,
        LastName:  optional.Of(tgUser.LastName),   // "" → nil
        Username:  optional.Of(tgUser.Username),   // "" → nil
    }
}
```

### 2. Filter Queries

```go
type UserFilter struct {
    IsActive  *bool
    IsBlocked *bool
    Role      *string
}

// Build filter with optional boolean fields
filter := UserFilter{
    IsActive:  optional.Of(true),   // *bool → true
    IsBlocked: optional.Of(false),  // *bool → false
}
```

### 3. API Request Building

```go
type UpdateUserRequest struct {
    Name      *string `json:"name,omitempty"`
    Email     *string `json:"email,omitempty"`
    AvatarURL *string `json:"avatar_url,omitempty"`
}

// Only non-empty fields will be sent
req := UpdateUserRequest{
    Name:      optional.Of(newName),      // "" → nil (omitted from JSON)
    Email:     optional.Of(newEmail),     // "" → nil (omitted from JSON)
    AvatarURL: optional.Of(newAvatarURL), // "" → nil (omitted from JSON)
}
```

### 4. Soft Delete Timestamps

```go
func (u *User) MarkAsDeleted() {
    u.DeletedAt = optional.Of(time.Now().UTC())
}

func (u *User) Restore() {
    u.DeletedAt = nil
}
```

## Typed Helpers (Alternative)

For non-generic codebases or specific types:

```go
func String(s string) *string {
    if s == "" {
        return nil
    }
    return &s
}

func Int(n int) *int {
    return &n
}

func Int64(n int64) *int64 {
    return &n
}

func Float64(n float64) *float64 {
    return &n
}

func Bool(b bool) *bool {
    return &b
}

func Time(t time.Time) *time.Time {
    if t.IsZero() {
        return nil
    }
    return &t
}
```

## Best Practices

| DO | DON'T |
|----|-------|
| Use for external data (APIs, DB) | Overuse for internal structs |
| Use for truly optional fields | Use for required fields |
| Check nil before dereferencing | Assume non-nil |
| Use `omitempty` with pointers | Send nil as explicit `null` |

## File Location

Place in `internal/optional/optional.go`:

```
internal/
└── optional/
    └── optional.go   ← here
```

## Dependencies

None — uses only standard library.

## Related

- [package-naming.md](package-naming.md) — Why not `common/optional.go`
