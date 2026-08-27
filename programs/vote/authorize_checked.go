package vote

import (
	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type AuthorizeChecked struct {
	AuthorizationType VoteAuthorize
	instruction
}

func NewAuthorizeCheckedInstruction(authorizationType VoteAuthorize, voteAccount, currentAuthority, newAuthority solana.PublicKey) *AuthorizeChecked {
	return &AuthorizeChecked{AuthorizationType: authorizationType, instruction: instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(voteAccount, true, false),
		solana.NewAccountMeta(solana.SysVarClockPubkey, false, false),
		solana.NewAccountMeta(currentAuthority, false, true),
		solana.NewAccountMeta(newAuthority, false, true),
	}}}
}
func (inst *AuthorizeChecked) Data() ([]byte, error) {
	enc := bin.NewEncoder(make([]byte, 0, 8))
	enc.WriteUint32(uint32(AuthorizeCheckedInstruction))
	if err := inst.AuthorizationType.write(enc); err != nil {
		return nil, err
	}
	return enc.Bytes(), enc.Err()
}
