# Pagination Pattern

Production pagination patterns: keyset (cursor) and offset.

## Keyset vs Offset Pagination

| Feature | Keyset (Cursor) | Offset |
|---------|-----------------|--------|
| Performance at scale | ✅ O(1) | ❌ O(n) degrades |
| Consistent results | ✅ No duplicates/skips | ❌ Drift on inserts |
| Random page access | ❌ Sequential only | ✅ Jump to any page |
| Implementation | Complex | Simple |

**Recommendation:** Use keyset for APIs, offset for admin panels.

## Keyset Pagination (Cursor-Based)

### Concept

Instead of `OFFSET 1000`, use `WHERE id > last_seen_id ORDER BY id LIMIT 20`.

```sql
-- First page
SELECT * FROM users ORDER BY id LIMIT 20;

-- Next page (cursor = last id from previous page)
SELECT * FROM users WHERE id > $1 ORDER BY id LIMIT 20;
```

### Single Column Keyset

```go
type PageRequest struct {
    Cursor string `json:"cursor"` // Base64 encoded cursor
    Limit  int    `json:"limit"`
}

type PageResponse[T any] struct {
    Items      []T    `json:"items"`
    NextCursor string `json:"next_cursor,omitempty"`
    HasMore    bool   `json:"has_more"`
}

// In repository
func (r *userRepo) ListAfter(ctx context.Context, cursor string, limit int) ([]*User, error) {
    query := squirrel.Select("*").
        From("users").
        Where(squirrel.Gt{"id": cursor}).
        OrderBy("id ASC").
        Limit(uint64(limit + 1)) // +1 to detect if there are more

    rows, err := query.RunWith(r.db).QueryContext(ctx)
    // ...
}
```

### Multi-Column Keyset

For sorting by multiple columns (e.g., `created_at`, `id`):

```go
type Cursor struct {
    CreatedAt time.Time `json:"created_at"`
    ID        string    `json:"id"`
}

func (r *userRepo) ListPaginated(ctx context.Context, cursor *Cursor, limit int) ([]*User, *Cursor, error) {
    qb := squirrel.Select("*").
        From("users").
        OrderBy("created_at DESC", "id DESC").
        Limit(uint64(limit + 1))

    if cursor != nil {
        // Multi-column comparison: (created_at, id) < (cursor_created_at, cursor_id)
        qb = qb.Where(
            squirrel.Or{
                squirrel.Lt{"created_at": cursor.CreatedAt},
                squirrel.And{
                    squirrel.Eq{"created_at": cursor.CreatedAt},
                    squirrel.Lt{"id": cursor.ID},
                },
            },
        )
    }

    sql, args, err := qb.PlaceholderFormat(squirrel.Dollar).ToSql()
    if err != nil {
        return nil, nil, err
    }

    rows, err := r.db.Query(ctx, sql, args...)
    if err != nil {
        return nil, nil, err
    }
    defer rows.Close()

    var users []*User
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.Name, &u.CreatedAt); err != nil {
            return nil, nil, err
        }
        users = append(users, &u)
    }

    // Check if there are more
    var nextCursor *Cursor
    if len(users) > limit {
        users = users[:limit]
        last := users[len(users)-1]
        nextCursor = &Cursor{
            CreatedAt: last.CreatedAt,
            ID:        last.ID,
        }
    }

    return users, nextCursor, nil
}
```

### Cursor Encoding

```go
import "encoding/base64"

func EncodeCursor(cursor *Cursor) string {
    if cursor == nil {
        return ""
    }
    data, _ := json.Marshal(cursor)
    return base64.URLEncoding.EncodeToString(data)
}

func DecodeCursor(s string) (*Cursor, error) {
    if s == "" {
        return nil, nil
    }
    data, err := base64.URLEncoding.DecodeString(s)
    if err != nil {
        return nil, fmt.Errorf("invalid cursor: %w", err)
    }
    var cursor Cursor
    if err := json.Unmarshal(data, &cursor); err != nil {
        return nil, fmt.Errorf("invalid cursor: %w", err)
    }
    return &cursor, nil
}
```

