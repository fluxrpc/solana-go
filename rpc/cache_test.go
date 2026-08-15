package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	solana "github.com/fluxrpc/solana-go"
)

// countingRPC serves getAccountInfo / getMultipleAccounts from an in-memory
// account set and counts every network fetch per key.
type countingRPC struct {
	mu       sync.Mutex
	accounts map[solana.PublicKey]uint64 // pubkey -> lamports; absent = not found
	slot     uint64
	fetches  map[solana.PublicKey]int
}

func (f *countingRPC) fetchCount(key solana.PublicKey) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.fetches[key]
}

func (f *countingRPC) accountJSON(key solana.PublicKey) any {
	lamports, ok := f.accounts[key]
	if !ok {
		return nil
	}
	return map[string]any{
		"lamports":   lamports,
		"owner":      "11111111111111111111111111111111",
		"data":       []any{"dGVzdA==", "base64"},
		"executable": false,
		"rentEpoch":  361,
		"space":      4,
	}
}

func newCountingRPC(t *testing.T) (*countingRPC, *Client) {
	t.Helper()
	f := &countingRPC{
		accounts: map[solana.PublicKey]uint64{},
		fetches:  map[solana.PublicKey]int{},
		slot:     1000,
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     uint64          `json:"id"`
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		var params []json.RawMessage
		json.Unmarshal(req.Params, &params)

		f.mu.Lock()
		var result any
		switch req.Method {
		case "getAccountInfo":
			var key solana.PublicKey
			json.Unmarshal(params[0], &key)
			f.fetches[key]++
			result = map[string]any{
				"context": map[string]any{"slot": f.slot},
				"value":   f.accountJSON(key),
			}
		case "getMultipleAccounts":
			var keys []solana.PublicKey
			json.Unmarshal(params[0], &keys)
			values := make([]any, len(keys))
			for i, key := range keys {
				f.fetches[key]++
				values[i] = f.accountJSON(key)
			}
			result = map[string]any{
				"context": map[string]any{"slot": f.slot},
				"value":   values,
			}
		default:
			t.Errorf("unexpected method %s", req.Method)
		}
		f.mu.Unlock()

		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
	}))
	t.Cleanup(server.Close)
	return f, New(server.URL)
}

func cacheKey(b byte) solana.PublicKey {
	return solana.PublicKey{b}
}

func TestCacheGetAccountInfo(t *testing.T) {
	fake, client := newCountingRPC(t)
	fake.accounts[cacheKey(1)] = 111
	client.EnableCache()
	defer client.DisableCache()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		res, err := client.GetAccountInfo(ctx, cacheKey(1))
		if err != nil {
			t.Fatal(err)
		}
		if res.Value.Lamports != 111 {
			t.Fatalf("lamports = %d", res.Value.Lamports)
		}
	}
	if got := fake.fetchCount(cacheKey(1)); got != 1 {
		t.Fatalf("network fetches = %d, want 1", got)
	}
	stats := client.CacheStats()
	if stats.Hits != 2 || stats.Misses != 1 || stats.Entries != 1 {
		t.Fatalf("stats = %+v", stats)
	}
}

func TestCacheDisabledPassthrough(t *testing.T) {
	fake, client := newCountingRPC(t)
	fake.accounts[cacheKey(1)] = 111
	ctx := context.Background()

	client.GetAccountInfo(ctx, cacheKey(1))
	client.GetAccountInfo(ctx, cacheKey(1))
	if got := fake.fetchCount(cacheKey(1)); got != 2 {
		t.Fatalf("network fetches = %d, want 2 (cache disabled)", got)
	}
	if stats := client.CacheStats(); stats != (CacheStats{}) {
		t.Fatalf("stats = %+v, want zero", stats)
	}

	// Toggling on and off again returns to passthrough.
	client.EnableCache()
	client.GetAccountInfo(ctx, cacheKey(1)) // miss -> fetch -> cache
	client.GetAccountInfo(ctx, cacheKey(1)) // hit
	client.DisableCache()
	client.GetAccountInfo(ctx, cacheKey(1)) // passthrough again
	if got := fake.fetchCount(cacheKey(1)); got != 4 {
		t.Fatalf("network fetches = %d, want 4", got)
	}
}

