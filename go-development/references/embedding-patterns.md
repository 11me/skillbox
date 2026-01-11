# Embedding Patterns

Composition through embedding in Go.

## What is Embedding?

Anonymous field in struct — type without field name:

```go
// Named field (not embedding)
type ReadWriter struct {
    reader *Reader
    writer *Writer
}

// Embedding (anonymous fields)
type ReadWriter struct {
    *Reader
    *Writer
}
```

## Method Promotion

Embedded type's methods are promoted to outer type:

```go
type Logger struct {
    prefix string
}

func (l *Logger) Log(msg string) {
    fmt.Printf("[%s] %s\n", l.prefix, msg)
}

type Service struct {
    *Logger  // Embedding
    name string
}

// Service gets Log method automatically
svc := &Service{Logger: &Logger{prefix: "SVC"}, name: "api"}
svc.Log("started")  // Calls Logger.Log
```

## Field Promotion

Embedded type's fields are also accessible:

```go
type Point struct {
    X, Y int
}

type Circle struct {
    Point  // Embedding
    Radius int
}

c := Circle{Point: Point{X: 10, Y: 20}, Radius: 5}
fmt.Println(c.X, c.Y)  // Access promoted fields
fmt.Println(c.Radius)
```

## Interface Embedding

Compose interfaces:

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

type Closer interface {
    Close() error
}

// Composed interface
type ReadWriteCloser interface {
    Reader
    Writer
    Closer
}
```

## Struct Embeds Interface

Struct can embed interface:

```go
type Conn struct {
    io.ReadWriteCloser  // Embed interface
    addr string
}

// Conn must be initialized with concrete implementation
conn := &Conn{
    ReadWriteCloser: tcpConn,
    addr:            "localhost:8080",
}
```

## Forwarding vs Embedding

```go
// Forwarding (explicit delegation)
type ReadWriter struct {
    reader *Reader
    writer *Writer
}

func (rw *ReadWriter) Read(p []byte) (int, error) {
    return rw.reader.Read(p)  // Manual forwarding
}

func (rw *ReadWriter) Write(p []byte) (int, error) {
    return rw.writer.Write(p)  // Manual forwarding
}

// Embedding (automatic promotion)
type ReadWriter struct {
    *Reader
    *Writer
}
// Read and Write methods promoted automatically
```

## Method Override

Outer type can override embedded methods:

```go
type Logger struct{}

func (l *Logger) Log(msg string) {
    fmt.Println(msg)
}

type PrefixLogger struct {
    *Logger
    Prefix string
}

// Override Log method
func (p *PrefixLogger) Log(msg string) {
    p.Logger.Log(p.Prefix + ": " + msg)  // Call embedded method
}
```

## Access Embedded Type

Access embedded type directly by type name:

```go
type Job struct {
    *log.Logger
    Command string
}

job := &Job{
    Logger:  log.New(os.Stdout, "JOB: ", 0),
    Command: "backup",
}

// Access Logger directly
job.Logger.SetPrefix("TASK: ")

// Or use promoted method
job.Println("starting")
```

## Name Conflicts

Outer field shadows inner:

```go
type Inner struct {
    Value int
}

type Outer struct {
    Inner
    Value string  // Shadows Inner.Value
}

o := Outer{Inner: Inner{Value: 42}, Value: "hello"}
fmt.Println(o.Value)        // "hello" (Outer.Value)
fmt.Println(o.Inner.Value)  // 42 (Inner.Value)
```

## Common Patterns

### Add Context to Logger

```go
type Job struct {
    *log.Logger
    Command string
}

func (j *Job) Printf(format string, args ...any) {
    j.Logger.Printf("%s: %s", j.Command, fmt.Sprintf(format, args...))
}
```

### Mutex in Struct

```go
type SafeCounter struct {
    sync.Mutex  // Embedding
    count int
}

func (c *SafeCounter) Inc() {
    c.Lock()
    defer c.Unlock()
    c.count++
}
```

### Extend http.Handler

```go
type LoggingHandler struct {
    http.Handler  // Embed interface
    logger *log.Logger
}

func (h *LoggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    h.logger.Printf("%s %s", r.Method, r.URL.Path)
    h.Handler.ServeHTTP(w, r)  // Delegate to embedded handler
}
```

## Embedding vs Composition

| Aspect | Embedding | Composition |
|--------|-----------|-------------|
| Syntax | Anonymous field | Named field |
| Method access | `outer.Method()` | `outer.field.Method()` |
| Method promotion | Automatic | Manual forwarding |
| Interface satisfaction | Inherited | Manual |
| Encapsulation | Weaker | Stronger |

## When to Use Embedding

| Use Embedding | Use Composition |
|---------------|-----------------|
| Type IS-A relationship | Type HAS-A relationship |
| Need method promotion | Need encapsulation |
| Implementing interfaces | Internal implementation detail |
| Decorating behavior | Hiding implementation |

## Quick Reference

```go
// Interface embedding
type ReadWriter interface {
    io.Reader
    io.Writer
}

// Struct embedding (pointer)
type Service struct {
    *Logger
    config Config
}

// Struct embedding (value)
type Point3D struct {
    Point2D
    Z int
}

// Embedding sync primitives
type SafeMap struct {
    sync.RWMutex
    data map[string]string
}
```
