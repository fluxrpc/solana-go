package vote

import (
	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type Authorize struct {
	NewAuthority      solana.PublicKey
	AuthorizationType VoteAuthorize
	instruction
}

func NewAuthorizeInstruction(newAuthority solana.PublicKey, authorizationType VoteAuthorize, voteAccount, currentAuthority solana.PublicKey) *Authorize {
	return &Authorize{NewAuthority: newAuthority, AuthorizationType: authorizationType, instruction: instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(voteAccount, true, false),
		solana.NewAccountMeta(solana.SysVarClockPubkey, false, false),
		solana.NewAccountMeta(currentAuthority, false, true),
	}}}
}
func (inst *Authorize) Data() ([]byte, error) {
	enc := bin.NewEncoder(make([]byte, 0, 40))
	enc.WriteUint32(uint32(AuthorizeInstruction))
	enc.WritePublicKey(inst.NewAuthority)
	if err := inst.AuthorizationType.write(enc); err != nil {
		return nil, err
	}
	return enc.Bytes(), enc.Err()
}
