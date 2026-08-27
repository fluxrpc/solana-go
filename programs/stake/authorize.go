package stake

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type Authorize struct {
	NewAuthorized  solana.PublicKey
	StakeAuthorize StakeAuthorize
	instruction
}

func NewAuthorizeInstruction(newAuthorized solana.PublicKey, authorityType StakeAuthorize, stakeAccount, authority solana.PublicKey, custodian *solana.PublicKey) *Authorize {
	accounts := make(solana.AccountMetaSlice, 0, 4)
	accounts = append(accounts,
		solana.NewAccountMeta(stakeAccount, true, false),
		solana.NewAccountMeta(solana.SysVarClockPubkey, false, false),
		solana.NewAccountMeta(authority, false, true),
	)
	if custodian != nil {
		accounts = append(
			accounts,
			solana.NewAccountMeta(*custodian, false, true),
		)
	}
	return &Authorize{NewAuthorized: newAuthorized, StakeAuthorize: authorityType, instruction: instruction{AccountMetaSlice: accounts}}
}

func (inst *Authorize) Data() ([]byte, error) {
	if !inst.StakeAuthorize.valid() {
		return nil, ErrInvalidStakeAuthorize
	}
	enc := binary.NewEncoder(make([]byte, 0, 40))
	enc.WriteUint32(uint32(AuthorizeInstruction))
	enc.WritePublicKey(inst.NewAuthorized)
	enc.WriteUint32(uint32(inst.StakeAuthorize))
	return enc.Bytes(), enc.Err()
}
