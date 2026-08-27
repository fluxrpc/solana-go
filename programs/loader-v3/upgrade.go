package loaderv3

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type Upgrade struct {
	CloseBuffer bool
	instruction
}

func NewUpgradeInstruction(
	closeBuffer bool,
	programData, program, buffer, spill, authority solana.PublicKey,
) *Upgrade {
	return &Upgrade{CloseBuffer: closeBuffer, instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
		solana.NewAccountMeta(programData, true, false),
		solana.NewAccountMeta(program, true, false),
		solana.NewAccountMeta(buffer, true, false),
		solana.NewAccountMeta(spill, true, false),
		solana.NewAccountMeta(solana.SysVarRentPubkey, false, false),
		solana.NewAccountMeta(solana.SysVarClockPubkey, false, false),
		solana.NewAccountMeta(authority, false, true),
	}}}
}

func (inst *Upgrade) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 5))
	enc.WriteUint32(uint32(UpgradeInstruction))
	enc.WriteBool(inst.CloseBuffer)
	return enc.Bytes(), enc.Err()
}
