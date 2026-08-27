package loaderv3

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type DeployWithMaxDataLen struct {
	MaxDataLen  uint64
	CloseBuffer bool
	instruction
}

func NewDeployWithMaxDataLenInstruction(
	maxDataLen uint64, closeBuffer bool,
	payer, programData, program, buffer, authority solana.PublicKey,
) *DeployWithMaxDataLen {
	return &DeployWithMaxDataLen{MaxDataLen: maxDataLen, CloseBuffer: closeBuffer, instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
		solana.NewAccountMeta(payer, true, true),
		solana.NewAccountMeta(programData, true, false),
		solana.NewAccountMeta(program, true, false),
		solana.NewAccountMeta(buffer, true, false),
		solana.NewAccountMeta(solana.SysVarRentPubkey, false, false),
		solana.NewAccountMeta(solana.SysVarClockPubkey, false, false),
		solana.NewAccountMeta(solana.SystemProgramID, false, false),
		solana.NewAccountMeta(authority, false, true),
	}}}
}

func (inst *DeployWithMaxDataLen) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 13))
	enc.WriteUint32(uint32(DeployWithMaxDataLenInstruction))
	enc.WriteUint64(inst.MaxDataLen)
	enc.WriteBool(inst.CloseBuffer)
	return enc.Bytes(), enc.Err()
}
