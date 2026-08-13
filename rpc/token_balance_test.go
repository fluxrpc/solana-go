package rpc

import (
	"encoding/json"
	"strings"
	"testing"
)

// Fixture: preTokenBalances entry of a getTransaction response, from
// upstream gagliardetto/solana-go rpc/client_test.go
// (TestClient_GetParsedTransaction).
const tokenBalanceFixture = `{"accountIndex":4,"mint":"E942z7FnS7GpswTvF5Vggvo7cMTbvZojjLbFgsrDVff1","owner":"G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo","programId":"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA","uiTokenAmount":{"amount":"47444666","decimals":6,"uiAmount":47.444666,"uiAmountString":"47.444666"}}`

func TestTokenBalanceJSON(t *testing.T) {
	balance := jsonRoundTrip[TokenBalance](t, []byte(tokenBalanceFixture))

	if balance.AccountIndex != 4 {
		t.Fatalf("AccountIndex = %d", balance.AccountIndex)
	}
	if balance.Mint.String() != "E942z7FnS7GpswTvF5Vggvo7cMTbvZojjLbFgsrDVff1" {
		t.Fatalf("Mint = %s", balance.Mint)
	}
	if balance.Owner == nil || balance.Owner.String() != "G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo" {
		t.Fatalf("Owner = %v", balance.Owner)
	}
	if balance.ProgramId == nil || balance.ProgramId.String() != "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA" {
		t.Fatalf("ProgramId = %v", balance.ProgramId)
	}
	amount := balance.UiTokenAmount
	if amount == nil {
		t.Fatal("UiTokenAmount is nil")
	}
	if amount.Amount != "47444666" || amount.Decimals != 6 || amount.UiAmountString != "47.444666" {
		t.Fatalf("UiTokenAmount = %+v", amount)
	}
	if amount.UiAmount == nil || *amount.UiAmount != 47.444666 {
		t.Fatalf("UiAmount = %v", amount.UiAmount)
	}
}

func TestTokenBalanceOptionalFields(t *testing.T) {
	// Old responses omit owner and programId; uiAmount is null for a zero
	// balance.
	in := `{"accountIndex":1,"mint":"E942z7FnS7GpswTvF5Vggvo7cMTbvZojjLbFgsrDVff1","uiTokenAmount":{"amount":"0","decimals":6,"uiAmount":null,"uiAmountString":"0"}}`
	balance := jsonRoundTrip[TokenBalance](t, []byte(in))

	if balance.Owner != nil || balance.ProgramId != nil {
		t.Fatalf("optional pubkeys not nil: %+v", balance)
	}
	if balance.UiTokenAmount.UiAmount != nil {
		t.Fatalf("UiAmount = %v, want nil", *balance.UiTokenAmount.UiAmount)
	}

	data, err := json.Marshal(balance)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "owner") || strings.Contains(string(data), "programId") {
		t.Fatalf("marshal did not omit nil owner/programId: %s", data)
	}
	// uiAmount has no omitempty: null must be preserved.
	if !strings.Contains(string(data), `"uiAmount":null`) {
		t.Fatalf("marshal dropped null uiAmount: %s", data)
	}
}
