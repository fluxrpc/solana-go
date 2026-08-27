package token2022

import (
	"encoding/hex"
	"math"
	"strings"
	"testing"
)

func TestConfidentialAccountBalanceFlow(t *testing.T) {
	service := ConfidentialTransferService{}
	service.Start()
	secretBytes := make([]byte, 32)
	secretBytes[0] = 1
	secret, err := service.ElGamalSecretKeyFromBytes(secretBytes)
	if err != nil {
		t.Fatal(err)
	}
	keypair, err := service.ElGamalKeypairFromSecret(secret)
	if err != nil {
		t.Fatal(err)
	}
	key := AEKey{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	available, err := service.EncryptAEAmountWithNonce(key, 100, [12]byte{1})
	if err != nil {
		t.Fatal(err)
	}
	low, err := service.EncryptElGamalWithOpening(keypair.PublicKey, 0x2345, PedersenOpening{})
	if err != nil {
		t.Fatal(err)
	}
	high, err := service.EncryptElGamalWithOpening(keypair.PublicKey, 3, PedersenOpening{})
	if err != nil {
		t.Fatal(err)
	}
	account := ConfidentialTransferAccount{
		PendingBalanceLo:            low,
		PendingBalanceHi:            high,
		DecryptableAvailableBalance: available,
	}
	availableAmount, err := service.DecryptConfidentialAvailableBalance(key, account)
	if err != nil {
		t.Fatal(err)
	}
	if availableAmount != 100 {
		t.Fatalf("available balance = %d, want 100", availableAmount)
	}
	pendingAmount, err := service.DecryptConfidentialPendingBalance(secret, account)
	if err != nil {
		t.Fatal(err)
	}
	if want := service.CombineAmount(0x2345, 3); pendingAmount != want {
		t.Fatalf("pending balance = %d, want %d", pendingAmount, want)
	}
	total, err := service.DecryptConfidentialTotalBalance(secret, key, account)
	if err != nil {
		t.Fatal(err)
	}
	if want := uint64(100) + service.CombineAmount(0x2345, 3); total != want {
		t.Fatalf("total balance = %d, want %d", total, want)
	}
	nonce := [12]byte{9, 8, 7}
	applied, err := service.NewDecryptableAvailableBalanceForApplyPendingWithNonce(secret, key, account, nonce)
	if err != nil {
		t.Fatal(err)
	}
	if appliedAmount, err := service.DecryptAEAmount(key, applied); err != nil || appliedAmount != total {
		t.Fatalf("applied balance = %d, %v, want %d", appliedAmount, err, total)
	}
	if got := [12]byte(applied[:12]); got != nonce {
		t.Fatalf("applied nonce = %x, want %x", got, nonce)
	}

	spent, err := service.NewDecryptableAvailableBalanceForSpendWithNonce(key, account, 40, nonce)
	if err != nil {
		t.Fatal(err)
	}
	if spentAmount, err := service.DecryptAEAmount(key, spent); err != nil || spentAmount != 60 {
		t.Fatalf("spent balance = %d, %v, want 60", spentAmount, err)
	}
	randomSpent, err := service.NewDecryptableAvailableBalanceForSpend(key, account, 10)
	if err != nil {
		t.Fatal(err)
	}
	if spentAmount, err := service.DecryptAEAmount(key, randomSpent); err != nil || spentAmount != 90 {
		t.Fatalf("random-nonce spent balance = %d, %v, want 90", spentAmount, err)
	}
	burned, err := service.NewDecryptableAvailableBalanceForBurnWithNonce(key, account, 40, nonce)
	if err != nil {
		t.Fatal(err)
	}
	if burnedAmount, err := service.DecryptAEAmount(key, burned); err != nil || burnedAmount != 60 {
		t.Fatalf("burned balance = %d, %v, want 60", burnedAmount, err)
	}
	if _, err := service.NewDecryptableAvailableBalanceForSpend(key, account, 101); err == nil || !strings.Contains(err.Error(), "insufficient funds") {
		t.Fatalf("spend error = %v", err)
	}
	if _, err := service.NewDecryptableAvailableBalanceForBurn(key, account, 101); err == nil || !strings.Contains(err.Error(), "overflow") {
		t.Fatalf("burn error = %v", err)
	}
}

func TestConfidentialTotalBalanceOverflow(t *testing.T) {
	service := ConfidentialTransferService{}
	service.Start()
	secretBytes := make([]byte, 32)
	secretBytes[0] = 1
	secret, err := service.ElGamalSecretKeyFromBytes(secretBytes)
	if err != nil {
		t.Fatal(err)
	}
	keypair, err := service.ElGamalKeypairFromSecret(secret)
	if err != nil {
		t.Fatal(err)
	}
	key := AEKey{1}
	available, err := service.EncryptAEAmountWithNonce(key, math.MaxUint64, [12]byte{})
	if err != nil {
		t.Fatal(err)
	}
	low, err := service.EncryptElGamalWithOpening(keypair.PublicKey, 1, PedersenOpening{})
	if err != nil {
		t.Fatal(err)
	}
	high, err := service.EncryptElGamalWithOpening(keypair.PublicKey, 0, PedersenOpening{})
	if err != nil {
		t.Fatal(err)
	}
	account := ConfidentialTransferAccount{
		PendingBalanceLo:            low,
		PendingBalanceHi:            high,
		DecryptableAvailableBalance: available,
	}
	if _, err := service.DecryptConfidentialTotalBalance(secret, key, account); err == nil || !strings.Contains(err.Error(), "overflow") {
		t.Fatalf("total balance error = %v", err)
	}
	if _, err := service.NewDecryptableAvailableBalanceForApplyPending(secret, key, account); err == nil || !strings.Contains(err.Error(), "overflow") {
		t.Fatalf("apply pending error = %v", err)
	}
}

func TestConfidentialSupplyAndWithheldFlow(t *testing.T) {
	service := ConfidentialTransferService{}
	service.Start()
	secretBytes := make([]byte, 32)
	secretBytes[0] = 1
	secret, err := service.ElGamalSecretKeyFromBytes(secretBytes)
	if err != nil {
		t.Fatal(err)
	}
	keypair, err := service.ElGamalKeypairFromSecret(secret)
	if err != nil {
		t.Fatal(err)
	}
	key := AEKey{15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0}
	decryptableSupply, err := service.EncryptAEAmountWithNonce(key, 1000, [12]byte{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	openingBytes := make([]byte, 32)
	openingBytes[0] = 2
	opening, err := service.PedersenOpeningFromBytes(openingBytes)
	if err != nil {
		t.Fatal(err)
	}
	confidentialSupply, err := service.EncryptElGamalWithOpening(keypair.PublicKey, 900, opening)
	if err != nil {
		t.Fatal(err)
	}
	state := ConfidentialMintBurn{
		ConfidentialSupply:  confidentialSupply,
		DecryptableSupply:   decryptableSupply,
		SupplyElGamalPubkey: keypair.PublicKey,
	}
	supply, err := service.DecryptCurrentConfidentialSupply(key, keypair, state)
	if err != nil {
		t.Fatal(err)
	}
	if supply != 900 {
		t.Fatalf("current supply = %d, want 900", supply)
	}
	nonce := [12]byte{4, 5, 6}
	updated, err := service.NewDecryptableSupplyWithNonce(key, keypair, state, 25, nonce)
	if err != nil {
		t.Fatal(err)
	}
	if updatedSupply, err := service.DecryptAEAmount(key, updated); err != nil || updatedSupply != 925 {
		t.Fatalf("updated supply = %d, %v, want 925", updatedSupply, err)
	}
	if got := [12]byte(updated[:12]); got != nonce {
		t.Fatalf("updated supply nonce = %x, want %x", got, nonce)
	}
	refreshed, err := service.NewDecryptableSupply(key, keypair, state, 0)
	if err != nil {
		t.Fatal(err)
	}
	if refreshedSupply, err := service.DecryptAEAmount(key, refreshed); err != nil || refreshedSupply != 900 {
		t.Fatalf("refreshed supply = %d, %v, want 900", refreshedSupply, err)
	}
	if _, err := service.NewDecryptableSupply(key, keypair, state, math.MaxUint64); err == nil || !strings.Contains(err.Error(), "overflow") {
		t.Fatalf("mint overflow error = %v", err)
	}

	withheldBytes, err := hex.DecodeString("f06bddcc8c6fc05e28e513c93c8745048b70923dae7cd67e02376100b909e87410f21a8723d8943e2b37207a4815638fcc0b5efc9dc3346445ca985c6d5c2207")
	if err != nil {
		t.Fatal(err)
	}
	withheld, err := service.ElGamalCiphertextFromBytes(withheldBytes)
	if err != nil {
		t.Fatal(err)
	}
	withheldAmount, err := service.DecryptConfidentialWithheldAmount(secret, withheld)
	if err != nil {
		t.Fatal(err)
	}
	if withheldAmount != 42 {
		t.Fatalf("withheld amount = %d, want 42", withheldAmount)
	}
}
