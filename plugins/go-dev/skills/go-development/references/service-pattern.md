# Service Layer Pattern

Using Service Registry for dependency injection.

## Service Registry

```go
package services

import (
    "myapp/internal/config"
    "myapp/internal/storage"
)

type Registry struct {
    conf    *config.Config
    storage storage.Storage
}

func NewRegistry(conf *config.Config, storage storage.Storage) *Registry {
    return &Registry{
        conf:    conf,
        storage: storage,
    }
}

func (r *Registry) UserService() *UserService {
    return NewUserService(r.storage, r.conf)
}

func (r *Registry) OrderService() *OrderService {
    return NewOrderService(r.storage, r.conf)
}
```

## Service Implementation

```go
package services

import (
    "context"

    "github.com/google/uuid"
    "myapp/internal/common"
    "myapp/internal/config"
    "myapp/internal/models"
    "myapp/internal/storage"
)

// Note: IDs are string type, not uuid.UUID.
// Generate new IDs with uuid.NewString().

type UserService struct {
    storage storage.Storage
    conf    *config.Config
}

func NewUserService(storage storage.Storage, conf *config.Config) *UserService {
    return &UserService{
        storage: storage,
        conf:    conf,
    }
}

func (s *UserService) GetByID(ctx context.Context, id string) (*models.User, error) {
    user, err := s.storage.Users().FindByID(ctx, id)
    if err != nil {
        return nil, err
    }
    return user, nil
}

func (s *UserService) Create(ctx context.Context, name, email string) (*models.User, error) {
    if name == "" {
        return nil, common.ValidationFailed("name is required")
    }

    user := &models.User{
        ID:    uuid.NewString(),
        Name:  name,
        Email: email,
    }

    err := s.storage.ExecReadCommitted(ctx, func(ctx context.Context) error {
        return s.storage.Users().Create(ctx, user)
    })
    if err != nil {
        return nil, err
    }

    return user, nil
}
```

## Storage Interface

```go
package storage

import "context"

type TxFunc func(ctx context.Context) error

type Storage interface {
    Users() UserRepository
    Orders() OrderRepository

    ExecReadCommitted(ctx context.Context, fn TxFunc) error
    ExecRepeatableRead(ctx context.Context, fn TxFunc) error
    ExecSerializable(ctx context.Context, fn TxFunc) error
}
```

## Benefits

- **No reflection** — explicit dependencies
- **Testable** — easy to mock Storage
- **Lazy initialization** — services created on demand
- **Consistent** — all services follow same pattern
