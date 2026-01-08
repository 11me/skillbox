# Blank Identifier

Using `_` in Go for intentional ignoring.

## Purpose

The blank identifier `_` discards values without creating unused variables.

## Ignore Return Values

```go
// Ignore error (use sparingly!)
data, _ := io.ReadAll(resp.Body)

// Ignore value, keep error
_, err := io.Copy(dst, src)
if err != nil {
    return err
}

// Ignore index in range
for _, value := range slice {
    sum += value
}

// Ignore value in range
for key := range m {
    keys = append(keys, key)
}
```

## Import for Side Effects

Import package only for its `init()` function:

```go
import (
    "database/sql"

    _ "github.com/lib/pq"           // PostgreSQL driver
    _ "github.com/go-sql-driver/mysql" // MySQL driver
    _ "net/http/pprof"              // Profiling handlers
)
```

## Compile-Time Interface Check

Verify type implements interface at compile time:

```go
// Fails to compile if *MyType doesn't implement Interface
var _ Interface = (*MyType)(nil)

// Common patterns
var _ io.Reader = (*MyReader)(nil)
var _ io.Writer = (*MyWriter)(nil)
var _ json.Marshaler = (*Config)(nil)
var _ fmt.Stringer = (*User)(nil)
var _ error = (*AppError)(nil)
var _ http.Handler = (*Server)(nil)
```

## Multiple Interface Checks

```go
var (
    _ io.Reader = (*Buffer)(nil)
    _ io.Writer = (*Buffer)(nil)
    _ io.Closer = (*Buffer)(nil)
)
```

## Silence Unused Import During Development

Temporary workaround (remove before commit):

```go
import (
    "fmt"
    "log"
    "os"
)

var _ = fmt.Println  // TODO: remove
var _ = log.Println  // TODO: remove

func main() {
    f, err := os.Open("test.txt")
    _ = f   // TODO: use f
    _ = err // TODO: handle err
}
```

## Multiple Assignment

```go
// Check map key existence
_, ok := m[key]
if !ok {
    // key not present
}

// Type assertion check
_, ok := value.(string)
if !ok {
    // not a string
}

// Channel receive with ok
_, ok := <-ch
if !ok {
    // channel closed
}
```

## Prevent Unused Variable Errors

```go
func example(a, b, c int) int {
    _ = b  // Acknowledge b is intentionally unused
    _ = c  // Acknowledge c is intentionally unused
    return a * 2
}
```

## Iota Skip Values

```go
const (
    _        = iota  // Skip 0
    KB int64 = 1 << (10 * iota)
    MB
    GB
    TB
)
```

## Pattern: Check Error Without Using

```go
// ❌ Bad - ignoring error silently
data, _ := json.Marshal(obj)

// ✅ Better - explicit about ignoring
data, err := json.Marshal(obj)
_ = err  // Intentionally ignored: marshaling known-good type

// ✅ Best - handle or document why safe
data, _ := json.Marshal(obj)  // Safe: obj is always valid JSON
```

## When to Use `_`

| Use Case | Example |
|----------|---------|
| Ignore index in range | `for _, v := range slice` |
| Import for side effects | `import _ "driver"` |
| Interface compliance check | `var _ Interface = (*Type)(nil)` |
| Ignore map ok | `v := m[key]` (zero value ok) |
| Skip iota values | `_ = iota` |
| Temporary unused vars | `_ = unusedVar` (during dev) |

## When NOT to Use `_`

| Bad Practice | Why |
|--------------|-----|
| `_, _ = f()` ignoring all returns | Hides bugs |
| `data, _ := io.ReadAll()` ignoring errors | Data corruption risk |
| `_ = err` without comment | Unclear why safe |
| Permanent `var _ = import` | Dead code |

## Quick Reference

```go
// Ignore value
_, err := doSomething()

// Ignore error (document why!)
result, _ := safeFn()  // Safe: never fails

// Import for init
import _ "package"

// Interface check
var _ Interface = (*Type)(nil)

// Skip iota
const (
    _ = iota
    One
    Two
)

// Ignore in range
for _, v := range items { }
for i := range items { }
```
