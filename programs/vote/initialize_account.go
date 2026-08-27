package vote

import (
	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type InitializeAccount struct {
	VoteInit
	instruction
}

func NewInitializeAccountInstruction(nodePubkey, authorizedVoter, authorizedWithdrawer solana.PublicKey, commission uint8, voteAccount solana.PublicKey) *InitializeAccount {
	return &InitializeAccount{VoteInit: VoteInit{NodePubkey: nodePubkey, AuthorizedVoter: authorizedVoter, AuthorizedWithdrawer: authorizedWithdrawer, Commission: commission}, instruction: instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(voteAccount, true, false),
		solana.NewAccountMeta(solana.SysVarRentPubkey, false, false),
		solana.NewAccountMeta(solana.SysVarClockPubkey, false, false),
		solana.NewAccountMeta(nodePubkey, false, true),
	}}}
}
func (inst *InitializeAccount) Data() ([]byte, error) {
	enc := bin.NewEncoder(make([]byte, 0, 101))
	enc.WriteUint32(uint32(InitializeAccountInstruction))
	inst.VoteInit.write(enc)
	return enc.Bytes(), enc.Err()
}
