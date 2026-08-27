package token2022

import (
	"fmt"
	"math/bits"

	solana "github.com/fluxrpc/solana-go"
)

type ConfidentialWithdrawProofBundle struct {
	Equality ZKProofData
	Range    ZKProofData
}

type ConfidentialAmountProofBundle struct {
	Equality            ZKProofData
	CiphertextValidity  ZKProofData
	Range               ZKProofData
	AuditorCiphertextLo ElGamalCiphertext
	AuditorCiphertextHi ElGamalCiphertext
}

type ConfidentialWithdrawPlan struct {
	Instructions                   []solana.Instruction
	Proofs                         ConfidentialWithdrawProofBundle
	NewDecryptableAvailableBalance AeCiphertext
}

type ConfidentialAmountPlan struct {
	Instructions                   []solana.Instruction
	Proofs                         ConfidentialAmountProofBundle
	NewDecryptableAvailableBalance AeCiphertext
}

type confidentialAmountProofValues struct {
	proofs        ConfidentialAmountProofBundle
	amountLo      uint64
	amountHi      uint64
	newAmount     uint64
	commitmentLo  PedersenCommitment
	commitmentHi  PedersenCommitment
	newCommitment PedersenCommitment
	openingLo     PedersenOpening
	openingHi     PedersenOpening
	newOpening    PedersenOpening
}

func (service ConfidentialTransferService) GenerateConfidentialWithdrawProofs(account ConfidentialTransferAccount, amount uint64, keypair ElGamalKeypair, key AEKey) (ConfidentialWithdrawProofBundle, uint64, error) {
	current, err := service.DecryptConfidentialAvailableBalance(key, account)
	if err != nil {
		return ConfidentialWithdrawProofBundle{}, 0, err
	}
	if current < amount {
		return ConfidentialWithdrawProofBundle{}, 0, fmt.Errorf("generate confidential withdraw proofs: insufficient funds")
	}
	remaining := current - amount
	opening, err := service.GeneratePedersenOpening()
	if err != nil {
		return ConfidentialWithdrawProofBundle{}, 0, fmt.Errorf("generate confidential withdraw proofs: %w", err)
	}
	commitment, err := service.CommitPedersen(remaining, opening)
	if err != nil {
		return ConfidentialWithdrawProofBundle{}, 0, fmt.Errorf("generate confidential withdraw proofs: %w", err)
	}
	ciphertext, err := service.SubtractElGamalAmount(account.AvailableBalance, amount)
	if err != nil {
		return ConfidentialWithdrawProofBundle{}, 0, fmt.Errorf("generate confidential withdraw proofs: %w", err)
	}
	equality, err := service.GenerateCiphertextCommitmentEqualityProof(keypair, ciphertext, commitment, opening, remaining)
	if err != nil {
		return ConfidentialWithdrawProofBundle{}, 0, err
	}
	rangeProof, err := service.GenerateBatchedRangeProof([]PedersenCommitment{commitment}, []uint64{remaining}, []uint8{64}, []PedersenOpening{opening})
	if err != nil {
		return ConfidentialWithdrawProofBundle{}, 0, err
	}
	proofs := ConfidentialWithdrawProofBundle{Equality: equality, Range: rangeProof}
	return proofs, remaining, nil
}

func (service ConfidentialTransferService) ConfidentialWithdrawWithProofs(tokenAccount, mint, authority solana.PublicKey, signers []solana.PublicKey, amount uint64, decimals uint8, account ConfidentialTransferAccount, keypair ElGamalKeypair, key AEKey) (ConfidentialWithdrawPlan, error) {
	proofs, remaining, err := service.GenerateConfidentialWithdrawProofs(account, amount, keypair, key)
	if err != nil {
		return ConfidentialWithdrawPlan{}, err
	}
	decryptable, err := service.EncryptAEAmount(key, remaining)
	if err != nil {
		return ConfidentialWithdrawPlan{}, err
	}
	locations := []ProofLocation{{InstructionOffset: 1, Proof: &proofs.Equality}, {InstructionOffset: 2, Proof: &proofs.Range}}
	instruction := service.ConfidentialWithdraw(tokenAccount, mint, authority, signers, amount, decimals, decryptable, locations[0], locations[1])
	instructions, err := service.ConfidentialInstructions(instruction, locations...)
	if err != nil {
		return ConfidentialWithdrawPlan{}, err
	}
	return ConfidentialWithdrawPlan{Instructions: instructions, Proofs: proofs, NewDecryptableAvailableBalance: decryptable}, nil
}

