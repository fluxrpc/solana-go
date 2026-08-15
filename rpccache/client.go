package rpccache

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/rpc"
)

// Defaults for New.
const (
	DefaultFreshFor        = 800 * time.Millisecond
	DefaultTTL             = time.Hour
	DefaultJanitorInterval = 5 * time.Minute
	DefaultShards          = 256
)

// Options configures a Client beyond the defaults.
type Options struct {
	// FreshFor is how long a fetched (non-streamed, non-immutable) account
	// is served from cache before it must be refetched.
	// Default DefaultFreshFor.
	FreshFor time.Duration

	// TTL evicts entries not read for this long. Default DefaultTTL.
	TTL time.Duration

	// JanitorInterval is how often eviction runs; <0 disables the janitor.
	// Default DefaultJanitorInterval.
	JanitorInterval time.Duration

	// Shards is the cache shard count, rounded up to a power of two.
	// Default DefaultShards.
	Shards int

	// Commitment is used for cache-miss fetches. Default CommitmentProcessed,
	// matching the freshness expectations of streamed data.
	Commitment rpc.CommitmentType
}

// Stats is a snapshot of cache effectiveness counters. Hits and Misses
// count individual accounts (one getMultipleAccounts call can produce
// both).
type Stats struct {
	Hits    uint64
	Misses  uint64
	Entries int
}

// Client serves account lookups from an in-memory cache backed by an
// rpc.Client for misses. Pipe realtime account updates into the cache (see
// StoreStreamed and the yellowstone package's PipeAccounts) and reads for
// locally-tracked accounts never leave the process.
//
// Accounts returned from the cache are shared between callers: treat them
// as read-only.
type Client struct {
	rpc        *rpc.Client
	cache      *cache
	freshFor   time.Duration
	ttl        time.Duration
	commitment rpc.CommitmentType

	hits   atomic.Uint64
	misses atomic.Uint64

	janitorStop chan struct{}
	closeOnce   sync.Once
}

// New wraps an rpc.Client with an account cache.
func New(rpcClient *rpc.Client, opts *Options) *Client {
	c := &Client{
		rpc:        rpcClient,
		freshFor:   DefaultFreshFor,
		ttl:        DefaultTTL,
		commitment: rpc.CommitmentProcessed,
	}
	shards := DefaultShards
	janitor := DefaultJanitorInterval
	if opts != nil {
		if opts.FreshFor > 0 {
			c.freshFor = opts.FreshFor
		}
		if opts.TTL > 0 {
			c.ttl = opts.TTL
		}
		if opts.JanitorInterval != 0 {
			janitor = opts.JanitorInterval
		}
		if opts.Shards > 0 {
			shards = opts.Shards
		}
		if opts.Commitment != "" {
			c.commitment = opts.Commitment
		}
	}
	c.cache = newCache(shards)

	if janitor > 0 {
		c.janitorStop = make(chan struct{})
		go c.janitor(janitor)
	}
	return c
}

func (c *Client) janitor(interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			c.cache.tidy(c.ttl)
		case <-c.janitorStop:
			return
		}
	}
}

// Close stops the janitor. It does not close the underlying rpc.Client.
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		if c.janitorStop != nil {
			close(c.janitorStop)
		}
	})
}

// Raw returns the underlying rpc.Client for uncached calls.
func (c *Client) Raw() *rpc.Client {
	return c.rpc
}

// lookup returns the cached account for key if it is servable.
func (c *Client) lookup(key solana.PublicKey, now int64) (*rpc.Account, bool) {
	e := c.cache.get(key)
	if e == nil || !e.fresh(now, c.freshFor) {
		return nil, false
	}
	e.lastRead.Store(now)
	return e.data, true
}

// GetAccountInfo returns the account, served from cache when it is fresh,
// streamed or immutable. Misses fetch via the underlying client and
// populate the cache. Returns rpc.ErrNotFound when the account does not
// exist (negative results are not cached).
func (c *Client) GetAccountInfo(ctx context.Context, account solana.PublicKey) (*rpc.GetAccountInfoResult, error) {
	if data, ok := c.lookup(account, time.Now().UnixNano()); ok {
		c.hits.Add(1)
		return &rpc.GetAccountInfoResult{Value: data}, nil
	}
	c.misses.Add(1)

	result, err := c.rpc.GetAccountInfoWithOpts(ctx, account, &rpc.GetAccountInfoOpts{
		Encoding:   solana.EncodingBase64,
		Commitment: c.commitment,
	})
	if err != nil {
		return nil, err
	}
	c.cache.set(account, result.Value, result.Context.Slot, false, false)
	return result, nil
}

