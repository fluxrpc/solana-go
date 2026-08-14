package rpc

// GetFeeForMessageOpts groups the optional configuration accepted by the
// getFeeForMessage RPC.
type GetFeeForMessageOpts struct {
	Commitment CommitmentType `json:"commitment,omitempty"`

	// The minimum slot that the request can be evaluated at.
	MinContextSlot *uint64 `json:"minContextSlot,omitempty"`
}

// GetFeeForMessageResult is the response of the getFeeForMessage RPC method.
type GetFeeForMessageResult struct {
	RPCContext

	// Fee corresponding to the message at the specified blockhash.
	Value *uint64 `json:"value"`
}
