package rpc

import (
	"testing"
)

// Fixture from upstream gagliardetto/solana-go rpc/client_test.go
// (TestClient_GetRecentPerformanceSamples), with the second sample carrying
// numNonVoteTransactions per TestClient_GetRecentPerformanceSamples_WithNonVoteTx.
const getRecentPerformanceSamplesFixture = `[{"numSlots":84,"numTransactions":90402,"samplePeriodSecs":60,"slot":83998844},{"numSlots":85,"numTransactions":91000,"numNonVoteTransactions":300,"samplePeriodSecs":60,"slot":83998760}]`

func TestGetRecentPerformanceSamplesResultJSON(t *testing.T) {
	out := jsonRoundTrip[[]*GetRecentPerformanceSamplesResult](t, []byte(getRecentPerformanceSamplesFixture))

	if len(out) != 2 {
		t.Fatalf("len = %d", len(out))
	}
	if out[0].Slot != 83998844 || out[0].NumTransactions != 90402 {
		t.Fatalf("sample[0] = %+v", out[0])
	}
	if out[0].NumNonVoteTransactions != nil {
		t.Fatalf("sample[0].NumNonVoteTransactions = %v", *out[0].NumNonVoteTransactions)
	}
	if out[0].SamplePeriodSecs != 60 || out[0].NumSlots != 84 {
		t.Fatalf("sample[0] = %+v", out[0])
	}
	if out[1].NumNonVoteTransactions == nil || *out[1].NumNonVoteTransactions != 300 {
		t.Fatalf("sample[1].NumNonVoteTransactions = %v", out[1].NumNonVoteTransactions)
	}
}
