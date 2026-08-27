package token2022

import solana "github.com/fluxrpc/solana-go"

type ConfidentialMintPlan struct {
	Instructions         []solana.Instruction
	Proofs               ConfidentialAmountProofBundle
	NewDecryptableSupply AeCiphertext
}

type ConfidentialRotateSupplyProofBundle struct {
	Equality            ZKProofData
	NewSupplyCiphertext ElGamalCiphertext
}

type ConfidentialRotateSupplyPlan struct {
	Instructions []solana.Instruction
	Proofs       ConfidentialRotateSupplyProofBundle
}

func (service ConfidentialTransferService) GenerateConfidentialMintProofs(state ConfidentialMintBurn, amount uint64, supplyKeypair ElGamalKeypair, supplyKey AEKey, destination ElGamalPubkey, auditor *ElGamalPubkey) (ConfidentialAmountProofBundle, uint64, error) {
	current, err := service.DecryptCurrentConfidentialSupply(supplyKey, supplyKeypair, state)
	if err != nil {
		return ConfidentialAmountProofBundle{}, 0, err
	}
	publicKeys := [3]ElGamalPubkey{destination, supplyKeypair.PublicKey, service.confidentialAuditor(auditor)}
	values, err := service.generateConfidentialAmountProofValues(state.ConfidentialSupply, current, amount, supplyKeypair, publicKeys, 1, true)
	if err != nil {
		return ConfidentialAmountProofBundle{}, 0, err
	}
	if err := service.generateConfidentialAmountRangeProof(&values); err != nil {
		return ConfidentialAmountProofBundle{}, 0, err
	}
	return values.proofs, values.newAmount, nil
}

func (service ConfidentialTransferService) ConfidentialMintWithProofs(tokenAccount, mint, authority solana.PublicKey, signers []solana.PublicKey, amount uint64, state ConfidentialMintBurn, supplyKeypair ElGamalKeypair, supplyKey AEKey, destination ElGamalPubkey, auditor *ElGamalPubkey) (ConfidentialMintPlan, error) {
	proofs, newSupply, err := service.GenerateConfidentialMintProofs(state, amount, supplyKeypair, supplyKey, destination, auditor)
	if err != nil {
		return ConfidentialMintPlan{}, err
	}
	decryptable, err := service.EncryptAEAmount(supplyKey, newSupply)
	if err != nil {
		return ConfidentialMintPlan{}, err
	}
	locations := []ProofLocation{{InstructionOffset: 1, Proof: &proofs.Equality}, {InstructionOffset: 2, Proof: &proofs.CiphertextValidity}, {InstructionOffset: 3, Proof: &proofs.Range}}
	instruction := service.ConfidentialMint(tokenAccount, mint, authority, signers, decryptable, proofs.AuditorCiphertextLo, proofs.AuditorCiphertextHi, locations[0], locations[1], locations[2])
	instructions, err := service.ConfidentialInstructions(instruction, locations...)
	if err != nil {
		return ConfidentialMintPlan{}, err
	}
	return ConfidentialMintPlan{Instructions: instructions, Proofs: proofs, NewDecryptableSupply: decryptable}, nil
}

func (service ConfidentialTransferService) GenerateConfidentialBurnProofs(account ConfidentialTransferAccount, amount uint64, sourceKeypair ElGamalKeypair, sourceKey AEKey, supply ElGamalPubkey, auditor *ElGamalPubkey) (ConfidentialAmountProofBundle, uint64, error) {
	current, err := service.DecryptConfidentialAvailableBalance(sourceKey, account)
	if err != nil {
		return ConfidentialAmountProofBundle{}, 0, err
	}
	publicKeys := [3]ElGamalPubkey{sourceKeypair.PublicKey, supply, service.confidentialAuditor(auditor)}
	values, err := service.generateConfidentialAmountProofValues(account.AvailableBalance, current, amount, sourceKeypair, publicKeys, 0, false)
	if err != nil {
		return ConfidentialAmountProofBundle{}, 0, err
	}
	if err := service.generateConfidentialAmountRangeProof(&values); err != nil {
		return ConfidentialAmountProofBundle{}, 0, err
	}
	return values.proofs, values.newAmount, nil
}

