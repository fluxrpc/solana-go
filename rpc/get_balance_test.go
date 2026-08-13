package rpc

import (
	"encoding/json"
	"reflect"
	"testing"
)

// GetBalanceOpts is a request-builder options struct without JSON tags (as
// upstream); the round trip below guards the field set.
func TestGetBalanceOptsRoundTrip(t *testing.T) {
	minContextSlot := uint64(83987501)
	opts := GetBalanceOpts{
		Commitment:     CommitmentConfirmed,
		MinContextSlot: &minContextSlot,
	}

	data, err := json.Marshal(opts)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back GetBalanceOpts
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(opts, back) {
		t.Fatalf("round trip mismatch:\nfirst:  %+v\nsecond: %+v", opts, back)
	}

	if back.Commitment != CommitmentConfirmed {
		t.Fatalf("Commitment = %s", back.Commitment)
	}
	if back.MinContextSlot == nil || *back.MinContextSlot != 83987501 {
		t.Fatalf("MinContextSlot = %v", back.MinContextSlot)
	}
}
