package stake

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type AuthorizeChecked struct {
	StakeAuthorize StakeAuthorize
	instruction
}

func NewAuthorizeCheckedInstruction(authorityType StakeAuthorize, stakeAccount, authority, newAuthority solana.PublicKey, custodian *solana.PublicKey) *AuthorizeChecked {
	accounts := make(solana.AccountMetaSlice, 0, 5)
	accounts = append(accounts,
		solana.NewAccountMeta(stakeAccount, true, false),
		solana.NewAccountMeta(solana.SysVarClockPubkey, false, false),
		solana.NewAccountMeta(authority, false, true),
		solana.NewAccountMeta(newAuthority, false, true),
	)
	if custodian != nil {
		accounts = append(
			accounts,
			solana.NewAccountMeta(*custodian, false, true),
		)
	}
	return &AuthorizeChecked{StakeAuthorize: authorityType, instruction: instruction{AccountMetaSlice: accounts}}
}

func (inst *AuthorizeChecked) Data() ([]byte, error) {
	if !inst.StakeAuthorize.valid() {
		return nil, ErrInvalidStakeAuthorize
	}
	enc := binary.NewEncoder(make([]byte, 0, 8))
	enc.WriteUint32(uint32(AuthorizeCheckedInstruction))
	enc.WriteUint32(uint32(inst.StakeAuthorize))
	return enc.Bytes(), enc.Err()
}
