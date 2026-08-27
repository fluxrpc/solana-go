package loaderv3

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type ExtendProgram struct {
	AdditionalBytes uint32
	instruction
}

func NewExtendProgramInstruction(
	additionalBytes uint32,
	programData, program solana.PublicKey,
) *ExtendProgram {
	return &ExtendProgram{AdditionalBytes: additionalBytes, instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
		solana.NewAccountMeta(programData, true, false),
		solana.NewAccountMeta(program, true, false),
	}}}
}

func NewExtendProgramWithPayerInstruction(
	additionalBytes uint32,
	programData, program, payer solana.PublicKey,
) *ExtendProgram {
	return &ExtendProgram{AdditionalBytes: additionalBytes, instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
		solana.NewAccountMeta(programData, true, false),
		solana.NewAccountMeta(program, true, false),
		solana.NewAccountMeta(solana.SystemProgramID, false, false),
		solana.NewAccountMeta(payer, true, true),
	}}}
}

func (inst *ExtendProgram) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 8))
	enc.WriteUint32(uint32(ExtendProgramInstruction))
	enc.WriteUint32(inst.AdditionalBytes)
	return enc.Bytes(), enc.Err()
}
