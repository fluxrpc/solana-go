package rpc

import (
	"encoding/json"
	"testing"
)

func TestGetProgramAccountsResultJSONRoundTrip(t *testing.T) {
	fixture := []byte(`[{"pubkey":"SysvarC1ock11111111111111111111111111111111","account":` + string(accountFixture) + `}]`)
	got := jsonRoundTrip[GetProgramAccountsResult](t, fixture)
	if len(got) != 1 || got[0].Account.Lamports != 88849814690250 {
		t.Fatalf("got %+v", got)
	}

	withContext := jsonRoundTrip[GetProgramAccountsWithContextResult](t, []byte(`{"context":{"slot":1},"value":`+string(fixture)+`}`))
	if len(withContext.Value) != 1 {
		t.Fatalf("got %+v", withContext)
	}
}

func TestGetProgramAccountsOptsMarshalOmitsEmpty(t *testing.T) {
	data, err := json.Marshal(GetProgramAccountsOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{}` {
		t.Fatalf("MarshalJSON() = %s, want {}", data)
	}

	yes := true
	opts := GetProgramAccountsOpts{
		Commitment:  CommitmentFinalized,
		Filters:     []RPCFilter{{DataSize: 165}},
		WithContext: &yes,
	}
	data, err = json.Marshal(opts)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"commitment":"finalized","filters":[{"dataSize":165}],"withContext":true}`
	if string(data) != want {
		t.Fatalf("MarshalJSON() = %s, want %s", data, want)
	}
}
