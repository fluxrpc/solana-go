package rpc

// GetEpochInfoOpts is the optional configuration for getEpochInfo.
type GetEpochInfoOpts struct {
	Commitment     CommitmentType `json:"commitment,omitempty"`
	MinContextSlot *uint64        `json:"minContextSlot,omitempty"`
}

// GetEpochInfoResult is the response of the getEpochInfo RPC method.
type GetEpochInfoResult struct {
	// The current slot.
	AbsoluteSlot uint64 `json:"absoluteSlot"`

	// The current block height.
	BlockHeight uint64 `json:"blockHeight"`

	// The current epoch.
	Epoch uint64 `json:"epoch"`

	// The current slot relative to the start of the current epoch.
	SlotIndex uint64 `json:"slotIndex"`

	// The number of slots in this epoch.
	SlotsInEpoch uint64 `json:"slotsInEpoch"`

	TransactionCount *uint64 `json:"transactionCount,omitempty"`
}
