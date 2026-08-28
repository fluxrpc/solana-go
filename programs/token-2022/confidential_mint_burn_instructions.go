package token2022

import (
	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

func (service ConfidentialTransferService) InitializeConfidentialMintBurn(mint solana.PublicKey, supplyPubkey ElGamalPubkey, decryptableSupply AeCiphertext) *ConfidentialMintBurnExtension {
	data := make([]byte, 0, 68)
	data = append(data, supplyPubkey[:]...)
	data = append(data, decryptableSupply[:]...)
	account := *solana.NewAccountMeta(mint, true, false)
	return service.RawConfidentialMintBurnInstruction(0, data, account)
}

func (service ConfidentialTransferService) RotateConfidentialSupplyKey(mint, authority solana.PublicKey, signers []solana.PublicKey, supplyPubkey ElGamalPubkey, proof ProofLocation) *ConfidentialMintBurnExtension {
	data := make([]byte, 0, 33)
	data = append(data, supplyPubkey[:]...)
	data = append(data, byte(proof.InstructionOffset))
	accounts := solana.AccountMetaSlice{solana.NewAccountMeta(mint, true, false)}
	accounts = append(accounts, service.proofAccounts(proof)...)
	accounts = append(accounts, service.authorityAccounts(authority, signers)...)
	return service.RawConfidentialMintBurnInstruction(1, data, service.accountValues(accounts)...)
}

func (service ConfidentialTransferService) UpdateConfidentialSupply(mint, authority solana.PublicKey, signers []solana.PublicKey, decryptableSupply AeCiphertext) *ConfidentialMintBurnExtension {
	accounts := solana.AccountMetaSlice{solana.NewAccountMeta(mint, true, false)}
	accounts = append(accounts, service.authorityAccounts(authority, signers)...)
	return service.RawConfidentialMintBurnInstruction(2, decryptableSupply[:], service.accountValues(accounts)...)
}

func (service ConfidentialTransferService) ConfidentialMint(tokenAccount, mint, authority solana.PublicKey, signers []solana.PublicKey, decryptableSupply AeCiphertext, auditorCiphertextLo, auditorCiphertextHi ElGamalCiphertext, equality, validity, rangeProof ProofLocation) *ConfidentialMintBurnExtension {
	data := service.mintBurnAmountData(decryptableSupply, auditorCiphertextLo, auditorCiphertextHi, equality, validity, rangeProof)
	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(tokenAccount, true, false),
		solana.NewAccountMeta(mint, true, false),
	}
	accounts = append(accounts, service.proofAccounts(equality, validity, rangeProof)...)
	accounts = append(accounts, service.authorityAccounts(authority, signers)...)
	return service.RawConfidentialMintBurnInstruction(3, data, service.accountValues(accounts)...)
}

func (service ConfidentialTransferService) ConfidentialBurn(tokenAccount, mint, authority solana.PublicKey, signers []solana.PublicKey, decryptableBalance AeCiphertext, auditorCiphertextLo, auditorCiphertextHi ElGamalCiphertext, equality, validity, rangeProof ProofLocation) *ConfidentialMintBurnExtension {
	data := service.mintBurnAmountData(decryptableBalance, auditorCiphertextLo, auditorCiphertextHi, equality, validity, rangeProof)
	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(tokenAccount, true, false),
		solana.NewAccountMeta(mint, true, false),
	}
	accounts = append(accounts, service.proofAccounts(equality, validity, rangeProof)...)
	accounts = append(accounts, service.authorityAccounts(authority, signers)...)
	return service.RawConfidentialMintBurnInstruction(4, data, service.accountValues(accounts)...)
}

func (service ConfidentialTransferService) ApplyPendingConfidentialBurn(mint, authority solana.PublicKey, signers []solana.PublicKey) *ConfidentialMintBurnExtension {
	accounts := solana.AccountMetaSlice{solana.NewAccountMeta(mint, true, false)}
	accounts = append(accounts, service.authorityAccounts(authority, signers)...)
	return service.RawConfidentialMintBurnInstruction(5, nil, service.accountValues(accounts)...)
}

func (ConfidentialTransferService) mintBurnAmountData(decryptable AeCiphertext, auditorCiphertextLo, auditorCiphertextHi ElGamalCiphertext, equality, validity, rangeProof ProofLocation) []byte {
	enc := bin.NewEncoder(make([]byte, 0, 167))
	enc.WriteBytes(decryptable[:])
	enc.WriteBytes(auditorCiphertextLo[:])
	enc.WriteBytes(auditorCiphertextHi[:])
	enc.WriteUint8(byte(equality.InstructionOffset))
	enc.WriteUint8(byte(validity.InstructionOffset))
	enc.WriteUint8(byte(rangeProof.InstructionOffset))
	return enc.Bytes()
}

func (ConfidentialTransferService) accountValues(accounts solana.AccountMetaSlice) []solana.AccountMeta {
	values := make([]solana.AccountMeta, len(accounts))
	for i, account := range accounts {
		values[i] = *account
	}
	return values
}
