package rpc

import (
	solana "github.com/fluxrpc/solana-go"
)

// GetSupplyOpts is the optional configuration object for the getSupply method.
type GetSupplyOpts struct {
	Commitment CommitmentType `json:"commitment,omitempty"`

	ExcludeNonCirculatingAccountsList bool `json:"excludeNonCirculatingAccountsList,omitempty"` // exclude non circulating accounts list from response
}

// GetSupplyResult is the result of the getSupply method.
type GetSupplyResult struct {
	RPCContext
	Value *SupplyResult `json:"value"`
}

// SupplyResult describes the current total, circulating and non-circulating
// supply.
type SupplyResult struct {
	// Total supply in lamports
	Total uint64 `json:"total"`

	// Circulating supply in lamports.
	Circulating uint64 `json:"circulating"`

	// Non-circulating supply in lamports.
	NonCirculating uint64 `json:"nonCirculating"`

	// An array of account addresses of non-circulating accounts.
	// If `excludeNonCirculatingAccountsList` is enabled, the returned array will be empty.
	NonCirculatingAccounts []solana.PublicKey `json:"nonCirculatingAccounts"`
}
