package rpc

// PriorizationFeeResult is a single entry of a getRecentPrioritizationFees
// response.
type PriorizationFeeResult struct {
	// Slot in which the fee was observed
	Slot uint64 `json:"slot"`

	// The per-compute-unit fee paid by at least one successfully landed transaction, specified in increments of 0.000001 lamports
	PrioritizationFee uint64 `json:"prioritizationFee"`
}
