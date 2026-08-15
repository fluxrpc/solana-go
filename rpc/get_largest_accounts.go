package rpc

import (
	solana "github.com/fluxrpc/solana-go"
)

// LargestAccountsFilterType filters getLargestAccounts results by account
// type.
type LargestAccountsFilterType string

const (
	// LargestAccountsFilterCirculating restricts results to circulating
	// accounts.
	LargestAccountsFilterCirculating LargestAccountsFilterType = "circulating"
	// LargestAccountsFilterNonCirculating restricts results to
	// non-circulating accounts.
	LargestAccountsFilterNonCirculating LargestAccountsFilterType = "nonCirculating"
)

// GetLargestAccountsOpts is the optional configuration object for the
// getLargestAccounts RPC method.
type GetLargestAccountsOpts struct {
	Commitment  CommitmentType            `json:"commitment,omitempty"`
	Filter      LargestAccountsFilterType `json:"filter,omitempty"`
	SortResults *bool                     `json:"sortResults,omitempty"`
}

// GetLargestAccountsResult is the response of the getLargestAccounts RPC
// method.
type GetLargestAccountsResult struct {
	RPCContext
	Value []LargestAccountsResult `json:"value"`
}

// LargestAccountsResult is one entry of a getLargestAccounts response.
type LargestAccountsResult struct {
	// Address of the account.
	Address solana.PublicKey `json:"address"`

	// Number of lamports in the account.
	Lamports uint64 `json:"lamports"`
}
