// Package examples demonstrates JSONB patterns for PostgreSQL.
//
// This file shows how to implement driver.Valuer and sql.Scanner
// interfaces for storing Go types in JSONB columns.
package examples

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/lib/pq"
)

// -----------------------------------------------------------------------------
// Basic JSONB Type
// -----------------------------------------------------------------------------

// Settings represents user preferences stored as JSONB.
type Settings struct {
	Theme       string   `json:"theme,omitempty"`
	Language    string   `json:"language,omitempty"`
	Timezone    string   `json:"timezone,omitempty"`
	Preferences []string `json:"preferences,omitempty"`
}

// Value implements driver.Valuer for INSERT/UPDATE operations.
// Converts Go struct to JSON bytes for PostgreSQL.
func (s Settings) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan implements sql.Scanner for SELECT operations.
// Converts PostgreSQL JSONB bytes to Go struct.
func (s *Settings) Scan(src any) error {
	if src == nil {
		return nil
	}
	data, ok := src.([]byte)
	if !ok {
		return errors.New("expected []byte for JSONB")
	}
	return json.Unmarshal(data, s)
}

// -----------------------------------------------------------------------------
// Filter Type for Queries
// -----------------------------------------------------------------------------

// GameFilter represents query filters stored as JSONB.
// Useful for saving/restoring user search criteria.
type GameFilter struct {
	IDs        []string `json:"ids,omitempty"`
	Status     *string  `json:"status,omitempty"`
	MinPlayers *int     `json:"min_players,omitempty"`
	MaxPlayers *int     `json:"max_players,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	CreatedBy  *string  `json:"created_by,omitempty"`
}

// Value implements driver.Valuer.
func (f GameFilter) Value() (driver.Value, error) {
	return json.Marshal(f)
}

// Scan implements sql.Scanner.
func (f *GameFilter) Scan(src any) error {
	if src == nil {
		return nil
	}
	data, ok := src.([]byte)
	if !ok {
		return errors.New("expected []byte for JSONB")
	}
	return json.Unmarshal(data, f)
}

// -----------------------------------------------------------------------------
// Generic List Type
// -----------------------------------------------------------------------------

// List is a generic slice type for PostgreSQL arrays.
// Works with any string-based type (string, Currency, Status, etc).
//
// Example:
//
//	type User struct {
//	    Roles List[string]  // PostgreSQL: roles TEXT[]
//	    Tags  List[string]  // PostgreSQL: tags TEXT[]
//	}
type List[T ~string] []T

// Value implements driver.Valuer — converts Go slice to PostgreSQL array.
func (l List[T]) Value() (driver.Value, error) {
	if l == nil {
		return nil, nil
	}
	strs := make([]string, len(l))
	for i, v := range l {
		strs[i] = string(v)
	}
	return pq.Array(strs).Value()
}

// Scan implements sql.Scanner — converts PostgreSQL array to Go slice.
func (l *List[T]) Scan(src any) error {
	if src == nil {
		*l = nil
		return nil
	}

	var strs []string
	if err := pq.Array(&strs).Scan(src); err != nil {
		return err
	}

	*l = make([]T, len(strs))
	for i, s := range strs {
		(*l)[i] = T(s)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Metadata Map Type
// -----------------------------------------------------------------------------

// Metadata stores arbitrary key-value pairs as JSONB.
type Metadata map[string]any

// Value implements driver.Valuer.
func (m Metadata) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan implements sql.Scanner.
func (m *Metadata) Scan(src any) error {
	if src == nil {
		*m = nil
		return nil
	}
	data, ok := src.([]byte)
	if !ok {
		return errors.New("expected []byte for JSONB")
	}
	return json.Unmarshal(data, m)
}

// Get returns value by key.
func (m Metadata) Get(key string) (any, bool) {
	if m == nil {
		return nil, false
	}
	v, ok := m[key]
	return v, ok
}

// GetString returns string value by key.
func (m Metadata) GetString(key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

// GetInt returns int value by key.
// Note: JSON numbers are decoded as float64 by default.
func (m Metadata) GetInt(key string) int {
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	default:
		return 0
	}
}

// -----------------------------------------------------------------------------
// Nullable JSONB
// -----------------------------------------------------------------------------

// NullableSettings handles NULL JSONB values.
// Use when the column can be NULL and you need to distinguish
// between NULL and empty object {}.
type NullableSettings struct {
	Settings
	Valid bool
}

// Value implements driver.Valuer.
func (n NullableSettings) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Settings.Value()
}

// Scan implements sql.Scanner.
func (n *NullableSettings) Scan(src any) error {
	if src == nil {
		n.Valid = false
		return nil
	}
	n.Valid = true
	return n.Settings.Scan(src)
}

// -----------------------------------------------------------------------------
// Complete Model Example
// -----------------------------------------------------------------------------

// UserWithJSONB demonstrates a model with various JSONB and array fields.
type UserWithJSONB struct {
	ID        string           `db:"id"`
	Name      string           `db:"name"`
	Settings  Settings         `db:"settings"`  // JSONB
	Metadata  Metadata         `db:"metadata"`  // JSONB (nullable)
	Roles     List[string]     `db:"roles"`     // TEXT[]
	Tags      List[string]     `db:"tags"`      // TEXT[]
	Prefs     NullableSettings `db:"prefs"`     // JSONB (nullable with Valid flag)
}

// -----------------------------------------------------------------------------
// Usage Example (conceptual, requires database connection)
// -----------------------------------------------------------------------------

/*
func ExampleJSONBUsage() {
    // Create user with JSONB fields
    user := &UserWithJSONB{
        ID:   uuid.NewString(),
        Name: "John Doe",
        Settings: Settings{
            Theme:    "dark",
            Language: "en",
            Timezone: "UTC",
        },
        Metadata: Metadata{
            "source":    "signup",
            "campaign":  "summer2024",
            "referrer":  "friend",
        },
        Roles: List[string]{"admin", "user"},
        Tags:  List[string]{"premium", "verified"},
    }

    // Insert — driver.Valuer converts to JSON automatically
    _, err := db.Exec(ctx, `
        INSERT INTO users (id, name, settings, metadata, roles, tags)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, user.ID, user.Name, user.Settings, user.Metadata, user.Roles, user.Tags)

    // Select — sql.Scanner converts from JSON automatically
    row := db.QueryRow(ctx, `SELECT * FROM users WHERE id = $1`, user.ID)
    var loaded UserWithJSONB
    err = row.Scan(
        &loaded.ID,
        &loaded.Name,
        &loaded.Settings,
        &loaded.Metadata,
        &loaded.Roles,
        &loaded.Tags,
    )

    // Query JSONB fields
    rows, err := db.Query(ctx, `
        SELECT * FROM users
        WHERE settings->>'theme' = $1
        AND metadata ? $2
    `, "dark", "campaign")
}
*/
