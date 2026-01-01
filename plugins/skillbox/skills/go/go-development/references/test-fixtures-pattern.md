# Test Fixtures Pattern

SQL-based test data management with goose for integration tests.

## When to Use

| Approach | Use When |
|----------|----------|
| **SQL Fixtures** | Repository tests, complex relationships, deterministic IDs |
| **Helper Functions** | Service/Handler tests with mocks, unique data per test |
| **Hybrid** | SQL for base data, helpers for test-specific mutations |

## Directory Structure

```
internal/storage/
├── main_test.go           # TestMain with testcontainers
├── user_test.go
├── order_test.go
└── testmigration/         # Test data fixtures
    ├── 100001_users.up.sql
    ├── 100001_users.down.sql
    ├── 100002_orders.up.sql
    └── 100002_orders.down.sql
```

## Naming Convention

### Migration Numbers

| Range | Purpose | Goose Table |
|-------|---------|-------------|
| 000001-099999 | Production schema | `goose_db_version` |
| 100001-199999 | Test data fixtures | `goose_db_test_version` |
| 200001-299999 | Performance test data | `goose_db_test_version` |

### File Names

```
{number}_{entity}_{description}.{up|down}.sql
```

Examples:
- `100001_users.up.sql` — base user fixtures
- `100002_orders_with_items.up.sql` — orders with relationships
- `100010_large_dataset.up.sql` — performance testing

## Deterministic UUIDs

Use predictable UUIDs for test data:

```sql
-- Users
'11111111-1111-1111-1111-111111111111' -- Alice (active)
'22222222-2222-2222-2222-222222222222' -- Bob (active)
'33333333-3333-3333-3333-333333333333' -- Inactive user
'44444444-4444-4444-4444-444444444444' -- Admin user

-- Orders
'aaaa1111-1111-1111-1111-111111111111' -- Pending order (Alice)
'bbbb2222-2222-2222-2222-222222222222' -- Completed order (Alice)
'cccc3333-3333-3333-3333-333333333333' -- Cancelled order (Bob)

-- Products
'dddd1111-1111-1111-1111-111111111111' -- Product A
'eeee2222-2222-2222-2222-222222222222' -- Product B
```

**Why deterministic IDs:**
- Tests are reproducible
- Easy to reference in test cases
- Clear documentation of test scenarios

## Example Fixtures

### Users Dataset

```sql
-- testmigration/100001_users.up.sql
INSERT INTO users (id, name, email, status, role, created_at) VALUES
    ('11111111-1111-1111-1111-111111111111', 'Alice', 'alice@test.local', 'active', 'user', NOW()),
    ('22222222-2222-2222-2222-222222222222', 'Bob', 'bob@test.local', 'active', 'user', NOW()),
    ('33333333-3333-3333-3333-333333333333', 'Inactive', 'inactive@test.local', 'inactive', 'user', NOW()),
    ('44444444-4444-4444-4444-444444444444', 'Admin', 'admin@test.local', 'active', 'admin', NOW());
```

```sql
-- testmigration/100001_users.down.sql
DELETE FROM users WHERE email LIKE '%@test.local';
```

### Orders with Relationships

```sql
-- testmigration/100002_orders.up.sql
-- Depends on: 100001_users

INSERT INTO orders (id, user_id, status, total, created_at) VALUES
    -- Alice has 2 orders
    ('aaaa1111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', 'pending', 150.00, NOW()),
    ('bbbb2222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 'completed', 250.00, NOW() - INTERVAL '1 day'),
    -- Bob has 1 cancelled order
    ('cccc3333-3333-3333-3333-333333333333', '22222222-2222-2222-2222-222222222222', 'cancelled', 75.50, NOW() - INTERVAL '2 days');

INSERT INTO order_items (id, order_id, product_id, quantity, price) VALUES
    ('item-1111-1111-1111-111111111111', 'aaaa1111-1111-1111-1111-111111111111', 'dddd1111-1111-1111-1111-111111111111', 2, 50.00),
    ('item-2222-2222-2222-222222222222', 'aaaa1111-1111-1111-1111-111111111111', 'eeee2222-2222-2222-2222-222222222222', 1, 50.00),
    ('item-3333-3333-3333-333333333333', 'bbbb2222-2222-2222-2222-222222222222', 'dddd1111-1111-1111-1111-111111111111', 5, 50.00);
```

```sql
-- testmigration/100002_orders.down.sql
DELETE FROM order_items WHERE order_id IN (
    'aaaa1111-1111-1111-1111-111111111111',
    'bbbb2222-2222-2222-2222-222222222222',
    'cccc3333-3333-3333-3333-333333333333'
);
DELETE FROM orders WHERE id IN (
    'aaaa1111-1111-1111-1111-111111111111',
    'bbbb2222-2222-2222-2222-222222222222',
    'cccc3333-3333-3333-3333-333333333333'
);
```

## Test Scenarios

