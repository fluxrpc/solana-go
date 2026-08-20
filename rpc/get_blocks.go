package rpc

// GetBlocksOpts is the optional configuration shared by getBlocks and
// getBlocksWithLimit.
type GetBlocksOpts struct {
	Commitment     CommitmentType `json:"commitment,omitempty"`
	MinContextSlot *uint64        `json:"minContextSlot,omitempty"`
}

// BlocksResult is the response of the getBlocks RPC method: an array of u64
// integers listing confirmed blocks in the requested slot range.
type BlocksResult []uint64
