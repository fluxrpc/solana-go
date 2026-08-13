package rpc

// GetInflationRewardOpts is the optional configuration object for the
// getInflationReward RPC method.
type GetInflationRewardOpts struct {
	Commitment CommitmentType

	// An epoch for which the reward occurs.
	// If omitted, the previous epoch will be used.
	Epoch *uint64

	// The minimum slot that the request can be evaluated at.
	MinContextSlot *uint64
}

// GetInflationRewardResult is one entry of the getInflationReward RPC
// response; null for addresses that received no reward.
type GetInflationRewardResult struct {
	// Epoch for which reward occurred.
	Epoch uint64 `json:"epoch"`

	// The slot in which the rewards are effective.
	EffectiveSlot uint64 `json:"effectiveSlot"`

	// Reward amount in lamports.
	Amount uint64 `json:"amount"`

	// Post balance of the account in lamports.
	PostBalance uint64 `json:"postBalance"`

	// Vote account commission when the reward was credited.
	Commission *uint8 `json:"commission,omitempty"`

	// Vote account commission in basis points when the reward was credited.
	CommissionBps *uint16 `json:"commissionBps,omitempty"`
}
