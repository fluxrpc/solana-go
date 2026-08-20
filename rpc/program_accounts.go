package rpc

import (
	"encoding/json"

	solana "github.com/fluxrpc/solana-go"
)

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

	// CursorFilter enables FluxRPC's deterministic, cursor-based pagination.
	CursorFilter *ProgramAccountsCursorFilter `json:"cursorFilter,omitempty"`

	// ChangedSinceSlot only returns accounts changed at or after this slot.
	// This is a FluxRPC extension.
	ChangedSinceSlot *uint64 `json:"changedSinceSlot,omitempty"`

	// ChangedUntilSlot only returns accounts changed at or before this slot.
	// This is a FluxRPC extension.
	ChangedUntilSlot *uint64 `json:"changedUntilSlot,omitempty"`

	// Async asks FluxRPC to stream the response as accounts become available.
	Async bool `json:"async,omitempty"`

	// SortedResults sorts results for deterministic pagination. This is a
	// FluxRPC extension whose wire name is "sortedResults".
	SortedResults *bool `json:"sortedResults,omitempty"`

	// SortResults is retained for source compatibility. FluxRPC serializes it
	// as "sortedResults"; prefer SortedResults in new code.
	SortResults *bool `json:"-"`

	// The minimum slot that the request can be evaluated at.
	MinContextSlot *uint64 `json:"minContextSlot,omitempty"`
}

// ProgramAccountsCursorFilter configures FluxRPC cursor-based pagination.
type ProgramAccountsCursorFilter struct {
	// Cursor is the last account public key returned by the previous page.
	Cursor *solana.PublicKey `json:"cursor,omitempty"`

	// Limit is the maximum number of accounts to return.
	Limit int `json:"limit"`
}

// MarshalJSON maps the deprecated SortResults field to FluxRPC's actual
// sortedResults wire field without emitting two competing options.
func (o GetProgramAccountsOpts) MarshalJSON() ([]byte, error) {
	type plain GetProgramAccountsOpts
	out := plain(o)
	if out.SortedResults == nil {
		out.SortedResults = o.SortResults
	}
	return json.Marshal(out)
}

// GetProgramAccountsResult is the response of the getProgramAccounts RPC method.
type GetProgramAccountsResult []*KeyedAccount

// GetProgramAccountsWithContextResult is the response of the
// getProgramAccounts RPC method when withContext is true.
type GetProgramAccountsWithContextResult struct {
	RPCContext
	Value GetProgramAccountsResult `json:"value"`
}
