package rpc

import (
	"testing"
)

// Fixture from upstream gagliardetto/solana-go rpc/client_test.go
// (TestClient_GetVersion).
const getVersionFixture = `{"feature-set":743297851,"solana-core":"1.7.3"}`

func TestGetVersionResultJSON(t *testing.T) {
	out := jsonRoundTrip[GetVersionResult](t, []byte(getVersionFixture))

	if out.SolanaCore != "1.7.3" {
		t.Fatalf("SolanaCore = %s", out.SolanaCore)
	}
	if out.FeatureSet != 743297851 {
		t.Fatalf("FeatureSet = %d", out.FeatureSet)
	}
}
