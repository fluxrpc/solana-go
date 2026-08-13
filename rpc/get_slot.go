package rpc

// GetSlotOpts is the optional configuration object for `getSlot`,
// mirroring the JSON-RPC spec for this method.
//
// See https://solana.com/docs/rpc/http/getslot.
type GetSlotOpts struct {
	// Commitment level to query the slot at.
	Commitment CommitmentType

	// MinContextSlot is the minimum slot at which the RPC node should
	// have processed the request. The validator returns a
	// `MinContextSlotNotReached` error to the caller if the local slot
	// has not yet caught up, instead of silently serving stale state.
	MinContextSlot *uint64
}
