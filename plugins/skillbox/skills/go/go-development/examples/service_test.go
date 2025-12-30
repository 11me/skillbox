package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"myapp/internal/common"
	"myapp/internal/models"
	"myapp/internal/services"
)

// Service tests use MOCKS for repositories.
// This tests business logic in isolation without database.

// MockUserRepository implements UserRepository interface using testify/mock.
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestUserService_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name      string
		inputName string
		email     string
		setupMock func(*MockUserRepository)
		wantErr   error
	}{
		{
			name:      "success",
			inputName: "Test User",
			email:     "test@example.com",
			setupMock: func(m *MockUserRepository) {
				// Check email doesn't exist
				m.On("GetByEmail", mock.Anything, "test@example.com").
					Return(nil, common.EntityNotFound("user not found"))

				// Create user
				m.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).
					Return(&models.User{
						ID:    uuid.New(),
						Name:  "Test User",
						Email: "test@example.com",
					}, nil)
			},
			wantErr: nil,
		},
		{
			name:      "validation error - empty name",
			inputName: "",
			email:     "test@example.com",
			setupMock: func(m *MockUserRepository) {
				// No mock calls expected - validation fails first
			},
			wantErr: common.ValidationFailed("name is required"),
		},
		{
			name:      "validation error - empty email",
			inputName: "Test User",
			email:     "",
			setupMock: func(m *MockUserRepository) {
				// No mock calls expected - validation fails first
			},
			wantErr: common.ValidationFailed("email is required"),
		},
		{
			name:      "conflict - email exists",
			inputName: "New User",
			email:     "existing@example.com",
			setupMock: func(m *MockUserRepository) {
				// Email already exists
				m.On("GetByEmail", mock.Anything, "existing@example.com").
					Return(&models.User{
						ID:    uuid.New(),
						Name:  "Existing User",
						Email: "existing@example.com",
					}, nil)
			},
			wantErr: common.StateConflict("user with this email already exists"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)

			svc := services.NewUserService(mockRepo, nil)
			user, err := svc.Create(ctx, tt.inputName, tt.email)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, tt.inputName, user.Name)
				assert.Equal(t, tt.email, user.Email)
				assert.NotEqual(t, uuid.Nil, user.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	existingID := uuid.New()
	nonExistingID := uuid.New()

	tests := []struct {
		name      string
		id        uuid.UUID
		setupMock func(*MockUserRepository)
		wantErr   bool
	}{
		{
			name: "success",
			id:   existingID,
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", mock.Anything, existingID).
					Return(&models.User{ID: existingID, Name: "Test"}, nil)
			},
			wantErr: false,
		},
		{
			name: "not found",
			id:   nonExistingID,
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", mock.Anything, nonExistingID).
					Return(nil, common.EntityNotFound("user not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)

			svc := services.NewUserService(mockRepo, nil)
			user, err := svc.GetByID(ctx, tt.id)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.id, user.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name      string
		id        uuid.UUID
		setupMock func(*MockUserRepository)
		wantErr   bool
	}{
		{
			name: "success",
			id:   userID,
			setupMock: func(m *MockUserRepository) {
				// First check user exists
				m.On("GetByID", mock.Anything, userID).
					Return(&models.User{ID: userID}, nil)
				// Then delete
				m.On("Delete", mock.Anything, userID).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "user not found",
			id:   userID,
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", mock.Anything, userID).
					Return(nil, common.EntityNotFound("user not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)

			svc := services.NewUserService(mockRepo, nil)
			err := svc.Delete(ctx, tt.id)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
