package rpc

// GetVersionResult is the result of the getVersion method.
type GetVersionResult struct {
	// Software version of solana-core.
	SolanaCore string `json:"solana-core"`

	// Unique identifier of the current software's feature set.
	FeatureSet int64 `json:"feature-set"`
}
