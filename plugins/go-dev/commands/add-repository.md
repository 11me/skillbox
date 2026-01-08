---
name: add-repository
description: Generate a new repository with interface for Go projects
argument-hint: "<EntityName>"
---

# /go-add-repository

Generate a new repository with interface and register it in the Storage.

## Usage

```
/go-add-repository <entity-name>
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `entity-name` | Yes | Entity name in PascalCase (e.g., `User`, `Order`, `Product`) |

## Prerequisites

- Go project with Repository pattern
- `internal/storage/storage.go` exists
- Model for the entity exists in `internal/models/`

## Generated Files

```
internal/models/
└── <name>.go         # Updated with {{Name}}Filter struct

internal/storage/
├── storage.go        # Updated with repository getter
└── <name>.go         # New repository file with get{{Name}}Condition()
```

## Steps

1. **Validate entity name**:
   - Must be PascalCase: `[A-Z][a-zA-Z0-9]*`
   - Must not already exist in `internal/storage/`

2. **Check for model**:
   - Verify `internal/models/<name>.go` exists
   - If not, suggest running `/go-add-model` first

3. **Generate Filter struct** (add to `internal/models/<name>.go`):

```go
// {{Name}}Filter defines filtering criteria for {{name}} queries.
type {{Name}}Filter struct {
    ID        []string    // Filter by IDs
    // TODO: Add entity-specific filter fields
    CreatedAt *DateFilter // Filter by creation date range
}
```

**Note:** Also ensure `DateFilter` exists in `internal/models/common.go`:
```go
type DateFilter struct {
    From *time.Time
    To   *time.Time
}
```

4. **Generate repository file** (`internal/storage/<name>.go`):

```go
package storage

import (
    "context"
    "fmt"

    "github.com/Masterminds/squirrel"
    "github.com/google/uuid"

    "{{MODULE}}/internal/common"
    "{{MODULE}}/internal/models"
)

type {{Name}}Repository interface {
    Create(ctx context.Context, entity *models.{{Name}}) error
    GetByID(ctx context.Context, id string) (*models.{{Name}}, error)
    FindByFilter(ctx context.Context, filter *models.{{Name}}Filter) ([]*models.{{Name}}, error)
    Update(ctx context.Context, entity *models.{{Name}}) error
    Delete(ctx context.Context, id string) error
    Serialize(ctx context.Context, label string) error
}

type {{name}}Repository struct {
    db QueryExecer
}

func new{{Name}}Repository(db QueryExecer) *{{name}}Repository {
    return &{{name}}Repository{db: db}
}

func (r *{{name}}Repository) Create(ctx context.Context, entity *models.{{Name}}) error {
    query, args, err := squirrel.
        Insert("{{table_name}}").
        Columns("id", "created_at", "updated_at").
        Values(entity.ID, entity.CreatedAt, entity.UpdatedAt).
        PlaceholderFormat(squirrel.Dollar).
        ToSql()
    if err != nil {
        return fmt.Errorf("build query: %w", err)
    }

    if _, err := r.db.Exec(ctx, query, args...); err != nil {
        return fmt.Errorf("exec: %w", err)
    }
    return nil
}

func (r *{{name}}Repository) GetByID(ctx context.Context, id uuid.UUID) (*models.{{Name}}, error) {
    query, args, err := squirrel.
        Select("id", "created_at", "updated_at").
        From("{{table_name}}").
        Where(squirrel.Eq{"id": id}).
        PlaceholderFormat(squirrel.Dollar).
        ToSql()
    if err != nil {
        return nil, fmt.Errorf("build query: %w", err)
    }

    row := r.db.QueryRow(ctx, query, args...)
    entity := &models.{{Name}}{}
    if err := row.Scan(&entity.ID, &entity.CreatedAt, &entity.UpdatedAt); err != nil {
        if err.Error() == "no rows in result set" {
            return nil, common.EntityNotFound("{{name}} not found")
        }
        return nil, fmt.Errorf("scan: %w", err)
    }
    return entity, nil
}

func (r *{{name}}Repository) Update(ctx context.Context, entity *models.{{Name}}) error {
    query, args, err := squirrel.
        Update("{{table_name}}").
        Set("updated_at", entity.UpdatedAt).
        Where(squirrel.Eq{"id": entity.ID}).
        PlaceholderFormat(squirrel.Dollar).
        ToSql()
    if err != nil {
        return fmt.Errorf("build query: %w", err)
    }

    if _, err := r.db.Exec(ctx, query, args...); err != nil {
        return fmt.Errorf("exec: %w", err)
    }
    return nil
}

