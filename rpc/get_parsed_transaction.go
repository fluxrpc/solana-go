package rpc

import (
	solana "github.com/fluxrpc/solana-go"
)

// GetParsedTransactionOpts is the optional configuration object for
// getTransaction when requesting the jsonParsed encoding.
type GetParsedTransactionOpts struct {
	// Desired commitment. "processed" is not supported. If parameter not provided, the default is "finalized".
	Commitment CommitmentType `json:"commitment,omitempty"`

	// Max transaction version to return in responses.
	// If the requested block contains a transaction with a higher version, an error will be returned.
	MaxSupportedTransactionVersion *uint64
}

// GetParsedTransactionResult is the result of getTransaction with the
// jsonParsed encoding.
type GetParsedTransactionResult struct {
	Slot        uint64
	BlockTime   *solana.UnixTimeSeconds
	Transaction *ParsedTransaction
	Meta        *ParsedTransactionMeta
	Version     TransactionVersion `json:"version"`
}
