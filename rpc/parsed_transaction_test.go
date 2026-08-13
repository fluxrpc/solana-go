package rpc

import (
	"encoding/json"
	"github.com/bytedance/sonic"
	"testing"
)

// Fixture: "transaction" object of a getTransaction response with
// jsonParsed encoding, from upstream gagliardetto/solana-go
// rpc/client_test.go (TestClient_GetParsedTransaction).
const parsedTransactionFixture = `{"message":{"accountKeys":[{"pubkey":"G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo","signer":true,"writable":true}],"addressTableLookups":null,"instructions":[{"parsed":{"info":{"destination":"9bFNrXNb2WTx8fMHXCheaZqkLZ3YCCaiqTftHxeintHy","lamports":100,"source":"G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo"},"type":"transfer"},"program":"system","programId":"11111111111111111111111111111111"},{"parsed":{"info":{"amount":"47444666","delegate":"7oPa2PHQdZmjSPqvpZN7MQxnC7Dcf3uL4oLqknGLk2S3","owner":"G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo","source":"BMnsyyG6S6zkaE3K5X3nbRMKdvBS5dT6HhcMozBVL7Ly"},"type":"approve"},"program":"spl-token","programId":"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"},{"accounts":["G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo"],"data":"2dmnzvSCNoP8bNbUnUtk7FTYod5czhUfk4E7LSPNMtK4V1FHgQVYeQ2GnsEtCKZCyLLHXvnkReP","programId":"wormDTUJ6AWPNvk59vGQbDvGJmqbDTdgWgAqcLBCgUb"}],"recentBlockhash":"9L8FEB81LfZ67ejxpMaaZmC9EmXBpV38dhNaiF9UbzZi"},"signatures":["2x1QBpfcEQetAx7zETLEmvVvjue9311s9AWroEvMAboFkqaHZVp1sUpTFXroc5Q6tkPmZK5pYfmPFteoZPVRLF89"]}`

// Fixture: transaction meta of the same jsonParsed getTransaction
// response, trimmed to one parsed inner instruction and one raw inner
// instruction, from upstream gagliardetto/solana-go rpc/client_test.go
// (TestClient_GetParsedTransaction).
const parsedTransactionMetaFixture = `{"err":null,"fee":10000,"innerInstructions":[{"index":2,"instructions":[{"parsed":{"info":{"account":"BMnsyyG6S6zkaE3K5X3nbRMKdvBS5dT6HhcMozBVL7Ly","amount":"47444666","authority":"7oPa2PHQdZmjSPqvpZN7MQxnC7Dcf3uL4oLqknGLk2S3","mint":"E942z7FnS7GpswTvF5Vggvo7cMTbvZojjLbFgsrDVff1"},"type":"burn"},"program":"spl-token","programId":"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"},{"accounts":["2yVjuQwpsvdsrywzsJJVs9Ueh4zayyo5DYJbBNc3DDpn","3KEmPDRc6WEvhomG8awhfv2k33HgeqfGJmE1dptFmzhR"],"data":"2Af7uakYAFq8MGzDZQhLpcgRrAP9WHnAaA61z8nFafM8rFGNsKkksFcD6dDnAebHD6LCZBXqP6iyo8mX8XnteCsiEagZSqRLbe1QTRBpzZmwtFBVwY4SLyqBMxXKX35SM7zKVA7GYiTa2UDCaDvqQ3SQdHvRNaF5AED3HcJpYC1eFGhPpSjESVZHPN2rYYZXwma","programId":"worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth"}]}],"loadedAddresses":{"readonly":[],"writable":[]},"logMessages":["Program 11111111111111111111111111111111 invoke [1]","Program 11111111111111111111111111111111 success"],"postBalances":[72226420],"postTokenBalances":[{"accountIndex":4,"mint":"E942z7FnS7GpswTvF5Vggvo7cMTbvZojjLbFgsrDVff1","owner":"G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo","programId":"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA","uiTokenAmount":{"amount":"0","decimals":6,"uiAmount":null,"uiAmountString":"0"}}],"preBalances":[74714380],"preTokenBalances":[],"rewards":[],"returnData":{"programId":"11111111111111111111111111111111","data":["","base64"]},"status":{"Ok":null}}`

