package benchcmp

// End-to-end client throughput: both clients issue real HTTP JSON-RPC calls
// against the same in-process server serving identical canned mainnet-shaped
// responses, so the comparison covers request building, transport reuse and
// response decoding.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	fluxrpc "github.com/fluxrpc/solana-go/rpc"
	gaglrpc "github.com/gagliardetto/solana-go/rpc"
)

func newBenchServer(b *testing.B) *httptest.Server {
	accountResponse := []byte(`{"jsonrpc":"2.0","id":1,"result":{"context":{"slot":341197053},"value":` + rpcAccountFixture + `}}`)
	blockResponse := []byte(`{"jsonrpc":"2.0","id":1,"result":` + rpcGetBlockResultFixture + `}`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string `json:"method"`
		}
		if err := jsonDecode(r, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch req.Method {
		case "getAccountInfo":
			w.Write(accountResponse)
		case "getBlock":
			w.Write(blockResponse)
		default:
			http.Error(w, "unknown method "+req.Method, http.StatusBadRequest)
		}
	}))
	b.Cleanup(server.Close)
	return server
}

func jsonDecode(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func BenchmarkRpcClient_GetAccountInfo(b *testing.B) {
	server := newBenchServer(b)
	ctx := context.Background()

	b.Run("flux", func(b *testing.B) {
		client := fluxrpc.New(server.URL)
		b.ReportAllocs()
		for b.Loop() {
			res, err := client.GetAccountInfo(ctx, fluxPdaMint)
			if err != nil || res.Value == nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("gagl", func(b *testing.B) {
		client := gaglrpc.New(server.URL)
		b.ReportAllocs()
		for b.Loop() {
			res, err := client.GetAccountInfo(ctx, gaglPdaMint)
			if err != nil || res.Value == nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkRpcClient_GetBlock(b *testing.B) {
	server := newBenchServer(b)
	ctx := context.Background()

	b.Run("flux", func(b *testing.B) {
		client := fluxrpc.New(server.URL)
		b.ReportAllocs()
		for b.Loop() {
			res, err := client.GetBlock(ctx, 83987984)
			if err != nil || len(res.Transactions) == 0 {
				b.Fatal(err)
			}
		}
	})
	b.Run("gagl", func(b *testing.B) {
		client := gaglrpc.New(server.URL)
		version := uint64(0)
		opts := &gaglrpc.GetBlockOpts{MaxSupportedTransactionVersion: &version}
		b.ReportAllocs()
		for b.Loop() {
			res, err := client.GetBlockWithOpts(ctx, 83987984, opts)
			if err != nil || len(res.Transactions) == 0 {
				b.Fatal(err)
			}
		}
	})
}
