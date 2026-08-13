package rpc

import (
	"testing"
)

// Fixture: first two entries of the upstream gagliardetto/solana-go
// rpc/client_test.go (TestClient_GetTokenLargestAccounts) response.
const getTokenLargestAccountsFixture = `{"context":{"slot":86069724},"value":[{"address":"7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932","amount":"100","decimals":0,"uiAmount":100,"uiAmountString":"100"},{"address":"H7YZoNkQq96FX6gwy1ZqVgunXhSm7hpSPtK7orjxgQDb","amount":"0","decimals":0,"uiAmount":0,"uiAmountString":"0"}]}`

func TestGetTokenLargestAccountsResultJSON(t *testing.T) {
	out := jsonRoundTrip[GetTokenLargestAccountsResult](t, []byte(getTokenLargestAccountsFixture))

	if out.Context.Slot != 86069724 {
		t.Fatalf("Context.Slot = %d", out.Context.Slot)
	}
	if len(out.Value) != 2 {
		t.Fatalf("Value = %+v", out.Value)
	}
	first := out.Value[0]
	if first.Address.String() != "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932" {
		t.Fatalf("Address = %s", first.Address)
	}
	if first.Amount != "100" || first.UiAmountString != "100" {
		t.Fatalf("amounts = %+v", first.UiTokenAmount)
	}
	if first.UiAmount == nil || *first.UiAmount != 100 {
		t.Fatalf("UiAmount = %v", first.UiAmount)
	}
}
