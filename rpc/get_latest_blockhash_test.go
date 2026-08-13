package rpc

import (
	"testing"
)

// Fixture: getLatestBlockhash response from upstream gagliardetto/solana-go
// rpc/client_test.go (TestClient_GetLatestBlockhash).
const getLatestBlockhashFixture = `{"context":{"slot":2792},"value":{"blockhash":"EkSnNWid2cvwEVnVx9aBqawnmiCNiDgp3gUdkDPTKN1N","lastValidBlockHeight":3090}}`

func TestGetLatestBlockhashResultJSON(t *testing.T) {
	result := jsonRoundTrip[GetLatestBlockhashResult](t, []byte(getLatestBlockhashFixture))

	if result.Context.Slot != 2792 {
		t.Fatalf("Context.Slot = %d", result.Context.Slot)
	}
	if result.Value == nil {
		t.Fatal("Value = nil")
	}
	if got := result.Value.Blockhash.String(); got != "EkSnNWid2cvwEVnVx9aBqawnmiCNiDgp3gUdkDPTKN1N" {
		t.Fatalf("Blockhash = %s", got)
	}
	if result.Value.LastValidBlockHeight != 3090 {
		t.Fatalf("LastValidBlockHeight = %d", result.Value.LastValidBlockHeight)
	}
}
