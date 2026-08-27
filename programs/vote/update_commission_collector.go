package vote

import (
	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type UpdateCommissionCollector struct {
	Kind CommissionKind
	instruction
}

func NewUpdateCommissionCollectorInstruction(kind CommissionKind, voteAccount, newCollector, withdrawAuthority solana.PublicKey) *UpdateCommissionCollector {
	return &UpdateCommissionCollector{Kind: kind, instruction: instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(voteAccount, true, false),
		solana.NewAccountMeta(newCollector, true, false),
		solana.NewAccountMeta(withdrawAuthority, false, true),
	}}}
}
func (inst *UpdateCommissionCollector) Data() ([]byte, error) {
	if err := inst.Kind.validate(); err != nil {
		return nil, err
	}
	enc := bin.NewEncoder(make([]byte, 0, 8))
	enc.WriteUint32(uint32(UpdateCommissionCollectorInstruction))
	enc.WriteUint32(uint32(inst.Kind))
	return enc.Bytes(), enc.Err()
}
