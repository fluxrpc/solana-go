package rpc

import (
	"testing"
)

func TestGetStakeMinimumDelegationOptsJSON(t *testing.T) {
	// Hand-built fixture: GetStakeMinimumDelegationOpts has no JSON tags, so
	// fields keep their Go names when marshaled.
	opts := jsonRoundTrip[GetStakeMinimumDelegationOpts](t, []byte(`{"Commitment":"finalized","MinContextSlot":22}`))

	if opts.Commitment != CommitmentType("finalized") {
		t.Fatalf("Commitment = %s", opts.Commitment)
	}
	if opts.MinContextSlot == nil || *opts.MinContextSlot != 22 {
		t.Fatalf("MinContextSlot = %v", opts.MinContextSlot)
	}
}
