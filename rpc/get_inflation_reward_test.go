package rpc

import (
	"encoding/json"
	"testing"
)

// Fixture: getInflationReward response from upstream gagliardetto/solana-go
// rpc/client_test.go (TestClient_GetInflationReward_WithCommissionBps).
const getInflationRewardFixture = `[{"epoch":100,"effectiveSlot":43200000,"amount":2500000,"postBalance":1002500000,"commission":5,"commissionBps":500}]`

func TestGetInflationRewardResultJSON(t *testing.T) {
	rewards := jsonRoundTrip[[]*GetInflationRewardResult](t, []byte(getInflationRewardFixture))

	if len(rewards) != 1 || rewards[0] == nil {
		t.Fatalf("rewards = %+v", rewards)
	}
	reward := rewards[0]
	if reward.Epoch != 100 {
		t.Fatalf("Epoch = %d", reward.Epoch)
	}
	if reward.EffectiveSlot != 43200000 {
		t.Fatalf("EffectiveSlot = %d", reward.EffectiveSlot)
	}
	if reward.Amount != 2500000 || reward.PostBalance != 1002500000 {
		t.Fatalf("Amount = %d, PostBalance = %d", reward.Amount, reward.PostBalance)
	}
	if reward.Commission == nil || *reward.Commission != 5 {
		t.Fatalf("Commission = %v", reward.Commission)
	}
	if reward.CommissionBps == nil || *reward.CommissionBps != 500 {
		t.Fatalf("CommissionBps = %v", reward.CommissionBps)
	}
}

func TestGetInflationRewardResultNullEntries(t *testing.T) {
	// Addresses without a reward come back as null entries, from upstream
	// gagliardetto/solana-go rpc/client_test.go (TestClient_GetInflationReward).
	var rewards []*GetInflationRewardResult
	if err := json.Unmarshal([]byte(`[null]`), &rewards); err != nil {
		t.Fatal(err)
	}
	if len(rewards) != 1 || rewards[0] != nil {
		t.Fatalf("rewards = %+v", rewards)
	}
}
