package rpc

// GetRecentPerformanceSamplesResult is a single performance sample returned by
// the getRecentPerformanceSamples method.
type GetRecentPerformanceSamplesResult struct {
	// Slot in which sample was taken at.
	Slot uint64 `json:"slot"`

	// Number of transactions in sample.
	NumTransactions uint64 `json:"numTransactions"`

	// Number of non-vote transactions in sample.
	NumNonVoteTransactions *uint64 `json:"numNonVoteTransactions,omitempty"`

	// Number of slots in sample.
	NumSlots uint64 `json:"numSlots"`

	// Number of seconds in a sample window.
	SamplePeriodSecs uint16 `json:"samplePeriodSecs"`
}
