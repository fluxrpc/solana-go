package vote

import (
	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type AuthorizeWithSeed struct {
	VoteAuthorizeWithSeedArgs
	instruction
}

func NewAuthorizeWithSeedInstruction(args VoteAuthorizeWithSeedArgs, voteAccount, authorityBase solana.PublicKey) (*AuthorizeWithSeed, error) {
	if len(args.CurrentAuthorityDerivedKeySeed) > solana.MaxSeedLength {
		return nil, solana.ErrMaxSeedLengthExceeded
	}
	return &AuthorizeWithSeed{VoteAuthorizeWithSeedArgs: args, instruction: instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(voteAccount, true, false),
		solana.NewAccountMeta(solana.SysVarClockPubkey, false, false),
		solana.NewAccountMeta(authorityBase, false, true),
	}}}, nil
}
func (inst *AuthorizeWithSeed) Data() ([]byte, error) {
	if len(inst.CurrentAuthorityDerivedKeySeed) > solana.MaxSeedLength {
		return nil, solana.ErrMaxSeedLengthExceeded
	}
	enc := bin.NewEncoder(make([]byte, 0, 112+len(inst.CurrentAuthorityDerivedKeySeed)))
	enc.WriteUint32(uint32(AuthorizeWithSeedInstruction))
	if err := inst.VoteAuthorizeWithSeedArgs.write(enc); err != nil {
		return nil, err
	}
	return enc.Bytes(), enc.Err()
}
