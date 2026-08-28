package token2022

import (
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

func TestConfidentialComposedProofsAndInstructionOrder(t *testing.T) {
	service := ConfidentialTransferService{}
	service.Start()
	sourceKeypair, err := service.ElGamalKeypairFromSecret(ElGamalSecretKey{1})
	if err != nil {
		t.Fatal(err)
	}
	destinationKeypair, err := service.ElGamalKeypairFromSecret(ElGamalSecretKey{2})
	if err != nil {
		t.Fatal(err)
	}
	auditorKeypair, err := service.ElGamalKeypairFromSecret(ElGamalSecretKey{3})
	if err != nil {
		t.Fatal(err)
	}
	supplyKeypair, err := service.ElGamalKeypairFromSecret(ElGamalSecretKey{4})
	if err != nil {
		t.Fatal(err)
	}
	withdrawKeypair, err := service.ElGamalKeypairFromSecret(ElGamalSecretKey{5})
	if err != nil {
		t.Fatal(err)
	}
	newSupplyKeypair, err := service.ElGamalKeypairFromSecret(ElGamalSecretKey{6})
	if err != nil {
		t.Fatal(err)
	}
	sourceAE := AEKey{7}
	destinationAE := AEKey{8}
	supplyAE := AEKey{9}
	sourceCiphertext, err := service.EncryptElGamalWithOpening(sourceKeypair.PublicKey, 1_000_000, PedersenOpening{10})
	if err != nil {
		t.Fatal(err)
	}
	sourceDecryptable, err := service.EncryptAEAmountWithNonce(sourceAE, 1_000_000, [12]byte{11})
	if err != nil {
		t.Fatal(err)
	}
	sourceState := ConfidentialTransferAccount{ElGamalPubkey: sourceKeypair.PublicKey, AvailableBalance: sourceCiphertext, DecryptableAvailableBalance: sourceDecryptable}
	destinationCiphertext, err := service.EncryptElGamalWithOpening(destinationKeypair.PublicKey, 100, PedersenOpening{12})
	if err != nil {
		t.Fatal(err)
	}
	destinationDecryptable, err := service.EncryptAEAmountWithNonce(destinationAE, 100, [12]byte{13})
	if err != nil {
		t.Fatal(err)
	}
	destinationState := ConfidentialTransferAccount{ElGamalPubkey: destinationKeypair.PublicKey, AvailableBalance: destinationCiphertext, DecryptableAvailableBalance: destinationDecryptable}
	supplyCiphertext, err := service.EncryptElGamalWithOpening(supplyKeypair.PublicKey, 500_000, PedersenOpening{14})
	if err != nil {
		t.Fatal(err)
	}
	supplyDecryptable, err := service.EncryptAEAmountWithNonce(supplyAE, 500_000, [12]byte{15})
	if err != nil {
		t.Fatal(err)
	}
	supplyState := ConfidentialMintBurn{ConfidentialSupply: supplyCiphertext, DecryptableSupply: supplyDecryptable, SupplyElGamalPubkey: supplyKeypair.PublicKey}

	withdraw, err := service.ConfidentialWithdrawWithProofs(token2022Key(1), token2022Key(2), token2022Key(3), nil, 1_000, 6, sourceState, sourceKeypair, sourceAE)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyZKProofData(withdraw.Proofs.Equality); err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyZKProofData(withdraw.Proofs.Range); err != nil {
		t.Fatal(err)
	}
	withdrawBalance, err := service.DecryptAEAmount(sourceAE, withdraw.NewDecryptableAvailableBalance)
	if err != nil || withdrawBalance != 999_000 {
		t.Fatalf("withdraw balance = %d, %v", withdrawBalance, err)
	}

	transfer, err := service.ConfidentialTransferWithProofs(token2022Key(1), token2022Key(2), token2022Key(4), token2022Key(3), nil, 12_345, sourceState, sourceKeypair, sourceAE, destinationKeypair.PublicKey, &auditorKeypair.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyZKProofData(transfer.Proofs.Equality); err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyZKProofData(transfer.Proofs.CiphertextValidity); err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyZKProofData(transfer.Proofs.Range); err != nil {
		t.Fatal(err)
	}
	transferAuditorAmount, err := service.DecryptPendingBalance(auditorKeypair.SecretKey, transfer.Proofs.AuditorCiphertextLo, transfer.Proofs.AuditorCiphertextHi)
	if err != nil || transferAuditorAmount != 12_345 {
		t.Fatalf("transfer auditor amount = %d, %v", transferAuditorAmount, err)
	}

	withFee, err := service.ConfidentialTransferWithFeeWithProofs(token2022Key(1), token2022Key(2), token2022Key(4), token2022Key(3), nil, 10_001, sourceState, sourceKeypair, sourceAE, destinationKeypair.PublicKey, withdrawKeypair.PublicKey, &auditorKeypair.PublicKey, 100, 50)
	if err != nil {
		t.Fatal(err)
	}
	if withFee.FeeAmount != 50 {
		t.Fatalf("fee amount = %d, want 50", withFee.FeeAmount)
	}
	if err := service.VerifyZKProofData(withFee.Proofs.Equality); err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyZKProofData(withFee.Proofs.TransferCiphertextValidity); err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyZKProofData(withFee.Proofs.FeeSigma); err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyZKProofData(withFee.Proofs.FeeCiphertextValidity); err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyZKProofData(withFee.Proofs.Range); err != nil {
		t.Fatal(err)
	}
	feeAuditorAmount, err := service.DecryptPendingBalance(auditorKeypair.SecretKey, withFee.Proofs.AuditorCiphertextLo, withFee.Proofs.AuditorCiphertextHi)
	if err != nil || feeAuditorAmount != 10_001 {
		t.Fatalf("fee transfer auditor amount = %d, %v", feeAuditorAmount, err)
	}

	mint, err := service.ConfidentialMintWithProofs(token2022Key(4), token2022Key(2), token2022Key(3), nil, 22_222, supplyState, supplyKeypair, supplyAE, destinationKeypair.PublicKey, &auditorKeypair.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyZKProofData(mint.Proofs.Equality); err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyZKProofData(mint.Proofs.CiphertextValidity); err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyZKProofData(mint.Proofs.Range); err != nil {
		t.Fatal(err)
	}
	mintSupply, err := service.DecryptAEAmount(supplyAE, mint.NewDecryptableSupply)
	if err != nil || mintSupply != 522_222 {
		t.Fatalf("mint supply = %d, %v", mintSupply, err)
	}

	burn, err := service.ConfidentialBurnWithProofs(token2022Key(1), token2022Key(2), token2022Key(3), nil, 1_111, sourceState, sourceKeypair, sourceAE, supplyKeypair.PublicKey, &auditorKeypair.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyZKProofData(burn.Proofs.Equality); err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyZKProofData(burn.Proofs.CiphertextValidity); err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyZKProofData(burn.Proofs.Range); err != nil {
		t.Fatal(err)
	}
	burnAuditorAmount, err := service.DecryptPendingBalance(auditorKeypair.SecretKey, burn.Proofs.AuditorCiphertextLo, burn.Proofs.AuditorCiphertextHi)
	if err != nil || burnAuditorAmount != 1_111 {
		t.Fatalf("burn auditor amount = %d, %v", burnAuditorAmount, err)
	}

	rotate, err := service.RotateConfidentialSupplyKeyWithProofs(token2022Key(2), token2022Key(3), nil, supplyState, supplyKeypair, supplyAE, newSupplyKeypair.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyZKProofData(rotate.Proofs.Equality); err != nil {
		t.Fatal(err)
	}
	rotatedSupply, err := service.DecryptElGamalU32(newSupplyKeypair.SecretKey, rotate.Proofs.NewSupplyCiphertext)
	if err != nil || rotatedSupply != 500_000 {
		t.Fatalf("rotated supply = %d, %v", rotatedSupply, err)
	}

	withheldCiphertext, err := service.EncryptElGamalWithOpening(withdrawKeypair.PublicKey, 37, PedersenOpening{16})
	if err != nil {
		t.Fatal(err)
	}
	withheld, err := service.WithdrawConfidentialWithheldFromMintWithProofs(token2022Key(2), token2022Key(4), token2022Key(3), nil, ConfidentialTransferFeeConfig{WithheldAmount: withheldCiphertext}, destinationState, withdrawKeypair, destinationAE)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyZKProofData(withheld.Proofs.Equality); err != nil {
		t.Fatal(err)
	}
	withheldDestinationAmount, err := service.DecryptElGamalU32(destinationKeypair.SecretKey, withheld.Proofs.DestinationCiphertext)
	if err != nil || withheldDestinationAmount != 37 {
		t.Fatalf("withheld destination amount = %d, %v", withheldDestinationAmount, err)
	}
	withheldOne, err := service.EncryptElGamalWithOpening(withdrawKeypair.PublicKey, 11, PedersenOpening{17})
	if err != nil {
		t.Fatal(err)
	}
	withheldTwo, err := service.EncryptElGamalWithOpening(withdrawKeypair.PublicKey, 12, PedersenOpening{18})
	if err != nil {
		t.Fatal(err)
	}
	withheldAccounts, err := service.WithdrawConfidentialWithheldFromAccountsWithProofs(token2022Key(2), token2022Key(4), token2022Key(3), nil, []ConfidentialWithheldAccount{
		{Address: token2022Key(5), State: ConfidentialTransferFeeAmount{WithheldAmount: withheldOne}},
		{Address: token2022Key(6), State: ConfidentialTransferFeeAmount{WithheldAmount: withheldTwo}},
	}, destinationState, withdrawKeypair, destinationAE)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyZKProofData(withheldAccounts.Proofs.Equality); err != nil {
		t.Fatal(err)
	}

	plans := []struct {
		name           string
		instructions   []solana.Instruction
		subInstruction byte
		discriminators []byte
	}{
		{name: "withdraw", instructions: withdraw.Instructions, subInstruction: 6, discriminators: []byte{byte(ConfidentialTransferExtensionInstruction), 3, 6}},
		{name: "transfer", instructions: transfer.Instructions, subInstruction: 7, discriminators: []byte{byte(ConfidentialTransferExtensionInstruction), 3, 12, 7}},
		{name: "transfer with fee", instructions: withFee.Instructions, subInstruction: 13, discriminators: []byte{byte(ConfidentialTransferExtensionInstruction), 3, 12, 5, 10, 8}},
		{name: "mint", instructions: mint.Instructions, subInstruction: 3, discriminators: []byte{byte(ConfidentialMintBurnExtensionInstruction), 3, 12, 7}},
		{name: "burn", instructions: burn.Instructions, subInstruction: 4, discriminators: []byte{byte(ConfidentialMintBurnExtensionInstruction), 3, 12, 7}},
		{name: "rotate", instructions: rotate.Instructions, subInstruction: 1, discriminators: []byte{byte(ConfidentialMintBurnExtensionInstruction), 2}},
		{name: "withheld mint", instructions: withheld.Instructions, subInstruction: 1, discriminators: []byte{byte(ConfidentialTransferFeeExtensionInstruction), 2}},
		{name: "withheld accounts", instructions: withheldAccounts.Instructions, subInstruction: 2, discriminators: []byte{byte(ConfidentialTransferFeeExtensionInstruction), 2}},
	}
	for _, plan := range plans {
		if len(plan.instructions) != len(plan.discriminators) {
			t.Fatalf("%s instruction count = %d, want %d", plan.name, len(plan.instructions), len(plan.discriminators))
		}
		for index, instruction := range plan.instructions {
			data, err := instruction.Data()
			if err != nil {
				t.Fatal(err)
			}
			if data[0] != plan.discriminators[index] {
				t.Fatalf("%s instruction %d discriminator = %d, want %d", plan.name, index, data[0], plan.discriminators[index])
			}
			if index == 0 && data[1] != plan.subInstruction {
				t.Fatalf("%s sub-instruction = %d, want %d", plan.name, data[1], plan.subInstruction)
			}
		}
	}
}

func TestWithdrawConfidentialWithheldAccountsCheckedCount(t *testing.T) {
	service := ConfidentialTransferService{}
	sources := make([]ConfidentialWithheldAccount, 256)
	if _, err := service.WithdrawConfidentialWithheldFromAccountsWithProofs(token2022Key(1), token2022Key(2), token2022Key(3), nil, sources, ConfidentialTransferAccount{}, ElGamalKeypair{}, AEKey{}); err == nil {
		t.Fatal("expected source account count error")
	}
}