func TestCacheFreshnessExpiry(t *testing.T) {
	fake, client := newCountingRPC(t)
	fake.accounts[cacheKey(1)] = 111
	client.EnableCacheWithOpts(&CacheOptions{FreshFor: 30 * time.Millisecond, JanitorInterval: -1})
	defer client.DisableCache()
	ctx := context.Background()

	client.GetAccountInfo(ctx, cacheKey(1))
	client.GetAccountInfo(ctx, cacheKey(1)) // hit
	time.Sleep(50 * time.Millisecond)
	client.GetAccountInfo(ctx, cacheKey(1)) // stale -> refetch
	if got := fake.fetchCount(cacheKey(1)); got != 2 {
		t.Fatalf("network fetches = %d, want 2", got)
	}
}

func TestCacheImmutableNeverExpires(t *testing.T) {
	fake, client := newCountingRPC(t)
	fake.accounts[cacheKey(1)] = 111
	client.EnableCacheWithOpts(&CacheOptions{FreshFor: time.Millisecond, JanitorInterval: -1})
	defer client.DisableCache()
	ctx := context.Background()

	client.GetAccountInfoImmutable(ctx, cacheKey(1))
	time.Sleep(10 * time.Millisecond)
	client.GetAccountInfoImmutable(ctx, cacheKey(1))
	client.GetAccountInfo(ctx, cacheKey(1)) // also served: entry is immutable
	if got := fake.fetchCount(cacheKey(1)); got != 1 {
		t.Fatalf("network fetches = %d, want 1", got)
	}
}

func TestCacheStreamedSlotOrdering(t *testing.T) {
	_, client := newCountingRPC(t)
	client.EnableCacheWithOpts(&CacheOptions{JanitorInterval: -1})
	defer client.DisableCache()
	ctx := context.Background()

	client.CacheStoreStreamed(cacheKey(1), &Account{Lamports: 10}, 100)

	// An older write (e.g. a late RPC response) must not clobber the data.
	client.CacheStore(cacheKey(1), &Account{Lamports: 5}, 90)
	res, err := client.GetAccountInfo(ctx, cacheKey(1))
	if err != nil {
		t.Fatal(err)
	}
	if res.Value.Lamports != 10 {
		t.Fatalf("lamports = %d, want 10 (older slot overwrote)", res.Value.Lamports)
	}

	// A newer streamed update wins.
	client.CacheStoreStreamed(cacheKey(1), &Account{Lamports: 20}, 101)
	res, _ = client.GetAccountInfo(ctx, cacheKey(1))
	if res.Value.Lamports != 20 {
		t.Fatalf("lamports = %d, want 20", res.Value.Lamports)
	}

	// The streamed flag survived the older plain store (sticky upgrade):
	// no freshness window applies, so no refetch should ever happen.
	if misses := client.CacheStats().Misses; misses != 0 {
		t.Fatalf("misses = %d, want 0", misses)
	}
}

func TestCacheGetMultipleAccountsPartialHit(t *testing.T) {
	fake, client := newCountingRPC(t)
	fake.accounts[cacheKey(1)] = 1
	fake.accounts[cacheKey(2)] = 2
	client.EnableCacheWithOpts(&CacheOptions{JanitorInterval: -1})
	defer client.DisableCache()
	ctx := context.Background()

	// Prime key 1 in cache.
	if _, err := client.GetAccountInfo(ctx, cacheKey(1)); err != nil {
		t.Fatal(err)
	}

	// key 1: cached; key 2: fetched once despite duplication; key 3: missing.
	res, err := client.GetMultipleAccounts(ctx, cacheKey(1), cacheKey(2), cacheKey(2), cacheKey(3))
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Value) != 4 {
		t.Fatalf("len = %d", len(res.Value))
	}
	if res.Value[0].Lamports != 1 || res.Value[1].Lamports != 2 || res.Value[2].Lamports != 2 {
		t.Fatalf("values = %+v", res.Value)
	}
	if res.Value[3] != nil {
		t.Fatalf("missing account = %+v", res.Value[3])
	}
	if fake.fetchCount(cacheKey(1)) != 1 || fake.fetchCount(cacheKey(2)) != 1 || fake.fetchCount(cacheKey(3)) != 1 {
		t.Fatalf("fetches = %d/%d/%d", fake.fetchCount(cacheKey(1)), fake.fetchCount(cacheKey(2)), fake.fetchCount(cacheKey(3)))
	}

	// Everything cached now except the missing account (nil is not cached).
	if _, err := client.GetMultipleAccounts(ctx, cacheKey(1), cacheKey(2), cacheKey(3)); err != nil {
		t.Fatal(err)
	}
	if fake.fetchCount(cacheKey(3)) != 2 {
		t.Fatalf("missing account fetches = %d, want 2", fake.fetchCount(cacheKey(3)))
	}
	if fake.fetchCount(cacheKey(2)) != 1 {
		t.Fatalf("cached account refetched")
	}
}

