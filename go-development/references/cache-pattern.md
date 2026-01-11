# Cache Pattern

Redis-based caching patterns for reducing database load.

## Problem

- Repeated database lookups for same data
- N+1 query problems in batch operations
- High latency for frequently accessed data
- Database overload under high traffic

## Caching Strategies

### 1. Cache-Aside (Lazy Loading)

Application manages cache explicitly. Most common pattern.

```
Read:  Check cache → miss → fetch DB → write cache → return
Write: Update DB → invalidate cache (or update cache)
```

**Pros:** Simple, only caches what's needed, tolerates cache failures
**Cons:** Initial request always hits DB, potential stale data

**Best for:** Read-heavy workloads, user profiles, product catalogs

### 2. Write-Through

Cache is updated synchronously with database.

```
Write: Update cache AND DB in same transaction
Read:  Always from cache (guaranteed fresh)
```

**Pros:** Data always consistent, no stale reads
**Cons:** Write latency increased, cache may hold unused data

**Best for:** Session data, shopping carts, consistency-critical data

### 3. Write-Behind (Write-Back)

Cache is updated immediately, database asynchronously.

```
Write: Update cache → queue async write to DB
Read:  Always from cache
```

**Pros:** Very fast writes, reduced DB load
**Cons:** Risk of data loss, eventual consistency only

**Best for:** Analytics events, logging, metrics, non-critical counters

## Cache Client Interface

Abstract cache operations for testability and flexibility:

```go
type Client interface {
    ExecBatch(ctx context.Context, name string, reqs ...Req) ([]Res, error)
    WithBatch(size int) Client
}

type Req interface {
    getID() string
    getCmdText() string
    prepareCmd() error
    handlePipe(context.Context, redis.Pipeliner)
    handleCmdr(redis.Cmder) Res
}

type Res interface {
    ID() string
    Val() any
    Err() error
}
```

## Request/Response Operations

### GET with deserialization

```go
func GetObj(key string, obj any) Req {
    return &getReq{id: generateID(), key: key, obj: obj}
}

type getReq struct {
    id   string
    key  string
    obj  any
    cmd  *redis.StringCmd
}

func (r *getReq) handlePipe(ctx context.Context, pipe redis.Pipeliner) {
    r.cmd = pipe.Get(ctx, r.key)
}

func (r *getReq) handleCmdr(cmdr redis.Cmder) Res {
    data, err := r.cmd.Bytes()
    if errors.Is(err, redis.Nil) {
        return &result{id: r.id, val: nil, err: nil} // Cache miss
    }
    if err != nil {
        return &result{id: r.id, val: nil, err: err}
    }
    if err := json.Unmarshal(data, r.obj); err != nil {
        return &result{id: r.id, val: nil, err: err}
    }
    return &result{id: r.id, val: r.obj, err: nil}
}
```

### SET with TTL

```go
func SetObjWithTTL(key string, obj any, ttl time.Duration) Req {
    return &setReq{id: generateID(), key: key, obj: obj, ttl: ttl}
}

type setReq struct {
    id   string
    key  string
    obj  any
    ttl  time.Duration
    data []byte
    cmd  *redis.StatusCmd
}

func (r *setReq) prepareCmd() error {
    data, err := json.Marshal(r.obj)
    if err != nil {
        return fmt.Errorf("marshal object: %w", err)
    }
    r.data = data
    return nil
}

func (r *setReq) handlePipe(ctx context.Context, pipe redis.Pipeliner) {
    r.cmd = pipe.Set(ctx, r.key, r.data, r.ttl)
}
```

### DELETE operations

```go
func DelObj(key string) Req {
    return &delReq{id: generateID(), key: key}
}

// Pattern-based deletion using Lua script
func DelByPattern(pattern string) Req {
    return &delPatternReq{id: generateID(), pattern: pattern}
}
```

## Batching Mechanism

Use Redis Pipeline for efficient batch operations:

