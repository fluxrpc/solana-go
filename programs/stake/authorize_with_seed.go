package stake

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type AuthorizeWithSeed struct {
	Args AuthorizeWithSeedArgs
	instruction
}

func NewAuthorizeWithSeedInstruction(args AuthorizeWithSeedArgs, stakeAccount, authorityBase solana.PublicKey, custodian *solana.PublicKey) (*AuthorizeWithSeed, error) {
	if !args.StakeAuthorize.valid() {
		return nil, ErrInvalidStakeAuthorize
	}
	if len(args.AuthoritySeed) > solana.MaxSeedLength {
		return nil, solana.ErrMaxSeedLengthExceeded
	}
	accounts := make(solana.AccountMetaSlice, 0, 4)
	accounts = append(accounts,
		solana.NewAccountMeta(stakeAccount, true, false),
		solana.NewAccountMeta(authorityBase, false, true),
		solana.NewAccountMeta(solana.SysVarClockPubkey, false, false),
	)
	if custodian != nil {
		accounts = append(
			accounts,
			solana.NewAccountMeta(*custodian, false, true),
		)
	}
	return &AuthorizeWithSeed{Args: args, instruction: instruction{AccountMetaSlice: accounts}}, nil
}

func (inst *AuthorizeWithSeed) Data() ([]byte, error) {
	if !inst.Args.StakeAuthorize.valid() {
		return nil, ErrInvalidStakeAuthorize
	}
	if len(inst.Args.AuthoritySeed) > solana.MaxSeedLength {
		return nil, solana.ErrMaxSeedLengthExceeded
	}
	enc := binary.NewEncoder(make([]byte, 0, 80+len(inst.Args.AuthoritySeed)))
	enc.WriteUint32(uint32(AuthorizeWithSeedInstruction))
	enc.WritePublicKey(inst.Args.NewAuthorized)
	enc.WriteUint32(uint32(inst.Args.StakeAuthorize))
	enc.WriteBincodeString(inst.Args.AuthoritySeed)
	enc.WritePublicKey(inst.Args.AuthorityOwner)
	return enc.Bytes(), enc.Err()
}