func TestCacheNotFoundNotCached(t *testing.T) {
	fake, client := newCountingRPC(t)
	client.EnableCacheWithOpts(&CacheOptions{JanitorInterval: -1})
	defer client.DisableCache()
	ctx := context.Background()

	if _, err := client.GetAccountInfo(ctx, cacheKey(9)); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
	if _, err := client.GetAccountInfo(ctx, cacheKey(9)); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
	if got := fake.fetchCount(cacheKey(9)); got != 2 {
		t.Fatalf("fetches = %d, want 2 (negative results must not cache)", got)
	}
}

func TestCacheClearAndTidy(t *testing.T) {
	_, client := newCountingRPC(t)
	client.EnableCacheWithOpts(&CacheOptions{JanitorInterval: -1})
	defer client.DisableCache()

	client.CacheStoreStreamed(cacheKey(1), &Account{Lamports: 1}, 1)
	client.CacheStoreImmutable(cacheKey(2), &Account{Lamports: 2}, 1)
	if !client.CacheHas(cacheKey(1)) || !client.CacheHas(cacheKey(2)) {
		t.Fatal("entries missing")
	}

	client.CacheClear(cacheKey(1))
	if client.CacheHas(cacheKey(1)) {
		t.Fatal("CacheClear did not remove the entry")
	}

	// Tidy evicts idle entries, immutable included.
	time.Sleep(10 * time.Millisecond)
	client.CacheTidy(time.Millisecond)
	if client.CacheHas(cacheKey(2)) {
		t.Fatal("CacheTidy did not evict the idle entry")
	}
}

var benchmarkCachedAccount *GetAccountInfoResult

func BenchmarkCacheGetAccountInfoHit(b *testing.B) {
	client := New("http://unused")
	client.EnableCacheWithOpts(&CacheOptions{JanitorInterval: -1})
	defer client.DisableCache()
	client.CacheStoreStreamed(cacheKey(1), &Account{Lamports: 1}, 1)
	ctx := context.Background()

	b.ReportAllocs()
	for b.Loop() {
		var err error
		benchmarkCachedAccount, err = client.GetAccountInfo(ctx, cacheKey(1))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCacheStoreStreamed(b *testing.B) {
	client := New("http://unused")
	client.EnableCacheWithOpts(&CacheOptions{JanitorInterval: -1})
	defer client.DisableCache()
	account := &Account{Lamports: 1}

	b.ReportAllocs()
	slot := uint64(0)
	for b.Loop() {
		slot++
		client.CacheStoreStreamed(cacheKey(byte(slot)), account, slot)
	}
}

func BenchmarkCacheGetMultipleAccountsAllHit(b *testing.B) {
	client := New("http://unused")
	client.EnableCacheWithOpts(&CacheOptions{JanitorInterval: -1})
	defer client.DisableCache()
	keys := make([]solana.PublicKey, 100)
	for i := range keys {
		keys[i] = cacheKey(byte(i + 1))
		client.CacheStoreStreamed(keys[i], &Account{Lamports: uint64(i)}, 1)
	}
	ctx := context.Background()

	b.ReportAllocs()
	for b.Loop() {
		res, err := client.GetMultipleAccounts(ctx, keys...)
		if err != nil || len(res.Value) != 100 {
			b.Fatal(err)
		}
	}
}
