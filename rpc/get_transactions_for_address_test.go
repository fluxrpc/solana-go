package rpc

import (
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

// Fixture from upstream gagliardetto/solana-go
// rpc/getTransactionsForAddress_test.go (TestClient_GetTransactionsForAddress),
// signatures detail level.
const getTransactionsForAddressFixture = `{"data":[{"slot":83994671,"transactionIndex":3,"blockTime":1625231961,"signature":"4Yig3yd33o2hyZV2qZBJkScDArwVmzurkxhBfKdqJeujTrdKHwrR3U8KR6LrhN5eWNTyugS5rkkYagVXCNnk7pks","err":null,"memo":null,"confirmationStatus":"finalized"},{"slot":83994656,"transactionIndex":1,"blockTime":1625231952,"signature":"3oQ7qqpJs5CtH1Xnnn8Ru5MtxkR3SZgshqzXwokuxFRArLihKdvCb9km6gbSiiUaNSHE7zVJqUVUZGfYuEaqWZPV","err":null,"memo":null,"confirmationStatus":"finalized"}],"paginationToken":"83994656:1"}`

func TestGetTransactionsForAddressResultJSON(t *testing.T) {
	out := jsonRoundTrip[GetTransactionsForAddressResult](t, []byte(getTransactionsForAddressFixture))

	if len(out.Data) != 2 {
		t.Fatalf("Data = %+v", out.Data)
	}
	first := out.Data[0]
	if first.Slot != 83994671 || first.TransactionIndex != 3 {
		t.Fatalf("Data[0] = %+v", first)
	}
	if first.BlockTime == nil || *first.BlockTime != solana.UnixTimeSeconds(1625231961) {
		t.Fatalf("BlockTime = %v", first.BlockTime)
	}
	if first.Signature.String() != "4Yig3yd33o2hyZV2qZBJkScDArwVmzurkxhBfKdqJeujTrdKHwrR3U8KR6LrhN5eWNTyugS5rkkYagVXCNnk7pks" {
		t.Fatalf("Signature = %s", first.Signature)
	}
	if first.ConfirmationStatus != ConfirmationStatusFinalized {
		t.Fatalf("ConfirmationStatus = %s", first.ConfirmationStatus)
	}
	if first.Err != nil || first.Memo != nil || first.Transaction != nil || first.Meta != nil {
		t.Fatalf("unexpected full-details fields: %+v", first)
	}
	if out.PaginationToken == nil || *out.PaginationToken != "83994656:1" {
		t.Fatalf("PaginationToken = %v", out.PaginationToken)
	}
}

func TestGetTransactionsForAddressOptsJSON(t *testing.T) {
	// Hand-built fixture mirroring the request object asserted in the
	// upstream TestClient_GetTransactionsForAddress.
	in := `{"transactionDetails":"signatures","sortOrder":"desc","limit":10,"paginationToken":"83999999:0","commitment":"finalized","minContextSlot":123456,"filters":{"status":"succeeded","slot":{"gte":83000000},"tokenTransfer":{"with":"7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932","direction":"in","amount":{"lt":500}}}}`
	opts := jsonRoundTrip[GetTransactionsForAddressOpts](t, []byte(in))

	if opts.TransactionDetails != TransactionDetailsType("signatures") {
		t.Fatalf("TransactionDetails = %s", opts.TransactionDetails)
	}
	if opts.SortOrder != TransactionsForAddressSortDesc {
		t.Fatalf("SortOrder = %s", opts.SortOrder)
	}
	if opts.Limit == nil || *opts.Limit != 10 {
		t.Fatalf("Limit = %v", opts.Limit)
	}
	if opts.Filters == nil || opts.Filters.Status != TransactionStatusSucceeded {
		t.Fatalf("Filters = %+v", opts.Filters)
	}
	if opts.Filters.Slot == nil || opts.Filters.Slot.Gte == nil || *opts.Filters.Slot.Gte != 83000000 {
		t.Fatalf("Filters.Slot = %+v", opts.Filters.Slot)
	}
	tt := opts.Filters.TokenTransfer
	if tt == nil || tt.Direction != TokenTransferIn || tt.Amount == nil || tt.Amount.Lt == nil || *tt.Amount.Lt != 500 {
		t.Fatalf("Filters.TokenTransfer = %+v", tt)
	}
}
