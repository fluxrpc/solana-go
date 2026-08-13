package rpc

// GetFeeForMessageOpts groups the optional configuration accepted by the
// getFeeForMessage RPC.
type GetFeeForMessageOpts struct {
	Commitment CommitmentType

	// The minimum slot that the request can be evaluated at.
	MinContextSlot *uint64
}

// GetFeeForMessageResult is the response of the getFeeForMessage RPC method.
type GetFeeForMessageResult struct {
	RPCContext

	// Fee corresponding to the message at the specified blockhash.
	Value *uint64 `json:"value"`
}
