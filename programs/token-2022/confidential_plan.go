package token2022

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	system "github.com/fluxrpc/solana-go/programs/system"
)

type ProofContextPlan struct {
	CreateAndVerify []solana.Instruction
	Close           solana.Instruction
}

func (service ConfidentialTransferService) ProofContextStateSize(discriminator uint8) (uint64, error) {
	contextSize, _, ok := service.zkProofSizes(discriminator)
	if !ok {
		return 0, fmt.Errorf("proof context state size: unknown discriminator %d", discriminator)
	}
	return uint64(contextSize + 33), nil
}

func (service ConfidentialTransferService) ProofContextInstructions(payer, contextState, contextAuthority, closeDestination solana.PublicKey, lamports uint64, data ZKProofData) (ProofContextPlan, error) {
	if err := service.validateZKProofData(data); err != nil {
		return ProofContextPlan{}, fmt.Errorf("build proof context instructions: %w", err)
	}
	space, err := service.ProofContextStateSize(data.Discriminator)
	if err != nil {
		return ProofContextPlan{}, err
	}
	create := system.NewCreateAccountInstruction(lamports, space, service.ProofProgramID, payer, contextState)
	verify := service.VerifyProof(data, &contextState, &contextAuthority)
	return ProofContextPlan{
		CreateAndVerify: []solana.Instruction{create, verify},
		Close:           service.CloseProofContext(contextState, closeDestination, contextAuthority),
	}, nil
}

func (service ConfidentialTransferService) ProofContextFromRecordInstructions(payer, contextState, contextAuthority, closeDestination, record solana.PublicKey, lamports uint64, discriminator uint8) (ProofContextPlan, error) {
	space, err := service.ProofContextStateSize(discriminator)
	if err != nil {
		return ProofContextPlan{}, err
	}
	create := system.NewCreateAccountInstruction(lamports, space, service.ProofProgramID, payer, contextState)
	verify := service.VerifyProofFromRecord(discriminator, record, &contextState, &contextAuthority)
	return ProofContextPlan{
		CreateAndVerify: []solana.Instruction{create, verify},
		Close:           service.CloseProofContext(contextState, closeDestination, contextAuthority),
	}, nil
}