## Offset Pagination

For admin panels and simple use cases:

```go
type OffsetRequest struct {
    Limit  int `json:"limit" validate:"min=1,max=100"`
    Offset int `json:"offset" validate:"min=0"`
}

type OffsetResponse[T any] struct {
    Items      []T   `json:"items"`
    TotalCount int64 `json:"total_count"`
    Limit      int   `json:"limit"`
    Offset     int   `json:"offset"`
}

func (r *userRepo) List(ctx context.Context, limit, offset int) ([]*User, int64, error) {
    // Count query
    countQuery := squirrel.Select("COUNT(*)").From("users")
    var total int64
    err := countQuery.RunWith(r.db).QueryRowContext(ctx).Scan(&total)
    if err != nil {
        return nil, 0, err
    }

    // Data query
    dataQuery := squirrel.Select("*").
        From("users").
        OrderBy("id").
        Limit(uint64(limit)).
        Offset(uint64(offset))

    rows, err := dataQuery.RunWith(r.db).QueryContext(ctx)
    // ...

    return users, total, nil
}
```

## API Response Formats

### Keyset Response

```json
{
    "items": [...],
    "next_cursor": "eyJjcmVhdGVkX2F0IjoiMjAyNC0wMS0wMVQxMjowMDowMFoiLCJpZCI6IjEyMzQifQ==",
    "has_more": true
}
```

### Offset Response

```json
{
    "items": [...],
    "total_count": 1250,
    "limit": 20,
    "offset": 40
}
```

## Handler Integration

```go
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
    cursorStr := r.URL.Query().Get("cursor")
    limit := getIntQuery(r, "limit", 20)

    cursor, err := DecodeCursor(cursorStr)
    if err != nil {
        h.error(w, http.StatusBadRequest, "invalid cursor")
        return
    }

    users, nextCursor, err := h.services.Users().List(r.Context(), cursor, limit)
    if err != nil {
        h.handleError(w, err)
        return
    }

    h.json(w, http.StatusOK, PageResponse[UserResponse]{
        Items:      mapUsers(users),
        NextCursor: EncodeCursor(nextCursor),
        HasMore:    nextCursor != nil,
    })
}
```

## Keyset with Nullable Columns

For columns that can be NULL, use COALESCE or handle NULLs explicitly:

```go
// Sorting by nullable "deleted_at" column
func (r *userRepo) ListWithDeleted(ctx context.Context, cursor *Cursor, limit int) ([]*User, error) {
    qb := squirrel.Select("*").
        From("users").
        OrderBy("COALESCE(deleted_at, '9999-12-31') DESC", "id DESC").
        Limit(uint64(limit + 1))

    if cursor != nil {
        // Handle NULL values in cursor
        if cursor.DeletedAt == nil {
            qb = qb.Where(
                squirrel.Or{
                    squirrel.NotEq{"deleted_at": nil},
                    squirrel.And{
                        squirrel.Eq{"deleted_at": nil},
                        squirrel.Lt{"id": cursor.ID},
                    },
                },
            )
        } else {
            qb = qb.Where(
                squirrel.Or{
                    squirrel.Lt{"deleted_at": cursor.DeletedAt},
                    squirrel.And{
                        squirrel.Eq{"deleted_at": cursor.DeletedAt},
                        squirrel.Lt{"id": cursor.ID},
                    },
                },
            )
        }
    }

    // ...
}
```

## Best Practices

### DO:
- ✅ Use keyset pagination for public APIs
- ✅ Always include a unique column (ID) in sort order
- ✅ Encode cursors to prevent tampering
- ✅ Limit max page size (e.g., 100)
- ✅ Return `has_more` flag for UI

### DON'T:
- ❌ Use offset for large datasets (>10K rows)
- ❌ Expose raw cursor values (security risk)
- ❌ Allow arbitrary ORDER BY from user input
- ❌ Count total for keyset (defeats the purpose)

## Related

- [repository-pattern.md](repository-pattern.md) — Repository pattern
- [http-handler-pattern.md](http-handler-pattern.md) — HTTP handlers
