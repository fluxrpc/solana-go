package rpc

import (
	"testing"
)

// Fixture from upstream gagliardetto/solana-go rpc/client_test.go
// (TestClient_GetRecentPrioritizationFees).
const getRecentPrioritizationFeesFixture = `[{"slot":348125,"prioritizationFee":0},{"slot":348126,"prioritizationFee":1000},{"slot":348127,"prioritizationFee":500}]`

func TestPrioritizationFeeResultJSON(t *testing.T) {
	out := jsonRoundTrip[[]PrioritizationFeeResult](t, []byte(getRecentPrioritizationFeesFixture))

	if len(out) != 3 {
		t.Fatalf("len = %d", len(out))
	}
	if out[0].Slot != 348125 || out[0].PrioritizationFee != 0 {
		t.Fatalf("fees[0] = %+v", out[0])
	}
	if out[1].PrioritizationFee != 1000 {
		t.Fatalf("fees[1] = %+v", out[1])
	}
	if out[2].Slot != 348127 || out[2].PrioritizationFee != 500 {
		t.Fatalf("fees[2] = %+v", out[2])
	}
}