func (service ConfidentialTransferService) GenerateConfidentialTransferProofs(account ConfidentialTransferAccount, amount uint64, keypair ElGamalKeypair, key AEKey, destination ElGamalPubkey, auditor *ElGamalPubkey) (ConfidentialAmountProofBundle, uint64, error) {
	current, err := service.DecryptConfidentialAvailableBalance(key, account)
	if err != nil {
		return ConfidentialAmountProofBundle{}, 0, err
	}
	publicKeys := [3]ElGamalPubkey{keypair.PublicKey, destination, service.confidentialAuditor(auditor)}
	values, err := service.generateConfidentialAmountProofValues(account.AvailableBalance, current, amount, keypair, publicKeys, 0, false)
	if err != nil {
		return ConfidentialAmountProofBundle{}, 0, err
	}
	if err := service.generateConfidentialAmountRangeProof(&values); err != nil {
		return ConfidentialAmountProofBundle{}, 0, err
	}
	return values.proofs, values.newAmount, nil
}

func (service ConfidentialTransferService) ConfidentialTransferWithProofs(source, mint, destination, authority solana.PublicKey, signers []solana.PublicKey, amount uint64, account ConfidentialTransferAccount, keypair ElGamalKeypair, key AEKey, destinationPubkey ElGamalPubkey, auditor *ElGamalPubkey) (ConfidentialAmountPlan, error) {
	proofs, remaining, err := service.GenerateConfidentialTransferProofs(account, amount, keypair, key, destinationPubkey, auditor)
	if err != nil {
		return ConfidentialAmountPlan{}, err
	}
	decryptable, err := service.EncryptAEAmount(key, remaining)
	if err != nil {
		return ConfidentialAmountPlan{}, err
	}
	locations := []ProofLocation{{InstructionOffset: 1, Proof: &proofs.Equality}, {InstructionOffset: 2, Proof: &proofs.CiphertextValidity}, {InstructionOffset: 3, Proof: &proofs.Range}}
	instruction := service.ConfidentialTransfer(source, mint, destination, authority, signers, decryptable, proofs.AuditorCiphertextLo, proofs.AuditorCiphertextHi, locations[0], locations[1], locations[2])
	instructions, err := service.ConfidentialInstructions(instruction, locations...)
	if err != nil {
		return ConfidentialAmountPlan{}, err
	}
	return ConfidentialAmountPlan{Instructions: instructions, Proofs: proofs, NewDecryptableAvailableBalance: decryptable}, nil
}

