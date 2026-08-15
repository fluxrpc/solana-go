package rpc

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	solana "github.com/fluxrpc/solana-go"
)

// Defaults for EnableCache.
const (
	DefaultCacheFreshFor        = 800 * time.Millisecond
	DefaultCacheTTL             = time.Hour
	DefaultCacheJanitorInterval = 5 * time.Minute
	DefaultCacheShards          = 256
)

// CacheOptions configures the account cache beyond the defaults.
type CacheOptions struct {
	// FreshFor is how long a fetched (non-streamed, non-immutable) account
	// is served from cache before it must be refetched.
	// Default DefaultCacheFreshFor.
	FreshFor time.Duration

	// TTL evicts entries not read for this long. Default DefaultCacheTTL.
	TTL time.Duration

	// JanitorInterval is how often eviction runs; <0 disables the janitor.
	// Default DefaultCacheJanitorInterval.
	JanitorInterval time.Duration

	// Shards is the cache shard count, rounded up to a power of two.
	// Default DefaultCacheShards.
	Shards int

	// Commitment is used for cache-miss fetches. Default
	// CommitmentProcessed, matching the freshness expectations of streamed
	// data.
	Commitment CommitmentType
}

// CacheStats is a snapshot of cache effectiveness counters. Hits and
// Misses count individual accounts (one getMultipleAccounts call can
// produce both).
type CacheStats struct {
	Hits    uint64
	Misses  uint64
	Entries int
}

// EnableCache turns on the in-memory account cache for this client (see
// the cache discussion in the package documentation). GetAccountInfo and
// GetMultipleAccounts are then served from cache whenever possible; the
// WithOpts variants always bypass it, since cached entries hold the full
// base64 account shape. Passing nil uses the defaults. Enabling replaces
// any previous cache and its contents.
func (c *Client) EnableCache(opts *CacheOptions) {
	newCacheState := newAccountCache(opts)
	if old := c.cache.Swap(newCacheState); old != nil {
		old.stop()
	}
}

// DisableCache turns the account cache off and discards its contents.
func (c *Client) DisableCache() {
	if old := c.cache.Swap(nil); old != nil {
		old.stop()
	}
}

// CacheStore caches an account under the regular freshness rules, e.g.
// data obtained out of band. slot orders concurrent writes: stale slots
// never overwrite newer data. No-op while the cache is disabled.
func (c *Client) CacheStore(account solana.PublicKey, data *Account, slot uint64) {
	if cache := c.cache.Load(); cache != nil {
		cache.set(account, data, slot, false, false)
	}
}

// CacheStoreImmutable caches an account without expiry (until idle-
// evicted). No-op while the cache is disabled.
func (c *Client) CacheStoreImmutable(account solana.PublicKey, data *Account, slot uint64) {
	if cache := c.cache.Load(); cache != nil {
		cache.set(account, data, slot, false, true)
	}
}

// CacheStoreStreamed caches an account fed by a realtime stream: it is
// served from cache indefinitely, on the premise that the stream keeps it
// current. Feed every update for the account (see the yellowstone
// package's PipeAccounts) or the cache will serve stale data. No-op while
// the cache is disabled.
func (c *Client) CacheStoreStreamed(account solana.PublicKey, data *Account, slot uint64) {
	if cache := c.cache.Load(); cache != nil {
		cache.set(account, data, slot, true, false)
	}
}

// CacheClear removes an account from the cache.
func (c *Client) CacheClear(account solana.PublicKey) {
	if cache := c.cache.Load(); cache != nil {
		cache.del(account)
	}
}

// CacheHas reports whether the account is cached (fresh or not).
func (c *Client) CacheHas(account solana.PublicKey) bool {
	cache := c.cache.Load()
	return cache != nil && cache.get(account) != nil
}

// CacheTidy evicts entries not read within ttl (the configured TTL when
// ttl is 0).
func (c *Client) CacheTidy(ttl time.Duration) {
	if cache := c.cache.Load(); cache != nil {
		if ttl <= 0 {
			ttl = cache.ttl
		}
		cache.tidy(ttl)
	}
}

// CacheStats returns a snapshot of the cache counters; zero while the
// cache is disabled.
func (c *Client) CacheStats() CacheStats {
	cache := c.cache.Load()
	if cache == nil {
		return CacheStats{}
	}
	return CacheStats{
		Hits:    cache.hits.Load(),
		Misses:  cache.misses.Load(),
		Entries: cache.len(),
	}
}

