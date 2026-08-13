package rpc

import (
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

// Fixture: getLargestAccounts response from upstream gagliardetto/solana-go
// rpc/client_test.go (TestClient_GetLargestAccounts), truncated to the
// first two accounts.
const getLargestAccountsFixture = `{"context":{"slot":83995022},"value":[{"address":"4Rf9mGD7FeYknun5JczX5nGLTfQuS1GRjNVfkEMKE92b","lamports":398178060209179300},{"address":"KchK7WTjPzq9QL5aCwnV1dLsT8rFjruS1Zfzamxus9G","lamports":215100454508495000}]}`

func TestGetLargestAccountsResultJSON(t *testing.T) {
	result := jsonRoundTrip[GetLargestAccountsResult](t, []byte(getLargestAccountsFixture))

	if result.Context.Slot != 83995022 {
		t.Fatalf("Context.Slot = %d", result.Context.Slot)
	}
	if len(result.Value) != 2 {
		t.Fatalf("len(Value) = %d", len(result.Value))
	}
	if result.Value[0].Address != solana.MustPublicKeyFromBase58("4Rf9mGD7FeYknun5JczX5nGLTfQuS1GRjNVfkEMKE92b") {
		t.Fatalf("Value[0].Address = %s", result.Value[0].Address)
	}
	if result.Value[0].Lamports != 398178060209179300 {
		t.Fatalf("Value[0].Lamports = %d", result.Value[0].Lamports)
	}
	if result.Value[1].Lamports != 215100454508495000 {
		t.Fatalf("Value[1].Lamports = %d", result.Value[1].Lamports)
	}
}

func TestLargestAccountsFilterConstants(t *testing.T) {
	if LargestAccountsFilterCirculating != "circulating" {
		t.Fatalf("LargestAccountsFilterCirculating = %q", LargestAccountsFilterCirculating)
	}
	if LargestAccountsFilterNonCirculating != "nonCirculating" {
		t.Fatalf("LargestAccountsFilterNonCirculating = %q", LargestAccountsFilterNonCirculating)
	}
}
