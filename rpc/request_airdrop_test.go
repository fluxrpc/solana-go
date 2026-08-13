package rpc

import (
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

func TestRequestAirdropOptsJSON(t *testing.T) {
	// Hand-built fixture: RequestAirdropOpts has no JSON tags, so fields keep
	// their Go names when marshaled. The blockhash is from the upstream
	// gagliardetto/solana-go rpc/client_test.go TestClient_RequestAirdropWithOpts.
	opts := jsonRoundTrip[RequestAirdropOpts](t, []byte(`{"Commitment":"confirmed","RecentBlockhash":"J7rBdM6AecPDEZp8aPq5iPSNKVkU5Q76F3oAV4eW5wsW"}`))

	if opts.Commitment != CommitmentType("confirmed") {
		t.Fatalf("Commitment = %s", opts.Commitment)
	}
	if opts.RecentBlockhash == nil ||
		opts.RecentBlockhash.String() != "J7rBdM6AecPDEZp8aPq5iPSNKVkU5Q76F3oAV4eW5wsW" {
		t.Fatalf("RecentBlockhash = %v", opts.RecentBlockhash)
	}
}

func TestRequestAirdropResponseJSON(t *testing.T) {
	// Response fixture from upstream gagliardetto/solana-go rpc/client_test.go
	// (TestClient_RequestAirdrop): the result is a plain signature string.
	in := `"3ZmWDnFJ5REjxtmtQRrczmVDraVZs7BpUFo3NRfnoQs6wvTJ2kTkw9YyGod291UHjK5Qg6w63Hqn7t6nrGMLWhga"`
	sig := jsonRoundTrip[solana.Signature](t, []byte(in))
	if sig.String() != "3ZmWDnFJ5REjxtmtQRrczmVDraVZs7BpUFo3NRfnoQs6wvTJ2kTkw9YyGod291UHjK5Qg6w63Hqn7t6nrGMLWhga" {
		t.Fatalf("signature = %s", sig)
	}
}
