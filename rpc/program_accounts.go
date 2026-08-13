package rpc

import solana "github.com/fluxrpc/solana-go"

// GetProgramAccountsOpts are the options of the getProgramAccounts RPC method.
type GetProgramAccountsOpts struct {
	Commitment CommitmentType `json:"commitment,omitempty"`

	Encoding solana.EncodingType `json:"encoding,omitempty"`

	// Limit the returned account data.
	DataSlice *DataSlice `json:"dataSlice,omitempty"`

	// Filter on accounts, implicit AND between filters.
	// Filter results using various filter objects;
	// account must meet all filter criteria to be included in results.
	Filters []RPCFilter `json:"filters,omitempty"`

	// Wrap the result in an RpcResponse JSON object with context.
	WithContext *bool `json:"withContext,omitempty"`

	// Sort the results (useful for deterministic pagination).
	SortResults *bool `json:"sortResults,omitempty"`

	// The minimum slot that the request can be evaluated at.
	MinContextSlot *uint64 `json:"minContextSlot,omitempty"`
}

// GetProgramAccountsResult is the response of the getProgramAccounts RPC method.
type GetProgramAccountsResult []*KeyedAccount

// GetProgramAccountsWithContextResult is the response of the
// getProgramAccounts RPC method when withContext is true.
type GetProgramAccountsWithContextResult struct {
	RPCContext
	Value GetProgramAccountsResult `json:"value"`
}
