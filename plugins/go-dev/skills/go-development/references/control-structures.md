# Control Structures

Go control flow patterns from Effective Go.

## Guard Clauses (Early Return)

Avoid unnecessary `else` — when body ends with `return`, `break`, `continue`, or `goto`.

```go
// ❌ Unnecessary else
func process(data []byte) error {
    if len(data) == 0 {
        return errors.New("empty data")
    } else {
        // process...
        return nil
    }
}

// ✅ Guard clause pattern
func process(data []byte) error {
    if len(data) == 0 {
        return errors.New("empty data")
    }
    // happy path continues
    return nil
}
```

## Multiple Guard Clauses

Handle errors as they arise, successful flow runs down the page:

```go
func ReadConfig(path string) (*Config, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    data, err := io.ReadAll(f)
    if err != nil {
        return nil, err
    }

    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}
```

## If with Initialization

Declare variables in `if` statement scope:

```go
// Variable scoped to if/else block
if err := file.Chmod(0644); err != nil {
    return err
}

// err not accessible here

// Comma-ok idiom
if v, ok := cache[key]; ok {
    return v
}
```

## For Loops

### Three Forms

```go
// C-style for
for i := 0; i < 10; i++ {
    sum += i
}

// While-like
for condition {
    // ...
}

// Infinite loop
for {
    // break or return to exit
}
```

### Range Clause

```go
// Key and value
for key, value := range m {
    fmt.Println(key, value)
}

// Key only
for key := range m {
    delete(m, key)
}

// Value only (discard key)
for _, value := range slice {
    sum += value
}

// Index only
for i := range slice {
    slice[i] = i * 2
}
```

### Range Over Strings (UTF-8)

```go
// Range decodes UTF-8 runes
for pos, char := range "日本語" {
    fmt.Printf("position %d: %c\n", pos, char)
}
// position 0: 日
// position 3: 本
// position 6: 語
```

### Parallel Assignment in For

```go
// Reverse a slice
for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
    a[i], a[j] = a[j], a[i]
}
```

## Switch Statements

### No Automatic Fall-through

```go
// Each case breaks automatically
switch c {
case ' ', '\t', '\n':
    return true
case 'a', 'b', 'c':
    return false
}
```

### Switch Without Expression

Acts as `if-else-if` chain:

```go
func classify(n int) string {
    switch {
    case n < 0:
        return "negative"
    case n == 0:
        return "zero"
    case n < 10:
        return "small"
    default:
        return "large"
    }
}
```

### Break to Label

Break out of outer loop from switch:

```go
Loop:
    for _, item := range items {
        switch item.Type {
        case "skip":
            continue Loop
        case "stop":
            break Loop  // breaks the for loop, not just switch
        default:
            process(item)
        }
    }
```

## Type Switch

Discover dynamic type of interface value:

```go
func describe(v any) string {
    switch t := v.(type) {
    case nil:
        return "nil"
    case int:
        return fmt.Sprintf("int: %d", t)
    case string:
        return fmt.Sprintf("string: %q", t)
    case bool:
        return fmt.Sprintf("bool: %t", t)
    case fmt.Stringer:
        return fmt.Sprintf("Stringer: %s", t.String())
    default:
        return fmt.Sprintf("unknown type: %T", t)
    }
}
```

### Type Assertion

```go
// Safe type assertion (comma-ok)
if str, ok := value.(string); ok {
    fmt.Printf("string: %s\n", str)
}

// Unsafe (panics if wrong type)
str := value.(string)
```

## Redeclaration with `:=`

Variable can be redeclared if at least one new variable is created:

```go
f, err := os.Open(name)
if err != nil {
    return err
}

// err is reassigned, d is new
d, err := f.Stat()
if err != nil {
    f.Close()
    return err
}
```

## Quick Reference

| Pattern | Use Case |
|---------|----------|
| Guard clause | Early return on error/invalid input |
| If with init | Scope variables to conditional block |
| Range key only | Iteration, deletion from maps |
| Range value only | Sum, transform without index |
| Switch no expression | Clean if-else-if chains |
| Type switch | Handle multiple types from interface |
| Break to label | Exit outer loop from nested switch |
