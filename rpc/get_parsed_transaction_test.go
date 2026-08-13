package rpc

import (
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

// Fixture from upstream gagliardetto/solana-go rpc/client_test.go
// (TestClient_GetParsedTransaction): a jsonParsed getTransaction response.
// A returnData object (shape from upstream rpc/types_test.go) is added so the
// zero value survives the marshal round trip, as in transactionMetaFixture.
const getParsedTransactionFixture = `{"blockTime":1660570006,"meta":{"err":null,"fee":10000,"innerInstructions":[{"index":2,"instructions":[{"parsed":{"info":{"account":"BMnsyyG6S6zkaE3K5X3nbRMKdvBS5dT6HhcMozBVL7Ly","amount":"47444666","authority":"7oPa2PHQdZmjSPqvpZN7MQxnC7Dcf3uL4oLqknGLk2S3","mint":"E942z7FnS7GpswTvF5Vggvo7cMTbvZojjLbFgsrDVff1"},"type":"burn"},"program":"spl-token","programId":"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"},{"parsed":{"info":{"destination":"9bFNrXNb2WTx8fMHXCheaZqkLZ3YCCaiqTftHxeintHy","lamports":100,"source":"G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo"},"type":"transfer"},"program":"system","programId":"11111111111111111111111111111111"},{"accounts":["2yVjuQwpsvdsrywzsJJVs9Ueh4zayyo5DYJbBNc3DDpn","3KEmPDRc6WEvhomG8awhfv2k33HgeqfGJmE1dptFmzhR"],"data":"2Af7uakYAFq8MGzDZQhLpcgRrAP9WHnAaA61z8nFafM8rFGNsKkksFcD6dDnAebHD6LCZBXqP6iyo8mX8XnteCsiEagZSqRLbe1QTRBpzZmwtFBVwY4SLyqBMxXKX35SM7zKVA7GYiTa2UDCaDvqQ3SQdHvRNaF5AED3HcJpYC1eFGhPpSjESVZHPN2rYYZXwma","programId":"worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth"}]}],"loadedAddresses":{"readonly":[],"writable":[]},"logMessages":["Program 11111111111111111111111111111111 invoke [1]","Program 11111111111111111111111111111111 success"],"postBalances":[72226420],"postTokenBalances":[{"accountIndex":4,"mint":"E942z7FnS7GpswTvF5Vggvo7cMTbvZojjLbFgsrDVff1","owner":"G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo","programId":"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA","uiTokenAmount":{"amount":"0","decimals":6,"uiAmount":null,"uiAmountString":"0"}}],"preBalances":[74714380],"preTokenBalances":[{"accountIndex":4,"mint":"E942z7FnS7GpswTvF5Vggvo7cMTbvZojjLbFgsrDVff1","owner":"G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo","programId":"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA","uiTokenAmount":{"amount":"47444666","decimals":6,"uiAmount":47.444666,"uiAmountString":"47.444666"}}],"returnData":{"programId":"11111111111111111111111111111111","data":["","base64"]},"rewards":[],"status":{"Ok":null}},"slot":146099091,"transaction":{"message":{"accountKeys":[{"pubkey":"G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo","signer":true,"writable":true}],"addressTableLookups":null,"instructions":[{"parsed":{"info":{"destination":"9bFNrXNb2WTx8fMHXCheaZqkLZ3YCCaiqTftHxeintHy","lamports":100,"source":"G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo"},"type":"transfer"},"program":"system","programId":"11111111111111111111111111111111"},{"parsed":{"info":{"amount":"47444666","delegate":"7oPa2PHQdZmjSPqvpZN7MQxnC7Dcf3uL4oLqknGLk2S3","owner":"G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo","source":"BMnsyyG6S6zkaE3K5X3nbRMKdvBS5dT6HhcMozBVL7Ly"},"type":"approve"},"program":"spl-token","programId":"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"},{"accounts":["G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo"],"data":"2dmnzvSCNoP8bNbUnUtk7FTYod5czhUfk4E7LSPNMtK4V1FHgQVYeQ2GnsEtCKZCyLLHXvnkReP","programId":"wormDTUJ6AWPNvk59vGQbDvGJmqbDTdgWgAqcLBCgUb"}],"recentBlockhash":"9L8FEB81LfZ67ejxpMaaZmC9EmXBpV38dhNaiF9UbzZi"},"signatures":["2x1QBpfcEQetAx7zETLEmvVvjue9311s9AWroEvMAboFkqaHZVp1sUpTFXroc5Q6tkPmZK5pYfmPFteoZPVRLF89"]}}`

func TestGetParsedTransactionResultJSON(t *testing.T) {
	out := jsonRoundTrip[GetParsedTransactionResult](t, []byte(getParsedTransactionFixture))

	if out.Slot != 146099091 {
		t.Fatalf("Slot = %d", out.Slot)
	}
	if out.BlockTime == nil || *out.BlockTime != solana.UnixTimeSeconds(1660570006) {
		t.Fatalf("BlockTime = %v", out.BlockTime)
	}
	if out.Meta == nil || out.Meta.Fee != 10000 {
		t.Fatalf("Meta = %+v", out.Meta)
	}
	if len(out.Meta.InnerInstructions) != 1 || len(out.Meta.InnerInstructions[0].Instructions) != 3 {
		t.Fatalf("InnerInstructions = %+v", out.Meta.InnerInstructions)
	}
	if out.Transaction == nil {
		t.Fatal("Transaction is nil")
	}
	if len(out.Transaction.Signatures) != 1 ||
		out.Transaction.Signatures[0].String() != "2x1QBpfcEQetAx7zETLEmvVvjue9311s9AWroEvMAboFkqaHZVp1sUpTFXroc5Q6tkPmZK5pYfmPFteoZPVRLF89" {
		t.Fatalf("Signatures = %v", out.Transaction.Signatures)
	}
	msg := out.Transaction.Message
	if msg.RecentBlockHash != "9L8FEB81LfZ67ejxpMaaZmC9EmXBpV38dhNaiF9UbzZi" {
		t.Fatalf("RecentBlockHash = %s", msg.RecentBlockHash)
	}
	if len(msg.AccountKeys) != 1 || !msg.AccountKeys[0].Signer || !msg.AccountKeys[0].Writable {
		t.Fatalf("AccountKeys = %+v", msg.AccountKeys)
	}
	if len(msg.Instructions) != 3 {
		t.Fatalf("Instructions = %+v", msg.Instructions)
	}
}
