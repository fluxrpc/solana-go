package associatedtokenaccount

import solana "github.com/fluxrpc/solana-go"

// RecoverNested recovers tokens held by an associated token account that is
// nested inside another one.
type RecoverNested struct{ instruction }

func NewRecoverNestedInstruction(nestedAssociatedTokenAccount, nestedMint, destinationAssociatedTokenAccount, ownerAssociatedTokenAccount, ownerMint, wallet, tokenProgram solana.PublicKey) *RecoverNested {
	return &RecoverNested{instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(nestedAssociatedTokenAccount, true, false),
		solana.NewAccountMeta(nestedMint, false, false),
		solana.NewAccountMeta(destinationAssociatedTokenAccount, true, false),
		solana.NewAccountMeta(ownerAssociatedTokenAccount, false, false),
		solana.NewAccountMeta(ownerMint, false, false),
		solana.NewAccountMeta(wallet, true, true),
		solana.NewAccountMeta(tokenProgram, false, false),
	}}}
}

func (*RecoverNested) Data() ([]byte, error) { return []byte{uint8(RecoverNestedInstruction)}, nil }