func (service ConfidentialTransferService) generateConfidentialAmountProofValues(currentCiphertext ElGamalCiphertext, currentAmount, amount uint64, equalityKeypair ElGamalKeypair, publicKeys [3]ElGamalPubkey, equalityHandle int, add bool) (confidentialAmountProofValues, error) {
	lo, hi, err := service.SplitAmount(amount)
	if err != nil {
		return confidentialAmountProofValues{}, err
	}
	groupedLo, openingLo, err := service.EncryptGroupedElGamal3(publicKeys, uint64(lo))
	if err != nil {
		return confidentialAmountProofValues{}, err
	}
	groupedHi, openingHi, err := service.EncryptGroupedElGamal3(publicKeys, uint64(hi))
	if err != nil {
		return confidentialAmountProofValues{}, err
	}
	amountCipherLo, err := service.GroupedElGamalCiphertext3Handle(groupedLo, equalityHandle)
	if err != nil {
		return confidentialAmountProofValues{}, err
	}
	amountCipherHi, err := service.GroupedElGamalCiphertext3Handle(groupedHi, equalityHandle)
	if err != nil {
		return confidentialAmountProofValues{}, err
	}
	amountCiphertext, err := service.CombineElGamalAmountCiphertexts(amountCipherLo, amountCipherHi)
	if err != nil {
		return confidentialAmountProofValues{}, err
	}
	newAmount, newCiphertext, err := service.confidentialAmountAfterOperation(currentAmount, amount, currentCiphertext, amountCiphertext, add)
	if err != nil {
		return confidentialAmountProofValues{}, err
	}
	newOpening, err := service.GeneratePedersenOpening()
	if err != nil {
		return confidentialAmountProofValues{}, err
	}
	newCommitment, err := service.CommitPedersen(newAmount, newOpening)
	if err != nil {
		return confidentialAmountProofValues{}, err
	}
	equality, err := service.GenerateCiphertextCommitmentEqualityProof(equalityKeypair, newCiphertext, newCommitment, newOpening, newAmount)
	if err != nil {
		return confidentialAmountProofValues{}, err
	}
	validity, err := service.GenerateBatchedGroupedCiphertext3HandlesValidityProof(publicKeys, [2]GroupedElGamalCiphertext3Handles{groupedLo, groupedHi}, [2]uint64{uint64(lo), uint64(hi)}, [2]PedersenOpening{openingLo, openingHi})
	if err != nil {
		return confidentialAmountProofValues{}, err
	}
	auditorLo, err := service.GroupedElGamalCiphertext3Handle(groupedLo, 2)
	if err != nil {
		return confidentialAmountProofValues{}, err
	}
	auditorHi, err := service.GroupedElGamalCiphertext3Handle(groupedHi, 2)
	if err != nil {
		return confidentialAmountProofValues{}, err
	}
	commitmentLo := PedersenCommitment{}
	commitmentHi := PedersenCommitment{}
	copy(commitmentLo[:], groupedLo[:32])
	copy(commitmentHi[:], groupedHi[:32])
	return confidentialAmountProofValues{
		proofs:   ConfidentialAmountProofBundle{Equality: equality, CiphertextValidity: validity, AuditorCiphertextLo: auditorLo, AuditorCiphertextHi: auditorHi},
		amountLo: uint64(lo), amountHi: uint64(hi), newAmount: newAmount,
		commitmentLo: commitmentLo, commitmentHi: commitmentHi, newCommitment: newCommitment,
		openingLo: openingLo, openingHi: openingHi, newOpening: newOpening,
	}, nil
}

func (service ConfidentialTransferService) confidentialAmountAfterOperation(currentAmount, amount uint64, currentCiphertext, amountCiphertext ElGamalCiphertext, add bool) (uint64, ElGamalCiphertext, error) {
	if add {
		newAmount, carry := bits.Add64(currentAmount, amount, 0)
		if carry != 0 {
			return 0, ElGamalCiphertext{}, fmt.Errorf("generate confidential amount proofs: amount overflow")
		}
		ciphertext, err := service.AddElGamalCiphertexts(currentCiphertext, amountCiphertext)
		return newAmount, ciphertext, err
	}
	if currentAmount < amount {
		return 0, ElGamalCiphertext{}, fmt.Errorf("generate confidential amount proofs: insufficient funds")
	}
	ciphertext, err := service.SubtractElGamalCiphertexts(currentCiphertext, amountCiphertext)
	return currentAmount - amount, ciphertext, err
}

func (service ConfidentialTransferService) generateConfidentialAmountRangeProof(values *confidentialAmountProofValues) error {
	paddingOpening, err := service.GeneratePedersenOpening()
	if err != nil {
		return err
	}
	paddingCommitment, err := service.CommitPedersen(0, paddingOpening)
	if err != nil {
		return err
	}
	proof, err := service.GenerateBatchedRangeProof(
		[]PedersenCommitment{values.newCommitment, values.commitmentLo, values.commitmentHi, paddingCommitment},
		[]uint64{values.newAmount, values.amountLo, values.amountHi, 0},
		[]uint8{64, 16, 32, 16},
		[]PedersenOpening{values.newOpening, values.openingLo, values.openingHi, paddingOpening},
	)
	if err != nil {
		return err
	}
	values.proofs.Range = proof
	return nil
}

func (ConfidentialTransferService) confidentialAuditor(auditor *ElGamalPubkey) ElGamalPubkey {
	if auditor == nil {
		return ElGamalPubkey{}
	}
	return *auditor
}
