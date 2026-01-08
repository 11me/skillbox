# Worker Pattern

Background worker patterns with graceful shutdown and panic recovery.

## Core Concepts

| Component | Purpose |
|-----------|---------|
| **Worker** | Processes items from a queue |
| **Queue** | Holds pending work items |
| **Handler** | Business logic for processing |
| **Recovery** | Catches panics, prevents crashes |

## Generic Queue Worker

Key patterns:

```go
package worker

import (
    "context"
    "fmt"
    "log/slog"
    "runtime/debug"
    "time"
)

// Handler processes a single work item.
type Handler[T any] func(ctx context.Context, item T) error

// Queue provides work items for processing.
type Queue[T any] interface {
    // Pop returns the next item or blocks until available.
    // Returns nil when queue is closed.
    Pop(ctx context.Context) (*T, error)

    // Complete marks item as successfully processed.
    Complete(ctx context.Context, item *T) error

    // Fail marks item as failed (for retry or DLQ).
    Fail(ctx context.Context, item *T, err error) error
}

// Worker processes items from a queue.
type Worker[T any] struct {
    name    string
    queue   Queue[T]
    handler Handler[T]
    logger  *slog.Logger

    pollInterval time.Duration
    maxRetries   int
}

// Config configures the worker.
type Config struct {
    PollInterval time.Duration
    MaxRetries   int
}

// DefaultConfig returns default worker configuration.
func DefaultConfig() Config {
    return Config{
        PollInterval: 1 * time.Second,
        MaxRetries:   3,
    }
}

// New creates a new worker.
func New[T any](
    name string,
    queue Queue[T],
    handler Handler[T],
    logger *slog.Logger,
    cfg Config,
) *Worker[T] {
    return &Worker[T]{
        name:         name,
        queue:        queue,
        handler:      handler,
        logger:       logger.With(slog.String("worker", name)),
        pollInterval: cfg.PollInterval,
        maxRetries:   cfg.MaxRetries,
    }
}

// Start begins processing items until context is cancelled.
func (w *Worker[T]) Start(ctx context.Context) error {
    w.logger.Info("starting worker")

    for {
        select {
        case <-ctx.Done():
            w.logger.Info("worker stopped")
            return ctx.Err()
        default:
            if err := w.processOne(ctx); err != nil {
                w.logger.Error("processing failed",
                    slog.String("error", err.Error()),
                )
            }
        }
    }
}

func (w *Worker[T]) processOne(ctx context.Context) error {
    item, err := w.queue.Pop(ctx)
    if err != nil {
        if ctx.Err() != nil {
            return ctx.Err()
        }
        return fmt.Errorf("pop: %w", err)
    }

    if item == nil {
        // No items available, wait before polling again
        time.Sleep(w.pollInterval)
        return nil
    }

    // Process with panic recovery
    handlerErr := w.safeHandle(ctx, item)

    if handlerErr != nil {
        if err := w.queue.Fail(ctx, item, handlerErr); err != nil {
            w.logger.Error("failed to mark item as failed",
                slog.String("error", err.Error()),
            )
        }
        return handlerErr
    }

    if err := w.queue.Complete(ctx, item); err != nil {
        return fmt.Errorf("complete: %w", err)
    }

    return nil
}

func (w *Worker[T]) safeHandle(ctx context.Context, item *T) (handlerErr error) {
    defer func() {
        if r := recover(); r != nil {
            w.logger.Error("panic in handler",
                slog.Any("panic", r),
                slog.String("stack", string(debug.Stack())),
            )
            handlerErr = fmt.Errorf("panic: %v", r)
        }
    }()

    return w.handler(ctx, *item)
}
```

## Simple In-Memory Queue

For development and testing:

```go
type MemoryQueue[T any] struct {
    items chan T
    done  chan struct{}
}

func NewMemoryQueue[T any](size int) *MemoryQueue[T] {
    return &MemoryQueue[T]{
        items: make(chan T, size),
        done:  make(chan struct{}),
    }
}

func (q *MemoryQueue[T]) Push(item T) error {
    select {
    case q.items <- item:
        return nil
    case <-q.done:
        return fmt.Errorf("queue closed")
    }
}

func (q *MemoryQueue[T]) Pop(ctx context.Context) (*T, error) {
    select {
    case item := <-q.items:
        return &item, nil
    case <-ctx.Done():
        return nil, ctx.Err()
    case <-q.done:
        return nil, nil
    }
}

func (q *MemoryQueue[T]) Complete(ctx context.Context, item *T) error {
    return nil // In-memory doesn't need completion tracking
}

func (q *MemoryQueue[T]) Fail(ctx context.Context, item *T, err error) error {
    return nil // In-memory doesn't have retry logic
}

func (q *MemoryQueue[T]) Close() {
    close(q.done)
}
```

