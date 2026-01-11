// Package pagination provides cursor-based (keyset) pagination utilities.
//
// This example shows:
// - Cursor encoding/decoding
// - Generic page response
// - Multi-column keyset pagination
// - Repository integration
package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// Note: IDs are string type, not uuid.UUID.

// ---------- Generic Page Types ----------

// PageRequest is the request for a paginated list.
type PageRequest struct {
	Cursor string // Base64 encoded cursor
	Limit  int    // Max items per page
}

// PageResponse is a generic paginated response.
type PageResponse[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// NewPageResponse creates a new page response.
func NewPageResponse[T any](items []T, nextCursor string) PageResponse[T] {
	return PageResponse[T]{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    nextCursor != "",
	}
}

// ---------- Cursor Types ----------

// IDCursor is a simple cursor using only ID.
type IDCursor struct {
	ID string `json:"id"`
}

// TimestampCursor is a cursor using timestamp and ID (for created_at ordering).
type TimestampCursor struct {
	Timestamp time.Time `json:"ts"`
	ID        string    `json:"id"`
}

// ---------- Cursor Encoding ----------

// EncodeCursor encodes any cursor struct to a base64 string.
func EncodeCursor[T any](cursor *T) string {
	if cursor == nil {
		return ""
	}
	data, err := json.Marshal(cursor)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(data)
}

// DecodeCursor decodes a base64 string to a cursor struct.
func DecodeCursor[T any](s string) (*T, error) {
	if s == "" {
		return nil, nil
	}
	data, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor encoding: %w", err)
	}
	var cursor T
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, fmt.Errorf("invalid cursor format: %w", err)
	}
	return &cursor, nil
}

// ---------- Pagination Helpers ----------

// Paginate handles the common pagination pattern:
// 1. Fetches limit+1 items
// 2. Determines if there are more items
// 3. Creates next cursor from last item
func Paginate[T any, C any](
	items []T,
	limit int,
	cursorFn func(T) C,
) ([]T, string) {
	if len(items) <= limit {
		return items, ""
	}

	// We have more items
	items = items[:limit]
	lastItem := items[len(items)-1]
	nextCursor := cursorFn(lastItem)

	return items, EncodeCursor(&nextCursor)
}

// ---------- Example Repository Usage ----------

// UserCursor is the cursor for user pagination.
type UserCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
}

// Example usage in repository:
//
//	func (r *userRepo) ListUsers(ctx context.Context, cursor *UserCursor, limit int) ([]*User, *UserCursor, error) {
//	    qb := squirrel.Select("*").
//	        From("users").
//	        OrderBy("created_at DESC", "id DESC").
//	        Limit(uint64(limit + 1))
//
//	    if cursor != nil {
//	        qb = qb.Where(
//	            squirrel.Or{
//	                squirrel.Lt{"created_at": cursor.CreatedAt},
//	                squirrel.And{
//	                    squirrel.Eq{"created_at": cursor.CreatedAt},
//	                    squirrel.Lt{"id": cursor.ID},
//	                },
//	            },
//	        )
//	    }
//
//	    sql, args, err := qb.PlaceholderFormat(squirrel.Dollar).ToSql()
//	    if err != nil {
//	        return nil, nil, err
//	    }
//
//	    rows, err := r.db.Query(ctx, sql, args...)
//	    // ... scan rows into users ...
//
//	    // Paginate results
//	    users, nextCursorStr := pagination.Paginate(users, limit, func(u *User) UserCursor {
//	        return UserCursor{CreatedAt: u.CreatedAt, ID: u.ID}
//	    })
//
//	    var nextCursor *UserCursor
//	    if nextCursorStr != "" {
//	        nextCursor, _ = pagination.DecodeCursor[UserCursor](nextCursorStr)
//	    }
//
//	    return users, nextCursor, nil
//	}

// ---------- Offset Pagination (for admin panels) ----------

// OffsetRequest is the request for offset-based pagination.
type OffsetRequest struct {
	Limit  int `json:"limit" validate:"min=1,max=100"`
	Offset int `json:"offset" validate:"min=0"`
}

// OffsetResponse is an offset-based paginated response.
type OffsetResponse[T any] struct {
	Items      []T   `json:"items"`
	TotalCount int64 `json:"total_count"`
	Limit      int   `json:"limit"`
	Offset     int   `json:"offset"`
}

// NewOffsetResponse creates a new offset response.
func NewOffsetResponse[T any](items []T, total int64, limit, offset int) OffsetResponse[T] {
	return OffsetResponse[T]{
		Items:      items,
		TotalCount: total,
		Limit:      limit,
		Offset:     offset,
	}
}

// DefaultLimit returns the default limit if not specified.
func DefaultLimit(limit, defaultVal, maxVal int) int {
	if limit <= 0 {
		return defaultVal
	}
	if limit > maxVal {
		return maxVal
	}
	return limit
}
