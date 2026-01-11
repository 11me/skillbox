# Defer Patterns

Guaranteed cleanup with `defer` in Go.

## Core Behavior

- Deferred function runs when enclosing function returns
- Arguments evaluated when `defer` executes, not when function runs
- Multiple defers execute in LIFO order (last-in, first-out)

## Basic Cleanup Pattern

```go
func ReadFile(path string) ([]byte, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()  // Runs before function returns

    return io.ReadAll(f)
}
```

## LIFO Execution Order

```go
func example() {
    defer fmt.Println("first")
    defer fmt.Println("second")
    defer fmt.Println("third")
}
// Output:
// third
// second
// first
```

## Common Use Cases

### File Handling

```go
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()

    // Process file...
    return nil
}
```

### Mutex Unlock

```go
func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.count++
}
```

### Database Transactions

```go
func (s *Store) Transfer(from, to string, amount int) error {
    tx, err := s.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()  // No-op if committed

    // Perform operations...

    return tx.Commit()
}
```

### HTTP Response Body

```go
func fetch(url string) ([]byte, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    return io.ReadAll(resp.Body)
}
```

### Rows Close

```go
func (r *Repo) FindAll(ctx context.Context) ([]*User, error) {
    rows, err := r.db.Query(ctx, "SELECT * FROM users")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []*User
    for rows.Next() {
        // ...
    }
    return users, rows.Err()
}
```

## Arguments Evaluated at Defer Time

```go
func example() {
    x := 10
    defer fmt.Println(x)  // Captures x=10
    x = 20
}
// Output: 10 (not 20)
```

## Defer with Closure

Use closure to capture current value:

```go
func example() {
    x := 10
    defer func() {
        fmt.Println(x)  // Reads x when executed
    }()
    x = 20
}
// Output: 20
```

## Modifying Named Return Values

```go
func double(n int) (result int) {
    defer func() {
        result *= 2  // Modifies return value
    }()
    return n
}

fmt.Println(double(5))  // Output: 10
```

## Panic Recovery

```go
func safeCall(fn func()) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic: %v", r)
        }
    }()
    fn()
    return nil
}
```

## Avoid: Defer in Loop

```go
// ❌ Files stay open until function returns
func processFiles(paths []string) error {
    for _, path := range paths {
        f, err := os.Open(path)
        if err != nil {
            return err
        }
        defer f.Close()  // All files open until loop ends
        // ...
    }
    return nil
}

// ✅ Close immediately or use helper
func processFiles(paths []string) error {
    for _, path := range paths {
        if err := processFile(path); err != nil {
            return err
        }
    }
    return nil
}

func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()
    // ...
    return nil
}
```

## Error Handling in Defer

```go
func WriteFile(path string, data []byte) (err error) {
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer func() {
        closeErr := f.Close()
        if err == nil {
            err = closeErr  // Only capture if no prior error
        }
    }()

    _, err = f.Write(data)
    return err
}
```

## Tracing Pattern

```go
func trace(name string) func() {
    start := time.Now()
    log.Printf("entering %s", name)
    return func() {
        log.Printf("leaving %s (%v)", name, time.Since(start))
    }
}

func doWork() {
    defer trace("doWork")()  // Note: () to call trace immediately
    // ...
}
```

## Quick Reference

| Pattern | Use Case |
|---------|----------|
| `defer f.Close()` | File/connection cleanup |
| `defer mu.Unlock()` | Mutex release |
| `defer tx.Rollback()` | Transaction safety |
| `defer rows.Close()` | Database rows cleanup |
| `defer resp.Body.Close()` | HTTP response cleanup |
| `defer recover()` | Panic handling |

## Rules

| DO | DON'T |
|----|-------|
| Defer immediately after acquiring resource | Defer in loops |
| Use closure to capture changing values | Assume defer captures by reference |
| Handle errors in defer when needed | Ignore Close() errors silently |
| Use for cleanup, not control flow | Use defer for normal logic |
