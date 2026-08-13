package rpc

import (
	"reflect"
	"testing"
)

// Fixture from upstream gagliardetto/solana-go rpc/client_test.go
// (TestClient_GetVoteAccounts).
const getVoteAccountsFixture = `{"current":[],"delinquent":[{"activatedStake":4997717120,"commission":100,"epochCredits":[[127,1124979,892885],[128,1435333,1124979],[129,1603147,1435333],[131,1739262,1603147],[132,1895556,1739262]],"epochVoteAccount":true,"lastVote":51699331,"nodePubkey":"z3roU4WgvZvYkAEAYmUGK4LkPK6qFii6uzgMAswGYjb","rootSlot":51699288,"votePubkey":"vot33MHDqT6nSwubGzqtc6m16ChcUywxV7tNULF19Vu"}]}`

func TestGetVoteAccountsResultJSON(t *testing.T) {
	out := jsonRoundTrip[GetVoteAccountsResult](t, []byte(getVoteAccountsFixture))

	if len(out.Current) != 0 {
		t.Fatalf("Current = %+v", out.Current)
	}
	if len(out.Delinquent) != 1 {
		t.Fatalf("Delinquent = %+v", out.Delinquent)
	}
	acc := out.Delinquent[0]
	if acc.VotePubkey.String() != "vot33MHDqT6nSwubGzqtc6m16ChcUywxV7tNULF19Vu" {
		t.Fatalf("VotePubkey = %s", acc.VotePubkey)
	}
	if acc.NodePubkey.String() != "z3roU4WgvZvYkAEAYmUGK4LkPK6qFii6uzgMAswGYjb" {
		t.Fatalf("NodePubkey = %s", acc.NodePubkey)
	}
	if acc.ActivatedStake != 4997717120 || acc.Commission != 100 || !acc.EpochVoteAccount {
		t.Fatalf("vote account = %+v", acc)
	}
	if acc.InflationRewardsCommissionBps != nil {
		t.Fatalf("InflationRewardsCommissionBps = %v", *acc.InflationRewardsCommissionBps)
	}
	if acc.LastVote != 51699331 || acc.RootSlot != 51699288 {
		t.Fatalf("vote account = %+v", acc)
	}
	if len(acc.EpochCredits) != 5 || !reflect.DeepEqual(acc.EpochCredits[0], []int64{127, 1124979, 892885}) {
		t.Fatalf("EpochCredits = %v", acc.EpochCredits)
	}
}
