# Allocation Patterns

`new` vs `make` and memory allocation in Go.

## new vs make

| Function | Types | Returns | Initializes |
|----------|-------|---------|-------------|
| `new(T)` | Any type | `*T` (pointer) | Zeroed memory |
| `make(T)` | Slice, map, channel only | `T` (value) | Initialized |

## new Function

Allocates zeroed memory, returns pointer:

```go
// new allocates zeroed storage
p := new(int)       // *int, points to 0
s := new(string)    // *string, points to ""
b := new(bool)      // *bool, points to false

// Equivalent to
var i int
p := &i

// For structs
type Config struct {
    Host string
    Port int
}

cfg := new(Config)  // *Config with zero values
// cfg.Host == "", cfg.Port == 0
```

## make Function

Creates and initializes slice, map, or channel:

```go
// Slice: make([]T, length, capacity)
s := make([]int, 10)       // len=10, cap=10
s := make([]int, 0, 100)   // len=0, cap=100

// Map: make(map[K]V, hint)
m := make(map[string]int)      // empty map
m := make(map[string]int, 100) // pre-sized for ~100 entries

// Channel: make(chan T, buffer)
ch := make(chan int)     // unbuffered
ch := make(chan int, 10) // buffered, capacity 10
```

## Decision Matrix

```
Need pointer to zero value?
├── YES → new(T) or &T{}
└── NO → Continue...

Creating slice, map, or channel?
├── YES → make(T)
└── NO → var declaration or composite literal
```

## Zero Values Are Useful

Design types so zero value is ready to use:

```go
// sync.Mutex — zero value is unlocked mutex
var mu sync.Mutex
mu.Lock()  // Works immediately

// bytes.Buffer — zero value is empty buffer
var buf bytes.Buffer
buf.WriteString("hello")  // Works immediately

// Your types should follow this pattern
type Counter struct {
    mu    sync.Mutex
    count int
}

var c Counter  // Ready to use
c.mu.Lock()
c.count++
c.mu.Unlock()
```

## Composite Literals

Create and initialize in one expression:

```go
// Struct literal
cfg := &Config{
    Host: "localhost",
    Port: 8080,
}

// Array literal
arr := [3]int{1, 2, 3}

// Slice literal
slice := []string{"a", "b", "c"}

// Map literal
m := map[string]int{
    "one":   1,
    "two":   2,
    "three": 3,
}
```

## Slice vs Array

```go
// Array — fixed size, value type
var arr [5]int           // Zero array
arr := [5]int{1, 2, 3}   // Partial init, rest are 0
arr := [...]int{1, 2, 3} // Size inferred: [3]int

// Slice — dynamic, reference type
var s []int              // nil slice
s := []int{}             // Empty slice (not nil)
s := make([]int, 0, 10)  // Empty with capacity
```

## nil vs Empty

```go
// nil slice
var s []int           // s == nil, len(s) == 0
s := []int(nil)       // Same

// Empty slice (not nil)
s := []int{}          // s != nil, len(s) == 0
s := make([]int, 0)   // Same

// For JSON marshaling:
// nil slice → null
// empty slice → []
```

## Allocation Patterns

### Pre-allocate When Size Known

```go
// ❌ Grows multiple times
var result []int
for _, v := range input {
    result = append(result, v*2)
}

// ✅ Single allocation
result := make([]int, 0, len(input))
for _, v := range input {
    result = append(result, v*2)
}

// ✅ Or direct assignment
result := make([]int, len(input))
for i, v := range input {
    result[i] = v * 2
}
```

### Map Pre-sizing

```go
// ❌ May rehash multiple times
m := make(map[string]int)
for _, item := range items {
    m[item.Key] = item.Value
}

// ✅ Pre-sized
m := make(map[string]int, len(items))
for _, item := range items {
    m[item.Key] = item.Value
}
```

## Constructors

Use `New` prefix for constructors:

```go
func NewClient(cfg Config) *Client {
    return &Client{
        config: cfg,
        http:   &http.Client{Timeout: 30 * time.Second},
    }
}

// With validation
func NewClient(cfg Config) (*Client, error) {
    if cfg.Host == "" {
        return nil, errors.New("host required")
    }
    return &Client{config: cfg}, nil
}
```

## Quick Reference

| Need | Use |
|------|-----|
| Pointer to zero struct | `new(T)` or `&T{}` |
| Slice with length | `make([]T, len)` |
| Slice with capacity | `make([]T, 0, cap)` |
| Empty map | `make(map[K]V)` |
| Pre-sized map | `make(map[K]V, hint)` |
| Unbuffered channel | `make(chan T)` |
| Buffered channel | `make(chan T, size)` |
| Struct with values | `&T{field: value}` |
