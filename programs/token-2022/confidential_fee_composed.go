package token2022

import (
	"fmt"
	"math/bits"

	solana "github.com/fluxrpc/solana-go"
)

type ConfidentialTransferWithFeeProofBundle struct {
	Equality                   ZKProofData
	TransferCiphertextValidity ZKProofData
	FeeSigma                   ZKProofData
	FeeCiphertextValidity      ZKProofData
	Range                      ZKProofData
	AuditorCiphertextLo        ElGamalCiphertext
	AuditorCiphertextHi        ElGamalCiphertext
}

type ConfidentialTransferWithFeePlan struct {
	Instructions                   []solana.Instruction
	Proofs                         ConfidentialTransferWithFeeProofBundle
	NewDecryptableAvailableBalance AeCiphertext
	FeeAmount                      uint64
}

type ConfidentialWithheldAccount struct {
	Address solana.PublicKey
	State   ConfidentialTransferFeeAmount
}

type ConfidentialWithheldProofBundle struct {
	Equality              ZKProofData
	DestinationCiphertext ElGamalCiphertext
}

type ConfidentialWithheldPlan struct {
	Instructions                   []solana.Instruction
	Proofs                         ConfidentialWithheldProofBundle
	NewDecryptableAvailableBalance AeCiphertext
}

func (service ConfidentialTransferService) GenerateConfidentialTransferWithFeeProofs(account ConfidentialTransferAccount, amount uint64, keypair ElGamalKeypair, key AEKey, destination, withdrawWithheldAuthority ElGamalPubkey, auditor *ElGamalPubkey, basisPoints uint16, maximumFee uint64) (ConfidentialTransferWithFeeProofBundle, uint64, uint64, error) {
	current, err := service.DecryptConfidentialAvailableBalance(key, account)
	if err != nil {
		return ConfidentialTransferWithFeeProofBundle{}, 0, 0, err
	}
	publicKeys := [3]ElGamalPubkey{keypair.PublicKey, destination, service.confidentialAuditor(auditor)}
	transfer, err := service.generateConfidentialAmountProofValues(account.AvailableBalance, current, amount, keypair, publicKeys, 0, false)
	if err != nil {
		return ConfidentialTransferWithFeeProofBundle{}, 0, 0, err
	}
	fee, claimedDelta, err := service.CalculateConfidentialTransferFee(amount, basisPoints, maximumFee)
	if err != nil {
		return ConfidentialTransferWithFeeProofBundle{}, 0, 0, err
	}
	feeLo, feeHi, err := service.SplitAmount(fee)
	if err != nil {
		return ConfidentialTransferWithFeeProofBundle{}, 0, 0, err
	}
	feeKeys := [2]ElGamalPubkey{destination, withdrawWithheldAuthority}
	feeCipherLo, feeOpeningLo, err := service.EncryptGroupedElGamal2(feeKeys, uint64(feeLo))
	if err != nil {
		return ConfidentialTransferWithFeeProofBundle{}, 0, 0, err
	}
	feeCipherHi, feeOpeningHi, err := service.EncryptGroupedElGamal2(feeKeys, uint64(feeHi))
	if err != nil {
		return ConfidentialTransferWithFeeProofBundle{}, 0, 0, err
	}
	feeCommitmentLo := PedersenCommitment{}
	feeCommitmentHi := PedersenCommitment{}
	copy(feeCommitmentLo[:], feeCipherLo[:32])
	copy(feeCommitmentHi[:], feeCipherHi[:32])
	transferCommitment, transferOpening, err := service.CombinePedersenAmount(transfer.commitmentLo, transfer.commitmentHi, transfer.openingLo, transfer.openingHi)
	if err != nil {
		return ConfidentialTransferWithFeeProofBundle{}, 0, 0, err
	}
	feeCommitment, feeOpening, err := service.CombinePedersenAmount(feeCommitmentLo, feeCommitmentHi, feeOpeningLo, feeOpeningHi)
	if err != nil {
		return ConfidentialTransferWithFeeProofBundle{}, 0, 0, err
	}
	netCommitment, err := service.SubtractPedersenCommitments(transferCommitment, feeCommitment)
	if err != nil {
		return ConfidentialTransferWithFeeProofBundle{}, 0, 0, err
	}
	netOpening, err := service.SubtractPedersenOpenings(transferOpening, feeOpening)
	if err != nil {
		return ConfidentialTransferWithFeeProofBundle{}, 0, 0, err
	}
	deltaCommitment, deltaOpening, err := service.confidentialFeeDelta(transferCommitment, transferOpening, feeCommitment, feeOpening, basisPoints)
	if err != nil {
		return ConfidentialTransferWithFeeProofBundle{}, 0, 0, err
	}
	claimedOpening, err := service.GeneratePedersenOpening()
	if err != nil {
		return ConfidentialTransferWithFeeProofBundle{}, 0, 0, err
	}
	claimedCommitment, err := service.CommitPedersen(claimedDelta, claimedOpening)
	if err != nil {
		return ConfidentialTransferWithFeeProofBundle{}, 0, 0, err
	}
	feeSigma, err := service.GeneratePercentageWithCapProof(feeCommitment, feeOpening, fee, deltaCommitment, deltaOpening, claimedDelta, claimedCommitment, claimedOpening, maximumFee)
	if err != nil {
		return ConfidentialTransferWithFeeProofBundle{}, 0, 0, err
	}
	feeValidity, err := service.GenerateBatchedGroupedCiphertext2HandlesValidityProof(feeKeys, [2]GroupedElGamalCiphertext2Handles{feeCipherLo, feeCipherHi}, [2]uint64{uint64(feeLo), uint64(feeHi)}, [2]PedersenOpening{feeOpeningLo, feeOpeningHi})
	if err != nil {
		return ConfidentialTransferWithFeeProofBundle{}, 0, 0, err
	}
	rangeProof, err := service.generateConfidentialTransferFeeRangeProof(transfer, claimedDelta, claimedCommitment, claimedOpening, feeCommitmentLo, feeCommitmentHi, feeOpeningLo, feeOpeningHi, uint64(feeLo), uint64(feeHi), amount-fee, netCommitment, netOpening)
	if err != nil {
		return ConfidentialTransferWithFeeProofBundle{}, 0, 0, err
	}
	proofs := ConfidentialTransferWithFeeProofBundle{
		Equality: transfer.proofs.Equality, TransferCiphertextValidity: transfer.proofs.CiphertextValidity,
		FeeSigma: feeSigma, FeeCiphertextValidity: feeValidity, Range: rangeProof,
		AuditorCiphertextLo: transfer.proofs.AuditorCiphertextLo, AuditorCiphertextHi: transfer.proofs.AuditorCiphertextHi,
	}
	return proofs, transfer.newAmount, fee, nil
}

