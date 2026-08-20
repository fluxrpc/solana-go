package rpc

import solana "github.com/fluxrpc/solana-go"

// GetTokenAccountsIndexOpts configures FluxRPC's mint holder-index lookup.
type GetTokenAccountsIndexOpts struct {
	// Limit requests one page. Omitting it asks FluxRPC to stream a full scan.
	// Values over 100,000 are clamped by the server.
	Limit *int `json:"limit,omitempty"`

	// Cursor is the opaque continuation token returned by the previous page.
	Cursor string `json:"cursor,omitempty"`
}

// GetTokenAccountsIndexResult is the context-wrapped result of FluxRPC's
// custom getTokenAccounts method.
type GetTokenAccountsIndexResult struct {
	RPCContext
	Value TokenAccountsIndexValue `json:"value"`
}

// TokenAccountsIndexValue contains indexed token accounts and a continuation
// cursor. Cursor is empty when the scan is exhausted.
type TokenAccountsIndexValue struct {
	Accounts []TokenAccountIndexEntry `json:"accounts"`
	Cursor   string                   `json:"cursor"`
}

// TokenAccountIndexEntry is one mint holder-index entry.
type TokenAccountIndexEntry struct {
	Pubkey solana.PublicKey `json:"pubKey"`
	Owner  solana.PublicKey `json:"owner"`
	Amount uint64           `json:"amount"`
}

// GetTokenAccountsCountOpts configures FluxRPC's mint holder count.
type GetTokenAccountsCountOpts struct {
	ExcludeZero bool `json:"excludeZero,omitempty"`
}
