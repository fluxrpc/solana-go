package rpc

import (
	solana "github.com/fluxrpc/solana-go"
)

// TokenBalance is the token balance of an account before or after a
// transaction, as reported in transaction meta.
type TokenBalance struct {
	// Index of the account in which the token balance is provided for.
	AccountIndex uint16 `json:"accountIndex"`

	// Pubkey of token balance's owner.
	Owner *solana.PublicKey `json:"owner,omitempty"`

	// Pubkey of token program.
	ProgramId *solana.PublicKey `json:"programId,omitempty"`

	// Pubkey of the token's mint.
	Mint          solana.PublicKey `json:"mint"`
	UiTokenAmount *UiTokenAmount   `json:"uiTokenAmount"`
}

// UiTokenAmount is a token amount in raw and display forms.
type UiTokenAmount struct {
	// Raw amount of tokens as a string, ignoring decimals.
	Amount string `json:"amount"`

	// Number of decimals configured for token's mint.
	Decimals uint8 `json:"decimals"`

	// DEPRECATED: Token amount as a float, accounting for decimals.
	UiAmount *float64 `json:"uiAmount"`

	// Token amount as a string, accounting for decimals.
	UiAmountString string `json:"uiAmountString"`
}
