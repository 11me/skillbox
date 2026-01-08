// Package examples demonstrates the Mapper pattern for Go repositories.
//
// Mappers bridge domain models and database columns, handling:
//   - Complex types (Money → amount + currency columns)
//   - Encryption/decryption of sensitive fields
//   - Type conversions between domain and storage layers
package examples

import (
	"time"
)

// MoneyAmount represents a decimal amount as a string.
type MoneyAmount string

// Currency represents a currency code.
type Currency string

// Money represents a monetary value with currency.
type Money struct {
	Amount   MoneyAmount `json:"amount"`
	Currency Currency    `json:"currency"`
}

// User is the domain model.
type User struct {
	ID        string
	Name      string
	Email     string
	Balance   *Money
	CreatedAt time.Time
}

// userMapper maps between User domain model and database columns.
// All fields are pointers to handle NULL values and partial updates.
type userMapper struct {
	id              *string
	name            *string
	email           *string
	balanceAmount   *string
	balanceCurrency *string
	createdAt       *time.Time
}

// NewUserMapper creates a mapper from a domain model.
// Pass nil to create an empty mapper for scanning.
func NewUserMapper(u *User) *userMapper {
	if u == nil {
		return &userMapper{}
	}

	m := &userMapper{
		id:        &u.ID,
		name:      &u.Name,
		email:     &u.Email,
		createdAt: &u.CreatedAt,
	}

	if u.Balance != nil {
		amount := string(u.Balance.Amount)
		currency := string(u.Balance.Currency)
		m.balanceAmount = &amount
		m.balanceCurrency = &currency
	}

	return m
}

// UserColumns returns column names in consistent order.
func UserColumns() []string {
	return []string{
		"id",
		"name",
		"email",
		"balance_amount",
		"balance_currency",
		"created_at",
	}
}

// Values returns values for INSERT statements in column order.
func (m *userMapper) Values() []any {
	return []any{
		m.id,
		m.name,
		m.email,
		m.balanceAmount,
		m.balanceCurrency,
		m.createdAt,
	}
}

// ScanValues returns pointers for scanning SELECT results.
func (m *userMapper) ScanValues() []any {
	return []any{
		&m.id,
		&m.name,
		&m.email,
		&m.balanceAmount,
		&m.balanceCurrency,
		&m.createdAt,
	}
}

// ToModel converts the mapper back to a domain model.
func (m *userMapper) ToModel() *User {
	if m.IsEmpty() {
		return nil
	}

	user := &User{}

	if m.id != nil {
		user.ID = *m.id
	}
	if m.name != nil {
		user.Name = *m.name
	}
	if m.email != nil {
		user.Email = *m.email
	}
	if m.createdAt != nil {
		user.CreatedAt = *m.createdAt
	}

	if m.balanceAmount != nil && m.balanceCurrency != nil {
		user.Balance = &Money{
			Amount:   MoneyAmount(*m.balanceAmount),
			Currency: Currency(*m.balanceCurrency),
		}
	}

	return user
}

// IsEmpty checks if the mapper has no data.
func (m *userMapper) IsEmpty() bool {
	return m.id == nil && m.name == nil && m.email == nil
}

// UserUpdate represents a partial update request.
type UserUpdate struct {
	Name    *string
	Email   *string
	Balance *Money
}

// userUpdateMapper handles partial updates.
type userUpdateMapper struct {
	name            *string
	email           *string
	balanceAmount   *string
	balanceCurrency *string
}

// NewUserUpdateMapper creates a mapper for partial updates.
func NewUserUpdateMapper(u *UserUpdate) *userUpdateMapper {
	if u == nil {
		return &userUpdateMapper{}
	}

	m := &userUpdateMapper{
		name:  u.Name,
		email: u.Email,
	}

	if u.Balance != nil {
		amount := string(u.Balance.Amount)
		currency := string(u.Balance.Currency)
		m.balanceAmount = &amount
		m.balanceCurrency = &currency
	}

	return m
}

