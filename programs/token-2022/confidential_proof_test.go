package token2022

import "testing"

func TestConfidentialInstructions(t *testing.T) {
	service := ConfidentialTransferService{}
	service.Start()
	first := ZKProofData{Discriminator: 4, Context: make([]byte, 32), Proof: make([]byte, 64)}
	second := ZKProofData{Discriminator: 6, Context: make([]byte, 264), Proof: make([]byte, 672)}
	locations := []ProofLocation{
		{InstructionOffset: 1, Proof: &first},
		{ContextState: token2022Key(1)},
		{InstructionOffset: 2, Proof: &second},
	}
	tokenInstruction := service.EmptyConfidentialTransferAccount(token2022Key(2), token2022Key(3), nil, locations[0])
	instructions, err := service.ConfidentialInstructions(tokenInstruction, locations...)
	if err != nil {
		t.Fatal(err)
	}
	if len(instructions) != 3 || instructions[1].ProgramID() != service.ProofProgramID || instructions[2].ProgramID() != service.ProofProgramID {
		t.Fatalf("instructions = %#v", instructions)
	}
	if _, err := service.ConfidentialInstructions(tokenInstruction, ProofLocation{InstructionOffset: 2, Proof: &first}); err == nil {
		t.Fatal("expected offset error")
	}
}
