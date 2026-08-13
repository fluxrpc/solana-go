package rpc

import (
	"encoding/json"
	"reflect"
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

// Fixture: getBlockProduction response. Context slot and the byIdentity
// entries are from upstream gagliardetto/solana-go rpc/client_test.go
// (TestClient_GetBlockProduction), truncated to three identities; the range
// object follows the documented shape at
// https://solana.com/docs/rpc/http/getblockproduction.
const getBlockProductionFixture = `{"context":{"slot":83992896},"value":{"byIdentity":{"121cur1YFVPZSoKQGNyjNr9sZZRa3eX2bSuYjXHtKD6":[44,38],"123vij84ecQEKUvQ7gYMKxKwKF6PbYSzCzzURYA4xULY":[52,49],"12QYHqRxPuTPfkBVLetEuGkLGHD9GhqM5coP67xK7wfG":[64,55]},"range":{"firstSlot":83961600,"lastSlot":83992895}}}`

func TestGetBlockProductionResultJSON(t *testing.T) {
	result := jsonRoundTrip[GetBlockProductionResult](t, []byte(getBlockProductionFixture))

	if result.Context.Slot != 83992896 {
		t.Fatalf("Context.Slot = %d", result.Context.Slot)
	}
	if len(result.Value.ByIdentity) != 3 {
		t.Fatalf("len(ByIdentity) = %d", len(result.Value.ByIdentity))
	}
	identity := solana.MustPublicKeyFromBase58("121cur1YFVPZSoKQGNyjNr9sZZRa3eX2bSuYjXHtKD6")
	if got := result.Value.ByIdentity[identity]; got != [2]int64{44, 38} {
		t.Fatalf("ByIdentity[121cur...] = %v", got)
	}
	if result.Value.Range.FirstSlot != 83961600 || result.Value.Range.LastSlot != 83992895 {
		t.Fatalf("Range = %+v", result.Value.Range)
	}
}

func TestIdentityToSlotsBlocksInvalidKey(t *testing.T) {
	var m IdentityToSlotsBlocks
	if err := json.Unmarshal([]byte(`{"not-a-key!":[1,2]}`), &m); err == nil {
		t.Fatal("expected error for invalid identity key")
	}
}

// GetBlockProductionOpts carries JSON tags upstream; check the wire shape
// and the round trip. Values from upstream gagliardetto/solana-go
// rpc/client_test.go (TestClient_GetBlockProductionWithOpts).
func TestGetBlockProductionOptsJSON(t *testing.T) {
	lastSlot := uint64(3)
	identity := solana.MustPublicKeyFromBase58("7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932")
	opts := GetBlockProductionOpts{
		Commitment: CommitmentFinalized,
		Range: &SlotRangeRequest{
			FirstSlot: 2,
			LastSlot:  &lastSlot,
		},
		Identity: &identity,
	}

	data, err := json.Marshal(opts)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var wire map[string]any
	if err := json.Unmarshal(data, &wire); err != nil {
		t.Fatalf("unmarshal wire: %v", err)
	}
	expected := map[string]any{
		"commitment": "finalized",
		"range": map[string]any{
			"firstSlot": float64(2),
			"lastSlot":  float64(3),
		},
		"identity": "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932",
	}
	if !reflect.DeepEqual(wire, expected) {
		t.Fatalf("wire shape mismatch:\ngot:  %+v\nwant: %+v", wire, expected)
	}

	var back GetBlockProductionOpts
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(opts, back) {
		t.Fatalf("round trip mismatch:\nfirst:  %+v\nsecond: %+v", opts, back)
	}
}
