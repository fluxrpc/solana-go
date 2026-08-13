package rpc

import (
	"reflect"
	"testing"
)

// Fixture: getBlocks response from upstream gagliardetto/solana-go
// rpc/client_test.go (TestClient_GetBlocks).
const getBlocksFixture = `[83993598,83993599,83993600]`

func TestBlocksResultJSON(t *testing.T) {
	blocks := jsonRoundTrip[BlocksResult](t, []byte(getBlocksFixture))

	if !reflect.DeepEqual(blocks, BlocksResult{83993598, 83993599, 83993600}) {
		t.Fatalf("BlocksResult = %v", blocks)
	}
}
