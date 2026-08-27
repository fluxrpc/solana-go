package vote

import (
	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type TowerSync struct {
	TowerSyncUpdate
	instruction
}

func NewTowerSyncInstruction(sync TowerSyncUpdate, voteAccount, voteAuthority solana.PublicKey) *TowerSync {
	return &TowerSync{TowerSyncUpdate: sync, instruction: instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(voteAccount, true, false),
		solana.NewAccountMeta(voteAuthority, false, true),
	}}}
}
func (inst *TowerSync) Data() ([]byte, error) {
	enc := bin.NewEncoder(make([]byte, 0, 87+len(inst.Lockouts)*3))
	enc.WriteUint32(uint32(TowerSyncInstruction))
	if err := inst.TowerSyncUpdate.write(enc); err != nil {
		return nil, err
	}
	return enc.Bytes(), enc.Err()
}
