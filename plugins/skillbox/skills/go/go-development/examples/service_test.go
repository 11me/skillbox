package services_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"myapp/internal/common"
	"myapp/internal/models"
	"myapp/internal/services"
)

type meta struct {
	name    string
	enabled bool
}

type fields struct {
	setupDatabase func(dbMock sqlmock.Sqlmock)
}

type args struct {
	ctx   context.Context
	name  string
	email string
}

type wants struct {
	user *models.User
	err  error
}

func TestUserService_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		meta   meta
		fields fields
		args   args
		wants  wants
	}{
		{
			meta: meta{name: "success", enabled: true},
			fields: fields{
				setupDatabase: func(dbMock sqlmock.Sqlmock) {
					// Check email doesn't exist
					dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, email`)).
						WithArgs("test@example.com").
						WillReturnError(common.EntityNotFound("user not found"))

					// Begin transaction
					dbMock.ExpectBegin()

					// Insert user
					dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO users`)).
						WithArgs(sqlmock.AnyArg(), "Test User", "test@example.com", sqlmock.AnyArg(), sqlmock.AnyArg()).
						WillReturnResult(sqlmock.NewResult(1, 1))

					dbMock.ExpectCommit()
				},
			},
			args: args{
				ctx:   context.Background(),
				name:  "Test User",
				email: "test@example.com",
			},
			wants: wants{
				user: &models.User{Name: "Test User", Email: "test@example.com"},
				err:  nil,
			},
		},
		{
			meta: meta{name: "validation error - empty name", enabled: true},
			fields: fields{
				setupDatabase: func(dbMock sqlmock.Sqlmock) {},
			},
			args: args{
				ctx:   context.Background(),
				name:  "",
				email: "test@example.com",
			},
			wants: wants{
				user: nil,
				err:  common.ValidationFailed("name is required"),
			},
		},
		{
			meta: meta{name: "validation error - empty email", enabled: true},
			fields: fields{
				setupDatabase: func(dbMock sqlmock.Sqlmock) {},
			},
			args: args{
				ctx:   context.Background(),
				name:  "Test User",
				email: "",
			},
			wants: wants{
				user: nil,
				err:  common.ValidationFailed("email is required"),
			},
		},
		{
			meta: meta{name: "conflict - email exists", enabled: true},
			fields: fields{
				setupDatabase: func(dbMock sqlmock.Sqlmock) {
					dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, email`)).
						WithArgs("existing@example.com").
						WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "created_at", "updated_at"}).
							AddRow(uuid.New(), "Existing", "existing@example.com", nil, nil))
				},
			},
			args: args{
				ctx:   context.Background(),
				name:  "New User",
				email: "existing@example.com",
			},
			wants: wants{
				user: nil,
				err:  common.StateConflict("user with this email already exists"),
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Capture for parallel
		t.Run(tt.meta.name, func(t *testing.T) {
			t.Parallel()
			if !tt.meta.enabled {
				t.SkipNow()
			}

			db, dbMock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.fields.setupDatabase(dbMock)

			storage := NewMockStorage(db)
			svc := services.NewUserService(storage, nil)

			user, err := svc.Create(tt.args.ctx, tt.args.name, tt.args.email)

			if tt.wants.err != nil {
				assert.Equal(t, tt.wants.err, err)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, tt.wants.user.Name, user.Name)
				assert.Equal(t, tt.wants.user.Email, user.Email)
				assert.NotEqual(t, uuid.Nil, user.ID)
			}

			assert.NoError(t, dbMock.ExpectationsWereMet())
		})
	}
}

// MockStorage implements storage.Storage for testing
type MockStorage struct {
	db *sql.DB
}

func NewMockStorage(db *sql.DB) *MockStorage {
	return &MockStorage{db: db}
}

// Implement Storage interface methods...
