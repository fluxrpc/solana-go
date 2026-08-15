package rpccache

import (
	"sync"
	"sync/atomic"
	"time"

	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/rpc"
)

// entry is one cached account. Data written at a higher slot always wins;
// the streamed and immutable flags are sticky upgrades (a plain refetch
// never demotes an entry fed by a stream).
type entry struct {
	data      *rpc.Account
	slot      uint64
	streamed  bool
	immutable bool
	cachedAt  int64        // unix nanos at last write
	lastRead  atomic.Int64 // unix nanos at last read; atomic so hits skip the write lock
}

// fresh reports whether the entry may be served under the given freshness
// window: streamed entries are kept current by the feed, immutable entries
// never change, and everything else expires after freshFor.
func (e *entry) fresh(now int64, freshFor time.Duration) bool {
	return e.streamed || e.immutable || now-e.cachedAt < int64(freshFor)
}

// cache is a sharded pubkey → account map. Shard count is a power of two;
// the shard index reads the key's tail bytes, which are uniformly
// distributed for both ed25519 keys and PDAs.
type cache struct {
	shards []shard
	mask   uint64
}

type shard struct {
	mu   sync.RWMutex
	data map[solana.PublicKey]*entry
}

func newCache(shardCount int) *cache {
	n := 1
	for n < shardCount {
		n <<= 1
	}
	c := &cache{shards: make([]shard, n), mask: uint64(n - 1)}
	for i := range c.shards {
		c.shards[i].data = make(map[solana.PublicKey]*entry)
	}
	return c
}

func (c *cache) shard(key solana.PublicKey) *shard {
	idx := (uint64(key[24]) | uint64(key[25])<<8 | uint64(key[26])<<16 | uint64(key[27])<<24 |
		uint64(key[28])<<32 | uint64(key[29])<<40 | uint64(key[30])<<48 | uint64(key[31])<<56) & c.mask
	return &c.shards[idx]
}

func (c *cache) get(key solana.PublicKey) *entry {
	s := c.shard(key)
	s.mu.RLock()
	e := s.data[key]
	s.mu.RUnlock()
	return e
}

// set stores data under key. Writes carrying an older slot than the current
// entry keep the newer data but still apply flag upgrades and refresh the
// write time of non-streamed entries.
func (c *cache) set(key solana.PublicKey, data *rpc.Account, slot uint64, streamed, immutable bool) {
	now := time.Now().UnixNano()
	s := c.shard(key)
	s.mu.Lock()
	defer s.mu.Unlock()

	prev, ok := s.data[key]
	if !ok {
		e := &entry{data: data, slot: slot, streamed: streamed, immutable: immutable, cachedAt: now}
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

func (c *cache) del(key solana.PublicKey) {
	s := c.shard(key)
	s.mu.Lock()
	delete(s.data, key)
	s.mu.Unlock()
}

func (c *cache) len() int {
	total := 0
	for i := range c.shards {
		s := &c.shards[i]
		s.mu.RLock()
		total += len(s.data)
		s.mu.RUnlock()
	}
	return total
}

// tidy evicts entries that have not been read within ttl. Immutable entries
// are evicted like any other: they are cheap to refetch and pinning them
// forever would leak.
func (c *cache) tidy(ttl time.Duration) {
	if ttl <= 0 {
		return
	}
	cutoff := time.Now().Add(-ttl).UnixNano()
	for i := range c.shards {
		s := &c.shards[i]
		s.mu.Lock()
		for k, e := range s.data {
			if e.lastRead.Load() < cutoff {
				delete(s.data, k)
			}
		}
		s.mu.Unlock()
	}
}
