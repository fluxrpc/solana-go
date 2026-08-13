package rpc

import (
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

// Fixture: getIdentity response from upstream gagliardetto/solana-go
// rpc/client_test.go (TestClient_GetIdentity).
const getIdentityFixture = `{"identity":"DMeohMfD3JzmYZA34jL9iiTXp5N7tpAR3rAoXMygdH3U"}`

func TestGetIdentityResultJSON(t *testing.T) {
	result := jsonRoundTrip[GetIdentityResult](t, []byte(getIdentityFixture))

	expected := solana.MustPublicKeyFromBase58("DMeohMfD3JzmYZA34jL9iiTXp5N7tpAR3rAoXMygdH3U")
	if result.Identity != expected {
		t.Fatalf("Identity = %s", result.Identity)
	}
}
