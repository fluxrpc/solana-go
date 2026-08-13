package rpc

import (
	"encoding/json"
	"github.com/bytedance/sonic"
	"reflect"
	"testing"
)

// Fixture: transaction meta of a getBlock (base64 encoding) response.
// Balances, logs, rewards from upstream gagliardetto/solana-go
// rpc/client_test.go (TestClient_GetBlock); token balances from
// TestClient_GetParsedTransaction; loadedAddresses/returnData shapes from
// rpc/types_test.go, with the inner instruction in compiled form.
const transactionMetaFixture = `{
	"err": null,
	"fee": 5000,
	"innerInstructions": [
		{"index": 0, "instructions": [
			{"programIdIndex": 4, "accounts": [1, 2, 3, 0], "data": "3yZe7d", "stackHeight": 2}
		]}
	],
	"logMessages": [
		"Program Vote111111111111111111111111111111111111111 invoke [1]",
		"Program Vote111111111111111111111111111111111111111 success"
	],
	"preBalances": [441866068495, 40905918933763, 1, 1, 1],
	"postBalances": [441866063495, 40905918933763, 1, 1, 1],
	"preTokenBalances": [{"accountIndex": 4, "mint": "E942z7FnS7GpswTvF5Vggvo7cMTbvZojjLbFgsrDVff1", "owner": "G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo", "programId": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA", "uiTokenAmount": {"amount": "47444666", "decimals": 6, "uiAmount": 47.444666, "uiAmountString": "47.444666"}}],
	"postTokenBalances": [{"accountIndex": 4, "mint": "E942z7FnS7GpswTvF5Vggvo7cMTbvZojjLbFgsrDVff1", "owner": "G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo", "programId": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA", "uiTokenAmount": {"amount": "0", "decimals": 6, "uiAmount": null, "uiAmountString": "0"}}],
	"rewards": [{"pubkey": "5rL3AaidKJa4ChSV3ys1SvpDg9L4amKiwYayGR5oL3dq", "lamports": 1595000, "postBalance": 482032983798, "rewardType": "Fee"}],
	"loadedAddresses": {"readonly": ["11111111111111111111111111111111"], "writable": ["4ejjNYBbaETZyqaiK8aDj2BWER8LKHgDcCnRrPC22YGg"]},
	"returnData": {"programId": "11111111111111111111111111111111", "data": ["aGVsbG8=", "base64"]},
	"status": {"Ok": null},
	"computeUnitsConsumed": 2100
}`

func TestTransactionMetaJSON(t *testing.T) {
	meta := jsonRoundTrip[TransactionMeta](t, []byte(transactionMetaFixture))

	if meta.Err != nil {
		t.Fatalf("Err = %v", meta.Err)
	}
	if meta.Fee != 5000 {
		t.Fatalf("Fee = %d", meta.Fee)
	}
	if !reflect.DeepEqual(meta.PreBalances, []uint64{441866068495, 40905918933763, 1, 1, 1}) {
		t.Fatalf("PreBalances = %v", meta.PreBalances)
	}
	if !reflect.DeepEqual(meta.PostBalances, []uint64{441866063495, 40905918933763, 1, 1, 1}) {
		t.Fatalf("PostBalances = %v", meta.PostBalances)
	}
	if len(meta.InnerInstructions) != 1 || len(meta.InnerInstructions[0].Instructions) != 1 {
		t.Fatalf("InnerInstructions = %+v", meta.InnerInstructions)
	}
	if got := meta.InnerInstructions[0].Instructions[0]; got.ProgramIDIndex != 4 || got.StackHeight != 2 {
		t.Fatalf("inner instruction = %+v", got)
	}
	if len(meta.PreTokenBalances) != 1 || len(meta.PostTokenBalances) != 1 {
		t.Fatalf("token balances = %+v / %+v", meta.PreTokenBalances, meta.PostTokenBalances)
	}
	if len(meta.LogMessages) != 2 {
		t.Fatalf("LogMessages = %v", meta.LogMessages)
	}
	if !reflect.DeepEqual(meta.Status, DeprecatedTransactionMetaStatus{"Ok": nil}) {
		t.Fatalf("Status = %v", meta.Status)
	}
	if len(meta.Rewards) != 1 || meta.Rewards[0].Lamports != 1595000 {
		t.Fatalf("Rewards = %+v", meta.Rewards)
	}
	if len(meta.LoadedAddresses.ReadOnly) != 1 || len(meta.LoadedAddresses.Writable) != 1 {
		t.Fatalf("LoadedAddresses = %+v", meta.LoadedAddresses)
	}
	if meta.ReturnData.ProgramId.String() != "11111111111111111111111111111111" {
		t.Fatalf("ReturnData = %+v", meta.ReturnData)
	}
	if meta.ComputeUnitsConsumed == nil || *meta.ComputeUnitsConsumed != 2100 {
		t.Fatalf("ComputeUnitsConsumed = %v", meta.ComputeUnitsConsumed)
	}
}

func TestTransactionMetaMinimal(t *testing.T) {
	// Pre token-balance / CU-metering era meta, from upstream
	// gagliardetto/solana-go rpc/getTransactionsForAddress_test.go.
	in := `{"err":null,"fee":5000,"preBalances":[1000],"postBalances":[995]}`
	var meta TransactionMeta
	if err := json.Unmarshal([]byte(in), &meta); err != nil {
		t.Fatal(err)
	}
	if meta.Err != nil || meta.Fee != 5000 {
		t.Fatalf("got %+v", meta)
	}
	if meta.InnerInstructions != nil || meta.PreTokenBalances != nil ||
		meta.PostTokenBalances != nil || meta.LogMessages != nil ||
		meta.Rewards != nil || meta.Status != nil {
		t.Fatalf("omitted fields not zero: %+v", meta)
	}
	if meta.ComputeUnitsConsumed != nil {
		t.Fatalf("ComputeUnitsConsumed = %v", *meta.ComputeUnitsConsumed)
	}
}

func TestTransactionMetaErr(t *testing.T) {
	// Failed transaction: err carries the TransactionError object and
	// status carries its deprecated envelope.
	in := `{"err":{"InstructionError":[0,{"Custom":6000}]},"fee":5000,"preBalances":[],"postBalances":[],"status":{"Err":{"InstructionError":[0,{"Custom":6000}]}}}`
	var meta TransactionMeta
	if err := json.Unmarshal([]byte(in), &meta); err != nil {
		t.Fatal(err)
	}

	errObj, ok := meta.Err.(map[string]any)
	if !ok {
		t.Fatalf("Err = %T %v", meta.Err, meta.Err)
	}
	if _, ok := errObj["InstructionError"]; !ok {
		t.Fatalf("Err = %v", meta.Err)
	}
	if _, ok := meta.Status["Err"]; !ok {
		t.Fatalf("Status = %v", meta.Status)
	}
}

var (
	benchmarkMeta     TransactionMeta
	benchmarkMetaJSON []byte
)

func BenchmarkTransactionMetaUnmarshalJSON(b *testing.B) {
	data := []byte(transactionMetaFixture)
	b.ReportAllocs()
	for b.Loop() {
		if err := sonic.Unmarshal(data, &benchmarkMeta); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTransactionMetaMarshalJSON(b *testing.B) {
	var meta TransactionMeta
	if err := json.Unmarshal([]byte(transactionMetaFixture), &meta); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		var err error
		benchmarkMetaJSON, err = json.Marshal(&meta)
		if err != nil {
			b.Fatal(err)
		}
	}
}
