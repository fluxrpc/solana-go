package rpc

import (
	"testing"
)

func TestGetSlotOptsJSON(t *testing.T) {
	// Hand-built fixture: GetSlotOpts has no JSON tags, so fields keep their
	// Go names when marshaled.
	opts := jsonRoundTrip[GetSlotOpts](t, []byte(`{"Commitment":"confirmed","MinContextSlot":123456}`))

	if opts.Commitment != CommitmentType("confirmed") {
		t.Fatalf("Commitment = %s", opts.Commitment)
	}
	if opts.MinContextSlot == nil || *opts.MinContextSlot != 123456 {
		t.Fatalf("MinContextSlot = %v", opts.MinContextSlot)
	}
}
