package vote

import (
	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type AuthorizeCheckedWithSeed struct {
	VoteAuthorizeCheckedWithSeedArgs
	instruction
}

func NewAuthorizeCheckedWithSeedInstruction(args VoteAuthorizeCheckedWithSeedArgs, voteAccount, authorityBase, newAuthority solana.PublicKey) (*AuthorizeCheckedWithSeed, error) {
	if len(args.CurrentAuthorityDerivedKeySeed) > solana.MaxSeedLength {
		return nil, solana.ErrMaxSeedLengthExceeded
	}
	return &AuthorizeCheckedWithSeed{VoteAuthorizeCheckedWithSeedArgs: args, instruction: instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(voteAccount, true, false),
		solana.NewAccountMeta(solana.SysVarClockPubkey, false, false),
		solana.NewAccountMeta(authorityBase, false, true),
		solana.NewAccountMeta(newAuthority, false, true),
	}}}, nil
}
func (inst *AuthorizeCheckedWithSeed) Data() ([]byte, error) {
	if len(inst.CurrentAuthorityDerivedKeySeed) > solana.MaxSeedLength {
		return nil, solana.ErrMaxSeedLengthExceeded
	}
	enc := bin.NewEncoder(make([]byte, 0, 80+len(inst.CurrentAuthorityDerivedKeySeed)))
	enc.WriteUint32(uint32(AuthorizeCheckedWithSeedInstruction))
	if err := inst.VoteAuthorizeCheckedWithSeedArgs.write(enc); err != nil {
		return nil, err
	}
	return enc.Bytes(), enc.Err()
}
