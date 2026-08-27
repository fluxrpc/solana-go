package vote

import (
	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type UpdateCommissionBps struct {
	CommissionBps uint16
	Kind          CommissionKind
	instruction
}

func NewUpdateCommissionBpsInstruction(commissionBps uint16, kind CommissionKind, voteAccount, withdrawAuthority solana.PublicKey) *UpdateCommissionBps {
	return &UpdateCommissionBps{CommissionBps: commissionBps, Kind: kind, instruction: instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(voteAccount, true, false),
		solana.NewAccountMeta(withdrawAuthority, false, true),
	}}}
}
func (inst *UpdateCommissionBps) Data() ([]byte, error) {
	if err := inst.Kind.validate(); err != nil {
		return nil, err
	}
	enc := bin.NewEncoder(make([]byte, 0, 10))
	enc.WriteUint32(uint32(UpdateCommissionBpsInstruction))
	enc.WriteUint16(inst.CommissionBps)
	enc.WriteUint32(uint32(inst.Kind))
	return enc.Bytes(), enc.Err()
}