// UpdateFields returns field-value pairs for UPDATE SET clause.
func (m *userUpdateMapper) UpdateFields() map[string]any {
	fields := make(map[string]any)

	if m.name != nil {
		fields["name"] = *m.name
	}
	if m.email != nil {
		fields["email"] = *m.email
	}
	if m.balanceAmount != nil {
		fields["balance_amount"] = *m.balanceAmount
	}
	if m.balanceCurrency != nil {
		fields["balance_currency"] = *m.balanceCurrency
	}

	return fields
}

// HasChanges checks if there are any fields to update.
func (m *userUpdateMapper) HasChanges() bool {
	return m.name != nil || m.email != nil ||
		m.balanceAmount != nil || m.balanceCurrency != nil
}

// Encryptor interface for sensitive data handling.
type Encryptor interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
	Hash(plaintext string) string
}

// secureUserMapper handles encrypted fields.
type secureUserMapper struct {
	id        *string
	name      *string
	email     *string
	emailHash *string
	encryptor Encryptor
}

// NewSecureUserMapper creates a mapper with encryption support.
func NewSecureUserMapper(u *User, enc Encryptor) *secureUserMapper {
	m := &secureUserMapper{encryptor: enc}
	if u == nil {
		return m
	}

	m.id = &u.ID
	m.name = &u.Name

	if u.Email != "" && enc != nil {
		encrypted, err := enc.Encrypt(u.Email)
		if err == nil {
			m.email = &encrypted
			hash := enc.Hash(u.Email)
			m.emailHash = &hash
		}
	}

	return m
}

// ToModel decrypts fields when converting back to domain model.
func (m *secureUserMapper) ToModel() *User {
	if m.id == nil {
		return nil
	}

	user := &User{}
	if m.id != nil {
		user.ID = *m.id
	}
	if m.name != nil {
		user.Name = *m.name
	}

	if m.email != nil && m.encryptor != nil {
		decrypted, err := m.encryptor.Decrypt(*m.email)
		if err == nil {
			user.Email = decrypted
		}
	}

	return user
}

// Usage:
//
//	// Save inserts or updates (upsert pattern).
//	func (r *userRepo) Save(ctx context.Context, user *User) error {
//	    m := NewUserMapper(user)
//
//	    sql, args, err := sq.Insert("users").
//	        Columns(UserColumns()...).
//	        Values(m.Values()...).
//	        Suffix(`ON CONFLICT (id) DO UPDATE SET
//	            name = EXCLUDED.name,
//	            email = EXCLUDED.email,
//	            balance_amount = EXCLUDED.balance_amount,
//	            balance_currency = EXCLUDED.balance_currency`).
//	        PlaceholderFormat(sq.Dollar).
//	        ToSql()
//	    if err != nil {
//	        return err
//	    }
//
//	    _, err = r.db.Exec(ctx, sql, args...)
//	    return err
//	}
//
//	// FindByID returns a user by ID.
//	func (r *userRepo) FindByID(ctx context.Context, id string) (*User, error) {
//	    sql, args, _ := sq.Select(UserColumns()...).
//	        From("users").
//	        Where(sq.Eq{"id": id}).
//	        PlaceholderFormat(sq.Dollar).
//	        ToSql()
//
//	    m := NewUserMapper(nil) // empty for scanning
//	    err := r.db.QueryRow(ctx, sql, args...).Scan(m.ScanValues()...)
//	    if err != nil {
//	        return nil, err
//	    }
//
//	    return m.ToModel(), nil
//	}
//
//	// PartialUpdate updates only specified fields.
//	func (r *userRepo) PartialUpdate(ctx context.Context, id string, update *UserUpdate) error {
//	    m := NewUserUpdateMapper(update)
//	    if !m.HasChanges() {
//	        return nil
//	    }
//
//	    qb := sq.Update("users").
//	        Where(sq.Eq{"id": id}).
//	        PlaceholderFormat(sq.Dollar)
//
//	    for col, val := range m.UpdateFields() {
//	        qb = qb.Set(col, val)
//	    }
//
//	    sql, args, _ := qb.ToSql()
//	    _, err := r.db.Exec(ctx, sql, args...)
//	    return err
//	}
