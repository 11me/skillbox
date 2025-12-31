package common

import "time"

// ---------- Generic Optional ----------

// Optional converts a value to a pointer, returning nil for "empty" values.
// Empty values: empty string "", zero time.Time.
// All other zero values (0, false) become valid pointers.
func Optional[T any](val T) *T {
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

// ---------- Typed Helpers (for non-generic codebases) ----------

// OptionalString returns nil for empty string, pointer otherwise.
func OptionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// OptionalInt returns pointer to int.
func OptionalInt(n int) *int {
	return &n
}

// OptionalInt64 returns pointer to int64.
func OptionalInt64(n int64) *int64 {
	return &n
}

// OptionalFloat64 returns pointer to float64.
func OptionalFloat64(n float64) *float64 {
	return &n
}

// OptionalBool returns pointer to bool.
func OptionalBool(b bool) *bool {
	return &b
}

// OptionalTime returns nil for zero time, pointer otherwise.
func OptionalTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// ---------- Usage Examples ----------

// Example: Building user from external API response
type TgUser struct {
	ID        int64
	FirstName string
	LastName  string // may be empty
	Username  string // may be empty
}

type User struct {
	ID        string
	FirstName string
	LastName  *string    // optional
	Username  *string    // optional
	DeletedAt *time.Time // optional
}

func NewUserFromTelegram(tgUser *TgUser) *User {
	return &User{
		ID:        "user-123",
		FirstName: tgUser.FirstName,
		LastName:  Optional(tgUser.LastName),  // "" → nil
		Username:  Optional(tgUser.Username),  // "" → nil
		DeletedAt: nil,
	}
}

// Example: Filter with optional boolean fields
type UserFilter struct {
	IsActive  *bool
	IsBlocked *bool
	Role      *string
}

func ActiveUsersFilter() UserFilter {
	return UserFilter{
		IsActive:  Optional(true),
		IsBlocked: Optional(false),
	}
}

// Example: Soft delete
func (u *User) MarkAsDeleted() {
	u.DeletedAt = Optional(time.Now().UTC())
}

func (u *User) Restore() {
	u.DeletedAt = nil
}

func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}