func (service ConfidentialTransferService) ConfidentialTransferWithFeeWithProofs(source, mint, destination, authority solana.PublicKey, signers []solana.PublicKey, amount uint64, account ConfidentialTransferAccount, keypair ElGamalKeypair, key AEKey, destinationPubkey, withdrawWithheldAuthority ElGamalPubkey, auditor *ElGamalPubkey, basisPoints uint16, maximumFee uint64) (ConfidentialTransferWithFeePlan, error) {
	proofs, remaining, fee, err := service.GenerateConfidentialTransferWithFeeProofs(account, amount, keypair, key, destinationPubkey, withdrawWithheldAuthority, auditor, basisPoints, maximumFee)
	if err != nil {
		return ConfidentialTransferWithFeePlan{}, err
	}
	decryptable, err := service.EncryptAEAmount(key, remaining)
	if err != nil {
		return ConfidentialTransferWithFeePlan{}, err
	}
	locations := []ProofLocation{
		{InstructionOffset: 1, Proof: &proofs.Equality},
		{InstructionOffset: 2, Proof: &proofs.TransferCiphertextValidity},
		{InstructionOffset: 3, Proof: &proofs.FeeSigma},
		{InstructionOffset: 4, Proof: &proofs.FeeCiphertextValidity},
		{InstructionOffset: 5, Proof: &proofs.Range},
	}
	instruction := service.ConfidentialTransferWithFee(source, mint, destination, authority, signers, decryptable, proofs.AuditorCiphertextLo, proofs.AuditorCiphertextHi, locations[0], locations[1], locations[2], locations[3], locations[4])
	instructions, err := service.ConfidentialInstructions(instruction, locations...)
	if err != nil {
		return ConfidentialTransferWithFeePlan{}, err
	}
	return ConfidentialTransferWithFeePlan{Instructions: instructions, Proofs: proofs, NewDecryptableAvailableBalance: decryptable, FeeAmount: fee}, nil
}

