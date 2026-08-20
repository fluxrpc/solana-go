package rpc

// GetTokenAccountBalanceOpts is the optional configuration object for
// `getTokenAccountBalance`, mirroring the JSON-RPC spec for this method.
//
// See https://solana.com/docs/rpc/http/gettokenaccountbalance.
type GetTokenAccountBalanceOpts struct {
	// Commitment level to query the balance at.
	Commitment CommitmentType `json:"commitment,omitempty"`
}

// GetTokenAccountBalanceResult is the result of the getTokenAccountBalance
// method.
type GetTokenAccountBalanceResult struct {
	RPCContext
	Value *UiTokenAmount `json:"value"`
}