func (service ConfidentialTransferService) ConfidentialBurnWithProofs(tokenAccount, mint, authority solana.PublicKey, signers []solana.PublicKey, amount uint64, account ConfidentialTransferAccount, sourceKeypair ElGamalKeypair, sourceKey AEKey, supply ElGamalPubkey, auditor *ElGamalPubkey) (ConfidentialAmountPlan, error) {
	proofs, remaining, err := service.GenerateConfidentialBurnProofs(account, amount, sourceKeypair, sourceKey, supply, auditor)
	if err != nil {
		return ConfidentialAmountPlan{}, err
	}
	decryptable, err := service.EncryptAEAmount(sourceKey, remaining)
	if err != nil {
		return ConfidentialAmountPlan{}, err
	}
	locations := []ProofLocation{{InstructionOffset: 1, Proof: &proofs.Equality}, {InstructionOffset: 2, Proof: &proofs.CiphertextValidity}, {InstructionOffset: 3, Proof: &proofs.Range}}
	instruction := service.ConfidentialBurn(tokenAccount, mint, authority, signers, decryptable, proofs.AuditorCiphertextLo, proofs.AuditorCiphertextHi, locations[0], locations[1], locations[2])
	instructions, err := service.ConfidentialInstructions(instruction, locations...)
	if err != nil {
		return ConfidentialAmountPlan{}, err
	}
	return ConfidentialAmountPlan{Instructions: instructions, Proofs: proofs, NewDecryptableAvailableBalance: decryptable}, nil
}

func (service ConfidentialTransferService) GenerateRotateConfidentialSupplyKeyProofs(state ConfidentialMintBurn, currentKeypair ElGamalKeypair, currentKey AEKey, newPubkey ElGamalPubkey) (ConfidentialRotateSupplyProofBundle, error) {
	current, err := service.DecryptCurrentConfidentialSupply(currentKey, currentKeypair, state)
	if err != nil {
		return ConfidentialRotateSupplyProofBundle{}, err
	}
	newCiphertext, newOpening, err := service.EncryptElGamal(newPubkey, current)
	if err != nil {
		return ConfidentialRotateSupplyProofBundle{}, err
	}
	proof, err := service.GenerateCiphertextCiphertextEqualityProof(currentKeypair, newPubkey, state.ConfidentialSupply, newCiphertext, newOpening, current)
	if err != nil {
		return ConfidentialRotateSupplyProofBundle{}, err
	}
	return ConfidentialRotateSupplyProofBundle{Equality: proof, NewSupplyCiphertext: newCiphertext}, nil
}

func (service ConfidentialTransferService) RotateConfidentialSupplyKeyWithProofs(mint, authority solana.PublicKey, signers []solana.PublicKey, state ConfidentialMintBurn, currentKeypair ElGamalKeypair, currentKey AEKey, newPubkey ElGamalPubkey) (ConfidentialRotateSupplyPlan, error) {
	proofs, err := service.GenerateRotateConfidentialSupplyKeyProofs(state, currentKeypair, currentKey, newPubkey)
	if err != nil {
		return ConfidentialRotateSupplyPlan{}, err
	}
	location := ProofLocation{InstructionOffset: 1, Proof: &proofs.Equality}
	instruction := service.RotateConfidentialSupplyKey(mint, authority, signers, newPubkey, location)
	instructions, err := service.ConfidentialInstructions(instruction, location)
	if err != nil {
		return ConfidentialRotateSupplyPlan{}, err
	}
	return ConfidentialRotateSupplyPlan{Instructions: instructions, Proofs: proofs}, nil
}