func (service ConfidentialTransferService) GenerateConfidentialWithheldProof(withheld ElGamalCiphertext, withdrawKeypair ElGamalKeypair, destination ElGamalPubkey) (ConfidentialWithheldProofBundle, uint64, error) {
	amount, err := service.DecryptConfidentialWithheldAmount(withdrawKeypair.SecretKey, withheld)
	if err != nil {
		return ConfidentialWithheldProofBundle{}, 0, err
	}
	destinationCiphertext, destinationOpening, err := service.EncryptElGamal(destination, uint64(amount))
	if err != nil {
		return ConfidentialWithheldProofBundle{}, 0, err
	}
	proof, err := service.GenerateCiphertextCiphertextEqualityProof(withdrawKeypair, destination, withheld, destinationCiphertext, destinationOpening, uint64(amount))
	if err != nil {
		return ConfidentialWithheldProofBundle{}, 0, err
	}
	return ConfidentialWithheldProofBundle{Equality: proof, DestinationCiphertext: destinationCiphertext}, uint64(amount), nil
}

func (service ConfidentialTransferService) WithdrawConfidentialWithheldFromMintWithProofs(mint, destination, authority solana.PublicKey, signers []solana.PublicKey, config ConfidentialTransferFeeConfig, destinationAccount ConfidentialTransferAccount, withdrawKeypair ElGamalKeypair, destinationKey AEKey) (ConfidentialWithheldPlan, error) {
	return service.withdrawConfidentialWithheldWithProofs(mint, destination, authority, signers, nil, config.WithheldAmount, destinationAccount, withdrawKeypair, destinationKey)
}

func (service ConfidentialTransferService) WithdrawConfidentialWithheldFromAccountsWithProofs(mint, destination, authority solana.PublicKey, signers []solana.PublicKey, sources []ConfidentialWithheldAccount, destinationAccount ConfidentialTransferAccount, withdrawKeypair ElGamalKeypair, destinationKey AEKey) (ConfidentialWithheldPlan, error) {
	if len(sources) > 255 {
		return ConfidentialWithheldPlan{}, fmt.Errorf("withdraw confidential withheld tokens: source account count exceeds 255")
	}
	addresses := make([]solana.PublicKey, len(sources))
	withheld := ElGamalCiphertext{}
	var err error
	for index, source := range sources {
		addresses[index] = source.Address
		withheld, err = service.AddElGamalCiphertexts(withheld, source.State.WithheldAmount)
		if err != nil {
			return ConfidentialWithheldPlan{}, err
		}
	}
	return service.withdrawConfidentialWithheldWithProofs(mint, destination, authority, signers, addresses, withheld, destinationAccount, withdrawKeypair, destinationKey)
}

func (service ConfidentialTransferService) withdrawConfidentialWithheldWithProofs(mint, destination, authority solana.PublicKey, signers, sources []solana.PublicKey, withheld ElGamalCiphertext, destinationAccount ConfidentialTransferAccount, withdrawKeypair ElGamalKeypair, destinationKey AEKey) (ConfidentialWithheldPlan, error) {
	proofs, amount, err := service.GenerateConfidentialWithheldProof(withheld, withdrawKeypair, destinationAccount.ElGamalPubkey)
	if err != nil {
		return ConfidentialWithheldPlan{}, err
	}
	current, err := service.DecryptConfidentialAvailableBalance(destinationKey, destinationAccount)
	if err != nil {
		return ConfidentialWithheldPlan{}, err
	}
	newBalance, carry := bits.Add64(current, amount, 0)
	if carry != 0 {
		return ConfidentialWithheldPlan{}, fmt.Errorf("withdraw confidential withheld tokens: balance overflow")
	}
	decryptable, err := service.EncryptAEAmount(destinationKey, newBalance)
	if err != nil {
		return ConfidentialWithheldPlan{}, err
	}
	location := ProofLocation{InstructionOffset: 1, Proof: &proofs.Equality}
	var instruction solana.Instruction
	if sources == nil {
		instruction = service.WithdrawConfidentialWithheldTokensFromMint(mint, destination, authority, signers, decryptable, location)
	} else {
		instruction = service.WithdrawConfidentialWithheldTokensFromAccounts(mint, destination, authority, signers, sources, decryptable, location)
	}
	instructions, err := service.ConfidentialInstructions(instruction, location)
	if err != nil {
		return ConfidentialWithheldPlan{}, err
	}
	return ConfidentialWithheldPlan{Instructions: instructions, Proofs: proofs, NewDecryptableAvailableBalance: decryptable}, nil
}

