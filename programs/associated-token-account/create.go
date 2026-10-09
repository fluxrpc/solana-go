package associatedtokenaccount

import solana "github.com/fluxrpc/solana-go"

type Create struct{ instruction }

func NewCreateInstruction(payer, associatedTokenAccount, wallet, mint, tokenProgram solana.PublicKey) *Create {
	return &Create{instruction{createAccounts(payer, associatedTokenAccount, wallet, mint, tokenProgram)}}
}

func (*Create) Data() ([]byte, error) { return []byte{uint8(CreateInstruction)}, nil }

type CreateIdempotent struct{ instruction }

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
