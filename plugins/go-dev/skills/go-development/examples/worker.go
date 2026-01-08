// Package worker provides a generic background worker with panic recovery.
//
// This example shows:
// - Generic queue worker pattern
// - Panic recovery with stack trace logging
// - Graceful shutdown
// - In-memory queue for testing
package worker

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"time"
)

// ---------- Queue Interface ----------

// Queue provides work items for processing.
type Queue[T any] interface {
	// Pop returns the next item or nil if none available.
	Pop(ctx context.Context) (*T, error)

	// Complete marks item as successfully processed.
	Complete(ctx context.Context, item *T) error

	// Fail marks item as failed.
	Fail(ctx context.Context, item *T, err error) error
}

// ---------- Worker ----------

// Handler processes a single work item.
type Handler[T any] func(ctx context.Context, item T) error

// Config configures the worker.
type Config struct {
	PollInterval time.Duration
}

// DefaultConfig returns default worker configuration.
func DefaultConfig() Config {
	return Config{
		PollInterval: 1 * time.Second,
	}
}

// Worker processes items from a queue.
type Worker[T any] struct {
	name    string
	queue   Queue[T]
	handler Handler[T]
	logger  *slog.Logger
	cfg     Config
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
		name:    name,
		queue:   queue,
		handler: handler,
		logger:  logger.With(slog.String("worker", name)),
		cfg:     cfg,
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
				// Log but don't exit on processing errors
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
		time.Sleep(w.cfg.PollInterval)
		return nil
	}

	// Process with panic recovery
	start := time.Now()
	handlerErr := w.safeHandle(ctx, item)
	elapsed := time.Since(start)

	if handlerErr != nil {
		w.logger.Error("item processing failed",
			slog.Duration("elapsed", elapsed),
			slog.String("error", handlerErr.Error()),
		)
		if err := w.queue.Fail(ctx, item, handlerErr); err != nil {
			w.logger.Error("failed to mark item as failed",
				slog.String("error", err.Error()),
			)
		}
		return handlerErr
	}

	w.logger.Debug("item processed",
		slog.Duration("elapsed", elapsed),
	)

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

// ---------- In-Memory Queue (for testing) ----------

// MemoryQueue is an in-memory queue for testing.
type MemoryQueue[T any] struct {
	items chan T
	done  chan struct{}
	mu    sync.Mutex
}

// NewMemoryQueue creates a new in-memory queue.
func NewMemoryQueue[T any](size int) *MemoryQueue[T] {
	return &MemoryQueue[T]{
		items: make(chan T, size),
		done:  make(chan struct{}),
	}
}

// Push adds an item to the queue.
func (q *MemoryQueue[T]) Push(item T) error {
	select {
	case q.items <- item:
		return nil
	case <-q.done:
		return fmt.Errorf("queue closed")
	}
}

// Pop returns the next item or nil if none available.
func (q *MemoryQueue[T]) Pop(ctx context.Context) (*T, error) {
	select {
	case item := <-q.items:
		return &item, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-q.done:
		return nil, nil
	default:
		// Non-blocking: no items available
		return nil, nil
	}
}

// Complete marks item as processed (no-op for in-memory).
func (q *MemoryQueue[T]) Complete(ctx context.Context, item *T) error {
	return nil
}

// Fail marks item as failed (no-op for in-memory).
func (q *MemoryQueue[T]) Fail(ctx context.Context, item *T, err error) error {
	return nil
}

// Close closes the queue.
func (q *MemoryQueue[T]) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	select {
	case <-q.done:
		// Already closed
	default:
		close(q.done)
	}
}

// Len returns the number of items in the queue.
func (q *MemoryQueue[T]) Len() int {
	return len(q.items)
}

// ---------- Worker Pool ----------

// Pool manages multiple workers.
type Pool[T any] struct {
	workers []*Worker[T]
	wg      sync.WaitGroup
}

// NewPool creates a new worker pool.
func NewPool[T any](
	count int,
	queue Queue[T],
	handler Handler[T],
	logger *slog.Logger,
	cfg Config,
) *Pool[T] {
	pool := &Pool[T]{
		workers: make([]*Worker[T], count),
	}

	for i := 0; i < count; i++ {
		pool.workers[i] = New(
			fmt.Sprintf("worker-%d", i),
			queue,
			handler,
			logger,
			cfg,
		)
	}

	return pool
}

// Start begins all workers.
func (p *Pool[T]) Start(ctx context.Context) {
	for _, w := range p.workers {
		p.wg.Add(1)
		go func(worker *Worker[T]) {
			defer p.wg.Done()
			worker.Start(ctx)
		}(w)
	}
}

// Wait blocks until all workers have stopped.
func (p *Pool[T]) Wait() {
	p.wg.Wait()
}

// ---------- Usage Example ----------

// Example usage:
//
//	func main() {
//	    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
//	    defer stop()
//
//	    logger := slog.Default()
//	    queue := worker.NewMemoryQueue[EmailTask](1000)
//
//	    // Start worker pool
//	    pool := worker.NewPool(5, queue, func(ctx context.Context, task EmailTask) error {
//	        return emailService.Send(ctx, task.To, task.Subject, task.Body)
//	    }, logger, worker.DefaultConfig())
//
//	    pool.Start(ctx)
//
//	    // Push some tasks
//	    queue.Push(EmailTask{To: "user@example.com", Subject: "Hello", Body: "..."})
//
//	    // Wait for shutdown
//	    pool.Wait()
//	}
//
//	type EmailTask struct {
//	    To      string
//	    Subject string
//	    Body    string
//	}