// GetAccountInfoImmutable is GetAccountInfo for accounts whose data never
// changes (e.g. finalized mints, program binaries): a fetched value is
// cached without expiry until evicted for idleness. With the cache
// disabled it behaves exactly like GetAccountInfo.
func (c *Client) GetAccountInfoImmutable(ctx context.Context, account solana.PublicKey) (*GetAccountInfoResult, error) {
	cache := c.cache.Load()
	if cache != nil {
		if data, ok := cache.lookup(account); ok {
			return &GetAccountInfoResult{Value: data}, nil
		}
	}

	result, err := c.getAccountInfoForCache(ctx, account, cache)
	if err != nil {
		return nil, err
	}
	if cache != nil {
		cache.set(account, result.Value, result.Context.Slot, false, true)
	}
	return result, nil
}

// cachedGetAccountInfo is the cache-aware path behind GetAccountInfo.
func (c *Client) cachedGetAccountInfo(ctx context.Context, account solana.PublicKey, cache *accountCache) (*GetAccountInfoResult, error) {
	if data, ok := cache.lookup(account); ok {
		return &GetAccountInfoResult{Value: data}, nil
	}

	result, err := c.getAccountInfoForCache(ctx, account, cache)
	if err != nil {
		return nil, err
	}
	cache.set(account, result.Value, result.Context.Slot, false, false)
	return result, nil
}

func (c *Client) getAccountInfoForCache(ctx context.Context, account solana.PublicKey, cache *accountCache) (*GetAccountInfoResult, error) {
	commitment := CommitmentProcessed
	if cache != nil {
		commitment = cache.commitment
	}
	return c.GetAccountInfoWithOpts(ctx, account, &GetAccountInfoOpts{
		Encoding:   solana.EncodingBase64,
		Commitment: commitment,
	})
}

// cachedGetMultipleAccounts is the cache-aware path behind
// GetMultipleAccounts: it serves what it can from cache and fetches the
// deduplicated misses in one call.
func (c *Client) cachedGetMultipleAccounts(ctx context.Context, accounts []solana.PublicKey, cache *accountCache) (*GetMultipleAccountsResult, error) {
	out := &GetMultipleAccountsResult{
		Value: make([]*Account, len(accounts)),
	}

	now := time.Now().UnixNano()
	var fetchKeys []solana.PublicKey
	positions := make(map[solana.PublicKey][]int)
	hits := uint64(0)
	for i, key := range accounts {
		if data, ok := cache.lookupAt(key, now); ok {
			out.Value[i] = data
			hits++
			continue
		}
		if _, seen := positions[key]; !seen {
			fetchKeys = append(fetchKeys, key)
		}
		positions[key] = append(positions[key], i)
	}
	cache.hits.Add(hits)
	cache.misses.Add(uint64(len(fetchKeys)))

	if len(fetchKeys) == 0 {
		return out, nil
	}

	fetched, err := c.GetMultipleAccountsWithOpts(ctx, fetchKeys, &GetMultipleAccountsOpts{
		Encoding:   solana.EncodingBase64,
		Commitment: cache.commitment,
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
			cache.set(key, value, fetched.Context.Slot, false, false)
		}
	}
	return out, nil
}

// accountCache is the sharded slot-aware account cache behind EnableCache.
type accountCache struct {
	shards     []cacheShard
	mask       uint64
	freshFor   time.Duration
	ttl        time.Duration
	commitment CommitmentType

	hits   atomic.Uint64
	misses atomic.Uint64

	janitorStop chan struct{}
	stopOnce    sync.Once
}

type cacheShard struct {
	mu   sync.RWMutex
	data map[solana.PublicKey]*cacheEntry
}

// cacheEntry is one cached account. Data written at a higher slot always
// wins; the streamed and immutable flags are sticky upgrades (a plain
// refetch never demotes an entry fed by a stream).
type cacheEntry struct {
	data      *Account
	slot      uint64
	streamed  bool
	immutable bool
	cachedAt  int64        // unix nanos at last write
	lastRead  atomic.Int64 // unix nanos at last read; atomic so hits skip the write lock
}

// fresh reports whether the entry may be served: streamed entries are kept
// current by the feed, immutable entries never change, and everything else
// expires after freshFor.
func (e *cacheEntry) fresh(now int64, freshFor time.Duration) bool {
	return e.streamed || e.immutable || now-e.cachedAt < int64(freshFor)
}

