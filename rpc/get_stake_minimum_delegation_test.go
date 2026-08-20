package rpc

import (
	"testing"
)

func TestGetStakeMinimumDelegationOptsJSON(t *testing.T) {
	opts := jsonRoundTrip[GetStakeMinimumDelegationOpts](t, []byte(`{"commitment":"finalized","minContextSlot":22}`))

	if opts.Commitment != CommitmentType("finalized") {
		t.Fatalf("Commitment = %s", opts.Commitment)
	}
	if opts.MinContextSlot == nil || *opts.MinContextSlot != 22 {
		t.Fatalf("MinContextSlot = %v", opts.MinContextSlot)
	}
}
