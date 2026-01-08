# Mapper Pattern

Bridging domain models and database representation with explicit mapping.

## Problem

Domain models often don't map 1:1 to database columns:

```go
// Domain model
type User struct {
    ID      string
    Name    string
    Balance *Money  // Complex type → 2 columns (amount, currency)
}

// Database: users (id, name, balance_amount, balance_currency)
```

## Solution

Create a mapper struct that handles the transformation:

```go
type userMapper struct {
    id              *string
    name            *string
    balanceAmount   *string    // Money.Amount → string column
    balanceCurrency *string    // Money.Currency → string column
    createdAt       *time.Time
}
```

## Core Mapper Interface

```go
// Mapper defines the contract for model-database mapping.
type Mapper[T any] interface {
    Values() []any       // For INSERT — returns values in column order
    ScanValues() []any   // For SELECT — returns pointers for scanning
    ToModel() *T         // Converts mapper back to domain model
    IsEmpty() bool       // Checks if all fields are nil/zero
}
```

## Complete Mapper Implementation

```go
package storage

import (
    "time"

    "myapp/internal/models"
)

// userMapper maps between User domain model and database columns.
type userMapper struct {
    id              *string
    name            *string
    email           *string
    balanceAmount   *string
    balanceCurrency *string
    createdAt       *time.Time
}

// NewUserMapper creates mapper from domain model.
func NewUserMapper(u *models.User) *userMapper {
    if u == nil {
        return &userMapper{}
    }

    m := &userMapper{
        id:        &u.ID,
        name:      &u.Name,
        email:     &u.Email,
        createdAt: &u.CreatedAt,
    }

    // Handle Money type — split into two columns
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
    return []string{"id", "name", "email", "balance_amount", "balance_currency", "created_at"}
}

// Values returns values for INSERT in column order.
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

// ScanValues returns pointers for SELECT scanning.
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

// ToModel converts mapper back to domain model.
func (m *userMapper) ToModel() *models.User {
    if m.IsEmpty() {
        return nil
    }

    user := &models.User{}

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

    // Reconstruct Money from split columns
    if m.balanceAmount != nil && m.balanceCurrency != nil {
        user.Balance = &models.Money{
            Amount:   models.MoneyAmount(*m.balanceAmount),
            Currency: models.Currency(*m.balanceCurrency),
        }
    }

    return user
}

// IsEmpty checks if mapper has no data.
func (m *userMapper) IsEmpty() bool {
    return m.id == nil && m.name == nil && m.email == nil
}
```

## Repository Usage

### Save (Upsert)

```go
func (r *userRepo) Save(ctx context.Context, user *models.User) error {
    m := NewUserMapper(user)

    sql, args, err := sq.Insert("users").
        Columns(UserColumns()...).
        Values(m.Values()...).
        Suffix(`ON CONFLICT (id) DO UPDATE SET
            name = EXCLUDED.name,
            email = EXCLUDED.email,
            balance_amount = EXCLUDED.balance_amount,
            balance_currency = EXCLUDED.balance_currency`).
        PlaceholderFormat(sq.Dollar).
        ToSql()
    if err != nil {
        return err
    }

    _, err = r.db.Exec(ctx, sql, args...)
    return err
}
```

### SELECT Single

```go
func (r *userRepo) FindByID(ctx context.Context, id string) (*models.User, error) {
    query := sq.Select(UserColumns()...).
        From("users").
        Where(sq.Eq{"id": id}).
        PlaceholderFormat(sq.Dollar)

    sql, args, err := query.ToSql()
    if err != nil {
        return nil, err
    }

    m := &userMapper{}
    err = r.db.QueryRow(ctx, sql, args...).Scan(m.ScanValues()...)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, ErrUserNotFound
        }
        return nil, err
    }

    return m.ToModel(), nil
}
```

### SELECT Multiple

```go
func (r *userRepo) Find(ctx context.Context, filter *UserFilter) ([]*models.User, error) {
    query := sq.Select(UserColumns()...).
        From("users").
        PlaceholderFormat(sq.Dollar)

    if filter.Name != nil {
        query = query.Where(sq.Eq{"name": *filter.Name})
    }

    sql, args, err := query.ToSql()
    if err != nil {
        return nil, err
    }

    rows, err := r.db.Query(ctx, sql, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []*models.User
    for rows.Next() {
        m := &userMapper{}
        if err := rows.Scan(m.ScanValues()...); err != nil {
            return nil, err
        }
        users = append(users, m.ToModel())
    }

    return users, rows.Err()
}
```

