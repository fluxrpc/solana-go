package rpc

import (
	"testing"
)

// Fixture from upstream gagliardetto/solana-go rpc/client_test.go
// (TestClient_GetTokenSupply).
const getTokenSupplyFixture = `{"context":{"slot":86069939},"value":{"amount":"100","decimals":0,"uiAmount":100,"uiAmountString":"100"}}`

func TestGetTokenSupplyResultJSON(t *testing.T) {
	out := jsonRoundTrip[GetTokenSupplyResult](t, []byte(getTokenSupplyFixture))

	if out.Context.Slot != 86069939 {
		t.Fatalf("Context.Slot = %d", out.Context.Slot)
	}
	if out.Value == nil {
		t.Fatal("Value is nil")
	}
	if out.Value.Amount != "100" || out.Value.Decimals != 0 {
		t.Fatalf("Value = %+v", out.Value)
	}
	if out.Value.UiAmount == nil || *out.Value.UiAmount != 100 {
		t.Fatalf("UiAmount = %v", out.Value.UiAmount)
	}
}
