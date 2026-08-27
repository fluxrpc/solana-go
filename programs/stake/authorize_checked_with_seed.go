package stake

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type AuthorizeCheckedWithSeed struct {
	Args AuthorizeCheckedWithSeedArgs
	instruction
}

func NewAuthorizeCheckedWithSeedInstruction(args AuthorizeCheckedWithSeedArgs, stakeAccount, authorityBase, newAuthority solana.PublicKey, custodian *solana.PublicKey) (*AuthorizeCheckedWithSeed, error) {
	if !args.StakeAuthorize.valid() {
		return nil, ErrInvalidStakeAuthorize
	}
	if len(args.AuthoritySeed) > solana.MaxSeedLength {
		return nil, solana.ErrMaxSeedLengthExceeded
	}
	accounts := make(solana.AccountMetaSlice, 0, 5)
	accounts = append(accounts,
		solana.NewAccountMeta(stakeAccount, true, false),
		solana.NewAccountMeta(authorityBase, false, true),
		solana.NewAccountMeta(solana.SysVarClockPubkey, false, false),
		solana.NewAccountMeta(newAuthority, false, true),
	)
	if custodian != nil {
		accounts = append(
			accounts,
			solana.NewAccountMeta(*custodian, false, true),
		)
	}
	return &AuthorizeCheckedWithSeed{Args: args, instruction: instruction{AccountMetaSlice: accounts}}, nil
}

func (inst *AuthorizeCheckedWithSeed) Data() ([]byte, error) {
	if !inst.Args.StakeAuthorize.valid() {
		return nil, ErrInvalidStakeAuthorize
	}
	if len(inst.Args.AuthoritySeed) > solana.MaxSeedLength {
		return nil, solana.ErrMaxSeedLengthExceeded
	}
	enc := binary.NewEncoder(make([]byte, 0, 48+len(inst.Args.AuthoritySeed)))
	enc.WriteUint32(uint32(AuthorizeCheckedWithSeedInstruction))
	enc.WriteUint32(uint32(inst.Args.StakeAuthorize))
	enc.WriteBincodeString(inst.Args.AuthoritySeed)
	enc.WritePublicKey(inst.Args.AuthorityOwner)
	return enc.Bytes(), enc.Err()
}
