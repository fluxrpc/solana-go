package rpc

import (
	"testing"
)

// Fixture: getInflationRate response from upstream gagliardetto/solana-go
// rpc/client_test.go (TestClient_GetInflationRate).
const getInflationRateFixture = `{"epoch":207,"foundation":0,"total":0.1403151524615605,"validator":0.1403151524615605}`

func TestGetInflationRateResultJSON(t *testing.T) {
	rate := jsonRoundTrip[GetInflationRateResult](t, []byte(getInflationRateFixture))

	if rate.Total != 0.1403151524615605 {
		t.Fatalf("Total = %v", rate.Total)
	}
	if rate.Validator != 0.1403151524615605 {
		t.Fatalf("Validator = %v", rate.Validator)
	}
	if rate.Foundation != 0 {
		t.Fatalf("Foundation = %v", rate.Foundation)
	}
	if rate.Epoch != 207 {
		t.Fatalf("Epoch = %d", rate.Epoch)
	}
}
