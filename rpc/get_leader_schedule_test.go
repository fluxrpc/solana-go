package rpc

import (
	"encoding/json"
	"reflect"
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

// Fixture: getLeaderSchedule response. First identity and its leading slot
// indices are from upstream gagliardetto/solana-go rpc/client_test.go
// (TestClient_GetLeaderSchedule), truncated to four indices; the second
// identity is from the upstream getBlockProduction fixture.
const getLeaderScheduleFixture = `{"DsaF77cCADh79q7HPfz5TrWPfEmD5Gw1c15zSm4eaFyt":[128,129,130,131],"121cur1YFVPZSoKQGNyjNr9sZZRa3eX2bSuYjXHtKD6":[0,1,2,3]}`

func TestGetLeaderScheduleResultJSON(t *testing.T) {
	schedule := jsonRoundTrip[GetLeaderScheduleResult](t, []byte(getLeaderScheduleFixture))

	if len(schedule) != 2 {
		t.Fatalf("len(schedule) = %d", len(schedule))
	}
	first := solana.MustPublicKeyFromBase58("DsaF77cCADh79q7HPfz5TrWPfEmD5Gw1c15zSm4eaFyt")
	if !reflect.DeepEqual(schedule[first], []uint64{128, 129, 130, 131}) {
		t.Fatalf("schedule[DsaF77...] = %v", schedule[first])
	}
	second := solana.MustPublicKeyFromBase58("121cur1YFVPZSoKQGNyjNr9sZZRa3eX2bSuYjXHtKD6")
	if !reflect.DeepEqual(schedule[second], []uint64{0, 1, 2, 3}) {
		t.Fatalf("schedule[121cur...] = %v", schedule[second])
	}
}

func TestGetLeaderScheduleResultNull(t *testing.T) {
	// The RPC returns null when the epoch of the requested slot is unknown.
	var schedule GetLeaderScheduleResult
	if err := json.Unmarshal([]byte(`null`), &schedule); err != nil {
		t.Fatal(err)
	}
	if schedule != nil {
		t.Fatalf("schedule = %v", schedule)
	}
}

func TestGetLeaderScheduleResultInvalidKey(t *testing.T) {
	var schedule GetLeaderScheduleResult
	if err := json.Unmarshal([]byte(`{"not-a-key!":[1]}`), &schedule); err == nil {
		t.Fatal("expected error for invalid identity key")
	}
}