## Database-Backed Queue

For production with persistence:

```go
type DBQueue[T any] struct {
    pool      *pgxpool.Pool
    tableName string
    logger    *slog.Logger
}

type QueueItem[T any] struct {
    ID        int64
    Payload   T
    Status    string    // pending, processing, completed, failed
    Attempts  int
    CreatedAt time.Time
    UpdatedAt time.Time
}

func (q *DBQueue[T]) Pop(ctx context.Context) (*QueueItem[T], error) {
    var item QueueItem[T]

    err := q.pool.QueryRow(ctx, `
        UPDATE `+q.tableName+`
        SET status = 'processing', updated_at = NOW(), attempts = attempts + 1
        WHERE id = (
            SELECT id FROM `+q.tableName+`
            WHERE status = 'pending'
            ORDER BY created_at
            FOR UPDATE SKIP LOCKED
            LIMIT 1
        )
        RETURNING id, payload, status, attempts, created_at, updated_at
    `).Scan(&item.ID, &item.Payload, &item.Status, &item.Attempts, &item.CreatedAt, &item.UpdatedAt)

    if err == pgx.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }

    return &item, nil
}

func (q *DBQueue[T]) Complete(ctx context.Context, item *QueueItem[T]) error {
    _, err := q.pool.Exec(ctx, `
        UPDATE `+q.tableName+`
        SET status = 'completed', updated_at = NOW()
        WHERE id = $1
    `, item.ID)
    return err
}

func (q *DBQueue[T]) Fail(ctx context.Context, item *QueueItem[T], handlerErr error) error {
    _, err := q.pool.Exec(ctx, `
        UPDATE `+q.tableName+`
        SET status = CASE WHEN attempts >= $1 THEN 'failed' ELSE 'pending' END,
            error = $2,
            updated_at = NOW()
        WHERE id = $3
    `, 3, handlerErr.Error(), item.ID)
    return err
}
```

## Worker Pool

Run multiple workers in parallel:

```go
func StartWorkerPool[T any](
    ctx context.Context,
    count int,
    factory func() *Worker[T],
) {
    var wg sync.WaitGroup

    for i := 0; i < count; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            worker := factory()
            worker.Start(ctx)
        }()
    }

    wg.Wait()
}

// Usage
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go StartWorkerPool(ctx, 5, func() *Worker[EmailTask] {
    return NewEmailWorker(queue, emailService, logger)
})
```

## Graceful Shutdown

```go
func main() {
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    // Start workers
    go worker.Start(ctx)

    // Wait for shutdown signal
    <-ctx.Done()

    // Give workers time to finish current item
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Workers should check ctx.Done() and return gracefully
    <-shutdownCtx.Done()
}
```

## Dual-Loop Pattern

For workers that need both processing and cleanup:

```go
func (w *Worker[T]) StartWithCleanup(ctx context.Context) error {
    g, ctx := errgroup.WithContext(ctx)

    // Main processing loop
    g.Go(func() error {
        return w.processLoop(ctx)
    })

    // Cleanup loop (stale items, dead-letter queue, etc.)
    g.Go(func() error {
        return w.cleanupLoop(ctx)
    })

    return g.Wait()
}

func (w *Worker[T]) cleanupLoop(ctx context.Context) error {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            if err := w.cleanupStale(ctx); err != nil {
                w.logger.Error("cleanup failed", slog.String("error", err.Error()))
            }
        }
    }
}

func (w *Worker[T]) cleanupStale(ctx context.Context) error {
    // Move stuck items back to pending
    // Move old failed items to dead-letter queue
    // etc.
    return nil
}
```

## Best Practices

### DO:
- ✅ Always recover from panics
- ✅ Log stack traces on panic
- ✅ Use context for cancellation
- ✅ Implement graceful shutdown
- ✅ Add metrics (processing time, success/failure rates)

### DON'T:
- ❌ Ignore errors from handler
- ❌ Process indefinitely without checking ctx.Done()
- ❌ Use unbounded queues in production
- ❌ Skip retry logic for transient errors

## Related

- [database-pattern.md](database-pattern.md) — Database patterns
- [error-handling.md](error-handling.md) — Error handling
