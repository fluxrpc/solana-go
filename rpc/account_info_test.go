package rpc

import (
	"bytes"
	"testing"
)

func TestGetAccountInfoResultJSONRoundTrip(t *testing.T) {
	fixture := []byte(`{"context":{"apiVersion":"2.0.15","slot":341197053},"value":` + string(accountFixture) + `}`)
	got := jsonRoundTrip[GetAccountInfoResult](t, fixture)

	if got.Context.Slot != 341197053 {
		t.Fatalf("Slot = %d", got.Context.Slot)
	}
	if !bytes.Equal(got.GetBinary(), []byte("test data")) {
		t.Fatalf("GetBinary() = %q", got.GetBinary())
	}
	if !bytes.Equal(got.Bytes(), got.GetBinary()) {
		t.Fatal("Bytes() != GetBinary()")
	}
}

func TestGetAccountInfoResultGetBinaryNilSafety(t *testing.T) {
	var nilResult *GetAccountInfoResult
	if nilResult.GetBinary() != nil {
		t.Fatal("nil result GetBinary() != nil")
	}
	if (&GetAccountInfoResult{}).GetBinary() != nil {
		t.Fatal("nil value GetBinary() != nil")
	}
	if (&GetAccountInfoResult{Value: &Account{}}).GetBinary() != nil {
		t.Fatal("nil data GetBinary() != nil")
	}
}

func TestSimpleResultWrappers(t *testing.T) {
	balance := jsonRoundTrip[GetBalanceResult](t, []byte(`{"context":{"slot":1114},"value":269}`))
	if balance.Value != 269 || balance.Context.Slot != 1114 {
		t.Fatalf("got %+v", balance)
	}

	stake := jsonRoundTrip[GetStakeMinimumDelegationResult](t, []byte(`{"context":{"slot":501},"value":1000000000}`))
	if stake.Value != 1000000000 {
		t.Fatalf("got %+v", stake)
	}

	blockhash := jsonRoundTrip[IsValidBlockhashResult](t, []byte(`{"context":{"slot":2483},"value":false}`))
	if blockhash.Value {
		t.Fatalf("got %+v", blockhash)
	}
}
