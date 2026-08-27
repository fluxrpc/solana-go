package vote

import (
	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type UpdateCommission struct {
	Commission uint8
	instruction
}

func NewUpdateCommissionInstruction(commission uint8, voteAccount, withdrawAuthority solana.PublicKey) *UpdateCommission {
	return &UpdateCommission{Commission: commission, instruction: instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(voteAccount, true, false),
		solana.NewAccountMeta(withdrawAuthority, false, true),
	}}}
}
func (inst *UpdateCommission) Data() ([]byte, error) {
	enc := bin.NewEncoder(make([]byte, 0, 5))
	enc.WriteUint32(uint32(UpdateCommissionInstruction))
	enc.WriteUint8(inst.Commission)
	return enc.Bytes(), enc.Err()
}
