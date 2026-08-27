package token2022

import (
	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

func (service ConfidentialTransferService) InitializeConfidentialTransferMint(mint solana.PublicKey, authority *solana.PublicKey, autoApprove bool, auditor *ElGamalPubkey) *ConfidentialTransferExtension {
	data := make([]byte, 0, 65)
	data = append(data, service.publicKeyBytes(authority)...)
	data = append(data, byte(0))
	if autoApprove {
		data[32] = 1
	}
	data = append(data, service.elGamalPubkeyBytes(auditor)...)
	return service.RawConfidentialTransferInstruction(0, data, *solana.NewAccountMeta(mint, true, false))
}

func (service ConfidentialTransferService) UpdateConfidentialTransferMint(mint, authority solana.PublicKey, signers []solana.PublicKey, autoApprove bool, auditor *ElGamalPubkey) *ConfidentialTransferExtension {
	data := make([]byte, 1, 33)
	if autoApprove {
		data[0] = 1
	}
	data = append(data, service.elGamalPubkeyBytes(auditor)...)
	accounts := solana.AccountMetaSlice{solana.NewAccountMeta(mint, true, false)}
	accounts = append(accounts, service.authorityAccounts(authority, signers)...)
	return service.RawConfidentialTransferInstruction(1, data, service.accountValues(accounts)...)
}

func (service ConfidentialTransferService) ConfigureConfidentialTransferAccount(tokenAccount, mint, authority solana.PublicKey, signers []solana.PublicKey, decryptableZeroBalance AeCiphertext, maximumPendingBalanceCreditCounter uint64, proof ProofLocation) *ConfidentialTransferExtension {
	enc := bin.NewEncoder(make([]byte, 0, 45))
	enc.WriteBytes(decryptableZeroBalance[:])
	enc.WriteUint64(maximumPendingBalanceCreditCounter)
	enc.WriteUint8(byte(proof.InstructionOffset))
	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(tokenAccount, true, false),
		solana.NewAccountMeta(mint, false, false),
	}
	accounts = append(accounts, service.proofAccounts(proof)...)
	accounts = append(accounts, service.authorityAccounts(authority, signers)...)
	return service.RawConfidentialTransferInstruction(2, enc.Bytes(), service.accountValues(accounts)...)
}

func (service ConfidentialTransferService) ApproveConfidentialTransferAccount(tokenAccount, mint, authority solana.PublicKey, signers []solana.PublicKey) *ConfidentialTransferExtension {
	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(tokenAccount, true, false),
		solana.NewAccountMeta(mint, false, false),
	}
	accounts = append(accounts, service.authorityAccounts(authority, signers)...)
	return service.RawConfidentialTransferInstruction(3, nil, service.accountValues(accounts)...)
}

func (service ConfidentialTransferService) EmptyConfidentialTransferAccount(tokenAccount, authority solana.PublicKey, signers []solana.PublicKey, proof ProofLocation) *ConfidentialTransferExtension {
	accounts := solana.AccountMetaSlice{solana.NewAccountMeta(tokenAccount, true, false)}
	accounts = append(accounts, service.proofAccounts(proof)...)
	accounts = append(accounts, service.authorityAccounts(authority, signers)...)
	return service.RawConfidentialTransferInstruction(4, []byte{byte(proof.InstructionOffset)}, service.accountValues(accounts)...)
}

func (service ConfidentialTransferService) ConfidentialDeposit(tokenAccount, mint, authority solana.PublicKey, signers []solana.PublicKey, amount uint64, decimals uint8) *ConfidentialTransferExtension {
	enc := bin.NewEncoder(make([]byte, 0, 9))
	enc.WriteUint64(amount)
	enc.WriteUint8(decimals)
	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(tokenAccount, true, false),
		solana.NewAccountMeta(mint, false, false),
	}
	accounts = append(accounts, service.authorityAccounts(authority, signers)...)
	return service.RawConfidentialTransferInstruction(5, enc.Bytes(), service.accountValues(accounts)...)
}

