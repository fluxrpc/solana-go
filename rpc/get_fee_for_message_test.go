package rpc

import (
	"encoding/json"
	"testing"
)

// Fixture: getFeeForMessage response from upstream gagliardetto/solana-go
// rpc/client_test.go (TestClient_GetFeeForMessage).
const getFeeForMessageFixture = `{"context":{"slot":5068},"value":5000}`

func TestGetFeeForMessageResultJSON(t *testing.T) {
	result := jsonRoundTrip[GetFeeForMessageResult](t, []byte(getFeeForMessageFixture))

	if result.Context.Slot != 5068 {
		t.Fatalf("Context.Slot = %d", result.Context.Slot)
	}
	if result.Value == nil || *result.Value != 5000 {
		t.Fatalf("Value = %v", result.Value)
	}
}

func TestGetFeeForMessageResultNullValue(t *testing.T) {
	// The RPC returns a null fee for an invalid blockhash.
	in := `{"context":{"slot":5068},"value":null}`
	var result GetFeeForMessageResult
	if err := json.Unmarshal([]byte(in), &result); err != nil {
		t.Fatal(err)
	}
	if result.Value != nil {
		t.Fatalf("Value = %v", *result.Value)
	}
}
