package rpc

import (
	"encoding/json"
	"strings"
	"testing"
)

// Fixture: rewards entry of a getBlock response, from upstream
// gagliardetto/solana-go rpc/client_test.go (TestClient_GetBlock).
const blockRewardFixture = `{"lamports":1595000,"postBalance":482032983798,"pubkey":"5rL3AaidKJa4ChSV3ys1SvpDg9L4amKiwYayGR5oL3dq","rewardType":"Fee"}`

func TestBlockRewardJSON(t *testing.T) {
	reward := jsonRoundTrip[BlockReward](t, []byte(blockRewardFixture))

	if reward.Pubkey.String() != "5rL3AaidKJa4ChSV3ys1SvpDg9L4amKiwYayGR5oL3dq" {
		t.Fatalf("Pubkey = %s", reward.Pubkey)
	}
	if reward.Lamports != 1595000 || reward.PostBalance != 482032983798 {
		t.Fatalf("got %+v", reward)
	}
	if reward.RewardType != RewardTypeFee {
		t.Fatalf("RewardType = %q", reward.RewardType)
	}

	// Commission absent from the fixture: stays nil and is omitted on
	// re-marshal.
	if reward.Commission != nil {
		t.Fatalf("Commission = %v, want nil", *reward.Commission)
	}
	data, err := json.Marshal(reward)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "commission") {
		t.Fatalf("marshal did not omit nil commission: %s", data)
	}
}

func TestBlockRewardCommission(t *testing.T) {
	// Voting rewards carry a commission.
	in := `{"pubkey":"5rL3AaidKJa4ChSV3ys1SvpDg9L4amKiwYayGR5oL3dq","lamports":-42,"postBalance":100,"rewardType":"Voting","commission":5}`
	reward := jsonRoundTrip[BlockReward](t, []byte(in))

	if reward.Lamports != -42 {
		t.Fatalf("Lamports = %d", reward.Lamports)
	}
	if reward.RewardType != RewardTypeVoting {
		t.Fatalf("RewardType = %q", reward.RewardType)
	}
	if reward.Commission == nil || *reward.Commission != 5 {
		t.Fatalf("Commission = %v", reward.Commission)
	}
}
