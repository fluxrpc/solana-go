package rpc

import (
	"encoding/json"
	"testing"
)

// Fixture: getHighestSnapshotSlot response from upstream
// gagliardetto/solana-go rpc/client_test.go
// (TestClient_GetHighestSnapshotSlot).
const getHighestSnapshotSlotFixture = `{"full":100,"incremental":110}`

func TestGetHighestSnapshotSlotResultJSON(t *testing.T) {
	result := jsonRoundTrip[GetHighestSnapshotSlotResult](t, []byte(getHighestSnapshotSlotFixture))

	if result.Full != 100 {
		t.Fatalf("Full = %d", result.Full)
	}
	if result.Incremental == nil || *result.Incremental != 110 {
		t.Fatalf("Incremental = %v", result.Incremental)
	}
}

func TestGetHighestSnapshotSlotResultNoIncremental(t *testing.T) {
	in := `{"full":100}`
	var result GetHighestSnapshotSlotResult
	if err := json.Unmarshal([]byte(in), &result); err != nil {
		t.Fatal(err)
	}
	if result.Full != 100 || result.Incremental != nil {
		t.Fatalf("result = %+v", result)
	}
}
