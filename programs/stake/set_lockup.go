package stake

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type SetLockup struct {
	Args LockupArgs
	instruction
}

func NewSetLockupInstruction(args LockupArgs, stakeAccount, authority solana.PublicKey) *SetLockup {
	return &SetLockup{Args: args, instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
		solana.NewAccountMeta(stakeAccount, true, false),
		solana.NewAccountMeta(authority, false, true),
	}}}
}

func (inst *SetLockup) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 55))
	enc.WriteUint32(uint32(SetLockupInstruction))
	inst.Args.write(enc)
	return enc.Bytes(), enc.Err()
}
