-- Example: testmigration/100001_users_dataset.up.sql
-- =====================================================
-- Test data fixtures for repository integration tests.
-- Place in: internal/storage/testmigration/
--
-- NAMING CONVENTION:
--   {number}_{entity}_{description}.{up|down}.sql
--   - 100001-199999: Reserved for test data
--   - Use @test.local domain for emails
--   - Use deterministic UUIDs for reproducibility
--
-- GOOSE TABLE:
--   Test fixtures use separate table: goose_db_test_version
--   (configured in applyTestData function)

-- +goose Up

-- Users with various statuses and roles
INSERT INTO users (id, name, email, status, role, created_at) VALUES
    -- Active users
    ('11111111-1111-1111-1111-111111111111', 'Alice', 'alice@test.local', 'active', 'user', NOW()),
    ('22222222-2222-2222-2222-222222222222', 'Bob', 'bob@test.local', 'active', 'user', NOW()),
    -- Inactive user (for testing status filters)
    ('33333333-3333-3333-3333-333333333333', 'Inactive User', 'inactive@test.local', 'inactive', 'user', NOW()),
    -- Admin user (for testing role-based logic)
    ('44444444-4444-4444-4444-444444444444', 'Admin', 'admin@test.local', 'active', 'admin', NOW());

-- +goose Down

-- Clean up test data
-- Use @test.local domain to easily identify test records
DELETE FROM users WHERE email LIKE '%@test.local';


-- =====================================================
-- Example: testmigration/100002_orders_dataset.up.sql
-- =====================================================
-- Depends on: 100001_users_dataset

-- +goose Up

-- Orders for testing relationships and aggregations
INSERT INTO orders (id, user_id, status, total, created_at) VALUES
    -- Alice has 2 orders (for testing GetUserOrders)
    ('aaaa1111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', 'pending', 150.00, NOW()),
    ('bbbb2222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 'completed', 250.00, NOW() - INTERVAL '1 day'),
    -- Bob has 1 cancelled order
    ('cccc3333-3333-3333-3333-333333333333', '22222222-2222-2222-2222-222222222222', 'cancelled', 75.50, NOW() - INTERVAL '2 days');
    -- Note: Inactive user (33333333...) has NO orders
    -- This is intentional for testing edge cases

-- Order items for testing nested relationships
INSERT INTO order_items (id, order_id, product_id, quantity, price) VALUES
    -- Alice's pending order items
    ('item-1111-1111-1111-111111111111', 'aaaa1111-1111-1111-1111-111111111111', 'prod-1111', 2, 50.00),
    ('item-2222-2222-2222-222222222222', 'aaaa1111-1111-1111-1111-111111111111', 'prod-2222', 1, 50.00),
    -- Alice's completed order items
    ('item-3333-3333-3333-333333333333', 'bbbb2222-2222-2222-2222-222222222222', 'prod-1111', 5, 50.00);

-- +goose Down

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
