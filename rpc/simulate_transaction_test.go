package rpc

import (
	"testing"
)

// Fixture from upstream gagliardetto/solana-go rpc/client_test.go
// (TestClient_SimulateTransaction).
const simulateTransactionFixture = `{"context":{"slot":218},"value":{"accounts":null,"logs":["Program 83astBRguLMdt2h5U1Tpdq5tjFoJ6noeGwaY3mDLVcri invoke [1]","Program 83astBRguLMdt2h5U1Tpdq5tjFoJ6noeGwaY3mDLVcri consumed 2366 of 1400000 compute units","Program return: 83astBRguLMdt2h5U1Tpdq5tjFoJ6noeGwaY3mDLVcri KgAAAAAAAAA=","Program 83astBRguLMdt2h5U1Tpdq5tjFoJ6noeGwaY3mDLVcri success"],"unitsConsumed":2366}}`

// Fixture from upstream gagliardetto/solana-go rpc/client_test.go
// (TestClient_SimulateTransaction_FullResult).
const simulateTransactionFullFixture = `{"context":{"slot":400},"value":{"logs":["Program log: ok"],"accounts":null,"unitsConsumed":3000,"loadedAccountsDataSize":1024,"fee":5000,"preBalances":[10000000,0],"postBalances":[9995000,0],"loadedAddresses":{"readonly":["11111111111111111111111111111111"],"writable":[]},"replacementBlockhash":{"blockhash":"EETubP5AKHgjPAhzPkToc6S4eibc4FFqQGnHR1Sh9rAr","lastValidBlockHeight":500}}}`

func TestSimulateTransactionResponseJSON(t *testing.T) {
	out := jsonRoundTrip[SimulateTransactionResponse](t, []byte(simulateTransactionFixture))

	if out.Context.Slot != 218 {
		t.Fatalf("Context.Slot = %d", out.Context.Slot)
	}
	if out.Value == nil {
		t.Fatal("Value is nil")
	}
	if out.Value.Err != nil {
		t.Fatalf("Err = %v", out.Value.Err)
	}
	if len(out.Value.Logs) != 4 {
		t.Fatalf("Logs = %v", out.Value.Logs)
	}
	if out.Value.Accounts != nil {
		t.Fatalf("Accounts = %v", out.Value.Accounts)
	}
	if out.Value.UnitsConsumed == nil || *out.Value.UnitsConsumed != 2366 {
		t.Fatalf("UnitsConsumed = %v", out.Value.UnitsConsumed)
	}
}

func TestSimulateTransactionResponseFullJSON(t *testing.T) {
	out := jsonRoundTrip[SimulateTransactionResponse](t, []byte(simulateTransactionFullFixture))

	v := out.Value
	if v == nil {
		t.Fatal("Value is nil")
	}
	if v.UnitsConsumed == nil || *v.UnitsConsumed != 3000 {
		t.Fatalf("UnitsConsumed = %v", v.UnitsConsumed)
	}
	if v.LoadedAccountsDataSize == nil || *v.LoadedAccountsDataSize != 1024 {
		t.Fatalf("LoadedAccountsDataSize = %v", v.LoadedAccountsDataSize)
	}
	if v.Fee == nil || *v.Fee != 5000 {
		t.Fatalf("Fee = %v", v.Fee)
	}
	if len(v.PreBalances) != 2 || v.PreBalances[0] != 10000000 ||
		len(v.PostBalances) != 2 || v.PostBalances[0] != 9995000 {
		t.Fatalf("balances = %v / %v", v.PreBalances, v.PostBalances)
	}
	if v.LoadedAddresses == nil || len(v.LoadedAddresses.ReadOnly) != 1 {
		t.Fatalf("LoadedAddresses = %+v", v.LoadedAddresses)
	}
	if v.ReplacementBlockhash == nil ||
		v.ReplacementBlockhash.Blockhash.String() != "EETubP5AKHgjPAhzPkToc6S4eibc4FFqQGnHR1Sh9rAr" ||
		v.ReplacementBlockhash.LastValidBlockHeight != 500 {
		t.Fatalf("ReplacementBlockhash = %+v", v.ReplacementBlockhash)
	}
}
