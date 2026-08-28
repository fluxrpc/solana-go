package token2022

import solana "github.com/fluxrpc/solana-go"

func (service ConfidentialTransferService) InitializeConfidentialTransferFee(mint solana.PublicKey, authority *solana.PublicKey, withdrawAuthority ElGamalPubkey) *ConfidentialTransferFeeExtension {
	data := make([]byte, 0, 64)
	data = append(data, service.publicKeyBytes(authority)...)
	data = append(data, withdrawAuthority[:]...)
	return service.RawConfidentialTransferFeeInstruction(0, data, *solana.NewAccountMeta(mint, true, false))
}

func (service ConfidentialTransferService) WithdrawConfidentialWithheldTokensFromMint(mint, destination, authority solana.PublicKey, signers []solana.PublicKey, decryptableBalance AeCiphertext, proof ProofLocation) *ConfidentialTransferFeeExtension {
	data := make([]byte, 1, 37)
	data[0] = byte(proof.InstructionOffset)
	data = append(data, decryptableBalance[:]...)
	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(mint, true, false),
		solana.NewAccountMeta(destination, true, false),
	}
	accounts = append(accounts, service.proofAccounts(proof)...)
	accounts = append(accounts, service.authorityAccounts(authority, signers)...)
	return service.RawConfidentialTransferFeeInstruction(1, data, service.accountValues(accounts)...)
}

func (service ConfidentialTransferService) WithdrawConfidentialWithheldTokensFromAccounts(mint, destination, authority solana.PublicKey, signers []solana.PublicKey, sourceAccounts []solana.PublicKey, decryptableBalance AeCiphertext, proof ProofLocation) *ConfidentialTransferFeeExtension {
	data := make([]byte, 2, 38)
	data[0] = byte(len(sourceAccounts))
	data[1] = byte(proof.InstructionOffset)
	data = append(data, decryptableBalance[:]...)
	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(mint, true, false),
		solana.NewAccountMeta(destination, true, false),
	}
	accounts = append(accounts, service.proofAccounts(proof)...)
	accounts = append(accounts, service.authorityAccounts(authority, signers)...)
	for _, source := range sourceAccounts {
		accounts = append(accounts, solana.NewAccountMeta(source, true, false))
	}
	return service.RawConfidentialTransferFeeInstruction(2, data, service.accountValues(accounts)...)
}

func (service ConfidentialTransferService) HarvestConfidentialWithheldTokensToMint(mint solana.PublicKey, sourceAccounts []solana.PublicKey) *ConfidentialTransferFeeExtension {
	accounts := solana.AccountMetaSlice{solana.NewAccountMeta(mint, true, false)}
	for _, source := range sourceAccounts {
		accounts = append(accounts, solana.NewAccountMeta(source, true, false))
	}
	return service.RawConfidentialTransferFeeInstruction(3, nil, service.accountValues(accounts)...)
}

func (service ConfidentialTransferService) EnableConfidentialHarvestToMint(mint, authority solana.PublicKey, signers []solana.PublicKey) *ConfidentialTransferFeeExtension {
	return service.confidentialHarvestInstruction(4, mint, authority, signers)
}

func (service ConfidentialTransferService) DisableConfidentialHarvestToMint(mint, authority solana.PublicKey, signers []solana.PublicKey) *ConfidentialTransferFeeExtension {
	return service.confidentialHarvestInstruction(5, mint, authority, signers)
}

func (service ConfidentialTransferService) confidentialHarvestInstruction(subInstruction uint8, mint, authority solana.PublicKey, signers []solana.PublicKey) *ConfidentialTransferFeeExtension {
	accounts := solana.AccountMetaSlice{solana.NewAccountMeta(mint, true, false)}
	accounts = append(accounts, service.authorityAccounts(authority, signers)...)
	return service.RawConfidentialTransferFeeInstruction(subInstruction, nil, service.accountValues(accounts)...)
}
