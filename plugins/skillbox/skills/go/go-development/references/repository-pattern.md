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
    Create(ctx context.Context, user *models.User) error
    Update(ctx context.Context, user *models.User) error
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

func (r *userRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
    query, args, err := psql.
        Select("id", "name", "email", "created_at", "updated_at").
        From("users").
        Where(sq.Eq{"id": id}).
        ToSql()
    if err != nil {
        return nil, err
    }

    row := r.db.Executor(ctx).QueryRow(ctx, query, args...)
    return scanUser(row)
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
    now := time.Now()
    user.CreatedAt = now
    user.UpdatedAt = now

    query, args, err := psql.
        Insert("users").
        Columns("id", "name", "email", "created_at", "updated_at").
        Values(user.ID, user.Name, user.Email, user.CreatedAt, user.UpdatedAt).
        ToSql()
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
