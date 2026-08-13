package rpc

import (
	"testing"
)

// Fixture: getInflationGovernor response from upstream
// gagliardetto/solana-go rpc/client_test.go
// (TestClient_GetInflationGovernor).
const getInflationGovernorFixture = `{"foundation":0,"foundationTerm":0,"initial":0.15,"taper":0.15,"terminal":0.015}`

func TestGetInflationGovernorResultJSON(t *testing.T) {
	governor := jsonRoundTrip[GetInflationGovernorResult](t, []byte(getInflationGovernorFixture))

	if governor.Initial != 0.15 {
		t.Fatalf("Initial = %v", governor.Initial)
	}
	if governor.Terminal != 0.015 {
		t.Fatalf("Terminal = %v", governor.Terminal)
	}
	if governor.Taper != 0.15 {
		t.Fatalf("Taper = %v", governor.Taper)
	}
	if governor.Foundation != 0 || governor.FoundationTerm != 0 {
		t.Fatalf("Foundation = %v, FoundationTerm = %v",
			governor.Foundation, governor.FoundationTerm)
	}
}
