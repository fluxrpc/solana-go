package rpc

import (
	solana "github.com/fluxrpc/solana-go"
)

// GetLatestBlockhashOpts is the optional configuration object for
// `getLatestBlockhash`, mirroring the JSON-RPC spec for this method.
//
// See https://solana.com/docs/rpc/http/getlatestblockhash.
type GetLatestBlockhashOpts struct {
	// Commitment level to query the blockhash at.
	Commitment CommitmentType `json:"commitment,omitempty"`

	// MinContextSlot is the minimum slot at which the RPC node should
	// have processed the request. The validator returns a
	// `MinContextSlotNotReached` error to the caller if the local slot
	// has not yet caught up, instead of silently serving stale state.
	MinContextSlot *uint64 `json:"minContextSlot,omitempty"`
}

// GetLatestBlockhashResult is the response of the getLatestBlockhash RPC
// method.
type GetLatestBlockhashResult struct {
	RPCContext
	Value *LatestBlockhashResult `json:"value"`
}

// LatestBlockhashResult is the value of a getLatestBlockhash response.
type LatestBlockhashResult struct {
	Blockhash            solana.Hash `json:"blockhash"`
	LastValidBlockHeight uint64      `json:"lastValidBlockHeight"` // Slot.
}
