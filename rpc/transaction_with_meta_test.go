package rpc

import (
	"encoding/json"
	"strings"
	"testing"
)

// legacyTxBase64 is a base64-encoded legacy system-transfer transaction,
// copied from the root package's message_test.go fixtures (1 signature,
// 3 account keys).
const legacyTxBase64 = "AfjEs3XhTc3hrxEvlnMPkm/cocvAUbFNbCl00qKnrFue6J53AhEqIFmcJJlJW3EDP5RmcMz+cNTTcZHW/WJYwAcBAAEDO8hh4VddzfcO5jbCt95jryl6y8ff65UcgukHNLWH+UQGgxCGGpgyfQVQV02EQYqm4QwzUt2qf9f1gVLM7rI4hwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA6ANIF55zOZWROWRkeh+lExxZBnKFqbvIxZDLE7EijjoBAgIAAQwCAAAAOTAAAAAAAAA="

// transactionWithMetaBase64Fixture is a getTransaction-style envelope
// (base64 encoding). Envelope shape from upstream gagliardetto/solana-go
// rpc/getTransactionsForAddress_test.go; meta is transactionMetaFixture.
var transactionWithMetaBase64Fixture = `{"slot":83994671,"blockTime":1625231961,"transaction":["` +
	legacyTxBase64 + `","base64"],"meta":` + transactionMetaFixture + `,"version":"legacy"}`

func TestTransactionWithMetaBase64(t *testing.T) {
	twm := jsonRoundTrip[TransactionWithMeta](t, []byte(transactionWithMetaBase64Fixture))

	if twm.Slot != 83994671 {
		t.Fatalf("Slot = %d", twm.Slot)
	}
	if twm.BlockTime == nil || int64(*twm.BlockTime) != 1625231961 {
		t.Fatalf("BlockTime = %v", twm.BlockTime)
	}
	if twm.Version != LegacyTransactionVersion {
		t.Fatalf("Version = %d", twm.Version)
	}
	if twm.Meta == nil || twm.Meta.Fee != 5000 {
		t.Fatalf("Meta = %+v", twm.Meta)
	}

	tx, err := twm.GetTransaction()
	if err != nil {
		t.Fatalf("GetTransaction: %v", err)
	}
	if len(tx.Signatures) != 1 {
		t.Fatalf("Signatures = %v", tx.Signatures)
	}
	if len(tx.Message.AccountKeys) != 3 {
		t.Fatalf("AccountKeys = %v", tx.Message.AccountKeys)
	}
	// The transfer's program key is the system program.
	if tx.Message.AccountKeys[2].String() != "11111111111111111111111111111111" {
		t.Fatalf("program key = %s", tx.Message.AccountKeys[2])
	}
	if got := twm.MustGetTransaction(); got.Signatures[0] != tx.Signatures[0] {
		t.Fatalf("MustGetTransaction mismatch")
	}

	// Binary data is not jsonParsed.
	if _, err := twm.GetParsedTransaction(); err == nil {
		t.Fatal("GetParsedTransaction accepted binary data")
	}
}

func TestTransactionWithMetaJSONEncoding(t *testing.T) {
	// Fixture: getTransaction response with "json" encoding, from upstream
	// gagliardetto/solana-go rpc/client_test.go (TestClient_GetTransaction).
	in := `{"blockTime":1624821990,"meta":{"err":null,"fee":5000,"innerInstructions":[],"logMessages":["Program Vote111111111111111111111111111111111111111 invoke [1]","Program Vote111111111111111111111111111111111111111 success"],"postBalances":[199247210749,90459349430703,1,1,1],"postTokenBalances":[],"preBalances":[199247215749,90459349430703,1,1,1],"preTokenBalances":[],"rewards":[],"status":{"Ok":null}},"slot":83311386,"transaction":{"message":{"accountKeys":["2ZZkgKcBfp4tW8qCLj2yjxRYh9CuvEVJWb6e2KKS91Mj","53R9tmVrTQwJAgaUCWEA7SiVf7eWAbaQarZ159ixt2D9","SysvarS1otHashes111111111111111111111111111","SysvarC1ock11111111111111111111111111111111","Vote111111111111111111111111111111111111111"],"header":{"numReadonlySignedAccounts":0,"numReadonlyUnsignedAccounts":3,"numRequiredSignatures":1},"instructions":[{"accounts":[1,2,3,0],"data":"3yZe7d","programIdIndex":4}],"recentBlockhash":"6o9C27iJ5rPi7wEpvQu1cFbB1WnRudtsPnbY8GvFWrgR"},"signatures":["QPzWhnwHnCwk3nj1zVCcjz1VP7EcAKouPg9Joietje3GnQTVQ5XyWxyPC3zHby8K5ahSn9SbQupauDbVRvv5DuL"]}}`

	var twm TransactionWithMeta
	if err := json.Unmarshal([]byte(in), &twm); err != nil {
		t.Fatal(err)
	}
	if twm.Slot != 83311386 {
		t.Fatalf("Slot = %d", twm.Slot)
	}
	// No "version" key: field keeps its zero value.
	if twm.Version != 0 {
		t.Fatalf("Version = %d", twm.Version)
	}

	tx, err := twm.GetTransaction()
	if err != nil {
		t.Fatalf("GetTransaction: %v", err)
	}
	if len(tx.Message.AccountKeys) != 5 {
		t.Fatalf("AccountKeys = %v", tx.Message.AccountKeys)
	}
	if tx.Signatures[0].String() != "QPzWhnwHnCwk3nj1zVCcjz1VP7EcAKouPg9Joietje3GnQTVQ5XyWxyPC3zHby8K5ahSn9SbQupauDbVRvv5DuL" {
		t.Fatalf("Signature = %s", tx.Signatures[0])
	}
	if tx.Message.RecentBlockhash.String() != "6o9C27iJ5rPi7wEpvQu1cFbB1WnRudtsPnbY8GvFWrgR" {
		t.Fatalf("RecentBlockhash = %s", tx.Message.RecentBlockhash)
	}
}