func (service ConfidentialTransferService) confidentialFeeDelta(transferCommitment PedersenCommitment, transferOpening PedersenOpening, feeCommitment PedersenCommitment, feeOpening PedersenOpening, basisPoints uint16) (PedersenCommitment, PedersenOpening, error) {
	scaledFeeCommitment, err := service.ScalePedersenCommitment(feeCommitment, 10_000)
	if err != nil {
		return PedersenCommitment{}, PedersenOpening{}, err
	}
	scaledTransferCommitment, err := service.ScalePedersenCommitment(transferCommitment, uint64(basisPoints))
	if err != nil {
		return PedersenCommitment{}, PedersenOpening{}, err
	}
	deltaCommitment, err := service.SubtractPedersenCommitments(scaledFeeCommitment, scaledTransferCommitment)
	if err != nil {
		return PedersenCommitment{}, PedersenOpening{}, err
	}
	scaledFeeOpening, err := service.ScalePedersenOpening(feeOpening, 10_000)
	if err != nil {
		return PedersenCommitment{}, PedersenOpening{}, err
	}
	scaledTransferOpening, err := service.ScalePedersenOpening(transferOpening, uint64(basisPoints))
	if err != nil {
		return PedersenCommitment{}, PedersenOpening{}, err
	}
	deltaOpening, err := service.SubtractPedersenOpenings(scaledFeeOpening, scaledTransferOpening)
	return deltaCommitment, deltaOpening, err
}

func (service ConfidentialTransferService) generateConfidentialTransferFeeRangeProof(transfer confidentialAmountProofValues, claimedDelta uint64, claimedCommitment PedersenCommitment, claimedOpening PedersenOpening, feeCommitmentLo, feeCommitmentHi PedersenCommitment, feeOpeningLo, feeOpeningHi PedersenOpening, feeLo, feeHi, netAmount uint64, netCommitment PedersenCommitment, netOpening PedersenOpening) (ZKProofData, error) {
	if claimedDelta > 9_999 {
		return ZKProofData{}, fmt.Errorf("generate confidential transfer fee range proof: invalid claimed delta")
	}
	zeroOpening := PedersenOpening{}
	maximumCommitment, err := service.CommitPedersen(9_999, zeroOpening)
	if err != nil {
		return ZKProofData{}, err
	}
	complementCommitment, err := service.SubtractPedersenCommitments(maximumCommitment, claimedCommitment)
	if err != nil {
		return ZKProofData{}, err
	}
	complementOpening, err := service.SubtractPedersenOpenings(zeroOpening, claimedOpening)
	if err != nil {
		return ZKProofData{}, err
	}
	return service.GenerateBatchedRangeProof(
		[]PedersenCommitment{transfer.newCommitment, transfer.commitmentLo, transfer.commitmentHi, claimedCommitment, complementCommitment, feeCommitmentLo, feeCommitmentHi, netCommitment},
		[]uint64{transfer.newAmount, transfer.amountLo, transfer.amountHi, claimedDelta, 9_999 - claimedDelta, feeLo, feeHi, netAmount},
		[]uint8{64, 16, 32, 16, 16, 16, 32, 64},
		[]PedersenOpening{transfer.newOpening, transfer.openingLo, transfer.openingHi, claimedOpening, complementOpening, feeOpeningLo, feeOpeningHi, netOpening},
	)
}
