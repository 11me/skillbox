# Channel Axioms

Four fundamental rules of Go channels (Dave Cheney).

## The Four Axioms

| Operation | nil channel | closed channel |
|-----------|-------------|----------------|
| **Send** | Blocks forever | **Panics** |
| **Receive** | Blocks forever | Returns zero value |
| **Close** | Panics | Panics |

## 1. Send to nil Channel Blocks Forever

```go
var ch chan int
ch <- 1  // blocks forever (deadlock)
```

Uninitialized channels are `nil`. Send operation waits indefinitely.

## 2. Receive from nil Channel Blocks Forever

```go
var ch chan int
v := <-ch  // blocks forever (deadlock)
```

## 3. Send to Closed Channel Panics

```go
ch := make(chan int)
close(ch)
ch <- 1  // panic: send on closed channel
```

**Warning:** There's no safe way to check if channel is closed before sending — any check creates a race condition.

## 4. Receive from Closed Channel Returns Zero Value

```go
ch := make(chan int, 3)
ch <- 1
ch <- 2
ch <- 3
close(ch)

for i := 0; i < 5; i++ {
    v, ok := <-ch
    fmt.Printf("v=%d, ok=%v\n", v, ok)
}
// Output:
// v=1, ok=true
// v=2, ok=true
// v=3, ok=true
// v=0, ok=false  ← closed, returns zero
// v=0, ok=false  ← still returns zero
```

**Key insight:** Closed channels never block — they always return immediately.

## Correct Pattern: Range Over Channel

```go
// Automatically stops when channel is closed
for v := range ch {
    process(v)
}

// Equivalent to:
for {
    v, ok := <-ch
    if !ok {
        break
    }
    process(v)
}
```

## Nil Channel in Select

Nil channels are **ignored** in select statements:

```go
func WaitMany(a, b <-chan struct{}) {
    for a != nil || b != nil {
        select {
        case <-a:
            a = nil  // disable this case
        case <-b:
            b = nil  // nil channels ignored in select
        }
    }
}
```

**Why this works:** When `a` closes first, it would loop infinitely (returns zero forever). Setting `a = nil` removes it from selection.

## Broadcast Signaling

Closing a channel signals ALL receivers simultaneously:

```go
const workers = 100
finish := make(chan struct{})
var wg sync.WaitGroup

for i := 0; i < workers; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        select {
        case <-time.After(time.Hour):
            // timeout
        case <-finish:
            // received close signal
            return
        }
    }()
}

close(finish)  // signals all 100 goroutines at once
wg.Wait()
```

**Pattern:** Use `chan struct{}` when you only need signaling, no data.

## chan struct{} for Signaling

```go
// Empty struct uses zero memory
done := make(chan struct{})

// Signal completion
close(done)

// Or send signal (if channel shouldn't be closed)
done <- struct{}{}
```

## Quick Reference

```go
// Nil channel
var ch chan T     // ch == nil
ch <- v           // blocks forever
<-ch              // blocks forever
close(ch)         // panics

// Closed channel
close(ch)
ch <- v           // panics
v := <-ch         // returns zero, ok=false
v, ok := <-ch     // v=zero, ok=false
close(ch)         // panics (double close)

// In select
select {
case <-nilCh:     // ignored
case <-closedCh:  // always ready
}
```

## Common Mistakes

| Mistake | Result |
|---------|--------|
| Send after close | Panic |
| Close from receiver | Race condition |
| Double close | Panic |
| Check `isClosed` before send | Race condition |
| Receive from nil without select | Deadlock |

## Best Practices

| DO | DON'T |
|----|-------|
| Close from sender only | Close from receiver |
| Use `range` to drain channel | Manual ok-check loops |
| Use `chan struct{}` for signals | Use `chan bool` for signals |
| Set to nil in select to disable | Leave closed channels in select |
| One goroutine owns channel close | Multiple goroutines close same channel |
