package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// =============================================================================
// Cache Client Interface
// =============================================================================

// Client defines cache operations with batch support.
type Client interface {
	ExecBatch(ctx context.Context, name string, reqs ...Req) ([]Res, error)
	WithBatch(size int) Client
}

// Req represents a cache request.
type Req interface {
	getID() string
	prepareCmd() error
	handlePipe(context.Context, redis.Pipeliner)
	handleCmdr(redis.Cmder) Res
}

// Res represents a cache response.
type Res interface {
	ID() string
	Val() any
	Err() error
}

// =============================================================================
// Redis Client Implementation
// =============================================================================

type RedisConfig struct {
	Server       string
	Database     int
	Password     string
	BatchTimeout time.Duration
	MaxBatchSize int
}

type redisClient struct {
	client       *redis.Client
	batchSize    int
	batchTimeout time.Duration
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

func (c *redisClient) ExecBatch(ctx context.Context, name string, reqs ...Req) ([]Res, error) {
	if len(reqs) == 0 {
		return nil, nil
	}

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

// =============================================================================
// Request/Response Types
// =============================================================================

type result struct {
	id  string
	val any
	err error
}

func (r *result) ID() string { return r.id }
func (r *result) Val() any   { return r.val }
func (r *result) Err() error { return r.err }

// GetObj creates a GET request with JSON deserialization.
func GetObj(key string, obj any) Req {
	return &getReq{id: generateID(), key: key, obj: obj}
}

type getReq struct {
	id  string
	key string
	obj any
	cmd *redis.StringCmd
}

func (r *getReq) getID() string                                    { return r.id }
func (r *getReq) prepareCmd() error                                { return nil }
func (r *getReq) handlePipe(ctx context.Context, pipe redis.Pipeliner) { r.cmd = pipe.Get(ctx, r.key) }
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

// SetObjWithTTL creates a SET request with TTL.
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

func (r *setReq) getID() string { return r.id }
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
func (r *setReq) handleCmdr(cmdr redis.Cmder) Res {
	return &result{id: r.id, val: nil, err: r.cmd.Err()}
}

// DelObj creates a DELETE request.
func DelObj(key string) Req {
	return &delReq{id: generateID(), key: key}
}

type delReq struct {
	id  string
	key string
	cmd *redis.IntCmd
}

func (r *delReq) getID() string                                    { return r.id }
func (r *delReq) prepareCmd() error                                { return nil }
func (r *delReq) handlePipe(ctx context.Context, pipe redis.Pipeliner) { r.cmd = pipe.Del(ctx, r.key) }
func (r *delReq) handleCmdr(cmdr redis.Cmder) Res {
	return &result{id: r.id, val: r.cmd.Val(), err: r.cmd.Err()}
}

func generateID() string {
	// Use UUID or similar in production
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// =============================================================================
// Cache-Aside Pattern: ItemFetcher Interface
// =============================================================================

// ItemFetcher provides cache-aside operations for a specific entity type.
type ItemFetcher interface {
	GetKey(itemID string) string
	GetNew() any
	ToList(items []any) any
	GetID(item any) string
	FetchMissed(ctx context.Context, missedIDs []string) ([]any, error)
}

// CachedItemProvider implements cache-aside pattern.
type CachedItemProvider struct {
	client  Client
	fetcher ItemFetcher
	name    string
	ttl     time.Duration
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

		// Step 4: Write back to cache (best-effort)
		setReqs := make([]Req, len(fetchedItems))
		for i, item := range fetchedItems {
			key := p.fetcher.GetKey(p.fetcher.GetID(item))
			setReqs[i] = SetObjWithTTL(key, item, p.ttl)
			items = append(items, item)
		}

		// Don't fail on cache write errors
		_, _ = p.client.ExecBatch(ctx, "set."+p.name, setReqs...)
	}

	return p.fetcher.ToList(items), nil
}

// =============================================================================
// Example: UserAccount Provider
// =============================================================================

type UserAccount struct {
	ID    string
	Name  string
	Email string
}

type UserAccountService interface {
	GetByIDs(ctx context.Context, ids []string) ([]*UserAccount, error)
}

type UserAccountProvider struct {
	svc UserAccountService
}

func NewUserAccountProvider(svc UserAccountService) *UserAccountProvider {
	return &UserAccountProvider{svc: svc}
}

func (p *UserAccountProvider) GetKey(accountID string) string {
	return "userAccount:" + accountID
}

func (p *UserAccountProvider) GetNew() any {
	return &UserAccount{}
}

func (p *UserAccountProvider) ToList(items []any) any {
	accounts := make([]*UserAccount, len(items))
	for i, item := range items {
		accounts[i] = item.(*UserAccount)
	}
	return accounts
}

func (p *UserAccountProvider) GetID(item any) string {
	return item.(*UserAccount).ID
}

func (p *UserAccountProvider) FetchMissed(ctx context.Context, missedIDs []string) ([]any, error) {
	accounts, err := p.svc.GetByIDs(ctx, missedIDs)
	if err != nil {
		return nil, err
	}

	items := make([]any, len(accounts))
	for i, acc := range accounts {
		items[i] = acc
	}
	return items, nil
}

// =============================================================================
// Usage Example
// =============================================================================

func ExampleUsage(ctx context.Context, cacheClient Client, userSvc UserAccountService) {
	// Create cached provider
	userProvider := NewCachedItemProvider(
		cacheClient.WithBatch(10),
		NewUserAccountProvider(userSvc),
		"users",
		5*time.Minute,
	)

	// Fetch users (cache-aside pattern)
	userIDs := []string{"user1", "user2", "user3"}
	result, err := userProvider.Fetch(ctx, userIDs)
	if err != nil {
		// Handle error
		return
	}

	users := result.([]*UserAccount)
	for _, user := range users {
		fmt.Printf("User: %s (%s)\n", user.Name, user.Email)
	}
}
