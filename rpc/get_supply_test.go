package rpc

import (
	"testing"
)

// Fixture from upstream gagliardetto/solana-go rpc/client_test.go
// (TestClient_GetSupply); the nonCirculatingAccounts list is truncated to
// three entries.
const getSupplyFixture = `{"context":{"slot":83999524},"value":{"circulating":1370901328666198300,"nonCirculating":154690270000000,"nonCirculatingAccounts":["Br3aeVGapRb2xTq17RU2pYZCoJpWA7bq6TKBCcYtMSmt","AzHQ8Bia1grVVbcGyci7wzueSWkgvu7YZVZ4B9rkL5P6","GpYnVDgB7dzvwSgsjQFeHznjG6Kt1DLBFYrKxjGU1LuD"],"total":1371056018936198300}}`

func TestGetSupplyResultJSON(t *testing.T) {
	out := jsonRoundTrip[GetSupplyResult](t, []byte(getSupplyFixture))

	if out.Context.Slot != 83999524 {
		t.Fatalf("Context.Slot = %d", out.Context.Slot)
	}
	if out.Value == nil {
		t.Fatal("Value is nil")
	}
	if out.Value.Total != 1371056018936198300 {
		t.Fatalf("Total = %d", out.Value.Total)
	}
	if out.Value.Circulating != 1370901328666198300 {
		t.Fatalf("Circulating = %d", out.Value.Circulating)
	}
	if out.Value.NonCirculating != 154690270000000 {
		t.Fatalf("NonCirculating = %d", out.Value.NonCirculating)
	}
	if len(out.Value.NonCirculatingAccounts) != 3 {
		t.Fatalf("NonCirculatingAccounts = %v", out.Value.NonCirculatingAccounts)
	}
	if out.Value.NonCirculatingAccounts[0].String() != "Br3aeVGapRb2xTq17RU2pYZCoJpWA7bq6TKBCcYtMSmt" {
		t.Fatalf("NonCirculatingAccounts[0] = %s", out.Value.NonCirculatingAccounts[0])
	}
}
