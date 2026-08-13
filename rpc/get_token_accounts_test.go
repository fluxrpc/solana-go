package rpc

import (
	"testing"
)

// Fixture from upstream gagliardetto/solana-go rpc/client_test.go
// (TestClient_GetTokenAccountsByDelegate), jsonParsed account data.
const getTokenAccountsFixture = `{"context":{"slot":1114},"value":[{"account":{"data":{"program":"spl-token","parsed":{"accountType":"account","info":{"tokenAmount":{"amount":"1","decimals":1,"uiAmount":0.1,"uiAmountString":"0.1"},"delegate":"4Nd1mBQtrMJVYVfKf2PJy9NZUZdTAsp7D4xWLs4gDB4T","delegatedAmount":1,"isInitialized":true,"isNative":false,"mint":"3wyAj7Rt1TWVPZVteFJPLa26JmLvdb1CAKEFZm3NY75E","owner":"CnPoSPKXu7wJqxe59Fs72tkBeALovhsCxYeFwPCQH9TD"}}},"executable":false,"lamports":1726080,"owner":"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA","rentEpoch":4,"space":0},"pubkey":"CnPoSPKXu7wJqxe59Fs72tkBeALovhsCxYeFwPCQH9TD"}]}`

func TestGetTokenAccountsResultJSON(t *testing.T) {
	out := jsonRoundTrip[GetTokenAccountsResult](t, []byte(getTokenAccountsFixture))

	if out.Context.Slot != 1114 {
		t.Fatalf("Context.Slot = %d", out.Context.Slot)
	}
	if len(out.Value) != 1 {
		t.Fatalf("Value = %+v", out.Value)
	}
	tokenAccount := out.Value[0]
	if tokenAccount.Pubkey.String() != "CnPoSPKXu7wJqxe59Fs72tkBeALovhsCxYeFwPCQH9TD" {
		t.Fatalf("Pubkey = %s", tokenAccount.Pubkey)
	}
	if tokenAccount.Account.Lamports != 1726080 {
		t.Fatalf("Lamports = %d", tokenAccount.Account.Lamports)
	}
	if tokenAccount.Account.Owner.String() != "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA" {
		t.Fatalf("Owner = %s", tokenAccount.Account.Owner)
	}
	if tokenAccount.Account.Data.GetRawJSON() == nil {
		t.Fatal("Data.GetRawJSON() is nil; expected jsonParsed data")
	}
}
