// Package models demonstrates Filter pattern for database queries.
// Filter structs live alongside their domain models.
package models

import (
	"time"
)

// =============================================================================
// Domain Model + Filter (in same file: internal/models/user.go)
// =============================================================================

// User represents a user in the system.
type User struct {
	ID        string
	Email     string
	Name      string
	Role      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserColumns returns column names for SELECT queries.
func UserColumns() []string {
	return []string{
		"id", "email", "name", "role",
		"is_active", "created_at", "updated_at",
	}
}

// UserFilter defines filtering criteria for user queries.
// Rules:
// - Slices for multi-value (supports IN clause)
// - Pointers for optional (nil = not filtered)
// - Nested types for ranges (DateFilter)
type UserFilter struct {
	ID        []string    // Filter by user IDs
	Email     []string    // Filter by emails
	Role      []string    // Filter by roles
	IsActive  *bool       // Filter by active status
	CreatedAt *DateFilter // Filter by creation date range
}

// =============================================================================
// Repository Implementation (in: internal/storage/user.go)
// =============================================================================

// Package storage (conceptually - this is just an example)

import (
	sq "github.com/Masterminds/squirrel"
)

// userRepo implements UserRepository.
type userRepo struct {
	db DB // your database interface
}

// getUserCondition converts UserFilter to SQL conditions.
// Private method - only used within repository.
func (r *userRepo) getUserCondition(filter *UserFilter) []sq.Sqlizer {
	conditions := make([]sq.Sqlizer, 0)

	// Always check for nil filter first
	if filter == nil {
		return conditions
	}

	// Slice fields: check len > 0
	// sq.Eq handles single value and IN clause automatically
	if len(filter.ID) > 0 {
		conditions = append(conditions, sq.Eq{"u.id": filter.ID})
	}

	if len(filter.Email) > 0 {
		conditions = append(conditions, sq.Eq{"u.email": filter.Email})
	}

	if len(filter.Role) > 0 {
		conditions = append(conditions, sq.Eq{"u.role": filter.Role})
	}

	// Pointer fields: check != nil
	if filter.IsActive != nil {
		conditions = append(conditions, sq.Eq{"u.is_active": *filter.IsActive})
	}

	// Nested filter: check outer nil, then inner fields
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

// GetUsers retrieves users matching the filter.
func (r *userRepo) GetUsers(ctx context.Context, filter *UserFilter) ([]*User, error) {
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
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var u User
		if err := rows.Scan(
			&u.ID, &u.Email, &u.Name, &u.Role,
			&u.IsActive, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, &u)
	}

	return users, rows.Err()
}

// CountUsers returns count of users matching the filter.
func (r *userRepo) CountUsers(ctx context.Context, filter *UserFilter) (int64, error) {
	conditions := r.getUserCondition(filter)

	query := sq.Select("COUNT(*)").
		From("users u").
		Where(sq.And(conditions)).
		PlaceholderFormat(sq.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("build query: %w", err)
	}

	var count int64
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}

	return count, nil
}

// =============================================================================
// Usage Examples
// =============================================================================

func exampleUsage() {
	// Get all active admins
	activeAdmins, _ := repo.GetUsers(ctx, &UserFilter{
		Role:     []string{"admin"},
		IsActive: ptr(true), // helper: func ptr[T any](v T) *T { return &v }
	})

	// Get users created in last 7 days
	lastWeek := time.Now().AddDate(0, 0, -7)
	recentUsers, _ := repo.GetUsers(ctx, &UserFilter{
		CreatedAt: &DateFilter{From: &lastWeek},
	})

	// Get specific users by IDs
	specificUsers, _ := repo.GetUsers(ctx, &UserFilter{
		ID: []string{"user-1", "user-2", "user-3"},
	})

	// Get all users (no filter)
	allUsers, _ := repo.GetUsers(ctx, nil)

	// Count inactive users
	inactiveCount, _ := repo.CountUsers(ctx, &UserFilter{
		IsActive: ptr(false),
	})
}
