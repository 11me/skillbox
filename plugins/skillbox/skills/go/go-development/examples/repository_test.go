package storage_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"myapp/internal/models"
	"myapp/internal/storage"
)

// Repository tests use REAL PostgreSQL database.
// DO NOT mock SQL queries in repository tests.
// The purpose of repository tests is to verify actual SQL works correctly.

func TestUserRepository_Create(t *testing.T) {
	t.Parallel()

	pool := connectDB(t)
	repo := storage.NewUserRepository(pool)

	ctx := context.Background()
	user := &models.User{
		Name:  "Test User",
		Email: fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8]),
	}

	// Create user
	created, err := repo.Create(ctx, user)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, created.ID)
	assert.Equal(t, user.Name, created.Name)
	assert.Equal(t, user.Email, created.Email)
	assert.False(t, created.CreatedAt.IsZero())

	// Verify in database
	found, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, user.Name, found.Name)
	assert.Equal(t, user.Email, found.Email)
}

func TestUserRepository_GetByID(t *testing.T) {
	t.Parallel()

	pool := connectDB(t)
	repo := storage.NewUserRepository(pool)

	ctx := context.Background()

	// Create a test user first
	user := createTestUser(t, pool)

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr bool
	}{
		{
			name:    "existing user",
			id:      user.ID,
			wantErr: false,
		},
		{
			name:    "non-existing user",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			found, err := repo.GetByID(ctx, tt.id)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, found)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.id, found.ID)
			}
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	t.Parallel()

	pool := connectDB(t)
	repo := storage.NewUserRepository(pool)

	ctx := context.Background()

	// Create a test user first
	user := createTestUser(t, pool)

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "existing email",
			email:   user.Email,
			wantErr: false,
		},
		{
			name:    "non-existing email",
			email:   "nonexistent@example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			found, err := repo.GetByEmail(ctx, tt.email)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, found)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.email, found.Email)
			}
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	t.Parallel()

	pool := connectDB(t)
	repo := storage.NewUserRepository(pool)

	ctx := context.Background()

	// Create a test user first
	user := createTestUser(t, pool)

	// Update user
	user.Name = "Updated Name"
	updated, err := repo.Update(ctx, user)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.True(t, updated.UpdatedAt.After(user.CreatedAt))

	// Verify in database
	found, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", found.Name)
}

func TestUserRepository_Delete(t *testing.T) {
	t.Parallel()

	pool := connectDB(t)
	repo := storage.NewUserRepository(pool)

	ctx := context.Background()

	// Create a test user first
	user := createTestUser(t, pool)

	// Delete user
	err := repo.Delete(ctx, user.ID)
	require.NoError(t, err)

	// Verify user is deleted
	found, err := repo.GetByID(ctx, user.ID)
	require.Error(t, err)
	assert.Nil(t, found)
}

func TestUserRepository_List(t *testing.T) {
	pool := connectDB(t) // Not parallel - modifies shared state
	repo := storage.NewUserRepository(pool)

	ctx := context.Background()

	// Clean up table before test
	truncateTable(t, pool, "users")

	// Create multiple test users
	for i := 0; i < 5; i++ {
		createTestUser(t, pool)
	}

	// List users with pagination
	users, err := repo.List(ctx, 10, 0)
	require.NoError(t, err)
	assert.Len(t, users, 5)

	// Test offset
	users, err = repo.List(ctx, 10, 3)
	require.NoError(t, err)
	assert.Len(t, users, 2)
}

// createTestUser creates a user in the database for testing.
func createTestUser(t *testing.T, pool any) *models.User {
	t.Helper()

	// Type assertion to get the actual pool type
	// In real code, pool would be *pgxpool.Pool
	repo := storage.NewUserRepository(pool)

	ctx := context.Background()
	user := &models.User{
		Name:  "Test User",
		Email: fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8]),
	}

	created, err := repo.Create(ctx, user)
	require.NoError(t, err)

	return created
}