func newAccountCache(opts *CacheOptions) *accountCache {
	cache := &accountCache{
		freshFor:   DefaultCacheFreshFor,
		ttl:        DefaultCacheTTL,
		commitment: CommitmentProcessed,
	}
	shards := DefaultCacheShards
	janitor := DefaultCacheJanitorInterval
	if opts != nil {
		if opts.FreshFor > 0 {
			cache.freshFor = opts.FreshFor
		}
		if opts.TTL > 0 {
			cache.ttl = opts.TTL
		}
		if opts.JanitorInterval != 0 {
			janitor = opts.JanitorInterval
		}
		if opts.Shards > 0 {
			shards = opts.Shards
		}
		if opts.Commitment != "" {
			cache.commitment = opts.Commitment
		}
	}

	n := 1
	for n < shards {
		n <<= 1
	}
	cache.shards = make([]cacheShard, n)
	cache.mask = uint64(n - 1)
	for i := range cache.shards {
		cache.shards[i].data = make(map[solana.PublicKey]*cacheEntry)
	}

	if janitor > 0 {
		cache.janitorStop = make(chan struct{})
		go cache.janitor(janitor)
	}
	return cache
}

func (ac *accountCache) janitor(interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			ac.tidy(ac.ttl)
		case <-ac.janitorStop:
			return
		}
	}
}

func (ac *accountCache) stop() {
	ac.stopOnce.Do(func() {
		if ac.janitorStop != nil {
			close(ac.janitorStop)
		}
	})
}

// shard picks by the key's tail bytes, which are uniformly distributed for
// both ed25519 keys and PDAs.
func (ac *accountCache) shard(key solana.PublicKey) *cacheShard {
	idx := (uint64(key[24]) | uint64(key[25])<<8 | uint64(key[26])<<16 | uint64(key[27])<<24 |
		uint64(key[28])<<32 | uint64(key[29])<<40 | uint64(key[30])<<48 | uint64(key[31])<<56) & ac.mask
	return &ac.shards[idx]
}

func (ac *accountCache) get(key solana.PublicKey) *cacheEntry {
	s := ac.shard(key)
	s.mu.RLock()
	e := s.data[key]
	s.mu.RUnlock()
	return e
}

// lookup returns the cached account for key if it is servable, maintaining
// the hit/miss counters and the read clock.
func (ac *accountCache) lookup(key solana.PublicKey) (*Account, bool) {
	data, ok := ac.lookupAt(key, time.Now().UnixNano())
	if ok {
		ac.hits.Add(1)
	} else {
		ac.misses.Add(1)
	}
	return data, ok
}

// lookupAt is lookup with a caller-provided clock and no counter updates,
// for batch paths that check many keys per call.
func (ac *accountCache) lookupAt(key solana.PublicKey, now int64) (*Account, bool) {
	e := ac.get(key)
	if e == nil || !e.fresh(now, ac.freshFor) {
		return nil, false
	}
	e.lastRead.Store(now)
	return e.data, true
}

// set stores data under key. Writes carrying an older slot than the
// current entry keep the newer data but still apply flag upgrades and
// refresh the write time.
func (ac *accountCache) set(key solana.PublicKey, data *Account, slot uint64, streamed, immutable bool) {
	now := time.Now().UnixNano()
	s := ac.shard(key)
	s.mu.Lock()
	defer s.mu.Unlock()

	prev, ok := s.data[key]
	if !ok {
		e := &cacheEntry{data: data, slot: slot, streamed: streamed, immutable: immutable, cachedAt: now}
		e.lastRead.Store(now)
		s.data[key] = e
		return
	}

	prev.streamed = prev.streamed || streamed
	prev.immutable = prev.immutable || immutable
	if slot >= prev.slot {
		prev.data = data
		prev.slot = slot
	}
	prev.cachedAt = now
}

func (ac *accountCache) del(key solana.PublicKey) {
	s := ac.shard(key)
	s.mu.Lock()
	delete(s.data, key)
	s.mu.Unlock()
}

func (ac *accountCache) len() int {
	total := 0
	for i := range ac.shards {
		s := &ac.shards[i]
		s.mu.RLock()
		total += len(s.data)
		s.mu.RUnlock()
	}
	return total
}

// tidy evicts entries that have not been read within ttl. Immutable
// entries are evicted like any other: they are cheap to refetch and
// pinning them forever would leak.
func (ac *accountCache) tidy(ttl time.Duration) {
	if ttl <= 0 {
		return
	}
	cutoff := time.Now().Add(-ttl).UnixNano()
	for i := range ac.shards {
		s := &ac.shards[i]
		s.mu.Lock()
		for k, e := range s.data {
			if e.lastRead.Load() < cutoff {
				delete(s.data, k)
			}
		}
		s.mu.Unlock()
	}
}