func TestParsedTransactionJSON(t *testing.T) {
	tx := jsonRoundTrip[ParsedTransaction](t, []byte(parsedTransactionFixture))

	if len(tx.Signatures) != 1 || tx.Signatures[0].String() != "2x1QBpfcEQetAx7zETLEmvVvjue9311s9AWroEvMAboFkqaHZVp1sUpTFXroc5Q6tkPmZK5pYfmPFteoZPVRLF89" {
		t.Fatalf("Signatures = %v", tx.Signatures)
	}
	msg := tx.Message
	if msg.RecentBlockHash != "9L8FEB81LfZ67ejxpMaaZmC9EmXBpV38dhNaiF9UbzZi" {
		t.Fatalf("RecentBlockHash = %s", msg.RecentBlockHash)
	}
	if len(msg.AccountKeys) != 1 {
		t.Fatalf("AccountKeys = %+v", msg.AccountKeys)
	}
	key := msg.AccountKeys[0]
	if key.PublicKey.String() != "G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo" || !key.Signer || !key.Writable {
		t.Fatalf("AccountKeys[0] = %+v", key)
	}
	if len(msg.Instructions) != 3 {
		t.Fatalf("Instructions = %+v", msg.Instructions)
	}

	// Parsed system transfer.
	transfer := msg.Instructions[0]
	if transfer.Program != "system" || transfer.ProgramId.String() != "11111111111111111111111111111111" {
		t.Fatalf("Instructions[0] = %+v", transfer)
	}
	info := transfer.Parsed.asInstructionInfo
	if info == nil || info.InstructionType != "transfer" {
		t.Fatalf("Parsed = %+v", transfer.Parsed)
	}
	if info.Info["lamports"] != float64(100) || info.Info["source"] != "G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo" {
		t.Fatalf("Info = %v", info.Info)
	}

	// Raw (unparsed) instruction: accounts and data instead of parsed.
	raw := msg.Instructions[2]
	if raw.Parsed != nil || raw.Program != "" {
		t.Fatalf("Instructions[2] = %+v", raw)
	}
	if len(raw.Accounts) != 1 || raw.Accounts[0].String() != "G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo" {
		t.Fatalf("Accounts = %v", raw.Accounts)
	}
	if len(raw.Data) == 0 {
		t.Fatal("Data is empty")
	}
}

func TestParsedTransactionMetaJSON(t *testing.T) {
	meta := jsonRoundTrip[ParsedTransactionMeta](t, []byte(parsedTransactionMetaFixture))

	if meta.Err != nil || meta.Fee != 10000 {
		t.Fatalf("got Err=%v Fee=%d", meta.Err, meta.Fee)
	}
	if len(meta.InnerInstructions) != 1 {
		t.Fatalf("InnerInstructions = %+v", meta.InnerInstructions)
	}
	inner := meta.InnerInstructions[0]
	if inner.Index != 2 || len(inner.Instructions) != 2 {
		t.Fatalf("inner = %+v", inner)
	}
	burn := inner.Instructions[0]
	if burn.Program != "spl-token" || burn.Parsed == nil || burn.Parsed.asInstructionInfo.InstructionType != "burn" {
		t.Fatalf("burn = %+v", burn)
	}
	if len(inner.Instructions[1].Accounts) != 2 {
		t.Fatalf("raw inner = %+v", inner.Instructions[1])
	}
	if len(meta.PostTokenBalances) != 1 || meta.PostTokenBalances[0].UiTokenAmount.UiAmount != nil {
		t.Fatalf("PostTokenBalances = %+v", meta.PostTokenBalances)
	}
	if meta.ComputeUnitsConsumed != nil {
		t.Fatalf("ComputeUnitsConsumed = %v", *meta.ComputeUnitsConsumed)
	}
}

func TestInstructionInfoEnvelopeJSON(t *testing.T) {
	// Object form.
	var envelope InstructionInfoEnvelope
	if err := json.Unmarshal([]byte(`{"type":"transfer","info":{"lamports":1}}`), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.asString != "" {
		t.Fatalf("asString = %q", envelope.asString)
	}
	if envelope.asInstructionInfo == nil || envelope.asInstructionInfo.InstructionType != "transfer" {
		t.Fatalf("asInstructionInfo = %+v", envelope.asInstructionInfo)
	}
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	var round InstructionInfoEnvelope
	if err := json.Unmarshal(data, &round); err != nil {
		t.Fatal(err)
	}
	if round.asInstructionInfo == nil || round.asInstructionInfo.Info["lamports"] != float64(1) {
		t.Fatalf("round = %+v", round.asInstructionInfo)
	}

	// String form.
	var str InstructionInfoEnvelope
	if err := json.Unmarshal([]byte(`"vote"`), &str); err != nil {
		t.Fatal(err)
	}
	if str.asString != "vote" || str.asInstructionInfo != nil {
		t.Fatalf("got %+v", str)
	}
	data, err = json.Marshal(str)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `"vote"` {
		t.Fatalf("Marshal = %s", data)
	}

	// Null is a no-op; other kinds are rejected.
	var null InstructionInfoEnvelope
	if err := json.Unmarshal([]byte(`null`), &null); err != nil {
		t.Fatal(err)
	}
	if null.asString != "" || null.asInstructionInfo != nil {
		t.Fatalf("got %+v", null)
	}
	var bad InstructionInfoEnvelope
	if err := bad.UnmarshalJSON([]byte(`123`)); err == nil {
		t.Fatal("UnmarshalJSON accepted a number")
	}
}

var benchmarkParsedTransaction ParsedTransaction

func BenchmarkParsedTransactionUnmarshalJSON(b *testing.B) {
	data := []byte(parsedTransactionFixture)
	b.ReportAllocs()
	for b.Loop() {
		if err := sonic.Unmarshal(data, &benchmarkParsedTransaction); err != nil {
			b.Fatal(err)
		}
	}
}
