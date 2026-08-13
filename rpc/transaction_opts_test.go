package rpc

import (
	"reflect"
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

func TestTransactionOptsToMapDefaults(t *testing.T) {
	opts := TransactionOpts{}
	got := opts.ToMap()

	want := M{
		"encoding":      "base64",
		"skipPreflight": false,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ToMap() = %v, want %v", got, want)
	}
}

func TestTransactionOptsToMap(t *testing.T) {
	maxRetries := uint(5)
	minContextSlot := uint64(1234)
	opts := TransactionOpts{
		Encoding:            solana.EncodingBase58,
		SkipPreflight:       true,
		PreflightCommitment: CommitmentProcessed,
		MaxRetries:          &maxRetries,
		MinContextSlot:      &minContextSlot,
	}
	got := opts.ToMap()

	want := M{
		"encoding":            solana.EncodingBase58,
		"skipPreflight":       true,
		"preflightCommitment": CommitmentProcessed,
		"maxRetries":          uint(5),
		"minContextSlot":      uint64(1234),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ToMap() = %v, want %v", got, want)
	}
}
