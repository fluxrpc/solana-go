package rpc

import (
	"encoding/json"
	"testing"
)

// Fixture: getBlockCommitment response from upstream gagliardetto/solana-go
// rpc/client_test.go (TestClient_GetBlockCommitment).
const getBlockCommitmentFixture = `{"commitment":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,44854495719374,0,51599979318189,5070972605440,140323113958535,169550804919131,272061505737107,860587424880950,1374732609383053,2334359721325133,4664454087479672,10122947678661428,52107037802932750],"totalStake":73611541921665680}`

func TestGetBlockCommitmentResultJSON(t *testing.T) {
	result := jsonRoundTrip[GetBlockCommitmentResult](t, []byte(getBlockCommitmentFixture))

	if len(result.Commitment) != 32 {
		t.Fatalf("len(Commitment) = %d", len(result.Commitment))
	}
	if result.Commitment[19] != 44854495719374 {
		t.Fatalf("Commitment[19] = %d", result.Commitment[19])
	}
	if result.Commitment[31] != 52107037802932750 {
		t.Fatalf("Commitment[31] = %d", result.Commitment[31])
	}
	if result.TotalStake != 73611541921665680 {
		t.Fatalf("TotalStake = %d", result.TotalStake)
	}
}

func TestGetBlockCommitmentResultUnknownBlock(t *testing.T) {
	// Unknown block: commitment is null.
	in := `{"commitment":null,"totalStake":73611541921665680}`
	var result GetBlockCommitmentResult
	if err := json.Unmarshal([]byte(in), &result); err != nil {
		t.Fatal(err)
	}
	if result.Commitment != nil {
		t.Fatalf("Commitment = %v", result.Commitment)
	}
	if result.TotalStake != 73611541921665680 {
		t.Fatalf("TotalStake = %d", result.TotalStake)
	}
}
