package rpc

import (
	"math/big"
	"testing"
)

// Fixture from upstream gagliardetto/solana-go rpc/client_test.go
// (TestClient_GetMultipleAccounts).
const getMultipleAccountsFixture = `{"context":{"slot":83996178},"value":[{"data":["","base64"],"executable":true,"lamports":19039980000,"owner":"11111111111111111111111111111111","rentEpoch":207}]}`

func TestGetMultipleAccountsResultJSON(t *testing.T) {
	out := jsonRoundTrip[GetMultipleAccountsResult](t, []byte(getMultipleAccountsFixture))

	if out.Context.Slot != 83996178 {
		t.Fatalf("Context.Slot = %d", out.Context.Slot)
	}
	if len(out.Value) != 1 {
		t.Fatalf("Value = %+v", out.Value)
	}
	acc := out.Value[0]
	if acc.Lamports != 19039980000 {
		t.Fatalf("Lamports = %d", acc.Lamports)
	}
	if !acc.Executable {
		t.Fatal("Executable = false")
	}
	if acc.Owner.String() != "11111111111111111111111111111111" {
		t.Fatalf("Owner = %s", acc.Owner)
	}
	if acc.RentEpoch.Cmp(big.NewInt(207)) != 0 {
		t.Fatalf("RentEpoch = %s", acc.RentEpoch)
	}
}
