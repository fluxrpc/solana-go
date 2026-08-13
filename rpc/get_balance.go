package rpc

// GetBalanceOpts is the optional configuration object for `getBalance`,
// mirroring the JSON-RPC spec for this method.
//
// See https://solana.com/docs/rpc/http/getbalance for the canonical field list.
type GetBalanceOpts struct {
	// Commitment level to query the balance at.
	Commitment CommitmentType

	// MinContextSlot is the minimum slot at which the RPC node should
	// have processed the request. The validator returns a
	// `MinContextSlotNotReached` error to the caller if the local slot
	// has not yet caught up, instead of silently serving stale state.
	MinContextSlot *uint64
}
