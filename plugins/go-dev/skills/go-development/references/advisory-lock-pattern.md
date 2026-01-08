# Advisory Lock Pattern

PostgreSQL advisory locks for preventing serializable transaction conflicts.

## Problem

PostgreSQL SERIALIZABLE isolation can fail with `SQLSTATE 40001` (serialization_failure) when concurrent transactions modify overlapping data. This requires application-level retry logic.

```
ERROR: could not serialize access due to concurrent update
SQLSTATE: 40001
```

## Why SERIALIZABLE Alone Isn't Enough

### The Mixed Isolation Problem

PostgreSQL's SERIALIZABLE isolation (SSI) works by building a **dependency graph** between transactions and detecting "dangerous structures" — cycles of rw-conflicts (read-write conflicts).

**Critical limitation:** SSI only tracks transactions that participate in the protocol.

#### How SSI Detects Conflicts

1. **SERIALIZABLE transactions** leave **predicate locks (SIREAD locks)** when reading
2. These locks are "markers" in memory: *"Transaction T1 read rows matching condition X"*
3. When another transaction modifies those rows, PostgreSQL detects the conflict

#### Why READ COMMITTED Breaks SSI

**READ COMMITTED transactions don't leave SIREAD locks.** They're invisible to SSI for reads.

**Write Skew Example:**
- Constraint: `balance(A) + balance(B) > 0`
- T1 (SERIALIZABLE): Reads A, B → decreases A. Leaves SIREAD lock.
- T2 (READ COMMITTED): Reads A, B → decreases B. **No SIREAD lock.**

**What PostgreSQL sees:**
- T1→T2: T1 read B, T2 modified B → Conflict detected ✓
- T2→T1: T2 read A, T1 modified A → **No detection** ✗ (T2 left no marker)

**Result:** Both commit, constraint violated, data corrupted.

### The Real-World Problem

Most production systems have:
- Legacy code using READ COMMITTED (default in PostgreSQL)
- New code using SERIALIZABLE for critical operations
- Background jobs, migrations, admin scripts — often READ COMMITTED

**Egor Rogov's verdict:** *"If you want SERIALIZABLE guarantees, ALL transactions touching logically related data must be SERIALIZABLE."*

### Advisory Locks as Universal Protection

Advisory locks work at a **lower level** than SSI — they're true exclusive locks that block all transactions regardless of isolation level.

| Protection | SERIALIZABLE only | Advisory Lock + SERIALIZABLE |
|------------|-------------------|------------------------------|
| Against other SERIALIZABLE | ✓ | ✓ |
| Against READ COMMITTED writes | ✗ | ✓ |
| Against legacy code | ✗ | ✓ |

**This is why we use advisory locks:** They provide protection even in mixed-isolation environments.

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
| Mixed isolation level environment | Yes — essential |
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
