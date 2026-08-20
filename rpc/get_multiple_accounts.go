package rpc

import solana "github.com/fluxrpc/solana-go"

// GetMultipleAccountsResult is the result of the getMultipleAccounts method.
type GetMultipleAccountsResult struct {
	RPCContext
	Value []*Account `json:"value"`
}

// GetMultipleAccountsOpts is the optional configuration object for
// getMultipleAccounts.
type GetMultipleAccountsOpts struct {
	Encoding       solana.EncodingType `json:"encoding,omitempty"`
	Commitment     CommitmentType      `json:"commitment,omitempty"`
	DataSlice      *DataSlice          `json:"dataSlice,omitempty"`
	MinContextSlot *uint64             `json:"minContextSlot,omitempty"`

	// Async asks FluxRPC to stream results as they become available.
	Async bool `json:"async,omitempty"`
}
