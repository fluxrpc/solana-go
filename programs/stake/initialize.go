package stake

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type Initialize struct {
	Authorized Authorized
	Lockup     Lockup
	instruction
}

func NewInitializeInstruction(authorized Authorized, lockup Lockup, stakeAccount solana.PublicKey) *Initialize {
	return &Initialize{Authorized: authorized, Lockup: lockup, instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
		solana.NewAccountMeta(stakeAccount, true, false),
		solana.NewAccountMeta(solana.SysVarRentPubkey, false, false),
	}}}
}

func (inst *Initialize) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 4+64+48))
	enc.WriteUint32(uint32(InitializeInstruction))
	inst.Authorized.write(enc)
	inst.Lockup.write(enc)
	return enc.Bytes(), enc.Err()
}
