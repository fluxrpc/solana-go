package rpc

import (
	solana "github.com/fluxrpc/solana-go"
)

// GetTokenLargestAccountsResult is the result of the getTokenLargestAccounts
// method.
type GetTokenLargestAccountsResult struct {
	RPCContext
	Value []*TokenLargestAccountsResult `json:"value"`
}

// TokenLargestAccountsResult is a single entry of a getTokenLargestAccounts
// response.
type TokenLargestAccountsResult struct {
	Address solana.PublicKey `json:"address"` // the address of the token account
	UiTokenAmount
}
