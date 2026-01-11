# Filter Pattern for Database Queries

Type-safe, composable filtering for repository queries using Squirrel.

## Overview

Filter pattern separates query criteria from query execution:

```go
// 1. Filter struct in models (what to filter)
type UserFilter struct {
    ID        []string
    IsActive  *bool
    CreatedAt *DateFilter
}

// 2. getCondition in repository (how to filter)
func (r *userRepo) getUserCondition(filter *UserFilter) []sq.Sqlizer

// 3. Usage in repository methods
func (r *userRepo) GetUsers(filter *UserFilter) ([]*User, error)
```

## Package Structure

```
internal/
├── models/
│   ├── common.go        ← DateFilter, OptionalRange (shared types)
│   └── user.go          ← User + UserFilter (together)
├── storage/
│   ├── common.go        ← SQL helpers (NewSqlMatch, etc.)
│   └── user.go          ← UserRepository + getUserCondition()
```

**Rule:** Filter lives with its model, getCondition() lives in repository.

## Filter Struct

### Design Principles

| Principle | Example | Why |
|-----------|---------|-----|
| Slices for multi-value | `ID []string` | Supports `IN (...)` queries |
| Pointers for optional | `*bool`, `*DateFilter` | `nil` = not filtered |
| No single values | `ID []string` not `ID string` | Consistency, always `sq.Eq` |
| Nested filters for ranges | `CreatedAt *DateFilter` | Clean range handling |

### Basic Filter

```go
// internal/models/user.go
type UserFilter struct {
    ID        []string    // Filter by IDs (IN clause)
    Email     []string    // Filter by emails
    IsActive  *bool       // Filter by active status
    Role      []string    // Filter by roles
    CreatedAt *DateFilter // Filter by date range
}
```

### Common Filter Types

```go
// internal/models/common.go
type DateFilter struct {
    From *time.Time
    To   *time.Time
}

type OptionalIntRange struct {
    Min *int
    Max *int
}

type OptionalFloatRange struct {
    Min *float64
    Max *float64
}
```

## getCondition Method

### Pattern

```go
// internal/storage/user.go
func (r *userRepo) getUserCondition(filter *UserFilter) []sq.Sqlizer {
    conditions := make([]sq.Sqlizer, 0)

    if filter == nil {
        return conditions
    }

    // Slice fields → sq.Eq (handles IN automatically)
    if len(filter.ID) > 0 {
        conditions = append(conditions, sq.Eq{"u.id": filter.ID})
    }

    if len(filter.Email) > 0 {
        conditions = append(conditions, sq.Eq{"u.email": filter.Email})
    }

    // Pointer fields → check nil first
    if filter.IsActive != nil {
        conditions = append(conditions, sq.Eq{"u.is_active": *filter.IsActive})
    }

    // Nested date filter
    if filter.CreatedAt != nil {
        if filter.CreatedAt.From != nil {
            conditions = append(conditions, sq.GtOrEq{"u.created_at": filter.CreatedAt.From})
        }
        if filter.CreatedAt.To != nil {
            conditions = append(conditions, sq.LtOrEq{"u.created_at": filter.CreatedAt.To})
        }
    }

    return conditions
}
```

### Key Rules

1. **Always check `filter == nil`** — return empty slice
2. **Check `len() > 0` for slices** — empty slice = not filtered
3. **Check `!= nil` for pointers** — nil = not filtered
4. **Return `[]sq.Sqlizer`** — composable with `sq.And()`
5. **Use table alias** — `u.id` not `id` for join safety

## Repository Interface

```go
// internal/storage/storage.go
type UserRepository interface {
    GetUsers(filter *UserFilter) ([]*User, error)
    GetUserByID(id string) (*User, error)
    CountUsers(filter *UserFilter) (int64, error)
}
```

## Repository Implementation