func TestTransactionWithMetaAccountsMode(t *testing.T) {
	// Fixture: transactions entry of a getBlock response with
	// transactionDetails=accounts, from upstream gagliardetto/solana-go
	// rpc/client_test.go (TestClient_GetBlock_AccountsDetails).
	in := `{"meta":{"err":null,"fee":5000,"innerInstructions":[],"logMessages":[],"postBalances":[441866063495,40905918933763,1],"postTokenBalances":[],"preBalances":[441866068495,40905918933763,1],"preTokenBalances":[],"rewards":[],"status":{"Ok":null}},"transaction":{"signatures":["D8emaP3CaepSGigD3TCrev7j67yPLMi82qfzTb9iZYPxHcCmm6sQBKTU4bzAee4445zbnbWduVAZ87WfbWbXoAU"],"accountKeys":[{"pubkey":"EVd8FFVB54svYdZdG6hH4F4hTbqre5mpQ7XyF5rKUmes","signer":true,"writable":true,"source":"transaction"},{"pubkey":"72miaovmbPqccdbAA861r2uxwB5yL1sMjrgbCnc4JfVT","signer":false,"writable":true,"source":"transaction"},{"pubkey":"Vote111111111111111111111111111111111111111","signer":false,"writable":false,"source":"lookupTable"}]},"version":0}`

	var twm TransactionWithMeta
	if err := json.Unmarshal([]byte(in), &twm); err != nil {
		t.Fatal(err)
	}
	if twm.Version != 0 {
		t.Fatalf("Version = %d", twm.Version)
	}

	// The transaction has no message: GetTransaction must refuse and point
	// at GetAccountKeys.
	if _, err := twm.GetTransaction(); err == nil || !strings.Contains(err.Error(), "GetAccountKeys") {
		t.Fatalf("GetTransaction error = %v", err)
	}

	keys, err := twm.GetAccountKeys()
	if err != nil {
		t.Fatalf("GetAccountKeys: %v", err)
	}
	if len(keys.Signatures) != 1 || len(keys.AccountKeys) != 3 {
		t.Fatalf("got %+v", keys)
	}
	first := keys.AccountKeys[0]
	if first.Pubkey.String() != "EVd8FFVB54svYdZdG6hH4F4hTbqre5mpQ7XyF5rKUmes" || !first.Signer || !first.Writable {
		t.Fatalf("first key = %+v", first)
	}
	if first.Source == nil || *first.Source != AccountKeySourceTransaction {
		t.Fatalf("first source = %v", first.Source)
	}
	last := keys.AccountKeys[2]
	if last.Signer || last.Writable || last.Source == nil || *last.Source != AccountKeySourceLookupTable {
		t.Fatalf("last key = %+v", last)
	}
}

func TestTransactionWithMetaNilAndNull(t *testing.T) {
	var twm TransactionWithMeta
	if _, err := twm.GetTransaction(); err == nil {
		t.Fatal("GetTransaction accepted nil transaction")
	}
	if _, err := twm.GetAccountKeys(); err == nil {
		t.Fatal("GetAccountKeys accepted nil transaction")
	}
	if _, err := twm.GetParsedTransaction(); err == nil {
		t.Fatal("GetParsedTransaction accepted nil transaction")
	}

	// Null blockTime and absent meta stay nil.
	in := `{"slot":1,"blockTime":null,"transaction":["` + legacyTxBase64 + `","base64"],"version":"legacy"}`
	if err := json.Unmarshal([]byte(in), &twm); err != nil {
		t.Fatal(err)
	}
	if twm.BlockTime != nil {
		t.Fatalf("BlockTime = %v", *twm.BlockTime)
	}
	if twm.Meta != nil {
		t.Fatalf("Meta = %+v", twm.Meta)
	}

	// Binary transactions are not accounts-mode.
	if _, err := twm.GetAccountKeys(); err == nil {
		t.Fatal("GetAccountKeys accepted binary transaction")
	}
}
