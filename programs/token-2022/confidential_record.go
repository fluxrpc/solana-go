package token2022

import (
	"encoding/binary"
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	system "github.com/fluxrpc/solana-go/programs/system"
)

type ConfidentialRecord struct {
	Version   uint8
	Authority solana.PublicKey
	Data      []byte
}

type ConfidentialRecordPlan struct {
	CreateAndWrite []solana.Instruction
	Close          solana.Instruction
	Space          uint64
}

func (ConfidentialTransferService) ConfidentialRecordDataOffset() uint32 {
	return 33
}

func (service ConfidentialTransferService) InitializeConfidentialRecord(record, authority solana.PublicKey) *ConfidentialAuxiliaryInstruction {
	return &ConfidentialAuxiliaryInstruction{
		programID: service.RecordProgramID,
		AccountMetaSlice: solana.AccountMetaSlice{
			solana.NewAccountMeta(record, true, false),
			solana.NewAccountMeta(authority, false, false),
		},
		data: []byte{0},
	}
}

func (service ConfidentialTransferService) WriteConfidentialRecord(record, authority solana.PublicKey, offset uint64, data []byte) *ConfidentialAuxiliaryInstruction {
	payload := make([]byte, 13, 13+len(data))
	payload[0] = 1
	binary.LittleEndian.PutUint64(payload[1:], offset)
	binary.LittleEndian.PutUint32(payload[9:], uint32(len(data)))
	payload = append(payload, data...)
	return &ConfidentialAuxiliaryInstruction{
		programID: service.RecordProgramID,
		AccountMetaSlice: solana.AccountMetaSlice{
			solana.NewAccountMeta(record, true, false),
			solana.NewAccountMeta(authority, false, true),
		},
		data: payload,
	}
}

func (service ConfidentialTransferService) SetConfidentialRecordAuthority(record, authority, newAuthority solana.PublicKey) *ConfidentialAuxiliaryInstruction {
	return &ConfidentialAuxiliaryInstruction{
		programID: service.RecordProgramID,
		AccountMetaSlice: solana.AccountMetaSlice{
			solana.NewAccountMeta(record, true, false),
			solana.NewAccountMeta(authority, false, true),
			solana.NewAccountMeta(newAuthority, false, false),
		},
		data: []byte{2},
	}
}

func (service ConfidentialTransferService) CloseConfidentialRecord(record, authority, destination solana.PublicKey) *ConfidentialAuxiliaryInstruction {
	return &ConfidentialAuxiliaryInstruction{
		programID: service.RecordProgramID,
		AccountMetaSlice: solana.AccountMetaSlice{
			solana.NewAccountMeta(record, true, false),
			solana.NewAccountMeta(authority, false, true),
			solana.NewAccountMeta(destination, true, false),
		},
		data: []byte{3},
	}
}

func (service ConfidentialTransferService) ReallocateConfidentialRecord(record, authority solana.PublicKey, dataLength uint64) *ConfidentialAuxiliaryInstruction {
	payload := make([]byte, 9)
	payload[0] = 4
	binary.LittleEndian.PutUint64(payload[1:], dataLength)
	return &ConfidentialAuxiliaryInstruction{
		programID: service.RecordProgramID,
		AccountMetaSlice: solana.AccountMetaSlice{
			solana.NewAccountMeta(record, true, false),
			solana.NewAccountMeta(authority, false, true),
		},
		data: payload,
	}
}

func (service ConfidentialTransferService) ConfidentialProofRecordInstructions(payer, record, authority, destination solana.PublicKey, lamports uint64, proofData []byte, chunkSize int) (ConfidentialRecordPlan, error) {
	if chunkSize <= 0 {
		return ConfidentialRecordPlan{}, fmt.Errorf("create confidential proof record: invalid chunk size %d", chunkSize)
	}
	space := uint64(len(proofData)) + uint64(service.ConfidentialRecordDataOffset())
	instructions := []solana.Instruction{
		system.NewCreateAccountInstruction(lamports, space, service.RecordProgramID, payer, record),
		service.InitializeConfidentialRecord(record, authority),
	}
	for offset := 0; offset < len(proofData); offset += chunkSize {
		end := len(proofData)
		if chunkSize < len(proofData)-offset {
			end = offset + chunkSize
		}
		instructions = append(instructions, service.WriteConfidentialRecord(record, authority, uint64(offset), proofData[offset:end]))
	}
	return ConfidentialRecordPlan{
		CreateAndWrite: instructions,
		Close:          service.CloseConfidentialRecord(record, authority, destination),
		Space:          space,
	}, nil
}

func (service ConfidentialTransferService) ConfidentialProofRecordInstructionsForProof(payer, record, authority, destination solana.PublicKey, lamports uint64, proof ZKProofData, chunkSize int) (ConfidentialRecordPlan, error) {
	if err := service.validateZKProofData(proof); err != nil {
		return ConfidentialRecordPlan{}, fmt.Errorf("create confidential proof record: %w", err)
	}
	data := make([]byte, 0, len(proof.Context)+len(proof.Proof))
	data = append(data, proof.Context...)
	data = append(data, proof.Proof...)
	return service.ConfidentialProofRecordInstructions(payer, record, authority, destination, lamports, data, chunkSize)
}

func (service ConfidentialTransferService) VerifyProofFromRecord(discriminator uint8, record solana.PublicKey, contextState, contextAuthority *solana.PublicKey) *ZKProofInstruction {
	return service.VerifyProofFromAccount(discriminator, record, service.ConfidentialRecordDataOffset(), contextState, contextAuthority)
}

func (service ConfidentialTransferService) DecodeConfidentialRecord(data []byte) (ConfidentialRecord, error) {
	offset := int(service.ConfidentialRecordDataOffset())
	if len(data) < offset {
		return ConfidentialRecord{}, fmt.Errorf("decode confidential record: expected at least %d bytes, got %d", offset, len(data))
	}
	record := ConfidentialRecord{Version: data[0], Data: data[offset:]}
	copy(record.Authority[:], data[1:offset])
	return record, nil
}
