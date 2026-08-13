package rpc

import (
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

// Fixture: first two entries of the upstream gagliardetto/solana-go
// rpc/client_test.go (TestClient_GetSignaturesForAddress) response.
const getSignaturesForAddressFixture = `[{"blockTime":1625231961,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4Yig3yd33o2hyZV2qZBJkScDArwVmzurkxhBfKdqJeujTrdKHwrR3U8KR6LrhN5eWNTyugS5rkkYagVXCNnk7pks","slot":83994671},{"blockTime":1625231952,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"3oQ7qqpJs5CtH1Xnnn8Ru5MtxkR3SZgshqzXwokuxFRArLihKdvCb9km6gbSiiUaNSHE7zVJqUVUZGfYuEaqWZPV","slot":83994656}]`

func TestGetSignaturesForAddressResponseJSON(t *testing.T) {
	out := jsonRoundTrip[[]*TransactionSignature](t, []byte(getSignaturesForAddressFixture))

	if len(out) != 2 {
		t.Fatalf("len = %d", len(out))
	}
	if out[0].Slot != 83994671 {
		t.Fatalf("Slot = %d", out[0].Slot)
	}
	if out[0].Signature.String() != "4Yig3yd33o2hyZV2qZBJkScDArwVmzurkxhBfKdqJeujTrdKHwrR3U8KR6LrhN5eWNTyugS5rkkYagVXCNnk7pks" {
		t.Fatalf("Signature = %s", out[0].Signature)
	}
	if out[0].ConfirmationStatus != ConfirmationStatusFinalized {
		t.Fatalf("ConfirmationStatus = %s", out[0].ConfirmationStatus)
	}
	if out[0].BlockTime == nil || *out[0].BlockTime != solana.UnixTimeSeconds(1625231961) {
		t.Fatalf("BlockTime = %v", out[0].BlockTime)
	}
}

func TestGetSignaturesForAddressOptsJSON(t *testing.T) {
	limit := 5
	opts := jsonRoundTrip[GetSignaturesForAddressOpts](t, []byte(`{"limit":5,"before":"4Yig3yd33o2hyZV2qZBJkScDArwVmzurkxhBfKdqJeujTrdKHwrR3U8KR6LrhN5eWNTyugS5rkkYagVXCNnk7pks","commitment":"finalized"}`))

	if opts.Limit == nil || *opts.Limit != limit {
		t.Fatalf("Limit = %v", opts.Limit)
	}
	if opts.Before.IsZero() {
		t.Fatal("Before is zero")
	}
	if opts.Commitment != "finalized" {
		t.Fatalf("Commitment = %s", opts.Commitment)
	}
}
