package rpc

import (
	"encoding/json"
	"reflect"
	"testing"
)

// The exact marshal assertion guards the option JSON names; the round trip
// guards the field values.
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
	wantJSON := `{"commitment":"confirmed","minContextSlot":83987501}`
	if string(data) != wantJSON {
		t.Fatalf("JSON = %s, want %s", data, wantJSON)
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
