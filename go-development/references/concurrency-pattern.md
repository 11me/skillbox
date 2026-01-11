# Concurrency Patterns

Go concurrency primitives: goroutines, channels, and sync package.

## Goroutine Lifecycle

**Rule:** Always know when goroutine exits.

```go
// Bad: fire and forget
func startWorker() {
    go func() {
        for {
            doWork() // runs forever, no way to stop
        }
    }()
}

// Good: controlled lifecycle
func startWorker(ctx context.Context) {
    go func() {
        for {
            select {
            case <-ctx.Done():
                return // clean exit
            default:
                doWork()
            }
        }
    }()
}
```

## Context Propagation

| Rule | Description |
|------|-------------|
| First parameter | `func Foo(ctx context.Context, ...)` |
| Never store in struct | Pass through function calls |
| Create at boundaries | HTTP handlers, main, tests |
| Derive with timeout | `context.WithTimeout(ctx, 5*time.Second)` |

```go
// Good
func (s *Service) CreateUser(ctx context.Context, req *CreateRequest) (*User, error) {
    // Pass context to downstream calls
    user, err := s.repo.Save(ctx, req.ToUser())
    if err != nil {
        return nil, err
    }

    // Check cancellation
    if ctx.Err() != nil {
        return nil, ctx.Err()
    }

    return user, nil
}

// Bad
type Service struct {
    ctx context.Context // Never store context in struct
}
```

## Channel Patterns

### Unbuffered vs Buffered

| Type | Use Case |
|------|----------|
| Unbuffered | Synchronization, handoff |
| Buffered | Async work, rate limiting |

```go
// Unbuffered: blocks until receiver ready
done := make(chan struct{})

// Buffered: doesn't block until full
jobs := make(chan Job, 100)
```

### Close from Sender

```go
func producer(ch chan<- int) {
    defer close(ch) // sender closes
    for i := 0; i < 10; i++ {
        ch <- i
    }
}

func consumer(ch <-chan int) {
    for v := range ch { // range exits when closed
        process(v)
    }
}
```

### Done Channel Pattern

```go
func worker(done <-chan struct{}, jobs <-chan Job) {
    for {
        select {
        case <-done:
            return
        case job := <-jobs:
            process(job)
        }
    }
}

// Usage
done := make(chan struct{})
go worker(done, jobs)

// Shutdown
close(done)
```

## Sync Primitives

### WaitGroup

```go
func processItems(items []Item) {
    var wg sync.WaitGroup

    for _, item := range items {
        wg.Add(1)
        go func(item Item) {
            defer wg.Done()
            process(item)
        }(item)
    }

    wg.Wait() // blocks until all done
}
```

### Mutex

```go
type SafeCounter struct {
    mu    sync.Mutex
    count int
}

func (c *SafeCounter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

func (c *SafeCounter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.count
}
```

### RWMutex

```go
type Cache struct {
    mu   sync.RWMutex
    data map[string]string
}

func (c *Cache) Get(key string) (string, bool) {
    c.mu.RLock() // multiple readers
    defer c.mu.RUnlock()
    v, ok := c.data[key]
    return v, ok
}

func (c *Cache) Set(key, value string) {
    c.mu.Lock() // exclusive writer
    defer c.mu.Unlock()
    c.data[key] = value
}
```

### Once

```go
var (
    instance *Singleton
    once     sync.Once
)

func GetInstance() *Singleton {
    once.Do(func() {
        instance = &Singleton{}
    })
    return instance
}
```

## Worker Pool

```go
func WorkerPool(ctx context.Context, jobs <-chan Job, workers int) <-chan Result {
    results := make(chan Result, workers)

    var wg sync.WaitGroup
    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for {
                select {
                case <-ctx.Done():
                    return
                case job, ok := <-jobs:
                    if !ok {
                        return
                    }
                    results <- process(job)
                }
            }
        }()
    }

    // Close results when all workers done
    go func() {
        wg.Wait()
        close(results)
    }()

    return results
}
```

## Fan-Out/Fan-In

```go
// Fan-out: distribute work to multiple goroutines
func fanOut(ctx context.Context, input <-chan int, workers int) []<-chan int {
    outputs := make([]<-chan int, workers)
    for i := 0; i < workers; i++ {
        outputs[i] = worker(ctx, input)
    }
    return outputs
}

// Fan-in: merge multiple channels into one
func fanIn(ctx context.Context, channels ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup

    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for v := range c {
                select {
                case <-ctx.Done():
                    return
                case out <- v:
                }
            }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}
```

## Semaphore Pattern

```go
func processWithLimit(items []Item, maxConcurrent int) {
    sem := make(chan struct{}, maxConcurrent)
    var wg sync.WaitGroup

    for _, item := range items {
        wg.Add(1)
        sem <- struct{}{} // acquire

        go func(item Item) {
            defer wg.Done()
            defer func() { <-sem }() // release

            process(item)
        }(item)
    }

    wg.Wait()
}
```

## Timeout Pattern

```go
func doWithTimeout(ctx context.Context, timeout time.Duration) error {
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()

    resultCh := make(chan error, 1)
    go func() {
        resultCh <- doExpensiveOperation(ctx)
    }()

    select {
    case err := <-resultCh:
        return err
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

## Errgroup

```go
import "golang.org/x/sync/errgroup"

func processAll(ctx context.Context, items []Item) error {
    g, ctx := errgroup.WithContext(ctx)

    for _, item := range items {
        item := item // capture for Go < 1.22
        g.Go(func() error {
            return process(ctx, item)
        })
    }

    return g.Wait() // returns first error
}
```

## Best Practices

| Do | Don't |
|----|-------|
| Pass context as first param | Store context in struct |
| Use WaitGroup for completion | Rely on sleep |
| Close channels from sender | Close from receiver |
| Use select with ctx.Done() | Ignore cancellation |
| Limit concurrent goroutines | Spawn unlimited goroutines |
| Use errgroup for error handling | Ignore goroutine errors |
