package rpc

import (
	"encoding/json"
	"testing"
)

// Fixture: getEpochInfo response from upstream gagliardetto/solana-go
// rpc/client_test.go (TestClient_GetEpochInfo).
const getEpochInfoFixture = `{"absoluteSlot":83994151,"blockHeight":69218302,"epoch":207,"slotIndex":93895,"slotsInEpoch":432000,"transactionCount":27287000257}`

func TestGetEpochInfoResultJSON(t *testing.T) {
	info := jsonRoundTrip[GetEpochInfoResult](t, []byte(getEpochInfoFixture))

	if info.AbsoluteSlot != 83994151 {
		t.Fatalf("AbsoluteSlot = %d", info.AbsoluteSlot)
	}
	if info.BlockHeight != 69218302 {
		t.Fatalf("BlockHeight = %d", info.BlockHeight)
	}
	if info.Epoch != 207 {
		t.Fatalf("Epoch = %d", info.Epoch)
	}
	if info.SlotIndex != 93895 || info.SlotsInEpoch != 432000 {
		t.Fatalf("SlotIndex = %d, SlotsInEpoch = %d", info.SlotIndex, info.SlotsInEpoch)
	}
	if info.TransactionCount == nil || *info.TransactionCount != 27287000257 {
		t.Fatalf("TransactionCount = %v", info.TransactionCount)
	}
}

func TestGetEpochInfoResultNoTransactionCount(t *testing.T) {
	in := `{"absoluteSlot":83994151,"blockHeight":69218302,"epoch":207,"slotIndex":93895,"slotsInEpoch":432000}`
	var info GetEpochInfoResult
	if err := json.Unmarshal([]byte(in), &info); err != nil {
		t.Fatal(err)
	}
	if info.TransactionCount != nil {
		t.Fatalf("TransactionCount = %v", *info.TransactionCount)
	}
}
