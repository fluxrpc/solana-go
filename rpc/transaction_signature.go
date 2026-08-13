package rpc

import (
	solana "github.com/fluxrpc/solana-go"
)

// ConfirmationStatusType is a transaction's cluster confirmation status.
type ConfirmationStatusType string

const (
	ConfirmationStatusProcessed ConfirmationStatusType = "processed"
	ConfirmationStatusConfirmed ConfirmationStatusType = "confirmed"
	ConfirmationStatusFinalized ConfirmationStatusType = "finalized"
)

// TransactionSignature is a transaction as listed in signature-oriented
// responses (e.g. getSignaturesForAddress).
type TransactionSignature struct {
	// Error if transaction failed, nil if transaction succeeded.
	Err any `json:"err"`

	// Memo associated with the transaction, nil if no memo is present.
	Memo *string `json:"memo"`

	// Transaction signature.
	Signature solana.Signature `json:"signature"`

	// The slot that contains the block with the transaction.
	Slot uint64 `json:"slot,omitempty"`

	// Estimated production time, as Unix timestamp (seconds since the Unix epoch)
	// of when transaction was processed. Nil if not available.
	BlockTime *solana.UnixTimeSeconds `json:"blockTime,omitempty"`

	ConfirmationStatus ConfirmationStatusType `json:"confirmationStatus,omitempty"`

	// The transaction's index within the block.
	TransactionIndex *uint32 `json:"transactionIndex,omitempty"`
}
