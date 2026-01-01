# Advisory Lock Pattern

PostgreSQL advisory locks for preventing serializable transaction conflicts.

## Problem

PostgreSQL SERIALIZABLE isolation can fail with `SQLSTATE 40001` (serialization_failure) when concurrent transactions modify overlapping data. This requires application-level retry logic.

```
ERROR: could not serialize access due to concurrent update
SQLSTATE: 40001
```

## Solution

Use `pg_advisory_xact_lock(hashtext(label))` — a transaction-scoped advisory lock:
- Acquires exclusive lock on a hash of the label string
- Automatically releases on commit/rollback
- Prevents concurrent access to the same logical resource

## Repository Interface

Add `Serialize` method to repositories that participate in serializable transactions:

```go
type UserRepository interface {
    // ... CRUD methods ...
    Serialize(ctx context.Context, label string) error
}
```

## Implementation

```go
// Serialize acquires an advisory lock for serializable transactions.
// Use inside WithTx with pgx.Serializable to prevent serialization conflicts.
func (r *userRepository) Serialize(ctx context.Context, label string) error {
    query, _, err := squirrel.
        Select("pg_advisory_xact_lock(hashtext(?))").
        PlaceholderFormat(squirrel.Dollar).
        ToSql()
    if err != nil {
        return fmt.Errorf("build lock query: %w", err)
    }

    if _, err := r.db.Exec(ctx, query, label); err != nil {
        return fmt.Errorf("acquire lock: %w", err)
    }
    return nil
}
```

## Usage Pattern: Lock-First

**Critical:** Always acquire the lock BEFORE reading data.

```go
// In service layer
err := s.db.WithTx(ctx, func(ctx context.Context) error {
    // 1. Acquire advisory lock FIRST
    if err := s.users.Serialize(ctx, fmt.Sprintf("TransferFunds:%s:%s", fromID, toID)); err != nil {
        return fmt.Errorf("serialize: %w", err)
    }

    // 2. Read current state (AFTER lock acquired)
    from, err := s.users.GetByID(ctx, fromID)
    if err != nil {
        return err
    }

    to, err := s.users.GetByID(ctx, toID)
    if err != nil {
        return err
    }

    // 3. Validate
    if from.Balance < amount {
        return errors.New("insufficient funds")
    }

    // 4. Update
    from.Balance -= amount
    to.Balance += amount

    if err := s.users.Update(ctx, from); err != nil {
        return err
    }
    if err := s.users.Update(ctx, to); err != nil {
        return err
    }

    return nil
}, pgx.Serializable)
```

## Lock Key Naming Convention

| Pattern | Example | Use Case |
|---------|---------|----------|
| `Operation:resourceID` | `"TransferFunds:user123"` | Single resource lock |
| `Operation:res1:res2` | `"Transfer:acc1:acc2"` | Multi-resource lock |
| `Namespace` | `"CardTokenSequence"` | Global sequence lock |

### Best Practices

- Use descriptive, unique labels
- Include operation name and resource IDs
- Keep labels reasonably short (hashtext has no limit, but readability matters)

```go
// Good
label := fmt.Sprintf("ProcessPayment:%s", paymentID)
label := fmt.Sprintf("UpdateBalance:%s:%s", fromID, toID)

// Bad
label := "lock"  // Too generic, conflicts with other operations
label := paymentID  // Missing operation context
```

## How It Works

1. `hashtext(label)` converts the string to a deterministic int64 hash
2. `pg_advisory_xact_lock(hash)` acquires an exclusive lock on that hash
3. Other transactions calling the same label will **block** until the lock is released
4. Lock is automatically released when transaction commits or rolls back

```sql
-- This is what the database executes
SELECT pg_advisory_xact_lock(hashtext('TransferFunds:user123'));
```

## When to Use

| Scenario | Use Advisory Lock? |
|----------|-------------------|
| Wallet balance updates | Yes |
| Counter increments (sequences) | Yes |
| Auction bidding | Yes |
| Simple CRUD without concurrency | No |
| Read-only queries | No |
| Transactions with natural row-level locks | Maybe |

## Anti-Patterns

### Don't: Lock after reading

```go
// WRONG: Race condition!
err := s.db.WithTx(ctx, func(ctx context.Context) error {
    user, err := s.users.GetByID(ctx, userID)  // Read first
    if err != nil {
        return err
    }

    s.users.Serialize(ctx, fmt.Sprintf("Update:%s", userID))  // Lock after read = WRONG

    user.Balance += 100
    return s.users.Update(ctx, user)
}, pgx.Serializable)
```

### Don't: Use generic labels

```go
// WRONG: All operations will serialize
s.users.Serialize(ctx, "global")  // Don't do this
```

### Don't: Skip the lock and rely only on retries

```go
// Suboptimal: Works, but wastes resources on conflicts
err := s.db.WithTx(ctx, func(ctx context.Context) error {
    // No Serialize() call
    // Transaction may fail and retry multiple times
}, pgx.Serializable)
```

## Integration with Retry Logic

The existing `isRetryable()` in `database-pattern.md` already handles SQLSTATE 40:

```go
func isRetryable(err error) bool {
    s := err.Error()
    return strings.Contains(s, "SQLSTATE 40") || // Transaction rollback
           strings.Contains(s, "SQLSTATE 08")    // Connection exception
}
```

Advisory locks reduce the frequency of retries by preventing conflicts upfront.

## Related Patterns

- [Database Pattern](database-pattern.md) — Transaction handling with `WithTx`
- [Repository Pattern](repository-pattern.md) — Repository interface design

## PostgreSQL Advisory Lock Types

| Function | Scope | Use Case |
|----------|-------|----------|
| `pg_advisory_xact_lock(key)` | Transaction | **Recommended** — auto-release |
| `pg_advisory_lock(key)` | Session | Manual release required |
| `pg_try_advisory_xact_lock(key)` | Transaction | Non-blocking, returns bool |

We use `pg_advisory_xact_lock` because it automatically releases on commit/rollback.
