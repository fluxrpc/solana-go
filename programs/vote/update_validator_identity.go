package vote

import solana "github.com/fluxrpc/solana-go"

type UpdateValidatorIdentity struct{ instruction }

func NewUpdateValidatorIdentityInstruction(voteAccount, newIdentity, withdrawAuthority solana.PublicKey) *UpdateValidatorIdentity {
	return &UpdateValidatorIdentity{instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(voteAccount, true, false),
		solana.NewAccountMeta(newIdentity, false, true),
		solana.NewAccountMeta(withdrawAuthority, false, true),
	}}}
}
func (*UpdateValidatorIdentity) Data() ([]byte, error) {
	return []byte{byte(UpdateValidatorIdentityInstruction), 0, 0, 0}, nil
}
