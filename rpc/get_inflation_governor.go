package rpc

// GetInflationGovernorResult is the response of the getInflationGovernor RPC
// method.
type GetInflationGovernorResult struct {
	// The initial inflation percentage from time 0.
	Initial float64 `json:"initial"`

	// Terminal inflation percentage.
	Terminal float64 `json:"terminal"`

	// Rate per year at which inflation is lowered. Rate reduction is derived using the target slot time in genesis config.
	Taper float64 `json:"taper"`

	// Percentage of total inflation allocated to the foundation.
	Foundation float64 `json:"foundation"`

	// Duration of foundation pool inflation in years.
	FoundationTerm float64 `json:"foundationTerm"`
}
