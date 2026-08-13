package rpc

import (
	solana "github.com/fluxrpc/solana-go"
)

// LargestAccountsFilterType filters getLargestAccounts results by account
// type.
type LargestAccountsFilterType string

const (
	LargestAccountsFilterCirculating    LargestAccountsFilterType = "circulating"
	LargestAccountsFilterNonCirculating LargestAccountsFilterType = "nonCirculating"
)

// GetLargestAccountsOpts is the optional configuration object for the
// getLargestAccounts RPC method.
type GetLargestAccountsOpts struct {
	Commitment  CommitmentType
	Filter      LargestAccountsFilterType
	SortResults *bool
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