// GetAccountInfoImmutable is GetAccountInfo for accounts whose data never
// changes (e.g. finalized mints, program binaries): a fetched value is
// cached without expiry until evicted for idleness.
func (c *Client) GetAccountInfoImmutable(ctx context.Context, account solana.PublicKey) (*rpc.GetAccountInfoResult, error) {
	if data, ok := c.lookup(account, time.Now().UnixNano()); ok {
		c.hits.Add(1)
		return &rpc.GetAccountInfoResult{Value: data}, nil
	}
	c.misses.Add(1)

	result, err := c.rpc.GetAccountInfoWithOpts(ctx, account, &rpc.GetAccountInfoOpts{
		Encoding:   solana.EncodingBase64,
		Commitment: c.commitment,
	})
	if err != nil {
		return nil, err
	}
	c.cache.set(account, result.Value, result.Context.Slot, false, true)
	return result, nil
}

// GetMultipleAccounts returns the accounts in the order requested, serving
// as many as possible from cache and fetching the rest in a single
// getMultipleAccounts call. Duplicate keys are fetched once. Missing
// accounts are nil entries, as in the plain RPC method (and are not
// cached).
func (c *Client) GetMultipleAccounts(ctx context.Context, accounts ...solana.PublicKey) (*rpc.GetMultipleAccountsResult, error) {
	out := &rpc.GetMultipleAccountsResult{
		Value: make([]*rpc.Account, len(accounts)),
	}

	now := time.Now().UnixNano()
	var fetchKeys []solana.PublicKey
	positions := make(map[solana.PublicKey][]int)
	hits := uint64(0)
	for i, key := range accounts {
		if data, ok := c.lookup(key, now); ok {
			out.Value[i] = data
			hits++
			continue
		}
		if _, seen := positions[key]; !seen {
			fetchKeys = append(fetchKeys, key)
		}
		positions[key] = append(positions[key], i)
	}
	c.hits.Add(hits)
	c.misses.Add(uint64(len(fetchKeys)))

	if len(fetchKeys) == 0 {
		return out, nil
	}

	fetched, err := c.rpc.GetMultipleAccountsWithOpts(ctx, fetchKeys, &rpc.GetMultipleAccountsOpts{
		Encoding:   solana.EncodingBase64,
		Commitment: c.commitment,
	})
	if err != nil {
		return nil, err
	}
	out.RPCContext = fetched.RPCContext

	for i, value := range fetched.Value {
		key := fetchKeys[i]
		for _, idx := range positions[key] {
			out.Value[idx] = value
		}
		if value != nil {
			c.cache.set(key, value, fetched.Context.Slot, false, false)
		}
	}
	return out, nil
}

// Store caches an account under the regular freshness rules, e.g. data
// obtained out of band. slot orders concurrent writes: stale slots never
// overwrite newer data.
func (c *Client) Store(account solana.PublicKey, data *rpc.Account, slot uint64) {
	c.cache.set(account, data, slot, false, false)
}

// StoreImmutable caches an account without expiry (until idle-evicted).
func (c *Client) StoreImmutable(account solana.PublicKey, data *rpc.Account, slot uint64) {
	c.cache.set(account, data, slot, false, true)
}

// StoreStreamed caches an account fed by a realtime stream: it is served
// from cache indefinitely, on the premise that the stream keeps it current.
// Feed every update for the account (see yellowstone.PipeAccounts) or the
// cache will serve stale data.
func (c *Client) StoreStreamed(account solana.PublicKey, data *rpc.Account, slot uint64) {
	c.cache.set(account, data, slot, true, false)
}

// Clear removes an account from the cache.
func (c *Client) Clear(account solana.PublicKey) {
	c.cache.del(account)
}

// Has reports whether the account is cached (fresh or not).
func (c *Client) Has(account solana.PublicKey) bool {
	return c.cache.get(account) != nil
}

// Tidy evicts entries not read within ttl (the configured TTL when 0).
func (c *Client) Tidy(ttl time.Duration) {
	if ttl <= 0 {
		ttl = c.ttl
	}
	c.cache.tidy(ttl)
}

// Stats returns a snapshot of the cache counters.
func (c *Client) Stats() Stats {
	return Stats{
		Hits:    c.hits.Load(),
		Misses:  c.misses.Load(),
		Entries: c.cache.len(),
	}
}
