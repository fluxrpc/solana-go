package rpc

// GetInflationRateResult is the response of the getInflationRate RPC method.
type GetInflationRateResult struct {
	// Total inflation.
	Total float64 `json:"total"`

	// Inflation allocated to validators.
	Validator float64 `json:"validator"`

	// Inflation allocated to the foundation.
	Foundation float64 `json:"foundation"`

	// Epoch for which these values are valid.
	Epoch uint64 `json:"epoch"`
}
