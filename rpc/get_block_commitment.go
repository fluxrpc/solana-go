package rpc

// GetBlockCommitmentResult is the response of the getBlockCommitment RPC
// method.
type GetBlockCommitmentResult struct {
	// nil if Unknown block, or array of u64 integers
	// logging the amount of cluster stake in lamports
	// that has voted on the block at each depth from 0 to `MAX_LOCKOUT_HISTORY` + 1
	Commitment []uint64 `json:"commitment"`

	// Total active stake, in lamports, of the current epoch.
	TotalStake uint64 `json:"totalStake"`
}
