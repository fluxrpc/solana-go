package rpc

// GetVersionResult is the result of the getVersion method.
type GetVersionResult struct {
	// Software version of solana-core.
	SolanaCore string `json:"solana-core"`

	// Unique identifier of the current software's feature set.
	FeatureSet int64 `json:"feature-set"`

	// Provider-specific extension (e.g. set by fluxrpc); not part of the
	// standard Solana RPC response.
	Custom any `json:"custom,omitempty"`
}
