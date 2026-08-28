package token2022

import solana "github.com/fluxrpc/solana-go"

type ConfidentialAuxiliaryInstruction struct {
	programID solana.PublicKey
	solana.AccountMetaSlice
	data []byte
}

func (inst *ConfidentialAuxiliaryInstruction) ProgramID() solana.PublicKey { return inst.programID }
func (inst *ConfidentialAuxiliaryInstruction) Accounts() []*solana.AccountMeta {
	return inst.AccountMetaSlice
}
func (inst *ConfidentialAuxiliaryInstruction) Data() ([]byte, error) { return inst.data, nil }
