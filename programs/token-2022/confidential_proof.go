package token2022

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
)

// ProofLocation selects an inline proof instruction or a pre-verified context account.
type ProofLocation struct {
	InstructionOffset int8
	ContextState      solana.PublicKey
	Proof             *ZKProofData
}

func (ConfidentialTransferService) proofAccounts(locations ...ProofLocation) solana.AccountMetaSlice {
	accounts := make(solana.AccountMetaSlice, 0, len(locations)+1)
	for _, location := range locations {
		if location.InstructionOffset != 0 {
			accounts = append(accounts, solana.NewAccountMeta(solana.SysVarInstructionsPubkey, false, false))
			break
		}
	}
	for _, location := range locations {
		if location.InstructionOffset == 0 {
			accounts = append(accounts, solana.NewAccountMeta(location.ContextState, false, false))
		}
	}
	return accounts
}

func (service ConfidentialTransferService) ConfidentialInstructions(instruction solana.Instruction, locations ...ProofLocation) ([]solana.Instruction, error) {
	instructions := []solana.Instruction{instruction}
	for _, location := range locations {
		if location.InstructionOffset == 0 {
			continue
		}
		if location.Proof == nil {
			return nil, fmt.Errorf("build confidential instructions: missing proof at offset %d", location.InstructionOffset)
		}
		if location.InstructionOffset != int8(len(instructions)) {
			return nil, fmt.Errorf("build confidential instructions: expected proof offset %d, got %d", len(instructions), location.InstructionOffset)
		}
		instructions = append(instructions, service.VerifyProof(*location.Proof, nil, nil))
	}
	return instructions, nil
}
