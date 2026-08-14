package rpc

import (
	solana "github.com/fluxrpc/solana-go"
)

// RequestAirdropOpts is the optional configuration object for the
// requestAirdrop method.
type RequestAirdropOpts struct {
	Commitment CommitmentType `json:"commitment,omitempty"`

	// Must be a recent blockhash as a base-58 encoded string.
	// If not provided, a recent blockhash is used.
	RecentBlockhash *solana.Hash `json:"recentBlockhash,omitempty"`
}
