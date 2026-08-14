package rpc

import (
	solana "github.com/fluxrpc/solana-go"
)

// GetSignaturesForAddressOpts is the optional configuration object for the
// getSignaturesForAddress method.
type GetSignaturesForAddressOpts struct {
	// (optional) Maximum transaction signatures to return (between 1 and 1,000, default: 1,000).
	Limit *int `json:"limit,omitempty"`

	// (optional) Start searching backwards from this transaction signature.
	// If not provided the search starts from the top of the highest max confirmed block.
	Before solana.Signature `json:"before,omitempty"`

	// (optional) Search until this transaction signature, if found before limit reached.
	Until solana.Signature `json:"until,omitempty"`

	// (optional) Commitment; "processed" is not supported.
	// If parameter not provided, the default is "finalized".
	Commitment CommitmentType `json:"commitment,omitempty"`

	// The minimum slot that the request can be evaluated at.
	// This parameter is optional.
	MinContextSlot *uint64 `json:"minContextSlot,omitempty"`
}
