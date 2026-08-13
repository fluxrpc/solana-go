package rpc

import (
	"testing"
)

// Fixture: getEpochSchedule response from upstream gagliardetto/solana-go
// rpc/client_test.go (TestClient_GetEpochSchedule).
const getEpochScheduleFixture = `{"firstNormalEpoch":14,"firstNormalSlot":524256,"leaderScheduleSlotOffset":432000,"slotsPerEpoch":432000,"warmup":true}`

func TestGetEpochScheduleResultJSON(t *testing.T) {
	schedule := jsonRoundTrip[GetEpochScheduleResult](t, []byte(getEpochScheduleFixture))

	if schedule.SlotsPerEpoch != 432000 {
		t.Fatalf("SlotsPerEpoch = %d", schedule.SlotsPerEpoch)
	}
	if schedule.LeaderScheduleSlotOffset != 432000 {
		t.Fatalf("LeaderScheduleSlotOffset = %d", schedule.LeaderScheduleSlotOffset)
	}
	if !schedule.Warmup {
		t.Fatal("Warmup = false")
	}
	if schedule.FirstNormalEpoch != 14 || schedule.FirstNormalSlot != 524256 {
		t.Fatalf("FirstNormalEpoch = %d, FirstNormalSlot = %d",
			schedule.FirstNormalEpoch, schedule.FirstNormalSlot)
	}
}