func (service ConfidentialTransferService) ConfidentialWithdraw(tokenAccount, mint, authority solana.PublicKey, signers []solana.PublicKey, amount uint64, decimals uint8, decryptableBalance AeCiphertext, equality, rangeProof ProofLocation) *ConfidentialTransferExtension {
	enc := bin.NewEncoder(make([]byte, 0, 47))
	enc.WriteUint64(amount)
	enc.WriteUint8(decimals)
	enc.WriteBytes(decryptableBalance[:])
	enc.WriteUint8(byte(equality.InstructionOffset))
	enc.WriteUint8(byte(rangeProof.InstructionOffset))
	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(tokenAccount, true, false),
		solana.NewAccountMeta(mint, false, false),
	}
	accounts = append(accounts, service.proofAccounts(equality, rangeProof)...)
	accounts = append(accounts, service.authorityAccounts(authority, signers)...)
	return service.RawConfidentialTransferInstruction(6, enc.Bytes(), service.accountValues(accounts)...)
}

func (service ConfidentialTransferService) ConfidentialTransfer(source, mint, destination, authority solana.PublicKey, signers []solana.PublicKey, decryptableBalance AeCiphertext, auditorCiphertextLo, auditorCiphertextHi ElGamalCiphertext, equality, validity, rangeProof ProofLocation) *ConfidentialTransferExtension {
	data := service.transferAmountData(decryptableBalance, auditorCiphertextLo, auditorCiphertextHi, equality, validity, rangeProof)
	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(source, true, false),
		solana.NewAccountMeta(mint, false, false),
		solana.NewAccountMeta(destination, true, false),
	}
	accounts = append(accounts, service.proofAccounts(equality, validity, rangeProof)...)
	accounts = append(accounts, service.authorityAccounts(authority, signers)...)
	return service.RawConfidentialTransferInstruction(7, data, service.accountValues(accounts)...)
}

func (service ConfidentialTransferService) ApplyConfidentialPendingBalance(tokenAccount, authority solana.PublicKey, signers []solana.PublicKey, expectedPendingBalanceCreditCounter uint64, decryptableBalance AeCiphertext) *ConfidentialTransferExtension {
	enc := bin.NewEncoder(make([]byte, 0, 44))
	enc.WriteUint64(expectedPendingBalanceCreditCounter)
	enc.WriteBytes(decryptableBalance[:])
	accounts := solana.AccountMetaSlice{solana.NewAccountMeta(tokenAccount, true, false)}
	accounts = append(accounts, service.authorityAccounts(authority, signers)...)
	return service.RawConfidentialTransferInstruction(8, enc.Bytes(), service.accountValues(accounts)...)
}

func (service ConfidentialTransferService) EnableConfidentialCredits(tokenAccount, authority solana.PublicKey, signers []solana.PublicKey) *ConfidentialTransferExtension {
	return service.balanceCreditsInstruction(9, tokenAccount, authority, signers)
}

func (service ConfidentialTransferService) DisableConfidentialCredits(tokenAccount, authority solana.PublicKey, signers []solana.PublicKey) *ConfidentialTransferExtension {
	return service.balanceCreditsInstruction(10, tokenAccount, authority, signers)
}

func (service ConfidentialTransferService) EnableNonConfidentialCredits(tokenAccount, authority solana.PublicKey, signers []solana.PublicKey) *ConfidentialTransferExtension {
	return service.balanceCreditsInstruction(11, tokenAccount, authority, signers)
}

func (service ConfidentialTransferService) DisableNonConfidentialCredits(tokenAccount, authority solana.PublicKey, signers []solana.PublicKey) *ConfidentialTransferExtension {
	return service.balanceCreditsInstruction(12, tokenAccount, authority, signers)
}