```go
type redisClient struct {
    client       *redis.Client
    batchSize    int
    batchTimeout time.Duration
    requests     chan batchItem
}

func (c *redisClient) ExecBatch(ctx context.Context, name string, reqs ...Req) ([]Res, error) {
    // Prepare all requests
    for _, req := range reqs {
        if err := req.prepareCmd(); err != nil {
            return nil, fmt.Errorf("prepare request: %w", err)
        }
    }

    // Execute via pipeline
    pipe := c.client.Pipeline()
    for _, req := range reqs {
        req.handlePipe(ctx, pipe)
    }

    cmds, err := pipe.Exec(ctx)
    if err != nil && !errors.Is(err, redis.Nil) {
        return nil, fmt.Errorf("exec pipeline: %w", err)
    }

    // Collect results
    results := make([]Res, len(reqs))
    for i, req := range reqs {
        results[i] = req.handleCmdr(cmds[i])
    }

    return results, nil
}

func (c *redisClient) WithBatch(size int) Client {
    return &redisClient{
        client:       c.client,
        batchSize:    size,
        batchTimeout: c.batchTimeout,
    }
}
```

### Async Batch Processing

For high-throughput scenarios:

```go
type batchItem struct {
    req    Req
    result chan Res
}

func (c *redisClient) startBatchProcessor(ctx context.Context) {
    go func() {
        batch := make([]batchItem, 0, c.batchSize)
        timer := time.NewTimer(c.batchTimeout)

        for {
            select {
            case item := <-c.requests:
                batch = append(batch, item)
                if len(batch) >= c.batchSize {
                    c.flushBatch(ctx, batch)
                    batch = batch[:0]
                    timer.Reset(c.batchTimeout)
                }
            case <-timer.C:
                if len(batch) > 0 {
                    c.flushBatch(ctx, batch)
                    batch = batch[:0]
                }
                timer.Reset(c.batchTimeout)
            case <-ctx.Done():
                return
            }
        }
    }()
}
```

## ItemFetcher Interface (Cache-Aside)

Generic interface for cache-aside pattern:

```go
type ItemFetcher interface {
    // GetKey returns cache key for item ID
    GetKey(itemID string) string

    // GetNew returns empty instance for unmarshaling
    GetNew() any

    // ToList converts items to typed slice
    ToList(items []any) any

    // GetID extracts ID from item
    GetID(item any) string

    // FetchMissed loads items from database
    FetchMissed(ctx context.Context, missedIDs []string) ([]any, error)
}
```

### Example Implementation

```go
type UserAccountProvider struct {
    svc *services.UserAccountService
}

func (p *UserAccountProvider) GetKey(accountID string) string {
    return "userAccount:" + accountID
}

func (p *UserAccountProvider) GetNew() any {
    return &domain.UserAccount{}
}

func (p *UserAccountProvider) ToList(items []any) any {
    accounts := make([]*domain.UserAccount, len(items))
    for i, item := range items {
        accounts[i] = item.(*domain.UserAccount)
    }
    return accounts
}

func (p *UserAccountProvider) GetID(item any) string {
    return item.(*domain.UserAccount).ID
}

func (p *UserAccountProvider) FetchMissed(ctx context.Context, missedIDs []string) ([]any, error) {
    filter := domain.UserAccountFilter{IDs: missedIDs}
    accounts, err := p.svc.GetAccounts(ctx, &filter)
    if err != nil {
        return nil, err
    }

    items := make([]any, len(accounts))
    for i, acc := range accounts {
        items[i] = acc
    }
    return items, nil
}
```

## CachedItemProvider

Automatic cache-aside with batch support:

```go
type CachedItemProvider struct {
    client   Client
    fetcher  ItemFetcher
    name     string
    ttl      time.Duration
}

func NewCachedItemProvider(client Client, fetcher ItemFetcher, name string, ttl time.Duration) *CachedItemProvider {
    return &CachedItemProvider{
        client:  client,
        fetcher: fetcher,
        name:    name,
        ttl:     ttl,
    }
}

func (p *CachedItemProvider) Fetch(ctx context.Context, itemIDs []string) (any, error) {
    if len(itemIDs) == 0 {
        return p.fetcher.ToList(nil), nil
    }

    // Step 1: Batch GET from cache
    getReqs := make([]Req, len(itemIDs))
    for i, id := range itemIDs {
        getReqs[i] = GetObj(p.fetcher.GetKey(id), p.fetcher.GetNew())
    }

    results, err := p.client.ExecBatch(ctx, "get."+p.name, getReqs...)
    if err != nil {
        return nil, fmt.Errorf("cache get: %w", err)
    }

    // Step 2: Identify hits and misses
    items := make([]any, 0, len(itemIDs))
    missedIDs := make([]string, 0)

    for i, res := range results {
        if res.Err() != nil {
            return nil, fmt.Errorf("cache result: %w", res.Err())
        }
        if res.Val() == nil {
            missedIDs = append(missedIDs, itemIDs[i])
        } else {
            items = append(items, res.Val())
        }
    }

    // Step 3: Fetch misses from database
    if len(missedIDs) > 0 {
        fetchedItems, err := p.fetcher.FetchMissed(ctx, missedIDs)
        if err != nil {
            return nil, fmt.Errorf("fetch missed: %w", err)
        }

        // Step 4: Write back to cache
        setReqs := make([]Req, len(fetchedItems))
        for i, item := range fetchedItems {
            key := p.fetcher.GetKey(p.fetcher.GetID(item))
            setReqs[i] = SetObjWithTTL(key, item, p.ttl)
            items = append(items, item)
        }

        if _, err := p.client.ExecBatch(ctx, "set."+p.name, setReqs...); err != nil {
            // Log but don't fail — cache write is best-effort
            // log.Warn("cache set failed", "error", err)
        }
    }

    return p.fetcher.ToList(items), nil
}
```

