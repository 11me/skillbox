# Repository Pattern

Repository + Mapper for data access.

## Repository Interface

```go
package storage

import (
    "context"

    "myapp/internal/models"
)

// Note: IDs are string type, not uuid.UUID.

type UserRepository interface {
    FindByID(ctx context.Context, id string) (*models.User, error)
    FindByEmail(ctx context.Context, email string) (*models.User, error)
    Save(ctx context.Context, users ...*models.User) error  // Upsert
    Delete(ctx context.Context, id string) error
}
```

## Repository Implementation

```go
package storage

import (
    "context"
    "errors"
    "time"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    sq "github.com/Masterminds/squirrel"
    "myapp/internal/common"
    "myapp/internal/models"
    "myapp/pkg/postgres"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type userRepository struct {
    db *postgres.Client
}

func NewUserRepository(db *postgres.Client) UserRepository {
    return &userRepository{db: db}
}

// Note: UserColumns() is defined in models package, not here.
// See mapper-pattern.md for Columns/Values pattern.

func (r *userRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
    query, args, err := psql.
        Select(models.UserColumns()...).
        From("users").
        Where(sq.Eq{"id": id}).
        ToSql()
    if err != nil {
        return nil, err
    }

    row := r.db.Executor(ctx).QueryRow(ctx, query, args...)
    return scanUser(row)
}

// Save inserts or updates users (upsert pattern).
func (r *userRepository) Save(ctx context.Context, users ...*models.User) error {
    if len(users) == 0 {
        return nil
    }

    now := time.Now()

    builder := psql.
        Insert("users").
        Columns(models.UserColumns()...).
        Suffix(`ON CONFLICT (id) DO UPDATE SET
            name = EXCLUDED.name,
            email = EXCLUDED.email,
            updated_at = EXCLUDED.updated_at`)

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

    query, args, err := builder.ToSql()
    if err != nil {
        return err
    }

    _, err = r.db.Executor(ctx).Exec(ctx, query, args...)
    return err
}
```

## Mapper (Scanner)

```go
func scanUser(row pgx.Row) (*models.User, error) {
    var u models.User
    err := row.Scan(
        &u.ID,
        &u.Name,
        &u.Email,
        &u.CreatedAt,
        &u.UpdatedAt,
    )
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, common.EntityNotFound("user not found")
        }
        return nil, err
    }
    return &u, nil
}

func scanUsers(rows pgx.Rows) ([]*models.User, error) {
    var users []*models.User
    for rows.Next() {
        var u models.User
        if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
            return nil, err
        }
        users = append(users, &u)
    }
    return users, rows.Err()
}
```

## Model

```go
package models

import "time"

// Note: IDs are string type.
// Generate new IDs with uuid.NewString() from github.com/google/uuid.

type User struct {
    ID        string
    Name      string
    Email     string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

## Storage Factory

```go
type storage struct {
    db    *postgres.Client
    users UserRepository
}

func NewStorage(db *postgres.Client) Storage {
    return &storage{
        db:    db,
        users: NewUserRepository(db),
    }
}

func (s *storage) Users() UserRepository {
    return s.users
}

func (s *storage) ExecReadCommitted(ctx context.Context, fn TxFunc) error {
    return s.db.ExecReadCommitted(ctx, fn)
}
```
