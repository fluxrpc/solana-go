package token2022

import (
	solana "github.com/fluxrpc/solana-go"
	system "github.com/fluxrpc/solana-go/programs/system"
)

func (service ConfidentialTransferService) ConfigureConfidentialTransferAccountWithKeys(tokenAccount, mint, authority solana.PublicKey, signers []solana.PublicKey, keypair ElGamalKeypair, key AEKey, maximumPendingBalanceCreditCounter uint64) ([]solana.Instruction, error) {
	decryptableZero, err := service.EncryptAEAmount(key, 0)
	if err != nil {
		return nil, err
	}
	proof, err := service.GeneratePubkeyValidityProof(keypair)
	if err != nil {
		return nil, err
	}
	location := ProofLocation{InstructionOffset: 1, Proof: &proof}
	instruction := service.ConfigureConfidentialTransferAccount(tokenAccount, mint, authority, signers, decryptableZero, maximumPendingBalanceCreditCounter, location)
	return service.ConfidentialInstructions(instruction, location)
}

func (service ConfidentialTransferService) EmptyConfidentialTransferAccountWithProof(tokenAccount, authority solana.PublicKey, signers []solana.PublicKey, keypair ElGamalKeypair, availableBalance ElGamalCiphertext) ([]solana.Instruction, error) {
	proof, err := service.GenerateZeroCiphertextProof(keypair, availableBalance)
	if err != nil {
		return nil, err
	}
	location := ProofLocation{InstructionOffset: 1, Proof: &proof}
	instruction := service.EmptyConfidentialTransferAccount(tokenAccount, authority, signers, location)
	return service.ConfidentialInstructions(instruction, location)
}

func (service ConfidentialTransferService) ApplyConfidentialPendingBalanceWithKeys(tokenAccount, authority solana.PublicKey, signers []solana.PublicKey, account ConfidentialTransferAccount, keypair ElGamalKeypair, key AEKey) (*ConfidentialTransferExtension, error) {
	decryptableBalance, err := service.NewDecryptableAvailableBalanceForApplyPending(keypair.SecretKey, key, account)
	if err != nil {
		return nil, err
	}
	return service.ApplyConfidentialPendingBalance(tokenAccount, authority, signers, account.PendingBalanceCreditCounter, decryptableBalance), nil
}

func (service ConfidentialTransferService) CreateElGamalRegistryWithKeypair(owner solana.PublicKey, keypair ElGamalKeypair) ([]solana.Instruction, error) {
	proof, err := service.GeneratePubkeyValidityProof(keypair)
	if err != nil {
		return nil, err
	}
	return service.CreateElGamalRegistry(owner, ProofLocation{InstructionOffset: 1, Proof: &proof})
}

func (service ConfidentialTransferService) CreateFundedElGamalRegistryWithKeypair(payer, owner solana.PublicKey, lamports uint64, keypair ElGamalKeypair) ([]solana.Instruction, error) {
	registry, _, err := service.ElGamalRegistryAddress(owner)
	if err != nil {
		return nil, err
	}
	instructions, err := service.CreateElGamalRegistryWithKeypair(owner, keypair)
	if err != nil {
		return nil, err
	}
	return append([]solana.Instruction{system.NewTransferInstruction(lamports, payer, registry)}, instructions...), nil
}

func (service ConfidentialTransferService) UpdateElGamalRegistryWithKeypair(owner solana.PublicKey, keypair ElGamalKeypair) ([]solana.Instruction, error) {
	proof, err := service.GeneratePubkeyValidityProof(keypair)
	if err != nil {
		return nil, err
	}
	return service.UpdateElGamalRegistry(owner, ProofLocation{InstructionOffset: 1, Proof: &proof})
}
