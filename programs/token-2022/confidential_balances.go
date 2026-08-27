package token2022

import (
	"fmt"
	"math/bits"
)

func (service ConfidentialTransferService) DecryptConfidentialAvailableBalance(key AEKey, account ConfidentialTransferAccount) (uint64, error) {
	amount, err := service.DecryptAEAmount(key, account.DecryptableAvailableBalance)
	if err != nil {
		return 0, fmt.Errorf("decrypt confidential available balance: %w", err)
	}
	return amount, nil
}

func (service ConfidentialTransferService) DecryptConfidentialPendingBalance(secret ElGamalSecretKey, account ConfidentialTransferAccount) (uint64, error) {
	amount, err := service.DecryptPendingBalance(secret, account.PendingBalanceLo, account.PendingBalanceHi)
	if err != nil {
		return 0, fmt.Errorf("decrypt confidential pending balance: %w", err)
	}
	return amount, nil
}

func (service ConfidentialTransferService) DecryptConfidentialTotalBalance(secret ElGamalSecretKey, key AEKey, account ConfidentialTransferAccount) (uint64, error) {
	pending, err := service.DecryptConfidentialPendingBalance(secret, account)
	if err != nil {
		return 0, err
	}
	available, err := service.DecryptConfidentialAvailableBalance(key, account)
	if err != nil {
		return 0, err
	}
	total, carry := bits.Add64(available, pending, 0)
	if carry != 0 {
		return 0, fmt.Errorf("decrypt confidential total balance: overflow")
	}
	return total, nil
}

func (service ConfidentialTransferService) NewDecryptableAvailableBalanceForApplyPending(secret ElGamalSecretKey, key AEKey, account ConfidentialTransferAccount) (AeCiphertext, error) {
	amount, err := service.availableBalanceAfterPending(secret, key, account)
	if err != nil {
		return AeCiphertext{}, err
	}
	return service.EncryptAEAmount(key, amount)
}

func (service ConfidentialTransferService) NewDecryptableAvailableBalanceForApplyPendingWithNonce(secret ElGamalSecretKey, key AEKey, account ConfidentialTransferAccount, nonce [12]byte) (AeCiphertext, error) {
	amount, err := service.availableBalanceAfterPending(secret, key, account)
	if err != nil {
		return AeCiphertext{}, err
	}
	return service.EncryptAEAmountWithNonce(key, amount, nonce)
}

func (service ConfidentialTransferService) NewDecryptableAvailableBalanceForSpend(key AEKey, account ConfidentialTransferAccount, amount uint64) (AeCiphertext, error) {
	balance, err := service.availableBalanceAfterSpend(key, account, amount, "insufficient funds")
	if err != nil {
		return AeCiphertext{}, err
	}
	return service.EncryptAEAmount(key, balance)
}

func (service ConfidentialTransferService) NewDecryptableAvailableBalanceForSpendWithNonce(key AEKey, account ConfidentialTransferAccount, amount uint64, nonce [12]byte) (AeCiphertext, error) {
	balance, err := service.availableBalanceAfterSpend(key, account, amount, "insufficient funds")
	if err != nil {
		return AeCiphertext{}, err
	}
	return service.EncryptAEAmountWithNonce(key, balance, nonce)
}

func (service ConfidentialTransferService) NewDecryptableAvailableBalanceForBurn(key AEKey, account ConfidentialTransferAccount, amount uint64) (AeCiphertext, error) {
	balance, err := service.availableBalanceAfterSpend(key, account, amount, "overflow")
	if err != nil {
		return AeCiphertext{}, err
	}
	return service.EncryptAEAmount(key, balance)
}

func (service ConfidentialTransferService) NewDecryptableAvailableBalanceForBurnWithNonce(key AEKey, account ConfidentialTransferAccount, amount uint64, nonce [12]byte) (AeCiphertext, error) {
	balance, err := service.availableBalanceAfterSpend(key, account, amount, "overflow")
	if err != nil {
		return AeCiphertext{}, err
	}
	return service.EncryptAEAmountWithNonce(key, balance, nonce)
}

