package rpc

import (
	"encoding/json"
	"testing"

	"github.com/bytedance/sonic"
	solana "github.com/fluxrpc/solana-go"
)

// Fixture from upstream gagliardetto/solana-go rpc/client_test.go
// (TestClient_GetTransaction): a vote transaction in the JSON encoding.
// A returnData object (shape from upstream rpc/types_test.go) is added so the
// zero value survives the marshal round trip, as in transactionMetaFixture.
const getTransactionJSONFixture = `{"blockTime":1624821990,"meta":{"err":null,"fee":5000,"innerInstructions":[],"logMessages":["Program Vote111111111111111111111111111111111111111 invoke [1]","Program Vote111111111111111111111111111111111111111 success"],"postBalances":[199247210749,90459349430703,1,1,1],"postTokenBalances":[],"preBalances":[199247215749,90459349430703,1,1,1],"preTokenBalances":[],"returnData":{"programId":"11111111111111111111111111111111","data":["","base64"]},"rewards":[],"status":{"Ok":null}},"slot":83311386,"transaction":{"message":{"accountKeys":["2ZZkgKcBfp4tW8qCLj2yjxRYh9CuvEVJWb6e2KKS91Mj","53R9tmVrTQwJAgaUCWEA7SiVf7eWAbaQarZ159ixt2D9","SysvarS1otHashes111111111111111111111111111","SysvarC1ock11111111111111111111111111111111","Vote111111111111111111111111111111111111111"],"header":{"numReadonlySignedAccounts":0,"numReadonlyUnsignedAccounts":3,"numRequiredSignatures":1},"instructions":[{"accounts":[1,2,3,0],"data":"3yZe7d","programIdIndex":4}],"recentBlockhash":"6o9C27iJ5rPi7wEpvQu1cFbB1WnRudtsPnbY8GvFWrgR"},"signatures":["QPzWhnwHnCwk3nj1zVCcjz1VP7EcAKouPg9Joietje3GnQTVQ5XyWxyPC3zHby8K5ahSn9SbQupauDbVRvv5DuL"]}}`

// Fixture: getTransaction response in the base64 encoding. The transaction
// blob and its meta come from the upstream gagliardetto/solana-go
// rpc/client_test.go getBlock fixture (TestClient_GetBlock, first vote
// transaction), reassembled into the getTransaction response shape, plus a
// returnData object (shape from upstream rpc/types_test.go) so the zero value
// survives the marshal round trip.
const getTransactionBase64Fixture = `{"blockTime":1625227950,"meta":{"err":null,"fee":5000,"innerInstructions":[],"logMessages":["Program Vote111111111111111111111111111111111111111 invoke [1]","Program Vote111111111111111111111111111111111111111 success"],"postBalances":[441866063495,40905918933763,1,1,1],"postTokenBalances":[],"preBalances":[441866068495,40905918933763,1,1,1],"preTokenBalances":[],"returnData":{"programId":"11111111111111111111111111111111","data":["","base64"]},"rewards":[],"status":{"Ok":null}},"slot":83987984,"transaction":["AQp2TH1spzjBAVM3alvnpaePFx3YEo9dvRglDuSChZUoTMD//2h0HY5+89LJjCdiGJ7Ph3+Fyvbeiz1uJF8gxw0BAAMFyH0KDkXtjL1xebUYflZxYGlpV+LvjazzZCb/mF2T67xZmkOUM/A0iDSEkFzD5m4Ol82vsojigvqxrmp7Z1vrQgan1RcZLwqvxvJl4/t3zHragsUp0L47E24tAFUgAAAABqfVFxjHdMkoVmOYaR1etoteuKObS21cc1VbIQAAAAAHYUgdNXR0u3xNdiTr072z2DVec9EQQ/wNo1OAAAAAAAMFYbeqrsxJ9/vZxtOaFi3rT2w9RF5Xi4jsyu61f3t1AQQEAQIDAAR0ZXN0","base64"],"version":"legacy"}`

func TestGetTransactionResultJSONEncoding(t *testing.T) {
	out := jsonRoundTrip[GetTransactionResult](t, []byte(getTransactionJSONFixture))

	if out.Slot != 83311386 {
		t.Fatalf("Slot = %d", out.Slot)
	}
	if out.BlockTime == nil || *out.BlockTime != solana.UnixTimeSeconds(1624821990) {
		t.Fatalf("BlockTime = %v", out.BlockTime)
	}
	if out.Meta == nil || out.Meta.Fee != 5000 {
		t.Fatalf("Meta = %+v", out.Meta)
	}
	if out.Transaction == nil {
		t.Fatal("Transaction is nil")
	}
	tx, err := out.Transaction.GetTransaction()
	if err != nil {
		t.Fatal(err)
	}
	if len(tx.Signatures) != 1 ||
		tx.Signatures[0].String() != "QPzWhnwHnCwk3nj1zVCcjz1VP7EcAKouPg9Joietje3GnQTVQ5XyWxyPC3zHby8K5ahSn9SbQupauDbVRvv5DuL" {
		t.Fatalf("Signatures = %v", tx.Signatures)
	}
	if len(tx.Message.AccountKeys) != 5 {
		t.Fatalf("AccountKeys = %v", tx.Message.AccountKeys)
	}
	// The JSON encoding does not carry binary content.
	if out.Transaction.GetBinary() != nil {
		t.Fatalf("GetBinary() = %v", out.Transaction.GetBinary())
	}
}

func TestGetTransactionResultBase64Encoding(t *testing.T) {
	out := jsonRoundTrip[GetTransactionResult](t, []byte(getTransactionBase64Fixture))

	if out.Slot != 83987984 {
		t.Fatalf("Slot = %d", out.Slot)
	}
	if out.Version != LegacyTransactionVersion {
		t.Fatalf("Version = %v", out.Version)
	}
	if out.Transaction == nil {
		t.Fatal("Transaction is nil")
	}
	if data := out.Transaction.GetData(); data.Encoding != solana.EncodingBase64 {
		t.Fatalf("Data.Encoding = %s", data.Encoding)
	}
	if len(out.Transaction.GetBinary()) == 0 {
		t.Fatal("GetBinary() is empty")
	}
	tx, err := out.Transaction.GetTransaction()
	if err != nil {
		t.Fatal(err)
	}
	if len(tx.Signatures) != 1 || len(tx.Message.AccountKeys) != 5 {
		t.Fatalf("decoded transaction = %+v", tx)
	}
	if tx.Message.AccountKeys[4].String() != "Vote111111111111111111111111111111111111111" {
		t.Fatalf("AccountKeys[4] = %s", tx.Message.AccountKeys[4])
	}
}

var (
	benchmarkGetTransactionResult GetTransactionResult
	benchmarkGetTransactionJSON   []byte
)

func BenchmarkGetTransactionResultUnmarshalJSON(b *testing.B) {
	data := []byte(getTransactionBase64Fixture)
	b.ReportAllocs()
	for b.Loop() {
		if err := sonic.Unmarshal(data, &benchmarkGetTransactionResult); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetTransactionResultMarshalJSON(b *testing.B) {
	var out GetTransactionResult
	if err := json.Unmarshal([]byte(getTransactionBase64Fixture), &out); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		var err error
		benchmarkGetTransactionJSON, err = json.Marshal(&out)
		if err != nil {
			b.Fatal(err)
		}
	}
}
