package system

import solana "github.com/fluxrpc/solana-go"

// AdvanceNonceAccount consumes a stored durable nonce and replaces it with
// its successor.
type AdvanceNonceAccount struct {
	instruction
}

// NewAdvanceNonceAccountInstruction creates an instruction that advances a
// durable nonce.
func NewAdvanceNonceAccountInstruction(
	nonceAccount solana.PublicKey,
	recentBlockhashesSysvar solana.PublicKey,
	nonceAuthority solana.PublicKey,
) *AdvanceNonceAccount {
	return &AdvanceNonceAccount{instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
		solana.NewAccountMeta(nonceAccount, true, false),
		solana.NewAccountMeta(recentBlockhashesSysvar, false, false),
		solana.NewAccountMeta(nonceAuthority, false, true),
	}}}
}

// Data returns the encoded System Program instruction data.
func (*AdvanceNonceAccount) Data() ([]byte, error) {
	return []byte{byte(AdvanceNonceAccountInstruction), 0, 0, 0}, nil
}
