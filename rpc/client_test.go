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

// testServer records incoming JSON-RPC requests and serves canned results.
type testServer struct {
	t       *testing.T
	mu      sync.Mutex
	lastReq struct {
		Method string          `json:"method"`
		Params json.RawMessage `json:"params"`
		ID     uint64          `json:"id"`
	}
	result any
	rpcErr *RPCError
	status int
	delay  time.Duration
}

func newTestClient(t *testing.T) (*Client, *testServer) {
	t.Helper()
	ts := &testServer{t: t, status: http.StatusOK}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ts.mu.Lock()
		defer ts.mu.Unlock()
		if err := json.NewDecoder(r.Body).Decode(&ts.lastReq); err != nil {
			t.Errorf("decoding request: %v", err)
		}
		if ts.delay > 0 {
			time.Sleep(ts.delay)
		}
		w.WriteHeader(ts.status)
		envelope := map[string]any{"jsonrpc": "2.0", "id": ts.lastReq.ID}
		if ts.rpcErr != nil {
			envelope["error"] = ts.rpcErr
		} else {
			envelope["result"] = ts.result
		}
		if err := json.NewEncoder(w).Encode(envelope); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	return New(server.URL), ts
}

func (ts *testServer) serveRaw(t *testing.T, raw string) {
	t.Helper()
	var result any
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		t.Fatal(err)
	}
	ts.result = result
}

func (ts *testServer) params(t *testing.T) []json.RawMessage {
	t.Helper()
	var params []json.RawMessage
	if err := json.Unmarshal(ts.lastReq.Params, &params); err != nil {
		t.Fatal(err)
	}
	return params
}

func TestClientGetAccountInfo(t *testing.T) {
	client, ts := newTestClient(t)
	ts.serveRaw(t, `{"context":{"slot":341197053},"value":`+string(accountFixture)+`}`)

	res, err := client.GetAccountInfo(context.Background(), solana.MustPublicKeyFromBase58(usdcMint))
	if err != nil {
		t.Fatal(err)
	}
	if res.Value.Lamports != 88849814690250 {
		t.Fatalf("got %+v", res.Value)
	}

	if ts.lastReq.Method != "getAccountInfo" {
		t.Fatalf("method = %s", ts.lastReq.Method)
	}
	params := ts.params(t)
	if len(params) != 2 {
		t.Fatalf("params = %s", ts.lastReq.Params)
	}
	if string(params[0]) != `"`+usdcMint+`"` {
		t.Fatalf("params[0] = %s", params[0])
	}
	if string(params[1]) != `{"encoding":"base64"}` {
		t.Fatalf("params[1] = %s", params[1])
	}
}

func TestClientNotFound(t *testing.T) {
	client, ts := newTestClient(t)
	ctx := context.Background()

	ts.result = nil // JSON null result
	if _, err := client.GetTransaction(ctx, solana.Signature{1}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetTransaction err = %v", err)
	}
	if _, err := client.GetBlock(ctx, 1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetBlock err = %v", err)
	}

	ts.serveRaw(t, `{"context":{"slot":1},"value":null}`)
	if _, err := client.GetAccountInfo(ctx, solana.PublicKey{1}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetAccountInfo err = %v", err)
	}
}

func TestClientRPCError(t *testing.T) {
	client, ts := newTestClient(t)
	ts.rpcErr = &RPCError{Code: -32601, Message: "nope"}

	_, err := client.GetSlot(context.Background(), CommitmentFinalized)
	var rpcErr *RPCError
	if !errors.As(err, &rpcErr) || rpcErr.Code != -32601 {
		t.Fatalf("err = %v", err)
	}
}

func TestClientHTTPError(t *testing.T) {
	client, ts := newTestClient(t)
	ts.status = http.StatusBadGateway

	if _, err := client.GetSlot(context.Background(), ""); err == nil {
		t.Fatal("expected error for http 502")
	}
}

func TestClientGetBlockDefaults(t *testing.T) {
	client, ts := newTestClient(t)
	ts.serveRaw(t, getBlockResultFixture)

	res, err := client.GetBlock(context.Background(), 83987984)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Transactions) != 2 {
		t.Fatalf("got %d transactions", len(res.Transactions))
	}

	params := ts.params(t)
	if string(params[0]) != "83987984" {
		t.Fatalf("params[0] = %s", params[0])
	}
	if string(params[1]) != `{"maxSupportedTransactionVersion":0}` {
		t.Fatalf("params[1] = %s", params[1])
	}
}

