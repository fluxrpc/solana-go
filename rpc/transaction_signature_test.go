package rpc

import (
	"encoding/json"
	"strings"
	"testing"
)

// Fixture: entry of a getSignaturesForAddress response, from upstream
// gagliardetto/solana-go rpc/client_test.go
// (TestClient_GetSignaturesForAddress).
const transactionSignatureFixture = `{"blockTime":1625231961,"confirmationStatus":"finalized","err":null,"memo":null,"signature":"4Yig3yd33o2hyZV2qZBJkScDArwVmzurkxhBfKdqJeujTrdKHwrR3U8KR6LrhN5eWNTyugS5rkkYagVXCNnk7pks","slot":83994671}`

func TestTransactionSignatureJSON(t *testing.T) {
	sig := jsonRoundTrip[TransactionSignature](t, []byte(transactionSignatureFixture))

	if sig.Err != nil {
		t.Fatalf("Err = %v", sig.Err)
	}
	if sig.Memo != nil {
		t.Fatalf("Memo = %v", *sig.Memo)
	}
	if sig.Signature.String() != "4Yig3yd33o2hyZV2qZBJkScDArwVmzurkxhBfKdqJeujTrdKHwrR3U8KR6LrhN5eWNTyugS5rkkYagVXCNnk7pks" {
		t.Fatalf("Signature = %s", sig.Signature)
	}
	if sig.Slot != 83994671 {
		t.Fatalf("Slot = %d", sig.Slot)
	}
	if sig.BlockTime == nil || int64(*sig.BlockTime) != 1625231961 {
		t.Fatalf("BlockTime = %v", sig.BlockTime)
	}
	if sig.ConfirmationStatus != ConfirmationStatusFinalized {
		t.Fatalf("ConfirmationStatus = %q", sig.ConfirmationStatus)
	}
	if sig.TransactionIndex != nil {
		t.Fatalf("TransactionIndex = %v", *sig.TransactionIndex)
	}
}

func TestTransactionSignatureOptionalFields(t *testing.T) {
	in := `{"err":{"InstructionError":[0,"InvalidAccountData"]},"memo":"[32] hello","signature":"4Yig3yd33o2hyZV2qZBJkScDArwVmzurkxhBfKdqJeujTrdKHwrR3U8KR6LrhN5eWNTyugS5rkkYagVXCNnk7pks","slot":83994671,"blockTime":1625231961,"confirmationStatus":"confirmed","transactionIndex":2}`
	sig := jsonRoundTrip[TransactionSignature](t, []byte(in))

	if sig.Err == nil {
		t.Fatal("Err = nil")
	}
	if sig.Memo == nil || *sig.Memo != "[32] hello" {
		t.Fatalf("Memo = %v", sig.Memo)
	}
	if sig.ConfirmationStatus != ConfirmationStatusConfirmed {
		t.Fatalf("ConfirmationStatus = %q", sig.ConfirmationStatus)
	}
	if sig.TransactionIndex == nil || *sig.TransactionIndex != 2 {
		t.Fatalf("TransactionIndex = %v", sig.TransactionIndex)
	}
}

func TestTransactionSignatureOmitsEmpty(t *testing.T) {
	// blockTime, confirmationStatus and transactionIndex are omitted when
	// unset; err and memo keep explicit nulls.
	data, err := json.Marshal(TransactionSignature{})
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"blockTime", "confirmationStatus", "transactionIndex"} {
		if strings.Contains(string(data), key) {
			t.Fatalf("marshal did not omit %s: %s", key, data)
		}
	}
	if !strings.Contains(string(data), `"err":null`) || !strings.Contains(string(data), `"memo":null`) {
		t.Fatalf("marshal dropped null err/memo: %s", data)
	}
}
