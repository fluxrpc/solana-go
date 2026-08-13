package rpc

// CommitmentType describes how finalized a block is at a point in time.
type CommitmentType string

const (
	// The node will query the most recent block confirmed by supermajority
	// of the cluster as having reached maximum lockout, meaning the cluster
	// has recognized this block as finalized.
	CommitmentFinalized CommitmentType = "finalized"

	// The node will query the most recent block that has been voted on by
	// supermajority of the cluster. It incorporates votes from gossip and
	// replay, but does not count votes on descendants of a block, only
	// direct votes on that block.
	CommitmentConfirmed CommitmentType = "confirmed"

	// The node will query its most recent block. Note that the block may
	// still be skipped by the cluster.
	CommitmentProcessed CommitmentType = "processed"
)