func TestClientSendTransaction(t *testing.T) {
	client, ts := newTestClient(t)
	key := testPrivateKey(t)
	tx := testUnsignedTransactionRPC(key)
	sig := tx.Signatures[0]
	ts.result = sig.String()

	got, err := client.SendTransaction(context.Background(), tx)
	if err != nil {
		t.Fatal(err)
	}
	if got != sig {
		t.Fatalf("signature = %s, want %s", got, sig)
	}

	if ts.lastReq.Method != "sendTransaction" {
		t.Fatalf("method = %s", ts.lastReq.Method)
	}
	params := ts.params(t)
	// The first param must be our tx in base64 wire form; the second must
	// carry the default encoding.
	var encoded string
	if err := json.Unmarshal(params[0], &encoded); err != nil {
		t.Fatal(err)
	}
	decoded, err := solana.TransactionFromBase64(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Signatures[0] != sig {
		t.Fatal("transaction did not round trip through the request")
	}
	var opts map[string]any
	if err := json.Unmarshal(params[1], &opts); err != nil {
		t.Fatal(err)
	}
	if opts["encoding"] != "base64" || opts["skipPreflight"] != false {
		t.Fatalf("opts = %v", opts)
	}
}

// testUnsignedTransactionRPC builds and signs a minimal transaction for
// client tests.
func testUnsignedTransactionRPC(key solana.PrivateKey) *solana.Transaction {
	payer := key.PublicKey()
	tx := &solana.Transaction{
		Message: solana.Message{
			AccountKeys: []solana.PublicKey{payer, {}},
			Header:      solana.MessageHeader{NumRequiredSignatures: 1, NumReadonlyUnsignedAccounts: 1},
			Instructions: []solana.CompiledInstruction{{
				ProgramIDIndex: 1,
				Accounts:       []uint16{0, 0},
				Data:           solana.Base58{2, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0},
			}},
			RecentBlockhash: solana.Hash{1},
		},
	}
	tx.Sign(func(pub solana.PublicKey) *solana.PrivateKey {
		return &key
	})
	return tx
}

func testPrivateKey(t testing.TB) solana.PrivateKey {
	t.Helper()
	key, err := solana.NewRandomPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func TestClientGetProgramAccountsStream(t *testing.T) {
	client, ts := newTestClient(t)
	ts.serveRaw(t, `{"context":{"slot":341197053},"value":[{"pubkey":"SysvarC1ock11111111111111111111111111111111","account":`+string(accountFixture)+`}]}`)

	var streamed []*KeyedAccount
	ctx, err := client.GetProgramAccountsStream(context.Background(), solana.SysVarPubkey, nil, func(ka *KeyedAccount) error {
		streamed = append(streamed, ka)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if ctx == nil || ctx.Slot != 341197053 || len(streamed) != 1 {
		t.Fatalf("ctx %+v, %d accounts", ctx, len(streamed))
	}

	// The request must force withContext so the slot is available.
	params := ts.params(t)
	var opts map[string]any
	if err := json.Unmarshal(params[1], &opts); err != nil {
		t.Fatal(err)
	}
	if opts["withContext"] != true {
		t.Fatalf("opts = %v", opts)
	}
}

func TestClientHeadersAndIDs(t *testing.T) {
	var gotHeader string
	var ids []uint64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Api-Key")
		var req struct {
			ID uint64 `json:"id"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		ids = append(ids, req.ID)
		w.Write([]byte(`{"jsonrpc":"2.0","result":1,"id":1}`))
	}))
	defer server.Close()

	client := New(server.URL)
	client.SetHeader("X-Api-Key", "secret")
	ctx := context.Background()
	client.GetSlot(ctx, "")
	client.GetSlot(ctx, "")

	if gotHeader != "secret" {
		t.Fatalf("header = %q", gotHeader)
	}
	if len(ids) != 2 || ids[0] == ids[1] {
		t.Fatalf("ids = %v", ids)
	}
}

func TestClientContextCancel(t *testing.T) {
	client, ts := newTestClient(t)
	ts.delay = 2 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := client.GetSlot(ctx, ""); err == nil {
		t.Fatal("expected context deadline error")
	}
}