```go
// internal/storage/user.go
func (r *userRepo) GetUsers(filter *UserFilter) ([]*User, error) {
    conditions := r.getUserCondition(filter)

    query := sq.Select(UserColumns()...).
        From("users u").
        Where(sq.And(conditions)).
        OrderBy("u.created_at DESC").
        PlaceholderFormat(sq.Dollar)

    sql, args, err := query.ToSql()
    if err != nil {
        return nil, fmt.Errorf("build query: %w", err)
    }

    rows, err := r.db.Query(ctx, sql, args...)
    // ... scan and return
}

func (r *userRepo) CountUsers(filter *UserFilter) (int64, error) {
    conditions := r.getUserCondition(filter)

    query := sq.Select("COUNT(*)").
        From("users u").
        Where(sq.And(conditions)).
        PlaceholderFormat(sq.Dollar)

    // ... execute and return count
}
```

## Advanced Patterns

### OR Conditions

```go
// Filter by home OR away team
if len(filter.TeamID) > 0 {
    conditions = append(conditions, sq.Or{
        sq.Eq{"m.home_team_id": filter.TeamID},
        sq.Eq{"m.away_team_id": filter.TeamID},
    })
}
```

### Enum/Lifecycle Filtering

```go
// Filter by multiple lifecycle stages
lifecycleConditions := make([]sq.Sqlizer, 0)
for _, stage := range filter.Lifecycle {
    switch stage {
    case LifecycleUpcoming:
        lifecycleConditions = append(lifecycleConditions, sq.Eq{"g.is_started": false})
    case LifecycleOngoing:
        lifecycleConditions = append(lifecycleConditions, sq.And{
            sq.Eq{"g.is_started": true},
            sq.Eq{"g.is_completed": false},
        })
    case LifecycleFinished:
        lifecycleConditions = append(lifecycleConditions, sq.Eq{"g.is_completed": true})
    }
}
if len(lifecycleConditions) > 0 {
    conditions = append(conditions, sq.Or(lifecycleConditions))
}
```

### Exclusion Filters

```go
type UserFilter struct {
    ID             []string
    ExcludeID      []string // IDs to exclude
    CreatorID      []string
    CreatorExclude bool     // Exclude instead of include
}

// In getCondition:
if len(filter.ExcludeID) > 0 {
    conditions = append(conditions, sq.NotEq{"u.id": filter.ExcludeID})
}

if len(filter.CreatorID) > 0 {
    if filter.CreatorExclude {
        conditions = append(conditions, sq.NotEq{"u.creator_id": filter.CreatorID})
    } else {
        conditions = append(conditions, sq.Eq{"u.creator_id": filter.CreatorID})
    }
}
```

## PostgreSQL Array Operations

For PostgreSQL array columns, use custom SQL helpers:

```go
// internal/storage/common.go

// SqlArrayContains checks if array contains all values (@> operator)
type SqlArrayContains struct {
    Field     string
    Values    []any
    ValueType string // e.g., "uuid[]", "text[]"
}

func NewSqlArrayContains(field string, values []any, valueType string) SqlArrayContains {
    return SqlArrayContains{Field: field, Values: values, ValueType: valueType}
}

func (s SqlArrayContains) ToSql() (string, []any, error) {
    placeholders := make([]string, len(s.Values))
    for i := range s.Values {
        placeholders[i] = fmt.Sprintf("$%d", i+1)
    }
    sql := fmt.Sprintf("%s @> ARRAY[%s]::%s", s.Field, strings.Join(placeholders, ","), s.ValueType)
    return sql, s.Values, nil
}

// Usage:
if len(filter.Label) > 0 {
    labels := make([]any, len(filter.Label))
    for i, l := range filter.Label {
        labels[i] = l
    }
    conditions = append(conditions, NewSqlArrayContains("g.labels", labels, "text[]"))
}
```

## Best Practices

### Do

- Keep Filter struct close to its model
- Use consistent naming: `XxxFilter`, `getXxxCondition()`
- Always handle nil filter gracefully
- Use table aliases in conditions
- Return `[]sq.Sqlizer` for composability

### Don't

- Put all filters in one file
- Use single values instead of slices
- Forget nil checks
- Hardcode table names without aliases
- Mix filter logic with business logic

## Dependencies

```bash
go get github.com/Masterminds/squirrel@latest
```

## See Also

- [Repository Pattern](repository-pattern.md)
- [Pagination Pattern](pagination-pattern.md)
