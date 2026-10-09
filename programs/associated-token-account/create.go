package associatedtokenaccount

import solana "github.com/fluxrpc/solana-go"

// Create creates an associated token account and fails if it already exists.
type Create struct{ instruction }

// NewCreateInstruction returns a Create instruction. associatedTokenAccount
// must be the address derived from wallet, mint and tokenProgram.
func NewCreateInstruction(payer, associatedTokenAccount, wallet, mint, tokenProgram solana.PublicKey) *Create {
	return &Create{instruction{createAccounts(payer, associatedTokenAccount, wallet, mint, tokenProgram)}}
}

func (*Create) Data() ([]byte, error) { return []byte{uint8(CreateInstruction)}, nil }

// CreateIdempotent creates an associated token account, and succeeds if it
// already exists with the expected owner and mint.
type CreateIdempotent struct{ instruction }

// NewCreateIdempotentInstruction returns a CreateIdempotent instruction.
// associatedTokenAccount must be the address derived from wallet, mint and
// tokenProgram.
func NewCreateIdempotentInstruction(payer, associatedTokenAccount, wallet, mint, tokenProgram solana.PublicKey) *CreateIdempotent {
	return &CreateIdempotent{instruction{createAccounts(payer, associatedTokenAccount, wallet, mint, tokenProgram)}}
}

func (*CreateIdempotent) Data() ([]byte, error) {
	return []byte{uint8(CreateIdempotentInstruction)}, nil
}

func createAccounts(payer, associatedTokenAccount, wallet, mint, tokenProgram solana.PublicKey) solana.AccountMetaSlice {
	return solana.AccountMetaSlice{
		solana.NewAccountMeta(payer, true, true),
		solana.NewAccountMeta(associatedTokenAccount, true, false),
		solana.NewAccountMeta(wallet, false, false),
		solana.NewAccountMeta(mint, false, false),
		solana.NewAccountMeta(solana.SystemProgramID, false, false),
		solana.NewAccountMeta(tokenProgram, false, false),
	}
}
