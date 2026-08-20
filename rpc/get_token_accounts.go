package rpc

import (
	"errors"

	solana "github.com/fluxrpc/solana-go"
)

// GetTokenAccountsConfig selects which token accounts to return, by mint or
// by owning token program. Exactly one of the fields must be set.
type GetTokenAccountsConfig struct {
	// Pubkey of the specific token Mint to limit accounts to.
	Mint *solana.PublicKey `json:"mint,omitempty"`

	// OR:

	// Pubkey of the Token program ID that owns the accounts.
	ProgramId *solana.PublicKey `json:"programId,omitempty"`
}

// Validate checks the Solana RPC requirement that exactly one token-account
// selector is present.
func (c *GetTokenAccountsConfig) Validate() error {
	if c == nil {
		return errors.New("token accounts config is required")
	}
	if (c.Mint == nil) == (c.ProgramId == nil) {
		return errors.New("token accounts config must set exactly one of mint or programId")
	}
	return nil
}

// GetTokenAccountsOpts is the optional configuration object for the
// getTokenAccountsByDelegate and getTokenAccountsByOwner methods.
type GetTokenAccountsOpts struct {
	Commitment CommitmentType `json:"commitment,omitempty"`

	Encoding solana.EncodingType `json:"encoding,omitempty"`

	DataSlice *DataSlice `json:"dataSlice,omitempty"`

	// The minimum slot that the request can be evaluated at.
	MinContextSlot *uint64 `json:"minContextSlot,omitempty"`
}

// GetTokenAccountsResult is the result of the getTokenAccountsByDelegate and
// getTokenAccountsByOwner methods.
type GetTokenAccountsResult struct {
	RPCContext
	Value []*TokenAccount `json:"value"`
}

// TokenAccount pairs a token account with its public key.
type TokenAccount struct {
	Pubkey  solana.PublicKey `json:"pubkey"`
	Account Account          `json:"account"`
}
