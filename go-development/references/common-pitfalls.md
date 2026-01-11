# Common Pitfalls

Frequent mistakes in Go and how to avoid them.

## Defer in Loops

Defer executes when **function** returns, not when loop iteration ends.

```go
// Bad: all files stay open until function returns
func processFiles(paths []string) error {
    for _, path := range paths {
        f, err := os.Open(path)
        if err != nil {
            return err
        }
        defer f.Close() // deferred until function returns!

        process(f)
    }
    return nil
}

// Good: wrap in closure
func processFiles(paths []string) error {
    for _, path := range paths {
        if err := func() error {
            f, err := os.Open(path)
            if err != nil {
                return err
            }
            defer f.Close() // closes after closure returns

            return process(f)
        }(); err != nil {
            return err
        }
    }
    return nil
}

// Good: extract function
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

    return process(f)
}
```

## Goroutine Leaks

Goroutine waiting forever on channel = memory leak.

```go
// Bad: goroutine leak if timeout occurs
func fetch(url string) ([]byte, error) {
    ch := make(chan []byte)

    go func() {
        data, _ := http.Get(url)
        ch <- data // blocks forever if no receiver
    }()

    select {
    case data := <-ch:
        return data, nil
    case <-time.After(5 * time.Second):
        return nil, errors.New("timeout")
        // goroutine still blocked on ch <- data
    }
}

// Good: use buffered channel
func fetch(url string) ([]byte, error) {
    ch := make(chan []byte, 1) // buffer of 1

    go func() {
        data, _ := http.Get(url)
        ch <- data // won't block even if no receiver
    }()

    select {
    case data := <-ch:
        return data, nil
    case <-time.After(5 * time.Second):
        return nil, errors.New("timeout")
    }
}

// Good: use context cancellation
func fetch(ctx context.Context, url string) ([]byte, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    return io.ReadAll(resp.Body)
}
```

## Loop Variable Capture (Go < 1.22)

**Note:** Fixed in Go 1.22+. For older versions:

```go
// Bad (Go < 1.22): all goroutines see same value
func process(items []Item) {
    for _, item := range items {
        go func() {
            handle(item) // captures loop variable
        }()
    }
}

// Good: pass as parameter
func process(items []Item) {
    for _, item := range items {
        go func(item Item) {
            handle(item)
        }(item)
    }
}

// Good: create local copy
func process(items []Item) {
    for _, item := range items {
        item := item // shadow with local copy
        go func() {
            handle(item)
        }()
    }
}
```

## Nil Map

Writing to nil map causes panic.

```go
// Bad: panic
var m map[string]int
m["key"] = 1 // panic: assignment to entry in nil map

// Good: initialize with make
m := make(map[string]int)
m["key"] = 1

// Good: use literal
m := map[string]int{}
m["key"] = 1

// Reading from nil is safe (returns zero value)
var m map[string]int
v := m["key"] // v = 0, ok
```

## Nil Slice

Nil slice is usable but has subtle differences.

```go
// Nil slice works with append
var s []int
s = append(s, 1, 2, 3) // works fine

// Nil vs empty slice in JSON
var nilSlice []int           // JSON: null
emptySlice := []int{}        // JSON: []
emptySlice2 := make([]int, 0) // JSON: []

// Length and capacity are 0
var s []int
len(s) // 0
cap(s) // 0
```

## Interface Nil

Interface is nil only if **both** type and value are nil.

```go
// Bad: unexpected non-nil
func getError() error {
    var err *MyError = nil
    return err // returns non-nil interface!
}

func main() {
    err := getError()
    if err != nil { // true!
        fmt.Println("error:", err)
    }
}

// Good: return nil directly
func getError() error {
    var err *MyError = nil
    if err == nil {
        return nil // return nil interface
    }
    return err
}

// Good: check concrete type
func isNil(err error) bool {
    if err == nil {
        return true
    }
    // Use reflection for nil pointer in interface
    v := reflect.ValueOf(err)
    return v.Kind() == reflect.Ptr && v.IsNil()
}
```

## String Iteration

Range over string iterates runes, not bytes.

```go
s := "hello"

// Iterates runes (may be > 1 byte each)
for i, r := range s {
    fmt.Printf("%d: %c\n", i, r)
}

// For byte iteration
for i := 0; i < len(s); i++ {
    fmt.Printf("%d: %c\n", i, s[i])
}

// UTF-8 example
s := "hi"
len(s)         // 8 (bytes)
len([]rune(s)) // 3 (runes)
```

## Append Overwrites

Append may modify underlying array.

```go
// Bad: modifies original
func addOne(s []int) []int {
    return append(s, 1)
}

original := make([]int, 0, 10)
original = append(original, 1, 2, 3)
modified := addOne(original[:2])
// original[2] may be overwritten!

// Good: force new allocation
func addOne(s []int) []int {
    result := make([]int, len(s), len(s)+1)
    copy(result, s)
    return append(result, 1)
}

// Good: use full slice expression
modified := addOne(original[:2:2]) // cap = len
```

## Time Comparison

Use Before/After/Equal, not == or <.

```go
// Bad: may fail due to monotonic clock
if t1 == t2 { ... }

// Good
if t1.Equal(t2) { ... }
if t1.Before(t2) { ... }
if t1.After(t2) { ... }

// For duration comparison
if time.Since(start) > 5*time.Second { ... }
```

## JSON Unexported Fields

Only exported fields are marshaled.

```go
type User struct {
    Name  string // exported: included
    email string // unexported: ignored
}

u := User{Name: "John", email: "john@example.com"}
data, _ := json.Marshal(u)
// {"Name":"John"} - email is missing

// Use tags for custom names
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}
```

## HTTP Body Not Closed

Always close response body.

```go
// Bad: resource leak
resp, err := http.Get(url)
if err != nil {
    return err
}
// body not closed!

// Good
resp, err := http.Get(url)
if err != nil {
    return err
}
defer resp.Body.Close()

// Read body even on error status
if resp.StatusCode != http.StatusOK {
    io.Copy(io.Discard, resp.Body) // drain body
    resp.Body.Close()
    return fmt.Errorf("status: %d", resp.StatusCode)
}
```

## Summary

| Pitfall | Solution |
|---------|----------|
| Defer in loops | Wrap in closure or extract function |
| Goroutine leaks | Buffered channels, context cancellation |
| Loop variable capture | Pass as param (Go < 1.22) |
| Nil map write | Initialize with make() |
| Interface nil check | Return nil directly, check concrete type |
| String iteration | Be aware: range = runes, index = bytes |
| Append overwrites | Use full slice expression or copy |
| Time comparison | Use Equal/Before/After methods |
| JSON unexported | Use exported fields with tags |
| HTTP body leak | Always defer Body.Close() |
