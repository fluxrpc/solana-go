package token2022

import (
	"encoding/binary"
	"fmt"

	solana "github.com/fluxrpc/solana-go"
)

type ElGamalRegistryInstructionData struct {
	InstructionOffset int8
}

type ConfidentialRecordInstructionData struct {
	Offset     uint64
	Data       []byte
	DataLength uint64
}

func (service ConfidentialTransferService) DecodeElGamalRegistryInstruction(accounts solana.AccountMetaSlice, data []byte) (*ConfidentialAuxiliaryInstruction, *ElGamalRegistryInstructionData, error) {
	if len(data) != 2 || data[0] > 1 {
		return nil, nil, fmt.Errorf("decode ElGamal registry instruction: invalid data")
	}
	decoded := &ElGamalRegistryInstructionData{InstructionOffset: int8(data[1])}
	return &ConfidentialAuxiliaryInstruction{programID: service.RegistryProgramID, AccountMetaSlice: accounts, data: data}, decoded, nil
}

func (service ConfidentialTransferService) DecodeConfidentialRecordInstruction(accounts solana.AccountMetaSlice, data []byte) (*ConfidentialAuxiliaryInstruction, *ConfidentialRecordInstructionData, error) {
	if len(data) == 0 || data[0] > 4 {
		return nil, nil, fmt.Errorf("decode confidential record instruction: invalid data")
	}
	decoded := &ConfidentialRecordInstructionData{}
	switch data[0] {
	case 0, 2, 3:
		if len(data) != 1 {
			return nil, nil, fmt.Errorf("decode confidential record instruction %d: invalid data", data[0])
		}
	case 1:
		if len(data) < 13 {
			return nil, nil, fmt.Errorf("decode confidential record write: invalid data")
		}
		decoded.Offset = binary.LittleEndian.Uint64(data[1:9])
		length := binary.LittleEndian.Uint32(data[9:13])
		if uint64(length) != uint64(len(data)-13) {
			return nil, nil, fmt.Errorf("decode confidential record write: invalid data length")
		}
		decoded.Data = data[13:]
	case 4:
		if len(data) != 9 {
			return nil, nil, fmt.Errorf("decode confidential record reallocate: invalid data")
		}
		decoded.DataLength = binary.LittleEndian.Uint64(data[1:])
	}
	return &ConfidentialAuxiliaryInstruction{programID: service.RecordProgramID, AccountMetaSlice: accounts, data: data}, decoded, nil
}
