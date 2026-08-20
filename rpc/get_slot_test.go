package rpc

import (
	"testing"
)

func TestGetSlotOptsJSON(t *testing.T) {
	opts := jsonRoundTrip[GetSlotOpts](t, []byte(`{"commitment":"confirmed","minContextSlot":123456}`))

	if opts.Commitment != CommitmentType("confirmed") {
		t.Fatalf("Commitment = %s", opts.Commitment)
	}
	if opts.MinContextSlot == nil || *opts.MinContextSlot != 123456 {
		t.Fatalf("MinContextSlot = %v", opts.MinContextSlot)
	}
}
