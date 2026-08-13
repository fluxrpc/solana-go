package rpc

import (
	"testing"
)

// Fixture from upstream gagliardetto/solana-go rpc/client_test.go
// (TestClient_GetTokenAccountBalance).
const getTokenAccountBalanceFixture = `{"context":{"slot":1114},"value":{"amount":"9864","decimals":2,"uiAmount":98.64,"uiAmountString":"98.64"}}`

func TestGetTokenAccountBalanceResultJSON(t *testing.T) {
	out := jsonRoundTrip[GetTokenAccountBalanceResult](t, []byte(getTokenAccountBalanceFixture))

	if out.Context.Slot != 1114 {
		t.Fatalf("Context.Slot = %d", out.Context.Slot)
	}
	if out.Value == nil {
		t.Fatal("Value is nil")
	}
	if out.Value.Amount != "9864" || out.Value.Decimals != 2 {
		t.Fatalf("Value = %+v", out.Value)
	}
	if out.Value.UiAmount == nil || *out.Value.UiAmount != 98.64 {
		t.Fatalf("UiAmount = %v", out.Value.UiAmount)
	}
	if out.Value.UiAmountString != "98.64" {
		t.Fatalf("UiAmountString = %s", out.Value.UiAmountString)
	}
}
