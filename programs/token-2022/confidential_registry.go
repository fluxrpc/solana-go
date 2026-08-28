package token2022

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
)

type ElGamalRegistry struct {
	Owner         solana.PublicKey
	ElGamalPubkey ElGamalPubkey
}

func (service ConfidentialTransferService) ElGamalRegistryAddress(owner solana.PublicKey) (solana.PublicKey, uint8, error) {
	return service.RegistryProgramID.FindProgramAddress([][]byte{[]byte("elgamal-registry"), owner[:]})
}

func (service ConfidentialTransferService) CreateElGamalRegistry(owner solana.PublicKey, proof ProofLocation) ([]solana.Instruction, error) {
	registry, _, err := service.ElGamalRegistryAddress(owner)
	if err != nil {
		return nil, fmt.Errorf("derive ElGamal registry: %w", err)
	}
	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(registry, true, false),
		solana.NewAccountMeta(owner, false, true),
		solana.NewAccountMeta(solana.SystemProgramID, false, false),
	}
	accounts = append(accounts, service.proofAccounts(proof)...)
	instruction := &ConfidentialAuxiliaryInstruction{
		programID:        service.RegistryProgramID,
		AccountMetaSlice: accounts,
		data:             []byte{0, byte(proof.InstructionOffset)},
	}
	return service.elGamalRegistryInstructions(instruction, proof)
}

func (service ConfidentialTransferService) UpdateElGamalRegistry(owner solana.PublicKey, proof ProofLocation) ([]solana.Instruction, error) {
	registry, _, err := service.ElGamalRegistryAddress(owner)
	if err != nil {
		return nil, fmt.Errorf("derive ElGamal registry: %w", err)
	}
	accounts := solana.AccountMetaSlice{solana.NewAccountMeta(registry, true, false)}
	accounts = append(accounts, service.proofAccounts(proof)...)
	accounts = append(accounts, solana.NewAccountMeta(owner, false, true))
	instruction := &ConfidentialAuxiliaryInstruction{
		programID:        service.RegistryProgramID,
		AccountMetaSlice: accounts,
		data:             []byte{1, byte(proof.InstructionOffset)},
	}
	return service.elGamalRegistryInstructions(instruction, proof)
}

func (service ConfidentialTransferService) elGamalRegistryInstructions(instruction solana.Instruction, proof ProofLocation) ([]solana.Instruction, error) {
	if proof.InstructionOffset != 0 && proof.Proof != nil && proof.Proof.Discriminator != 4 {
		return nil, fmt.Errorf("build ElGamal registry instructions: expected pubkey validity proof, got %d", proof.Proof.Discriminator)
	}
	return service.ConfidentialInstructions(instruction, proof)
}

func (service ConfidentialTransferService) DecodeElGamalRegistry(data []byte) (ElGamalRegistry, error) {
	if err := service.validateStateSize(data, 64, "ElGamal registry"); err != nil {
		return ElGamalRegistry{}, err
	}
	state := ElGamalRegistry{}
	copy(state.Owner[:], data[:32])
	copy(state.ElGamalPubkey[:], data[32:])
	return state, nil
}