### Known Data Queries

```go
func TestUserRepository_GetByID(t *testing.T) {
    t.Parallel()

    pool := connectDB(t)
    repo := storage.NewUserRepository(pool)

    tests := []struct {
        name     string
        id       string
        wantName string
        wantErr  bool
    }{
        {name: "alice", id: "11111111-1111-1111-1111-111111111111", wantName: "Alice"},
        {name: "bob", id: "22222222-2222-2222-2222-222222222222", wantName: "Bob"},
        {name: "not-found", id: "99999999-9999-9999-9999-999999999999", wantErr: true},
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            user, err := repo.GetByID(context.Background(), tt.id)
            if tt.wantErr {
                require.ErrorIs(t, err, errs.ErrNotFound)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.wantName, user.Name)
        })
    }
}
```

### Aggregate Queries

```go
func TestOrderRepository_GetUserOrders(t *testing.T) {
    t.Parallel()

    pool := connectDB(t)
    repo := storage.NewOrderRepository(pool)

    tests := []struct {
        name      string
        userID    string
        wantCount int
    }{
        {name: "alice-2-orders", userID: "11111111-1111-1111-1111-111111111111", wantCount: 2},
        {name: "bob-1-order", userID: "22222222-2222-2222-2222-222222222222", wantCount: 1},
        {name: "inactive-no-orders", userID: "33333333-3333-3333-3333-333333333333", wantCount: 0},
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            orders, err := repo.GetUserOrders(context.Background(), tt.userID)
            require.NoError(t, err)
            assert.Len(t, orders, tt.wantCount)
        })
    }
}
```

### Filter Queries

```go
func TestUserRepository_FindByStatus(t *testing.T) {
    t.Parallel()

    pool := connectDB(t)
    repo := storage.NewUserRepository(pool)

    // Known from fixtures: 3 active, 1 inactive
    active, err := repo.FindByStatus(context.Background(), "active")
    require.NoError(t, err)
    assert.Len(t, active, 3)

    inactive, err := repo.FindByStatus(context.Background(), "inactive")
    require.NoError(t, err)
    assert.Len(t, inactive, 1)
    assert.Equal(t, "33333333-3333-3333-3333-333333333333", inactive[0].ID)
}
```

### Constraint Violations

```go
func TestUserRepository_DeleteWithOrders(t *testing.T) {
    t.Parallel()

    pool := connectDB(t)
    repo := storage.NewUserRepository(pool)

    // Alice has orders - delete should fail
    err := repo.Delete(context.Background(), "11111111-1111-1111-1111-111111111111")
    require.Error(t, err)
    assert.Contains(t, err.Error(), "foreign key")

    // Inactive user has no orders - delete should succeed
    err = repo.Delete(context.Background(), "33333333-3333-3333-3333-333333333333")
    require.NoError(t, err)
}
```

## Cleanup Strategies

### Option 1: Fresh Fixtures Per Suite (Recommended)

Fixtures are applied once in TestMain. Tests should NOT modify shared data.

```go
func TestMain(m *testing.M) {
    // ... setup ...
    applyMigrations(pgConnURL)
    applyTestData(pgConnURL)
    code = m.Run()
}

// Tests only READ fixture data
func TestUserRepository_GetByID(t *testing.T) {
    t.Parallel() // Safe - no data modification
    // ...
}
```

### Option 2: Truncate Before Modification

When test MUST modify data:

```go
func TestUserRepository_Update(t *testing.T) {
    // NOT parallel - modifies shared data
    pool := connectDB(t)

    // Reset to known state
    truncateTable(t, pool, "users")
    seedTestUsers(t, pool) // Re-insert fixtures

    repo := storage.NewUserRepository(pool)
    // ... test update logic ...
}
```

### Option 3: Transaction Rollback

Wrap test in transaction, rollback after:

```go
func TestUserRepository_Update(t *testing.T) {
    pool := connectDB(t)

    tx, err := pool.Begin(context.Background())
    require.NoError(t, err)
    defer tx.Rollback(context.Background())

    repo := storage.NewUserRepository(tx)
    // ... test with tx, changes rolled back automatically ...
}
```

## Best Practices

| Practice | Why |
|----------|-----|
| Deterministic UUIDs | Reproducible tests, easy debugging |
| `@test.local` domain | Easy to identify and clean up test data |
| Separate goose table | Avoid conflicts with schema migrations |
| Document relationships | Make fixture dependencies clear |
| Provide down migrations | Clean rollback support |
| Keep fixtures minimal | Only data needed for tests |

## Anti-Patterns

```
❌ Random UUIDs in fixtures — tests become non-deterministic
❌ Real email domains — could accidentally send emails
❌ Mixing schema and data in one migration — hard to maintain
❌ Huge datasets by default — slows down test suite
❌ Missing down migrations — can't rollback cleanly
❌ Modifying fixtures in parallel tests — race conditions
```
