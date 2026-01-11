# Interface Design

Go interface patterns from Effective Go.

## Core Principle

> If something can do **this**, it can be used **here**.

Interfaces specify behavior, not data. Types implement interfaces implicitly.

## Small Interfaces

Prefer single-method interfaces — they are composable:

```go
// Standard library patterns
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

type Closer interface {
    Close() error
}

// Composed interfaces
type ReadWriter interface {
    Reader
    Writer
}

type ReadCloser interface {
    Reader
    Closer
}

type ReadWriteCloser interface {
    Reader
    Writer
    Closer
}
```

## Interface Naming

One-method interfaces: method name + `-er` suffix:

| Method | Interface |
|--------|-----------|
| `Read` | `Reader` |
| `Write` | `Writer` |
| `Close` | `Closer` |
| `Format` | `Formatter` |
| `String` | `Stringer` |

## Accept Interfaces, Return Structs

```go
// ✅ Accept interface — flexible for callers
func Process(r io.Reader) error {
    data, err := io.ReadAll(r)
    // ...
}

// ✅ Return concrete type — callers know what they get
func NewBuffer() *bytes.Buffer {
    return &bytes.Buffer{}
}
```

## Implicit Implementation

No `implements` keyword — just define the methods:

```go
type MyReader struct {
    data []byte
    pos  int
}

// MyReader now implements io.Reader
func (r *MyReader) Read(p []byte) (n int, err error) {
    if r.pos >= len(r.data) {
        return 0, io.EOF
    }
    n = copy(p, r.data[r.pos:])
    r.pos += n
    return n, nil
}
```

## Compile-Time Interface Check

Verify implementation at compile time:

```go
// Fails to compile if *MyReader doesn't implement io.Reader
var _ io.Reader = (*MyReader)(nil)

// Common patterns
var _ json.Marshaler = (*Config)(nil)
var _ fmt.Stringer = (*User)(nil)
var _ error = (*AppError)(nil)
```

## Type Assertion

Extract concrete type from interface:

```go
// Safe (comma-ok idiom)
if str, ok := value.(string); ok {
    fmt.Println("string:", str)
}

// Check for interface implementation
if stringer, ok := value.(fmt.Stringer); ok {
    fmt.Println(stringer.String())
}
```

## Type Switch

Handle multiple types:

```go
func stringify(v any) string {
    switch t := v.(type) {
    case string:
        return t
    case fmt.Stringer:
        return t.String()
    case error:
        return t.Error()
    default:
        return fmt.Sprintf("%v", t)
    }
}
```

## Interface Satisfaction for Methods

Method sets determine interface satisfaction:

| Receiver | Value Methods | Pointer Methods |
|----------|---------------|-----------------|
| `T` | ✅ | ❌ |
| `*T` | ✅ | ✅ |

```go
type Counter int

func (c *Counter) Increment() { *c++ }  // Pointer receiver
func (c Counter) Value() int { return int(c) }  // Value receiver

// *Counter has both methods
// Counter has only Value()
```

## Empty Interface (`any`)

Use sparingly — loses type safety:

```go
// ❌ Avoid when possible
func Process(data any) { }

// ✅ Prefer specific interface
func Process(data io.Reader) { }

// OK for truly generic cases (logging, formatting)
func Printf(format string, args ...any) { }
```

## Interface Design Guidelines

| DO | DON'T |
|----|-------|
| Keep interfaces small (1-3 methods) | Create large interfaces |
| Define interfaces at point of use | Define interfaces in implementation package |
| Accept interfaces in function params | Return interfaces (usually) |
| Use stdlib interfaces when applicable | Reinvent Reader/Writer/Closer |
| Check interface at compile time | Discover missing methods at runtime |

## Standard Library Interfaces

Know and use these:

```go
// io package
io.Reader
io.Writer
io.Closer
io.ReadWriter
io.ReadCloser

// fmt package
fmt.Stringer      // String() string
fmt.GoStringer    // GoString() string

// sort package
sort.Interface    // Len, Less, Swap

// encoding/json
json.Marshaler    // MarshalJSON() ([]byte, error)
json.Unmarshaler  // UnmarshalJSON([]byte) error

// context
context.Context   // Deadline, Done, Err, Value
```

## Handler Pattern

Function type implementing interface:

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}

// Function adapter
type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
    f(w, r)
}

// Usage: convert function to Handler
http.Handle("/", http.HandlerFunc(myHandler))
```
