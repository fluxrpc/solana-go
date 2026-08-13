package rpc

import (
	"encoding/json"
	"github.com/bytedance/sonic"
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

// Fixture: getBlock (base64 transaction encoding) response with two vote
// transactions, from upstream gagliardetto/solana-go rpc/client_test.go
// (TestClient_GetBlock), with a returnData object (shape from upstream
// rpc/types_test.go) added to each meta so the fixture round-trips: an
// absent returnData decodes to nil content but re-marshals as ["",""].
const getBlockResultFixture = `{"blockHeight":69213636,"blockTime":1625227950,"blockhash":"5M77sHdwzH6rckuQwF8HL1w52n7hjrh4GVTFiF6T8QyB","parentSlot":83987983,"previousBlockhash":"Aq9jSXe1jRzfiaBcRFLe4wm7j499vWVEeFQrq5nnXfZN","rewards":[{"lamports":1595000,"postBalance":482032983798,"pubkey":"5rL3AaidKJa4ChSV3ys1SvpDg9L4amKiwYayGR5oL3dq","rewardType":"Fee"}],"transactions":[{"meta":{"err":null,"fee":5000,"innerInstructions":[],"logMessages":["Program Vote111111111111111111111111111111111111111 invoke [1]","Program Vote111111111111111111111111111111111111111 success"],"postBalances":[441866063495,40905918933763,1,1,1],"postTokenBalances":[],"preBalances":[441866068495,40905918933763,1,1,1],"preTokenBalances":[],"rewards":[],"returnData":{"programId":"11111111111111111111111111111111","data":["aGVsbG8=","base64"]},"status":{"Ok":null}},"transaction":["AQp2TH1spzjBAVM3alvnpaePFx3YEo9dvRglDuSChZUoTMD\/\/2h0HY5+89LJjCdiGJ7Ph3+Fyvbeiz1uJF8gxw0BAAMFyH0KDkXtjL1xebUYflZxYGlpV+LvjazzZCb\/mF2T67xZmkOUM\/A0iDSEkFzD5m4Ol82vsojigvqxrmp7Z1vrQgan1RcZLwqvxvJl4\/t3zHragsUp0L47E24tAFUgAAAABqfVFxjHdMkoVmOYaR1etoteuKObS21cc1VbIQAAAAAHYUgdNXR0u3xNdiTr072z2DVec9EQQ\/wNo1OAAAAAAAMFYbeqrsxJ9\/vZxtOaFi3rT2w9RF5Xi4jsyu61f3t1AQQEAQIDAAR0ZXN0","base64"]},{"meta":{"err":null,"fee":5000,"innerInstructions":[],"logMessages":["Program Vote111111111111111111111111111111111111111 invoke [1]","Program Vote111111111111111111111111111111111111111 success"],"postBalances":[334759887662,151357332545078,1,1,1],"postTokenBalances":[],"preBalances":[334759892662,151357332545078,1,1,1],"preTokenBalances":[],"rewards":[],"returnData":{"programId":"11111111111111111111111111111111","data":["aGVsbG8=","base64"]},"status":{"Ok":null}},"transaction":["ATA7DkBatbe2JB43QV+QRj2yoXSMXXttYFggDxZYOBfsRyYuGtzrbUevivclchxVccRIPlRP9PtS\/9NPXlwmhwwBAAMFSDrhjiNPuNqc4BWwitZz7xJ2NIXtv6XZtwtEOmgLj3n3NQ+OONLFlsu0LoUBSDsp40i9jOjZJBsliMtvTfdV+gan1RcZLwqvxvJl4\/t3zHragsUp0L47E24tAFUgAAAABqfVFxjHdMkoVmOYaR1etoteuKObS21cc1VbIQAAAAAHYUgdNXR0u3xNdiTr072z2DVec9EQQ\/wNo1OAAAAAAAKlcZMqS\/Oh0v+kOq2Ipg73NqbvKBRGQJDK8\/01K+MBAQQEAQIDAAR0ZXN0","base64"]}]}`

func TestGetBlockResultJSON(t *testing.T) {
	block := jsonRoundTrip[GetBlockResult](t, []byte(getBlockResultFixture))

	if got := block.Blockhash.String(); got != "5M77sHdwzH6rckuQwF8HL1w52n7hjrh4GVTFiF6T8QyB" {
		t.Fatalf("Blockhash = %s", got)
	}
	if got := block.PreviousBlockhash.String(); got != "Aq9jSXe1jRzfiaBcRFLe4wm7j499vWVEeFQrq5nnXfZN" {
		t.Fatalf("PreviousBlockhash = %s", got)
	}
	if block.ParentSlot != 83987983 {
		t.Fatalf("ParentSlot = %d", block.ParentSlot)
	}
	if block.BlockHeight == nil || *block.BlockHeight != 69213636 {
		t.Fatalf("BlockHeight = %v", block.BlockHeight)
	}
	if block.BlockTime == nil || *block.BlockTime != solana.UnixTimeSeconds(1625227950) {
		t.Fatalf("BlockTime = %v", block.BlockTime)
	}
	if len(block.Rewards) != 1 || block.Rewards[0].Lamports != 1595000 {
		t.Fatalf("Rewards = %+v", block.Rewards)
	}
	if len(block.Transactions) != 2 {
		t.Fatalf("Transactions = %d", len(block.Transactions))
	}
	for i, tx := range block.Transactions {
		if tx.Meta == nil || tx.Meta.Fee != 5000 {
			t.Fatalf("transaction %d meta = %+v", i, tx.Meta)
		}
		if len(tx.Transaction.GetBinary()) == 0 {
			t.Fatalf("transaction %d has no binary payload", i)
		}
	}
	if block.NumRewardPartitions != nil {
		t.Fatalf("NumRewardPartitions = %v", *block.NumRewardPartitions)
	}
}

// Fixture: getBlock (jsonParsed encoding) response. Header fields and
// reward from upstream gagliardetto/solana-go rpc/client_test.go
// (TestClient_GetBlock); the transaction is the jsonParsed form of its vote
// transaction, with account keys and the instruction resolved as documented
// at https://solana.com/docs/rpc/http/getblock, and a returnData object
// added for round-trip stability (see getBlockResultFixture).
const getParsedBlockResultFixture = `{
	"blockHeight": 69213636,
	"blockTime": 1625227950,
	"blockhash": "5M77sHdwzH6rckuQwF8HL1w52n7hjrh4GVTFiF6T8QyB",
	"parentSlot": 83987983,
	"previousBlockhash": "Aq9jSXe1jRzfiaBcRFLe4wm7j499vWVEeFQrq5nnXfZN",
	"rewards": [{"lamports": 1595000, "postBalance": 482032983798, "pubkey": "5rL3AaidKJa4ChSV3ys1SvpDg9L4amKiwYayGR5oL3dq", "rewardType": "Fee"}],
	"transactions": [
		{
			"meta": {"err": null, "fee": 5000, "preBalances": [441866068495, 40905918933763, 1, 1, 1], "postBalances": [441866063495, 40905918933763, 1, 1, 1], "logMessages": ["Program Vote111111111111111111111111111111111111111 invoke [1]", "Program Vote111111111111111111111111111111111111111 success"], "returnData": {"programId": "11111111111111111111111111111111", "data": ["aGVsbG8=", "base64"]}, "status": {"Ok": null}},
			"transaction": {
				"signatures": ["D8emaP3CaepSGigD3TCrev7j67yPLMi82qfzTb9iZYPxHcCmm6sQBKTU4bzAee4445zbnbWduVAZ87WfbWbXoAU"],
				"message": {
					"accountKeys": [
						{"pubkey": "EVd8FFVB54svYdZdG6hH4F4hTbqre5mpQ7XyF5rKUmes", "signer": true, "writable": true},
						{"pubkey": "72miaovmbPqccdbAA861r2uxwB5yL1sMjrgbCnc4JfVT", "signer": false, "writable": true},
						{"pubkey": "Vote111111111111111111111111111111111111111", "signer": false, "writable": false}
					],
					"instructions": [
						{"program": "vote", "programId": "Vote111111111111111111111111111111111111111", "parsed": {"info": {"voteAccount": "72miaovmbPqccdbAA861r2uxwB5yL1sMjrgbCnc4JfVT"}, "type": "vote"}, "stackHeight": 1}
					],
					"recentBlockhash": "CnyzpJmBydX1X2FyXXzsPFc5WPT9UFdLVkEhnvW33at"
				}
			}
		}
	]
}`

func TestGetParsedBlockResultJSON(t *testing.T) {
	block := jsonRoundTrip[GetParsedBlockResult](t, []byte(getParsedBlockResultFixture))

	if got := block.Blockhash.String(); got != "5M77sHdwzH6rckuQwF8HL1w52n7hjrh4GVTFiF6T8QyB" {
		t.Fatalf("Blockhash = %s", got)
	}
	if block.ParentSlot != 83987983 {
		t.Fatalf("ParentSlot = %d", block.ParentSlot)
	}
	if len(block.Transactions) != 1 {
		t.Fatalf("Transactions = %d", len(block.Transactions))
	}
	twm := block.Transactions[0]
	if twm.Meta == nil || twm.Meta.Fee != 5000 {
		t.Fatalf("Meta = %+v", twm.Meta)
	}
	if twm.Transaction == nil || len(twm.Transaction.Signatures) != 1 {
		t.Fatalf("Transaction = %+v", twm.Transaction)
	}
	msg := twm.Transaction.Message
	if len(msg.AccountKeys) != 3 || !msg.AccountKeys[0].Signer || msg.AccountKeys[2].Writable {
		t.Fatalf("AccountKeys = %+v", msg.AccountKeys)
	}
	if len(msg.Instructions) != 1 || msg.Instructions[0].Program != "vote" {
		t.Fatalf("Instructions = %+v", msg.Instructions)
	}
}

func TestGetBlockOptsValidate(t *testing.T) {
	var nilOpts *GetBlockOpts
	if err := nilOpts.Validate(); err != nil {
		t.Fatalf("nil opts: %v", err)
	}
	if err := (&GetBlockOpts{}).Validate(); err != nil {
		t.Fatalf("empty encoding: %v", err)
	}
	for _, encoding := range []solana.EncodingType{
		solana.EncodingJSON,
		solana.EncodingJSONParsed,
		solana.EncodingBase58,
		solana.EncodingBase64,
	} {
		if err := (&GetBlockOpts{Encoding: encoding}).Validate(); err != nil {
			t.Fatalf("encoding %q: %v", encoding, err)
		}
	}
	// base64+zstd is rejected up front: this SDK cannot decode the response.
	for _, encoding := range []solana.EncodingType{solana.EncodingBase64Zstd, "base99"} {
		if err := (&GetBlockOpts{Encoding: encoding}).Validate(); err == nil {
			t.Fatalf("encoding %q unexpectedly accepted", encoding)
		}
	}
}

func TestTransactionDetailsConstants(t *testing.T) {
	for expected, got := range map[string]TransactionDetailsType{
		"full":       TransactionDetailsFull,
		"signatures": TransactionDetailsSignatures,
		"none":       TransactionDetailsNone,
		"accounts":   TransactionDetailsAccounts,
	} {
		if string(got) != expected {
			t.Fatalf("TransactionDetailsType %q != %q", got, expected)
		}
	}
	if MaxSupportedTransactionVersion0 != 0 || MaxSupportedTransactionVersion1 != 1 {
		t.Fatalf("MaxSupportedTransactionVersion = %d / %d",
			MaxSupportedTransactionVersion0, MaxSupportedTransactionVersion1)
	}
}

var (
	benchmarkGetBlockResult     GetBlockResult
	benchmarkGetBlockResultJSON []byte
)

func BenchmarkGetBlockResultUnmarshalJSON(b *testing.B) {
	data := []byte(getBlockResultFixture)
	b.ReportAllocs()
	for b.Loop() {
		if err := sonic.Unmarshal(data, &benchmarkGetBlockResult); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetBlockResultMarshalJSON(b *testing.B) {
	var block GetBlockResult
	if err := json.Unmarshal([]byte(getBlockResultFixture), &block); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		var err error
		benchmarkGetBlockResultJSON, err = json.Marshal(&block)
		if err != nil {
			b.Fatal(err)
		}
	}
}
