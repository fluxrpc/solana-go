package vote

import (
	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type TowerSyncSwitch struct {
	TowerSyncUpdate
	ProofHash solana.Hash
	instruction
}

func NewTowerSyncSwitchInstruction(sync TowerSyncUpdate, proofHash solana.Hash, voteAccount, voteAuthority solana.PublicKey) *TowerSyncSwitch {
	return &TowerSyncSwitch{TowerSyncUpdate: sync, ProofHash: proofHash, instruction: instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(voteAccount, true, false),
		solana.NewAccountMeta(voteAuthority, false, true),
	}}}
}
func (inst *TowerSyncSwitch) Data() ([]byte, error) {
	enc := bin.NewEncoder(make([]byte, 0, 119+len(inst.Lockouts)*3))
	enc.WriteUint32(uint32(TowerSyncSwitchInstruction))
	if err := inst.TowerSyncUpdate.write(enc); err != nil {
		return nil, err
	}
	enc.WriteHash(inst.ProofHash)
	return enc.Bytes(), enc.Err()
}
