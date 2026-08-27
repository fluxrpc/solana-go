package token2022

import (
	"encoding/binary"

	solana "github.com/fluxrpc/solana-go"
)

type ZKProofData struct {
	Discriminator uint8
	Context       []byte
	Proof         []byte
}

type ZKProofInstruction struct {
	programID solana.PublicKey
	solana.AccountMetaSlice
	data    []byte
	Decoded *ZKProofInstructionData
}

func (inst *ZKProofInstruction) ProgramID() solana.PublicKey     { return inst.programID }
func (inst *ZKProofInstruction) Accounts() []*solana.AccountMeta { return inst.AccountMetaSlice }
func (inst *ZKProofInstruction) Data() ([]byte, error)           { return inst.data, nil }

func (service ConfidentialTransferService) VerifyProof(data ZKProofData, contextState, contextAuthority *solana.PublicKey) *ZKProofInstruction {
	payload := make([]byte, 1, 1+len(data.Context)+len(data.Proof))
	payload[0] = data.Discriminator
	payload = append(payload, data.Context...)
	payload = append(payload, data.Proof...)
	accounts := make(solana.AccountMetaSlice, 0, 2)
	if contextState != nil {
		accounts = append(accounts,
			solana.NewAccountMeta(*contextState, true, false),
			solana.NewAccountMeta(*contextAuthority, false, false),
		)
	}
	return &ZKProofInstruction{programID: service.ProofProgramID, AccountMetaSlice: accounts, data: payload}
}

func (service ConfidentialTransferService) VerifyProofFromAccount(discriminator uint8, proofAccount solana.PublicKey, offset uint32, contextState, contextAuthority *solana.PublicKey) *ZKProofInstruction {
	payload := make([]byte, 5)
	payload[0] = discriminator
	binary.LittleEndian.PutUint32(payload[1:], offset)
	accounts := solana.AccountMetaSlice{solana.NewAccountMeta(proofAccount, false, false)}
	if contextState != nil {
		accounts = append(accounts,
			solana.NewAccountMeta(*contextState, true, false),
			solana.NewAccountMeta(*contextAuthority, false, false),
		)
	}
	return &ZKProofInstruction{programID: service.ProofProgramID, AccountMetaSlice: accounts, data: payload}
}

func (service ConfidentialTransferService) CloseProofContext(contextState, destination, authority solana.PublicKey) *ZKProofInstruction {
	return &ZKProofInstruction{
		programID: service.ProofProgramID,
		AccountMetaSlice: solana.AccountMetaSlice{
			solana.NewAccountMeta(contextState, true, false),
			solana.NewAccountMeta(destination, true, false),
			solana.NewAccountMeta(authority, false, true),
		},
		data: []byte{0},
	}
}