## Configuration

```go
type RedisConfig struct {
    Server       string        `envconfig:"REDIS_SERVER" default:"localhost:6379"`
    Database     int           `envconfig:"REDIS_DB" default:"0"`
    Password     string        `envconfig:"REDIS_PASS"`
    BatchTimeout time.Duration `envconfig:"REDIS_BATCH_TIMEOUT" default:"200ms"`
    MaxBatchSize int           `envconfig:"REDIS_MAX_BATCH_SIZE" default:"100"`
}

func NewRedisClient(ctx context.Context, conf *RedisConfig) (Client, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     conf.Server,
        Password: conf.Password,
        DB:       conf.Database,
    })

    if err := client.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("redis ping: %w", err)
    }

    return &redisClient{
        client:       client,
        batchSize:    conf.MaxBatchSize,
        batchTimeout: conf.BatchTimeout,
    }, nil
}
```

## When to Use

| Scenario | Strategy | TTL |
|----------|----------|-----|
| User profiles | Cache-Aside | 5-15 min |
| Session data | Write-Through | Session duration |
| Product catalog | Cache-Aside | 1-5 min |
| Analytics events | Write-Behind | N/A |
| Real-time counters | Write-Through | No TTL |
| Configuration | Cache-Aside | 1 min |
| Rapidly changing data | No cache | — |

## Anti-Patterns

### Don't: No TTL

```go
// WRONG: Data stays forever, becomes stale
SetObj(key, obj) // No TTL!

// CORRECT: Always set TTL
SetObjWithTTL(key, obj, 5*time.Minute)
```

### Don't: Cache Stampede

When cache expires, all requests hit database simultaneously.

```go
// WRONG: All goroutines fetch from DB
if cached == nil {
    data = fetchFromDB() // 100 concurrent calls!
    cache.Set(key, data)
}

// CORRECT: Use singleflight
var group singleflight.Group

result, err, _ := group.Do(key, func() (any, error) {
    return fetchFromDB()
})
```

### Don't: Cache Errors

```go
// WRONG: Caching error response
result, err := fetchFromDB()
cache.Set(key, result) // Even if err != nil!

// CORRECT: Only cache successful results
result, err := fetchFromDB()
if err == nil {
    cache.Set(key, result)
}
```

### Don't: Ignore Cache Failures

```go
// WRONG: Fail entire request on cache error
if err := cache.Set(key, data); err != nil {
    return nil, err // Request fails!
}

// CORRECT: Log and continue
if err := cache.Set(key, data); err != nil {
    log.Warn("cache set failed", "error", err)
}
return data, nil // Request succeeds
```

## Cache Key Naming Convention

| Pattern | Example | Use Case |
|---------|---------|----------|
| `entity:id` | `user:123` | Single entity |
| `entity:id:field` | `user:123:profile` | Entity subset |
| `query:hash` | `users:abc123` | Query result |
| `list:entity` | `list:active_users` | Collection |

```go
// Good
key := fmt.Sprintf("user:%s", userID)
key := fmt.Sprintf("tournament:%s:standings", tournamentID)

// Bad
key := userID                    // No namespace
key := fmt.Sprintf("data:%s", userID) // Generic namespace
```

## Related Patterns

- [Database Pattern](database-pattern.md) — Transaction handling
- [Repository Pattern](repository-pattern.md) — Data access layer
- [Advisory Lock Pattern](advisory-lock-pattern.md) — Concurrency control
