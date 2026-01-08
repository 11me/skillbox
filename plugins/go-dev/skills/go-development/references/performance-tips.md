# Performance Tips

Practical Go performance optimizations.

## String Building

String concatenation in loop creates many allocations.

```go
// Bad: O(n^2) allocations
func buildString(items []string) string {
    var result string
    for _, item := range items {
        result += item + "," // new allocation each time
    }
    return result
}

// Good: strings.Builder
func buildString(items []string) string {
    var b strings.Builder
    b.Grow(len(items) * 10) // pre-allocate estimate

    for i, item := range items {
        if i > 0 {
            b.WriteByte(',')
        }
        b.WriteString(item)
    }
    return b.String()
}

// Good: strings.Join for simple cases
func buildString(items []string) string {
    return strings.Join(items, ",")
}
```

## Slice Pre-allocation

```go
// Bad: multiple reallocations
func transform(items []Item) []Result {
    var results []Result
    for _, item := range items {
        results = append(results, process(item))
    }
    return results
}

// Good: pre-allocate with capacity
func transform(items []Item) []Result {
    results := make([]Result, 0, len(items))
    for _, item := range items {
        results = append(results, process(item))
    }
    return results
}

// Good: pre-allocate with length (if filling sequentially)
func transform(items []Item) []Result {
    results := make([]Result, len(items))
    for i, item := range items {
        results[i] = process(item)
    }
    return results
}
```

## Map Pre-allocation

```go
// Bad: multiple rehashes
m := make(map[string]int)
for _, item := range items {
    m[item.Key] = item.Value
}

// Good: pre-allocate
m := make(map[string]int, len(items))
for _, item := range items {
    m[item.Key] = item.Value
}
```

## sync.Pool

Reuse expensive objects to reduce GC pressure.

```go
var bufferPool = sync.Pool{
    New: func() any {
        return new(bytes.Buffer)
    },
}

func process(data []byte) string {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()

    buf.Write(data)
    // ... process buffer
    return buf.String()
}
```

**Good for:**
- Buffers, byte slices
- Encoders/decoders
- Temporary large structs

**Not good for:**
- Small objects (overhead > benefit)
- Long-lived objects
- Objects with cleanup requirements

## Avoid Allocations

### Pass Pointers for Large Structs

```go
// Bad: copies 1KB+ struct
func process(config Config) { ... }

// Good: pass pointer
func process(config *Config) { ... }
```

### Use Value Receivers for Small Structs

```go
// Good: small struct, value receiver avoids indirection
type Point struct {
    X, Y int
}

func (p Point) Distance() float64 {
    return math.Sqrt(float64(p.X*p.X + p.Y*p.Y))
}
```

### Avoid interface{}/any When Type is Known

```go
// Bad: allocation + type assertion
func sum(values []any) int {
    var total int
    for _, v := range values {
        total += v.(int)
    }
    return total
}

// Good: use concrete type
func sum(values []int) int {
    var total int
    for _, v := range values {
        total += v
    }
    return total
}
```

## Avoid Unnecessary Conversions

```go
// Bad: converts to string for comparison
if string(data) == "hello" { ... }

// Good: compare bytes
if bytes.Equal(data, []byte("hello")) { ... }

// Bad: unnecessary []byte conversion
json.Unmarshal([]byte(jsonString), &v)

// Good: use decoder for string
json.NewDecoder(strings.NewReader(jsonString)).Decode(&v)
```

## Struct Field Alignment

Order struct fields by size (largest first) to reduce padding.

```go
// Bad: 24 bytes (with padding)
type BadStruct struct {
    a bool   // 1 byte + 7 padding
    b int64  // 8 bytes
    c bool   // 1 byte + 7 padding
}

// Good: 16 bytes
type GoodStruct struct {
    b int64  // 8 bytes
    a bool   // 1 byte
    c bool   // 1 byte + 6 padding
}
```

## HTTP Client Reuse

```go
// Bad: new client per request
func fetch(url string) (*http.Response, error) {
    client := &http.Client{Timeout: 10 * time.Second}
    return client.Get(url)
}

// Good: reuse client (connection pooling)
var httpClient = &http.Client{
    Timeout: 10 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}

func fetch(url string) (*http.Response, error) {
    return httpClient.Get(url)
}
```

## Benchmarking

```go
func BenchmarkProcess(b *testing.B) {
    data := prepareTestData()

    b.ResetTimer() // exclude setup time
    for i := 0; i < b.N; i++ {
        process(data)
    }
}

// Run with:
// go test -bench=. -benchmem
```

**Output:**
```
BenchmarkProcess-8   1000000   1234 ns/op   256 B/op   3 allocs/op
```

## Profiling

```go
import _ "net/http/pprof"

func main() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()

    // ... application
}
```

**Analyze:**
```bash
# CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Memory profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profile
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

## Common Commands

```bash
# Run benchmarks
go test -bench=. -benchmem ./...

# Compare benchmarks
go test -bench=. -count=10 > old.txt
# ... make changes
go test -bench=. -count=10 > new.txt
benchstat old.txt new.txt

# Check for race conditions
go test -race ./...

# Build with optimizations disabled (for debugging)
go build -gcflags="-N -l"

# Check escape analysis
go build -gcflags="-m" 2>&1 | grep "escapes"
```

## Summary

| Area | Optimization |
|------|-------------|
| Strings | Use strings.Builder, strings.Join |
| Slices | Pre-allocate with make([]T, 0, cap) |
| Maps | Pre-allocate with make(map[K]V, size) |
| Objects | sync.Pool for expensive temporary objects |
| Structs | Order fields by size (largest first) |
| HTTP | Reuse http.Client (connection pooling) |
| Interfaces | Avoid unnecessary any/interface{} |
| Conversions | Avoid unnecessary string ↔ []byte |
| Profiling | Use pprof, benchmarks with -benchmem |