### Partial Update

Use separate update mapper for partial updates when you need to update only specific fields:

```go
// userUpdateMapper handles partial updates.
type userUpdateMapper struct {
    name            *string
    email           *string
    balanceAmount   *string
    balanceCurrency *string
}

func (m *userUpdateMapper) UpdateFields() map[string]any {
    fields := make(map[string]any)
    if m.name != nil {
        fields["name"] = *m.name
    }
    if m.email != nil {
        fields["email"] = *m.email
    }
    // ... other fields
    return fields
}

func (r *userRepo) PartialUpdate(ctx context.Context, id string, update *UserUpdate) error {
    m := NewUserUpdateMapper(update)
    if !m.HasChanges() {
        return nil
    }

    qb := sq.Update("users").
        Where(sq.Eq{"id": id}).
        PlaceholderFormat(sq.Dollar)

    for col, val := range m.UpdateFields() {
        qb = qb.Set(col, val)
    }

    sql, args, _ := qb.ToSql()
    _, err := r.db.Exec(ctx, sql, args...)
    return err
}
```

## Mapper with Encryption

For sensitive data that needs encryption at rest:

```go
type userMapper struct {
    id         *string
    name       *string
    email      *string  // plain
    emailHash  *string  // encrypted for search
    ssn        *string  // encrypted
    encryptor  Encryptor
}

func NewUserMapperWithEncryption(u *models.User, enc Encryptor) *userMapper {
    m := &userMapper{encryptor: enc}
    if u == nil {
        return m
    }

    m.id = &u.ID
    m.name = &u.Name

    // Encrypt sensitive fields
    if u.Email != "" {
        encrypted, _ := enc.Encrypt(u.Email)
        m.email = &encrypted
        hash := enc.Hash(u.Email)
        m.emailHash = &hash
    }

    if u.SSN != "" {
        encrypted, _ := enc.Encrypt(u.SSN)
        m.ssn = &encrypted
    }

    return m
}

func (m *userMapper) ToModel() *models.User {
    if m.IsEmpty() {
        return nil
    }

    user := &models.User{}
    if m.id != nil {
        user.ID = *m.id
    }
    if m.name != nil {
        user.Name = *m.name
    }

    // Decrypt sensitive fields
    if m.email != nil {
        decrypted, _ := m.encryptor.Decrypt(*m.email)
        user.Email = decrypted
    }
    if m.ssn != nil {
        decrypted, _ := m.encryptor.Decrypt(*m.ssn)
        user.SSN = decrypted
    }

    return user
}
```

## Mapper with pgx.CollectRows

Alternative approach using pgx's built-in collection:

```go
func (r *userRepo) Find(ctx context.Context, filter *UserFilter) ([]*models.User, error) {
    sql, args, err := buildQuery(filter)
    if err != nil {
        return nil, err
    }

    rows, err := r.db.Query(ctx, sql, args...)
    if err != nil {
        return nil, err
    }

    // Collect into mappers
    mappers, err := pgx.CollectRows(rows, pgx.RowToStructByName[userMapper])
    if err != nil {
        return nil, err
    }

    // Convert to domain models
    users := make([]*models.User, 0, len(mappers))
    for _, m := range mappers {
        users = append(users, m.ToModel())
    }

    return users, nil
}
```

## When to Use Mappers

| Scenario | Use Mapper? |
|----------|-------------|
| Simple 1:1 field mapping | No — use `pgx.RowToStructByName` |
| Complex types (Money, Address) | Yes |
| Encrypted fields | Yes |
| Computed fields | Yes |
| Different naming conventions | Maybe — consider struct tags first |
| Multiple representations | Yes |

## File Organization

```
internal/
├── models/
│   ├── user.go           # Domain model
│   └── money.go          # Value objects
└── storage/
    ├── user.go           # Repository
    └── user_mapper.go    # Mapper (or keep in user.go if small)
```

## Best Practices

### DO:
- Keep mappers close to repositories (same package)
- Use pointers for nullable fields
- Provide `Columns()` function for consistent ordering
- Handle nil/empty cases explicitly
- Use `IsEmpty()` to detect uninitialized mappers

### DON'T:
- Don't expose mappers outside storage package
- Don't add business logic to mappers
- Don't skip encryption/decryption steps
- Don't mix mapper responsibilities (one mapper per entity)

## Related

- [repository-pattern.md](repository-pattern.md) — Repository pattern
- [money-pattern.md](money-pattern.md) — Money handling
- [database-pattern.md](database-pattern.md) — Database client
