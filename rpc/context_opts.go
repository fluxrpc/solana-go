package rpc

// GetBlockHeightOpts is the optional configuration for getBlockHeight.
type GetBlockHeightOpts struct {
	Commitment     CommitmentType `json:"commitment,omitempty"`
	MinContextSlot *uint64        `json:"minContextSlot,omitempty"`
}

// GetTransactionCountOpts is the optional configuration for
// getTransactionCount.
type GetTransactionCountOpts struct {
	Commitment     CommitmentType `json:"commitment,omitempty"`
	MinContextSlot *uint64        `json:"minContextSlot,omitempty"`
}

// GetSlotLeaderOpts is the optional configuration for getSlotLeader.
type GetSlotLeaderOpts struct {
	Commitment     CommitmentType `json:"commitment,omitempty"`
	MinContextSlot *uint64        `json:"minContextSlot,omitempty"`
}
