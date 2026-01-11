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

// Registry holds all services
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

// UserService handles user business logic
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
	return s.storage.Users().FindByID(ctx, id)
}

func (s *UserService) Create(ctx context.Context, name, email string) (*models.User, error) {
	if name == "" {
		return nil, common.ValidationFailed("name is required")
	}
	if email == "" {
		return nil, common.ValidationFailed("email is required")
	}

	existing, err := s.storage.Users().FindByEmail(ctx, email)
	if err != nil && !common.IsNotFound(err) {
		return nil, err
	}
	if existing != nil {
		return nil, common.StateConflict("user with this email already exists")
	}

	user := &models.User{
		ID:    uuid.NewString(),
		Name:  name,
		Email: email,
	}

	err = s.storage.ExecReadCommitted(ctx, func(ctx context.Context) error {
		return s.storage.Users().Create(ctx, user)
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Update(ctx context.Context, id string, name, email string) (*models.User, error) {
	user, err := s.storage.Users().FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		user.Name = name
	}
	if email != "" {
		user.Email = email
	}

	err = s.storage.ExecReadCommitted(ctx, func(ctx context.Context) error {
		return s.storage.Users().Update(ctx, user)
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	return s.storage.ExecReadCommitted(ctx, func(ctx context.Context) error {
		return s.storage.Users().Delete(ctx, id)
	})
}