func (service ConfidentialTransferService) DecryptConfidentialWithheldAmount(secret ElGamalSecretKey, withheld ElGamalCiphertext) (uint32, error) {
	amount, err := service.DecryptElGamalU32(secret, withheld)
	if err != nil {
		return 0, fmt.Errorf("decrypt confidential withheld amount: %w", err)
	}
	return amount, nil
}

func (service ConfidentialTransferService) DecryptCurrentConfidentialSupply(key AEKey, keypair ElGamalKeypair, state ConfidentialMintBurn) (uint64, error) {
	decryptableSupply, err := service.DecryptAEAmount(key, state.DecryptableSupply)
	if err != nil {
		return 0, fmt.Errorf("decrypt current confidential supply: %w", err)
	}
	decryptableCiphertext, _, err := service.EncryptElGamal(keypair.PublicKey, decryptableSupply)
	if err != nil {
		return 0, fmt.Errorf("decrypt current confidential supply: %w", err)
	}
	deltaCiphertext, err := service.SubtractElGamalCiphertexts(decryptableCiphertext, state.ConfidentialSupply)
	if err != nil {
		return 0, fmt.Errorf("decrypt current confidential supply: %w", err)
	}
	delta, err := service.DecryptElGamalU32(keypair.SecretKey, deltaCiphertext)
	if err != nil {
		return 0, fmt.Errorf("decrypt current confidential supply: %w", err)
	}
	if decryptableSupply < uint64(delta) {
		return 0, fmt.Errorf("decrypt current confidential supply: overflow")
	}
	return decryptableSupply - uint64(delta), nil
}

func (service ConfidentialTransferService) NewDecryptableSupply(key AEKey, keypair ElGamalKeypair, state ConfidentialMintBurn, mintAmount uint64) (AeCiphertext, error) {
	supply, err := service.confidentialSupplyAfterMint(key, keypair, state, mintAmount)
	if err != nil {
		return AeCiphertext{}, err
	}
	return service.EncryptAEAmount(key, supply)
}

func (service ConfidentialTransferService) NewDecryptableSupplyWithNonce(key AEKey, keypair ElGamalKeypair, state ConfidentialMintBurn, mintAmount uint64, nonce [12]byte) (AeCiphertext, error) {
	supply, err := service.confidentialSupplyAfterMint(key, keypair, state, mintAmount)
	if err != nil {
		return AeCiphertext{}, err
	}
	return service.EncryptAEAmountWithNonce(key, supply, nonce)
}

func (service ConfidentialTransferService) availableBalanceAfterPending(secret ElGamalSecretKey, key AEKey, account ConfidentialTransferAccount) (uint64, error) {
	pending, err := service.DecryptConfidentialPendingBalance(secret, account)
	if err != nil {
		return 0, err
	}
	available, err := service.DecryptConfidentialAvailableBalance(key, account)
	if err != nil {
		return 0, err
	}
	balance, carry := bits.Add64(available, pending, 0)
	if carry != 0 {
		return 0, fmt.Errorf("new decryptable available balance for apply pending: overflow")
	}
	return balance, nil
}

func (service ConfidentialTransferService) availableBalanceAfterSpend(key AEKey, account ConfidentialTransferAccount, amount uint64, underflow string) (uint64, error) {
	available, err := service.DecryptConfidentialAvailableBalance(key, account)
	if err != nil {
		return 0, err
	}
	if available < amount {
		return 0, fmt.Errorf("new decryptable available balance: %s", underflow)
	}
	return available - amount, nil
}

func (service ConfidentialTransferService) confidentialSupplyAfterMint(key AEKey, keypair ElGamalKeypair, state ConfidentialMintBurn, mintAmount uint64) (uint64, error) {
	current, err := service.DecryptCurrentConfidentialSupply(key, keypair, state)
	if err != nil {
		return 0, err
	}
	supply, carry := bits.Add64(current, mintAmount, 0)
	if carry != 0 {
		return 0, fmt.Errorf("new decryptable confidential supply: overflow")
	}
	return supply, nil
}