func (r *{{name}}Repository) Delete(ctx context.Context, id uuid.UUID) error {
    query, args, err := squirrel.
        Delete("{{table_name}}").
        Where(squirrel.Eq{"id": id}).
        PlaceholderFormat(squirrel.Dollar).
        ToSql()
    if err != nil {
        return fmt.Errorf("build query: %w", err)
    }

    if _, err := r.db.Exec(ctx, query, args...); err != nil {
        return fmt.Errorf("exec: %w", err)
    }
    return nil
}

// Serialize acquires an advisory lock for serializable transactions.
// Use inside WithTx with pgx.Serializable to prevent serialization conflicts.
// See: advisory-lock-pattern.md
func (r *{{name}}Repository) Serialize(ctx context.Context, label string) error {
    query, _, err := squirrel.
        Select("pg_advisory_xact_lock(hashtext(?))").
        PlaceholderFormat(squirrel.Dollar).
        ToSql()
    if err != nil {
        return fmt.Errorf("build lock query: %w", err)
    }

    if _, err := r.db.Exec(ctx, query, label); err != nil {
        return fmt.Errorf("acquire lock: %w", err)
    }
    return nil
}

// get{{Name}}Condition converts filter to SQL conditions.
func (r *{{name}}Repository) get{{Name}}Condition(filter *models.{{Name}}Filter) []squirrel.Sqlizer {
    conditions := make([]squirrel.Sqlizer, 0)

    if filter == nil {
        return conditions
    }

    if len(filter.ID) > 0 {
        conditions = append(conditions, squirrel.Eq{"id": filter.ID})
    }

    // TODO: Add entity-specific filter conditions

    if filter.CreatedAt != nil {
        if filter.CreatedAt.From != nil {
            conditions = append(conditions, squirrel.GtOrEq{"created_at": filter.CreatedAt.From})
        }
        if filter.CreatedAt.To != nil {
            conditions = append(conditions, squirrel.LtOrEq{"created_at": filter.CreatedAt.To})
        }
    }

    return conditions
}

func (r *{{name}}Repository) FindByFilter(ctx context.Context, filter *models.{{Name}}Filter) ([]*models.{{Name}}, error) {
    conditions := r.get{{Name}}Condition(filter)

    query, args, err := squirrel.
        Select("id", "created_at", "updated_at").
        From("{{table_name}}").
        Where(squirrel.And(conditions)).
        OrderBy("created_at DESC").
        PlaceholderFormat(squirrel.Dollar).
        ToSql()
    if err != nil {
        return nil, fmt.Errorf("build query: %w", err)
    }

    rows, err := r.db.Query(ctx, query, args...)
    if err != nil {
        return nil, fmt.Errorf("query: %w", err)
    }
    defer rows.Close()

    var entities []*models.{{Name}}
    for rows.Next() {
        entity := &models.{{Name}}{}
        if err := rows.Scan(&entity.ID, &entity.CreatedAt, &entity.UpdatedAt); err != nil {
            return nil, fmt.Errorf("scan: %w", err)
        }
        entities = append(entities, entity)
    }
    return entities, nil
}
```

5. **Update Storage interface** (`internal/storage/storage.go`):

```go
type Storage interface {
    // ... existing methods
    {{Name}}s() {{Name}}Repository
}

// In storage struct
func (s *storage) {{Name}}s() {{Name}}Repository {
    return new{{Name}}Repository(s.getDB(ctx))
}
```

6. **Report created files** and suggest next steps.

## Example

```
/go-add-repository Product
```

Creates:
- `internal/models/product.go` with `ProductFilter` struct
- `internal/storage/product.go` with:
  - `ProductRepository` interface
  - `productRepository` implementation
  - `getProductCondition()` filter method
  - `FindByFilter()` method using filter pattern

Updates `storage.go`:
```go
func (s *storage) Products() ProductRepository {
    return newProductRepository(s.getDB(ctx))
}
```

## Final Validation (REQUIRED)

After generating files, run:

```bash
golangci-lint run --fast internal/storage/
```

Fix any issues before reporting completion.

## Next Steps After Generation

1. Add entity-specific fields to model
2. Update repository methods with actual columns
3. Create migration for the table
4. Add tests for repository
