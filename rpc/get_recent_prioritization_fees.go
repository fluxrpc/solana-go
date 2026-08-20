package rpc

// PrioritizationFeeResult is a single entry of a getRecentPrioritizationFees
// response.
type PrioritizationFeeResult struct {
	// Slot is the slot in which the fee was observed.
	Slot uint64 `json:"slot"`

	// PrioritizationFee is the per-compute-unit fee paid by at least one
	// successfully landed transaction, in increments of 0.000001 lamports.
	PrioritizationFee uint64 `json:"prioritizationFee"`
}
