package rpc

// GetTokenAccountBalanceOpts is the optional configuration object for
// `getTokenAccountBalance`, mirroring the JSON-RPC spec for this method.
//
// See https://solana.com/docs/rpc/http/gettokenaccountbalance.
type GetTokenAccountBalanceOpts struct {
	// Commitment level to query the balance at.
	Commitment CommitmentType

	// MinContextSlot is the minimum slot at which the RPC node should
	// have processed the request. The validator returns a
	// `MinContextSlotNotReached` error to the caller if the local slot
	// has not yet caught up, instead of silently serving stale state.
	MinContextSlot *uint64
}

// GetTokenAccountBalanceResult is the result of the getTokenAccountBalance
// method.
type GetTokenAccountBalanceResult struct {
	RPCContext
	Value *UiTokenAmount `json:"value"`
}
