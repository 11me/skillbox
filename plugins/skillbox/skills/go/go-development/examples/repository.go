// Package storage provides database repositories.
//
// This example shows:
// - Repository using pg.Client interface
// - pgx.CollectOneRow/CollectRows for row scanning
// - Squirrel query builder with Dollar placeholders
// - Upsert pattern with ON CONFLICT
// - IDs as string (not uuid.UUID)
package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	sq "github.com/Masterminds/squirrel"

	"myapp/internal/models"
	"myapp/pkg/pg"
)

// ---------- Errors ----------

var ErrUserNotFound = errors.New("user not found")

// ---------- Repository Interface ----------

type Users interface {
	FindByID(ctx context.Context, id string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Find(ctx context.Context, filter *models.UserFilter) ([]*models.User, error)
	Save(ctx context.Context, users ...*models.User) error
	Delete(ctx context.Context, id string) error
}

// ---------- Repository Implementation ----------

type userStorage struct {
	client pg.Client
}

// NewUserStorage creates a new user repository.
func NewUserStorage(client pg.Client) Users {
	return &userStorage{client: client}
}

// ---------- Read Operations ----------

func (s *userStorage) FindByID(ctx context.Context, id string) (*models.User, error) {
	sql, args, err := sq.
		Select("id", "name", "email", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := s.client.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query user: %w", err)
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("collect user: %w", err)
	}

	return &user, nil
}

func (s *userStorage) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	sql, args, err := sq.
		Select("id", "name", "email", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"email": email}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := s.client.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query user: %w", err)
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("collect user: %w", err)
	}

	return &user, nil
}

func (s *userStorage) Find(ctx context.Context, filter *models.UserFilter) ([]*models.User, error) {
	builder := sq.
		Select("id", "name", "email", "created_at", "updated_at").
		From("users").
		PlaceholderFormat(sq.Dollar)

	// Apply filters
	if filter != nil {
		if filter.Name != nil {
			builder = builder.Where(sq.ILike{"name": "%" + *filter.Name + "%"})
		}
		if filter.Email != nil {
			builder = builder.Where(sq.Eq{"email": *filter.Email})
		}
		if filter.Limit > 0 {
			builder = builder.Limit(uint64(filter.Limit))
		}
		if filter.Offset > 0 {
			builder = builder.Offset(uint64(filter.Offset))
		}
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := s.client.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}

	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.User])
	if err != nil {
		return nil, fmt.Errorf("collect users: %w", err)
	}

	// Convert to pointers
	result := make([]*models.User, len(users))
	for i := range users {
		result[i] = &users[i]
	}

	return result, nil
}

// ---------- Write Operations ----------

// Save inserts or updates users (upsert pattern).
func (s *userStorage) Save(ctx context.Context, users ...*models.User) error {
	if len(users) == 0 {
		return nil
	}

	now := time.Now()

	builder := sq.
		Insert("users").
		Columns("id", "name", "email", "created_at", "updated_at").
		Suffix(`ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			email = EXCLUDED.email,
			updated_at = EXCLUDED.updated_at`).
		PlaceholderFormat(sq.Dollar)

	for _, user := range users {
		if user.ID == "" {
			user.ID = uuid.NewString()
		}
		if user.CreatedAt.IsZero() {
			user.CreatedAt = now
		}
		user.UpdatedAt = now

		builder = builder.Values(
			user.ID,
			user.Name,
			user.Email,
			user.CreatedAt,
			user.UpdatedAt,
		)
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	_, err = s.client.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("save users: %w", err)
	}

	return nil
}

func (s *userStorage) Delete(ctx context.Context, id string) error {
	sql, args, err := sq.
		Delete("users").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	_, err = s.client.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	return nil
}

// ---------- Usage Example ----------

// Example usage in service:
//
//	type UserService struct {
//	    storage Storage
//	}
//
//	func (svc *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*models.User, error) {
//	    var user *models.User
//
//	    err := svc.storage.ExecSerializable(ctx, func(ctx context.Context) error {
//	        // Check if email exists (uses transaction automatically)
//	        existing, err := svc.storage.Users().FindByEmail(ctx, req.Email)
//	        if err != nil && !errors.Is(err, ErrUserNotFound) {
//	            return fmt.Errorf("check existing: %w", err)
//	        }
//	        if existing != nil {
//	            return errors.New("email already exists")
//	        }
//
//	        // Create user (uses same transaction)
//	        user = &models.User{
//	            ID:    uuid.NewString(),  // Generate string ID
//	            Name:  req.Name,
//	            Email: req.Email,
//	        }
//
//	        return svc.storage.Users().Save(ctx, user)
//	    })
//
//	    if err != nil {
//	        return nil, err
//	    }
//
//	    return user, nil
//	}
