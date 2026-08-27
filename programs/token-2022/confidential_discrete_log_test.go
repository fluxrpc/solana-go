package token2022

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestConfidentialDiscreteLogRustFixture(t *testing.T) {
	service := ConfidentialTransferService{}
	service.Start()
	targetBytes, err := hex.DecodeString("e00af9c74d9edb8ebcc160ceec97d531cbd6e2956f9e9162b8e9eda260e82e43")
	if err != nil {
		t.Fatal(err)
	}
	target, err := service.PedersenCommitmentFromBytes(targetBytes)
	if err != nil {
		t.Fatal(err)
	}
	amount, err := service.DecodePedersenCommitmentU32(target)
	if err != nil {
		t.Fatal(err)
	}
	if amount != 42 {
		t.Fatalf("decoded amount = %d, want 42", amount)
	}

	secretBytes, err := hex.DecodeString("0100000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	secret, err := service.ElGamalSecretKeyFromBytes(secretBytes)
	if err != nil {
		t.Fatal(err)
	}
	ciphertextBytes, err := hex.DecodeString("f06bddcc8c6fc05e28e513c93c8745048b70923dae7cd67e02376100b909e87410f21a8723d8943e2b37207a4815638fcc0b5efc9dc3346445ca985c6d5c2207")
	if err != nil {
		t.Fatal(err)
	}
	ciphertext, err := service.ElGamalCiphertextFromBytes(ciphertextBytes)
	if err != nil {
		t.Fatal(err)
	}
	amount, err = service.DecryptElGamalU32(secret, ciphertext)
	if err != nil {
		t.Fatal(err)
	}
	if amount != 42 {
		t.Fatalf("decrypted amount = %d, want 42", amount)
	}
}

func TestConfidentialDiscreteLogGroupedRustFixture(t *testing.T) {
	service := ConfidentialTransferService{}
	service.Start()
	secretBytes := make([]byte, 32)
	secretBytes[0] = 4
	secret, err := service.ElGamalSecretKeyFromBytes(secretBytes)
	if err != nil {
		t.Fatal(err)
	}
	ciphertextBytes, err := hex.DecodeString("f06bddcc8c6fc05e28e513c93c8745048b70923dae7cd67e02376100b909e87410f21a8723d8943e2b37207a4815638fcc0b5efc9dc3346445ca985c6d5c2207a0aca2678c338c9023b57c6b65b3bc8bf43a4c542c11484c3495c2d21d4c6455f05bc1df2831717c2992d85b57e0cf3d123fd6c254257de5f784be369747b249")
	if err != nil {
		t.Fatal(err)
	}
	ciphertext, err := service.GroupedElGamalCiphertext3FromBytes(ciphertextBytes)
	if err != nil {
		t.Fatal(err)
	}
	amount, err := service.DecryptGroupedElGamal3U32(secret, ciphertext, 2)
	if err != nil {
		t.Fatal(err)
	}
	if amount != 42 {
		t.Fatalf("decrypted grouped amount = %d, want 42", amount)
	}
}

func TestConfidentialDiscreteLogPendingBalance(t *testing.T) {
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
	low, err := service.EncryptElGamalWithOpening(keypair.PublicKey, 65535, PedersenOpening{})
	if err != nil {
		t.Fatal(err)
	}
	high, err := service.EncryptElGamalWithOpening(keypair.PublicKey, 7, PedersenOpening{})
	if err != nil {
		t.Fatal(err)
	}
	amount, err := service.DecryptPendingBalance(secret, low, high)
	if err != nil {
		t.Fatal(err)
	}
	if want := service.CombineAmount(65535, 7); amount != want {
		t.Fatalf("pending balance = %d, want %d", amount, want)
	}

	invalidLow, err := service.EncryptElGamalWithOpening(keypair.PublicKey, 65536, PedersenOpening{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.DecryptPendingBalance(secret, invalidLow, high); err == nil || !strings.Contains(err.Error(), "low amount exceeds 16 bits") {
		t.Fatalf("pending balance low error = %v", err)
	}
}

func TestConfidentialDiscreteLogBoundsAndTamper(t *testing.T) {
	service := ConfidentialTransferService{}
	service.Start()
	table := service.discreteLog
	maximum, err := service.CommitPedersen(uint64(^uint32(0)), PedersenOpening{})
	if err != nil {
		t.Fatal(err)
	}
	amount, err := service.decodePedersenCommitmentU32(maximum, table)
	if err != nil {
		t.Fatal(err)
	}
	if amount != ^uint32(0) {
		t.Fatalf("maximum decoded amount = %d", amount)
	}

	target, err := service.CommitPedersen(uint64(1)<<32, PedersenOpening{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.decodePedersenCommitmentU32(target, table); err == nil || !strings.Contains(err.Error(), "outside u32 range") {
		t.Fatalf("out-of-range error = %v", err)
	}

	invalid := PedersenCommitment{}
	for index := range invalid {
		invalid[index] = 0xff
	}
	if _, err := service.DecodePedersenCommitmentU32(invalid); err == nil || !strings.Contains(err.Error(), "decode Pedersen decrypt target") {
		t.Fatalf("tampered target error = %v", err)
	}
}
