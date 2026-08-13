package rpc

// GetStakeMinimumDelegationOpts groups the optional configuration accepted by
// the getStakeMinimumDelegation RPC.
type GetStakeMinimumDelegationOpts struct {
	Commitment CommitmentType

	// The minimum slot that the request can be evaluated at.
	MinContextSlot *uint64
}
