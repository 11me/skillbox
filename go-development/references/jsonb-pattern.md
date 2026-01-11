# JSONB Pattern

Storing complex Go types in PostgreSQL JSONB columns.

## Core Interfaces

To store a Go type in JSONB, implement two interfaces:

```go
// driver.Valuer — Go → Database (INSERT/UPDATE)
type Valuer interface {
    Value() (driver.Value, error)
}

// sql.Scanner — Database → Go (SELECT)
type Scanner interface {
    Scan(src any) error
}
```

## Basic JSONB Type

```go
package models

import (
    "database/sql/driver"
    "encoding/json"
    "errors"
)

// Settings stored as JSONB in PostgreSQL.
type Settings struct {
    Theme       string   `json:"theme,omitempty"`
    Language    string   `json:"language,omitempty"`
    Timezone    string   `json:"timezone,omitempty"`
    Preferences []string `json:"preferences,omitempty"`
}

// Value implements driver.Valuer for INSERT/UPDATE.
func (s Settings) Value() (driver.Value, error) {
    return json.Marshal(s)
}

// Scan implements sql.Scanner for SELECT.
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
```

## Filter Type for Queries

Common pattern for API filters stored in JSONB:

```go
// GameFilter represents query parameters stored as JSONB.
type GameFilter struct {
    IDs        []string `json:"ids,omitempty"`
    Status     *string  `json:"status,omitempty"`
    MinPlayers *int     `json:"min_players,omitempty"`
    MaxPlayers *int     `json:"max_players,omitempty"`
    Tags       []string `json:"tags,omitempty"`
    CreatedBy  *string  `json:"created_by,omitempty"`
}

func (f GameFilter) Value() (driver.Value, error) {
    return json.Marshal(f)
}

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
```

## Generic List Type

For PostgreSQL arrays that need custom serialization:

```go
// List is a generic slice type for PostgreSQL arrays.
// Works with any string-based type (string, Currency, Status, etc).
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
```

Usage:

```go
type User struct {
    ID    string
    Name  string
    Roles List[string]  // PostgreSQL: roles TEXT[]
    Tags  List[string]  // PostgreSQL: tags TEXT[]
}
```

## Metadata Map Type

For arbitrary key-value metadata:

```go
// Metadata stores arbitrary key-value pairs as JSONB.
type Metadata map[string]any

func (m Metadata) Value() (driver.Value, error) {
    if m == nil {
        return nil, nil
    }
    return json.Marshal(m)
}

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

// Get returns value by key with type assertion.
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
```

## Nullable JSONB

For optional JSONB columns:

```go
// NullableSettings handles NULL JSONB values.
type NullableSettings struct {
    Settings
    Valid bool
}

func (n NullableSettings) Value() (driver.Value, error) {
    if !n.Valid {
        return nil, nil
    }
    return n.Settings.Value()
}

func (n *NullableSettings) Scan(src any) error {
    if src == nil {
        n.Valid = false
        return nil
    }
    n.Valid = true
    return n.Settings.Scan(src)
}
```

## Database Schema

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    settings JSONB DEFAULT '{}',
    metadata JSONB,
    roles TEXT[],
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for JSONB queries
CREATE INDEX idx_users_settings_theme ON users ((settings->>'theme'));
CREATE INDEX idx_users_metadata ON users USING GIN (metadata);
```

## Repository Usage

```go
func (r *userRepo) Create(ctx context.Context, user *User) error {
    query := `
        INSERT INTO users (id, name, settings, metadata, roles)
        VALUES ($1, $2, $3, $4, $5)
    `
    _, err := r.db.Exec(ctx, query,
        user.ID,
        user.Name,
        user.Settings,   // driver.Valuer handles serialization
        user.Metadata,   // driver.Valuer handles serialization
        user.Roles,      // List[T] handles array conversion
    )
    return err
}

func (r *userRepo) FindByID(ctx context.Context, id string) (*User, error) {
    query := `
        SELECT id, name, settings, metadata, roles, created_at
        FROM users WHERE id = $1
    `
    row := r.db.QueryRow(ctx, query, id)

    var user User
    err := row.Scan(
        &user.ID,
        &user.Name,
        &user.Settings,   // sql.Scanner handles deserialization
        &user.Metadata,   // sql.Scanner handles deserialization
        &user.Roles,      // List[T] handles array conversion
        &user.CreatedAt,
    )
    if err != nil {
        return nil, err
    }
    return &user, nil
}
```

## Querying JSONB

```go
// Find users by settings theme
func (r *userRepo) FindByTheme(ctx context.Context, theme string) ([]*User, error) {
    query := `
        SELECT id, name, settings, metadata, roles, created_at
        FROM users
        WHERE settings->>'theme' = $1
    `
    rows, err := r.db.Query(ctx, query, theme)
    if err != nil {
        return nil, err
    }
    return pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[User])
}

// Find users with specific metadata key
func (r *userRepo) FindByMetadataKey(ctx context.Context, key string) ([]*User, error) {
    query := `
        SELECT id, name, settings, metadata, roles, created_at
        FROM users
        WHERE metadata ? $1
    `
    rows, err := r.db.Query(ctx, query, key)
    if err != nil {
        return nil, err
    }
    return pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[User])
}
```

## With squirrel

```go
func (r *userRepo) Find(ctx context.Context, filter *UserFilter) ([]*User, error) {
    qb := sq.Select("id", "name", "settings", "metadata", "roles", "created_at").
        From("users").
        PlaceholderFormat(sq.Dollar)

    if filter.Theme != nil {
        qb = qb.Where("settings->>'theme' = ?", *filter.Theme)
    }
    if filter.HasMetadataKey != nil {
        qb = qb.Where("metadata ? ?", *filter.HasMetadataKey)
    }
    if len(filter.Roles) > 0 {
        qb = qb.Where("roles && ?", pq.Array(filter.Roles))  // array overlap
    }

    sql, args, err := qb.ToSql()
    if err != nil {
        return nil, err
    }

    rows, err := r.db.Query(ctx, sql, args...)
    if err != nil {
        return nil, err
    }
    return pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[User])
}
```

## Best Practices

### DO:
- Use `json:"-"` for fields that shouldn't be serialized
- Use `omitempty` for optional fields to keep JSON compact
- Add GIN indexes for JSONB columns used in queries
- Use specific JSONB operators (`->`, `->>`, `?`, `@>`) for efficient queries
- Handle `nil` in both Value() and Scan() methods

### DON'T:
- Don't store large documents (>1MB) in JSONB — use separate tables
- Don't use JSONB for data you need to query frequently — normalize it
- Don't forget NULL handling in Scan()
- Don't assume JSONB preserves key order (it doesn't)

## JSONB Operators Reference

| Operator | Description | Example |
|----------|-------------|---------|
| `->` | Get JSON element | `settings->'theme'` returns JSON |
| `->>` | Get JSON element as text | `settings->>'theme'` returns TEXT |
| `?` | Key exists | `metadata ? 'key'` |
| `?&` | All keys exist | `metadata ?& array['a','b']` |
| `?\|` | Any key exists | `metadata ?\| array['a','b']` |
| `@>` | Contains | `metadata @> '{"key":"val"}'` |
| `<@` | Contained by | `'{"key":"val"}' <@ metadata` |
| `\|\|` | Concatenate | `settings \|\| '{"new":"val"}'` |
| `-` | Remove key | `metadata - 'key'` |

## Related

- [database-pattern.md](database-pattern.md) — Database client pattern
- [repository-pattern.md](repository-pattern.md) — Repository pattern
- [mapper-pattern.md](mapper-pattern.md) — Mapper pattern for complex types
