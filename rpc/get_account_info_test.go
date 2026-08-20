package rpc

import (
	"encoding/json"
	"reflect"
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

// Values are from upstream gagliardetto/solana-go rpc/client_test.go
// (TestClient_GetAccountInfoWithOpts).
func TestGetAccountInfoOptsRoundTrip(t *testing.T) {
	offset := uint64(22)
	length := uint64(33)
	minContextSlot := uint64(83986105)
	opts := GetAccountInfoOpts{
		Encoding:   solana.EncodingBase64Zstd,
		Commitment: CommitmentFinalized,
		DataSlice: &DataSlice{
			Offset: &offset,
			Length: &length,
		},
		MinContextSlot: &minContextSlot,
	}

	data, err := json.Marshal(opts)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	wantJSON := `{"encoding":"base64+zstd","commitment":"finalized","dataSlice":{"offset":22,"length":33},"minContextSlot":83986105}`
	if string(data) != wantJSON {
		t.Fatalf("JSON = %s, want %s", data, wantJSON)
	}
	var back GetAccountInfoOpts
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(opts, back) {
		t.Fatalf("round trip mismatch:\nfirst:  %+v\nsecond: %+v", opts, back)
	}

	if back.Encoding != solana.EncodingBase64Zstd {
		t.Fatalf("Encoding = %s", back.Encoding)
	}
	if back.DataSlice == nil || *back.DataSlice.Offset != 22 || *back.DataSlice.Length != 33 {
		t.Fatalf("DataSlice = %+v", back.DataSlice)
	}
	if back.MinContextSlot == nil || *back.MinContextSlot != 83986105 {
		t.Fatalf("MinContextSlot = %v", back.MinContextSlot)
	}
}
