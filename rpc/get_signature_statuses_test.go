package rpc

import (
	"testing"
)

// Fixture from upstream gagliardetto/solana-go rpc/client_test.go
// (TestClient_GetSignatureStatuses).
const getSignatureStatusesFixture = `{"context":{"slot":83999323},"value":[{"confirmationStatus":"finalized","confirmations":null,"err":null,"slot":82233105,"status":{"Ok":null}},{"confirmationStatus":"finalized","confirmations":null,"err":null,"slot":82232349,"status":{"Ok":null}}]}`

func TestGetSignatureStatusesResultJSON(t *testing.T) {
	out := jsonRoundTrip[GetSignatureStatusesResult](t, []byte(getSignatureStatusesFixture))

	if out.Context.Slot != 83999323 {
		t.Fatalf("Context.Slot = %d", out.Context.Slot)
	}
	if len(out.Value) != 2 {
		t.Fatalf("Value = %+v", out.Value)
	}
	first := out.Value[0]
	if first.Slot != 82233105 {
		t.Fatalf("Slot = %d", first.Slot)
	}
	if first.Confirmations != nil {
		t.Fatalf("Confirmations = %v", *first.Confirmations)
	}
	if first.Err != nil {
		t.Fatalf("Err = %v", first.Err)
	}
	if first.ConfirmationStatus != ConfirmationStatusFinalized {
		t.Fatalf("ConfirmationStatus = %s", first.ConfirmationStatus)
	}
	if _, ok := first.Status["Ok"]; !ok {
		t.Fatalf("Status = %v", first.Status)
	}
}