func (service ConfidentialTransferService) ConfidentialTransferWithFee(source, mint, destination, authority solana.PublicKey, signers []solana.PublicKey, decryptableBalance AeCiphertext, auditorCiphertextLo, auditorCiphertextHi ElGamalCiphertext, equality, transferValidity, feeSigma, feeValidity, rangeProof ProofLocation) *ConfidentialTransferExtension {
	enc := bin.NewEncoder(make([]byte, 0, 169))
	enc.WriteBytes(decryptableBalance[:])
	enc.WriteBytes(auditorCiphertextLo[:])
	enc.WriteBytes(auditorCiphertextHi[:])
	enc.WriteUint8(byte(equality.InstructionOffset))
	enc.WriteUint8(byte(transferValidity.InstructionOffset))
	enc.WriteUint8(byte(feeSigma.InstructionOffset))
	enc.WriteUint8(byte(feeValidity.InstructionOffset))
	enc.WriteUint8(byte(rangeProof.InstructionOffset))
	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(source, true, false),
		solana.NewAccountMeta(mint, false, false),
		solana.NewAccountMeta(destination, true, false),
	}
	accounts = append(accounts, service.proofAccounts(equality, transferValidity, feeSigma, feeValidity, rangeProof)...)
	accounts = append(accounts, service.authorityAccounts(authority, signers)...)
	return service.RawConfidentialTransferInstruction(13, enc.Bytes(), service.accountValues(accounts)...)
}

func (service ConfidentialTransferService) ConfigureConfidentialTransferAccountWithRegistry(tokenAccount, mint, registry solana.PublicKey, payer *solana.PublicKey) *ConfidentialTransferExtension {
	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(tokenAccount, true, false),
		solana.NewAccountMeta(mint, false, false),
		solana.NewAccountMeta(registry, false, false),
	}
	if payer != nil {
		accounts = append(accounts,
			solana.NewAccountMeta(*payer, true, true),
			solana.NewAccountMeta(solana.SystemProgramID, false, false),
		)
	}
	return service.RawConfidentialTransferInstruction(14, nil, service.accountValues(accounts)...)
}

func (service ConfidentialTransferService) balanceCreditsInstruction(subInstruction uint8, tokenAccount, authority solana.PublicKey, signers []solana.PublicKey) *ConfidentialTransferExtension {
	accounts := solana.AccountMetaSlice{solana.NewAccountMeta(tokenAccount, true, false)}
	accounts = append(accounts, service.authorityAccounts(authority, signers)...)
	return service.RawConfidentialTransferInstruction(subInstruction, nil, service.accountValues(accounts)...)
}

func (ConfidentialTransferService) authorityAccounts(authority solana.PublicKey, signers []solana.PublicKey) solana.AccountMetaSlice {
	accounts := solana.AccountMetaSlice{solana.NewAccountMeta(authority, false, len(signers) == 0)}
	for _, signer := range signers {
		accounts = append(accounts, solana.NewAccountMeta(signer, false, true))
	}
	return accounts
}

func (ConfidentialTransferService) transferAmountData(decryptable AeCiphertext, auditorCiphertextLo, auditorCiphertextHi ElGamalCiphertext, equality, validity, rangeProof ProofLocation) []byte {
	data := make([]byte, 0, 167)
	data = append(data, decryptable[:]...)
	data = append(data, auditorCiphertextLo[:]...)
	data = append(data, auditorCiphertextHi[:]...)
	data = append(data, byte(equality.InstructionOffset), byte(validity.InstructionOffset), byte(rangeProof.InstructionOffset))
	return data
}

func (ConfidentialTransferService) publicKeyBytes(key *solana.PublicKey) []byte {
	if key == nil {
		return make([]byte, 32)
	}
	return key[:]
}

func (ConfidentialTransferService) elGamalPubkeyBytes(key *ElGamalPubkey) []byte {
	if key == nil {
		return make([]byte, 32)
	}
	return key[:]
}
