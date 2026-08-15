package rpccache

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
	"github.com/fluxrpc/solana-go/rpc"
)

// fakeRPC serves getAccountInfo / getMultipleAccounts from an in-memory
// account set and counts every network fetch per key.
type fakeRPC struct {
	mu       sync.Mutex
	accounts map[solana.PublicKey]uint64 // pubkey -> lamports; absent = not found
	slot     uint64
	fetches  map[solana.PublicKey]int
	calls    int
}

func (f *fakeRPC) fetchCount(key solana.PublicKey) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.fetches[key]
}

func (f *fakeRPC) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func (f *fakeRPC) accountJSON(key solana.PublicKey) any {
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

func newFakeRPC(t *testing.T) (*fakeRPC, *rpc.Client) {
	t.Helper()
	f := &fakeRPC{
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
		f.calls++
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
	return f, rpc.New(server.URL)
}

func key(b byte) solana.PublicKey {
	return solana.PublicKey{b}
}

func TestGetAccountInfoCaching(t *testing.T) {
	fake, rpcClient := newFakeRPC(t)
	fake.accounts[key(1)] = 111
	client := New(rpcClient, nil)
	defer client.Close()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		res, err := client.GetAccountInfo(ctx, key(1))
		if err != nil {
			t.Fatal(err)
		}
		if res.Value.Lamports != 111 {
			t.Fatalf("lamports = %d", res.Value.Lamports)
		}
	}
	if got := fake.fetchCount(key(1)); got != 1 {
		t.Fatalf("network fetches = %d, want 1", got)
	}
	stats := client.Stats()
	if stats.Hits != 2 || stats.Misses != 1 || stats.Entries != 1 {
		t.Fatalf("stats = %+v", stats)
	}
}

func TestFreshnessExpiry(t *testing.T) {
	fake, rpcClient := newFakeRPC(t)
	fake.accounts[key(1)] = 111
	client := New(rpcClient, &Options{FreshFor: 30 * time.Millisecond, JanitorInterval: -1})
	defer client.Close()
	ctx := context.Background()

	client.GetAccountInfo(ctx, key(1))
	client.GetAccountInfo(ctx, key(1)) // hit
	time.Sleep(50 * time.Millisecond)
	client.GetAccountInfo(ctx, key(1)) // stale -> refetch
	if got := fake.fetchCount(key(1)); got != 2 {
		t.Fatalf("network fetches = %d, want 2", got)
	}
}

func TestImmutableNeverExpires(t *testing.T) {
	fake, rpcClient := newFakeRPC(t)
	fake.accounts[key(1)] = 111
	client := New(rpcClient, &Options{FreshFor: time.Millisecond, JanitorInterval: -1})
	defer client.Close()
	ctx := context.Background()

	client.GetAccountInfoImmutable(ctx, key(1))
	time.Sleep(10 * time.Millisecond)
	client.GetAccountInfoImmutable(ctx, key(1))
	client.GetAccountInfo(ctx, key(1)) // also served: entry is immutable
	if got := fake.fetchCount(key(1)); got != 1 {
		t.Fatalf("network fetches = %d, want 1", got)
	}
}

func TestStreamedSlotOrdering(t *testing.T) {
	_, rpcClient := newFakeRPC(t)
	client := New(rpcClient, &Options{JanitorInterval: -1})
	defer client.Close()
	ctx := context.Background()

	client.StoreStreamed(key(1), &rpc.Account{Lamports: 10}, 100)

	// An older write (e.g. a late RPC response) must not clobber the data.
	client.Store(key(1), &rpc.Account{Lamports: 5}, 90)
	res, err := client.GetAccountInfo(ctx, key(1))
	if err != nil {
		t.Fatal(err)
	}
	if res.Value.Lamports != 10 {
		t.Fatalf("lamports = %d, want 10 (older slot overwrote)", res.Value.Lamports)
	}

	// A newer streamed update wins.
	client.StoreStreamed(key(1), &rpc.Account{Lamports: 20}, 101)
	res, _ = client.GetAccountInfo(ctx, key(1))
	if res.Value.Lamports != 20 {
		t.Fatalf("lamports = %d, want 20", res.Value.Lamports)
	}

	// The streamed flag survived the older plain Store (sticky upgrade):
	// no freshness window applies, so no refetch should ever happen.
	if calls := client.Stats().Misses; calls != 0 {
		t.Fatalf("misses = %d, want 0", calls)
	}
}

func TestGetMultipleAccountsPartialHit(t *testing.T) {
	fake, rpcClient := newFakeRPC(t)
	fake.accounts[key(1)] = 1
	fake.accounts[key(2)] = 2
	client := New(rpcClient, &Options{JanitorInterval: -1})
	defer client.Close()
	ctx := context.Background()

	// Prime key 1 in cache.
	if _, err := client.GetAccountInfo(ctx, key(1)); err != nil {
		t.Fatal(err)
	}

	// key(1): cached; key(2): fetched once despite duplication; key(3): missing.
	res, err := client.GetMultipleAccounts(ctx, key(1), key(2), key(2), key(3))
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
	if fake.fetchCount(key(1)) != 1 || fake.fetchCount(key(2)) != 1 || fake.fetchCount(key(3)) != 1 {
		t.Fatalf("fetches = %d/%d/%d", fake.fetchCount(key(1)), fake.fetchCount(key(2)), fake.fetchCount(key(3)))
	}

	// Everything cached now except the missing account (nil is not cached).
	res, err = client.GetMultipleAccounts(ctx, key(1), key(2), key(3))
	if err != nil {
		t.Fatal(err)
	}
	if fake.fetchCount(key(3)) != 2 {
		t.Fatalf("missing account fetches = %d, want 2", fake.fetchCount(key(3)))
	}
	if fake.fetchCount(key(2)) != 1 {
		t.Fatalf("cached account refetched")
	}
	_ = res
}

func TestNotFoundNotCached(t *testing.T) {
	fake, rpcClient := newFakeRPC(t)
	client := New(rpcClient, &Options{JanitorInterval: -1})
	defer client.Close()
	ctx := context.Background()

	if _, err := client.GetAccountInfo(ctx, key(9)); !errors.Is(err, rpc.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
	if _, err := client.GetAccountInfo(ctx, key(9)); !errors.Is(err, rpc.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
	if got := fake.fetchCount(key(9)); got != 2 {
		t.Fatalf("fetches = %d, want 2 (negative results must not cache)", got)
	}
}

func TestClearAndTidy(t *testing.T) {
	_, rpcClient := newFakeRPC(t)
	client := New(rpcClient, &Options{JanitorInterval: -1})
	defer client.Close()

	client.StoreStreamed(key(1), &rpc.Account{Lamports: 1}, 1)
	client.StoreImmutable(key(2), &rpc.Account{Lamports: 2}, 1)
	if !client.Has(key(1)) || !client.Has(key(2)) {
		t.Fatal("entries missing")
	}

	client.Clear(key(1))
	if client.Has(key(1)) {
		t.Fatal("Clear did not remove the entry")
	}

	// Tidy evicts idle entries, immutable included.
	time.Sleep(10 * time.Millisecond)
	client.Tidy(time.Millisecond)
	if client.Has(key(2)) {
		t.Fatal("Tidy did not evict the idle entry")
	}
}

var benchmarkAccountResult *rpc.GetAccountInfoResult

func BenchmarkGetAccountInfoCacheHit(b *testing.B) {
	client := New(nil, &Options{JanitorInterval: -1})
	defer client.Close()
	client.StoreStreamed(key(1), &rpc.Account{Lamports: 1}, 1)
	ctx := context.Background()

	b.ReportAllocs()
	for b.Loop() {
		var err error
		benchmarkAccountResult, err = client.GetAccountInfo(ctx, key(1))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStoreStreamed(b *testing.B) {
	client := New(nil, &Options{JanitorInterval: -1})
	defer client.Close()
	account := &rpc.Account{Lamports: 1}

	b.ReportAllocs()
	slot := uint64(0)
	for b.Loop() {
		slot++
		client.StoreStreamed(key(byte(slot)), account, slot)
	}
}

func BenchmarkGetMultipleAccountsAllHit(b *testing.B) {
	client := New(nil, &Options{JanitorInterval: -1})
	defer client.Close()
	keys := make([]solana.PublicKey, 100)
	for i := range keys {
		keys[i] = key(byte(i + 1))
		client.StoreStreamed(keys[i], &rpc.Account{Lamports: uint64(i)}, 1)
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
