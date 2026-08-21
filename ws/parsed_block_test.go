package ws

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/rpc"
)

const parsedBlockNotification = `{"context":{"slot":112301554},"value":{"slot":112301554,"block":{"blockhash":"DZDb5RGDJp9DHkAVFF7WrujcaAAdMFRcaWSBv1PsPdvp","previousBlockhash":"C1qgvE2C5sjkTz1nQhNSbg6ZYK4hhWFbSD9CzjHYBSTS","parentSlot":112301553,"blockTime":1639926816,"blockHeight":101210751,"transactions":[{"transaction":{"message":{"accountKeys":[{"pubkey":"G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo","signer":true,"writable":true}],"addressTableLookups":null,"instructions":[{"parsed":{"info":{"destination":"9bFNrXNb2WTx8fMHXCheaZqkLZ3YCCaiqTftHxeintHy","lamports":100,"source":"G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo"},"type":"transfer"},"program":"system","programId":"11111111111111111111111111111111"}],"recentBlockhash":"9L8FEB81LfZ67ejxpMaaZmC9EmXBpV38dhNaiF9UbzZi"},"signatures":["2x1QBpfcEQetAx7zETLEmvVvjue9311s9AWroEvMAboFkqaHZVp1sUpTFXroc5Q6tkPmZK5pYfmPFteoZPVRLF89"]},"meta":null}]}}}`

func TestParsedBlockSubscribe(t *testing.T) {
	ts := newWSTestServer(t)
	client := testConnect(t, ts, nil)
	ctx := context.Background()

	sub, err := client.ParsedBlockSubscribe(ctx, rpc.CommitmentConfirmed)
	if err != nil {
		t.Fatal(err)
	}
	req := ts.nextReq(t)
	if req.Method != "blockSubscribe" {
		t.Fatalf("method = %s", req.Method)
	}
	var params []json.RawMessage
	if err := json.Unmarshal(req.Params, &params); err != nil {
		t.Fatal(err)
	}
	if string(params[0]) != `"all"` {
		t.Fatalf("params[0] = %s", params[0])
	}
	var opts map[string]json.RawMessage
	if err := json.Unmarshal(params[1], &opts); err != nil {
		t.Fatal(err)
	}
	if string(opts["encoding"]) != `"jsonParsed"` || string(opts["commitment"]) != `"confirmed"` ||
		string(opts["maxSupportedTransactionVersion"]) != `1` {
		t.Fatalf("opts = %v", opts)
	}

	ts.push(1, "blockNotification", parsedBlockNotification)
	got, err := sub.Recv(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got.Value.Slot != 112301554 || got.Value.Block == nil {
		t.Fatalf("notification = %+v", got)
	}
	block := got.Value.Block
	if block.Blockhash.String() != "DZDb5RGDJp9DHkAVFF7WrujcaAAdMFRcaWSBv1PsPdvp" ||
		block.ParentSlot != 112301553 || len(block.Transactions) != 1 {
		t.Fatalf("block = %+v", block)
	}
	tx := block.Transactions[0].Transaction
	if len(tx.Message.Instructions) != 1 || tx.Message.Instructions[0].Program != "system" {
		t.Fatalf("parsed tx = %+v", tx)
	}
	if err := sub.Unsubscribe(ctx); err != nil {
		t.Fatal(err)
	}
	if req := ts.nextReq(t); req.Method != "blockUnsubscribe" {
		t.Fatalf("method = %s", req.Method)
	}
}

func TestParsedBlockSubscribeMentionsAndOpts(t *testing.T) {
	ts := newWSTestServer(t)
	client := testConnect(t, ts, nil)
	ctx := context.Background()

	if _, err := client.ParsedBlockSubscribeMentions(ctx, solana.TokenProgramID, ""); err != nil {
		t.Fatal(err)
	}
	req := ts.nextReq(t)
	if !strings.Contains(string(req.Params), `"mentionsAccountOrProgram"`) {
		t.Fatalf("params = %s", req.Params)
	}

	// WithOpts forces jsonParsed even if the caller sets another encoding.
	if _, err := client.ParsedBlockSubscribeWithOpts(ctx, "all", rpc.M{
		"encoding":           solana.EncodingBase64,
		"transactionDetails": "signatures",
	}); err != nil {
		t.Fatal(err)
	}
	req = ts.nextReq(t)
	if !strings.Contains(string(req.Params), `"jsonParsed"`) ||
		!strings.Contains(string(req.Params), `"signatures"`) {
		t.Fatalf("params = %s", req.Params)
	}
}
