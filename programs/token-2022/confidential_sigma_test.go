package token2022

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestPubkeyValidityProofRustInteroperability(t *testing.T) {
	service := ConfidentialTransferService{}
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
	random := make([]byte, 64)
	random[0] = 2
	data, err := service.generatePubkeyValidityProofWithReader(keypair, bytes.NewReader(random))
	if err != nil {
		t.Fatal(err)
	}
	if data.Discriminator != 4 || len(data.Context) != 32 || len(data.Proof) != 64 {
		t.Fatalf("proof data sizes = %d, %d, %d", data.Discriminator, len(data.Context), len(data.Proof))
	}
	if got, want := hex.EncodeToString(data.Context), "8c9240b456a9e6dc65c377a1048d745f94a08cdb7f44cbcd7b46f34048871134"; got != want {
		t.Fatalf("context = %s, want %s", got, want)
	}
	if got, want := hex.EncodeToString(data.Proof), "10f21a8723d8943e2b37207a4815638fcc0b5efc9dc3346445ca985c6d5c2207aa894270d8eeb1a85ffe06f052810378dbf28363168f22f99988214c6ca80b09"; got != want {
		t.Fatalf("proof = %s, want %s", got, want)
	}
	if err := service.VerifyPubkeyValidityProof(data); err != nil {
		t.Fatal(err)
	}
	rustContext, err := hex.DecodeString("8c9240b456a9e6dc65c377a1048d745f94a08cdb7f44cbcd7b46f34048871134")
	if err != nil {
		t.Fatal(err)
	}
	rustProof, err := hex.DecodeString("10f21a8723d8943e2b37207a4815638fcc0b5efc9dc3346445ca985c6d5c2207aa894270d8eeb1a85ffe06f052810378dbf28363168f22f99988214c6ca80b09")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyPubkeyValidityProof(ZKProofData{Discriminator: 4, Context: rustContext, Proof: rustProof}); err != nil {
		t.Fatalf("verify Rust proof: %v", err)
	}
	tampered := ZKProofData{Discriminator: data.Discriminator, Context: append([]byte(nil), data.Context...), Proof: append([]byte(nil), data.Proof...)}
	tampered.Proof[63] ^= 1
	if err := service.VerifyPubkeyValidityProof(tampered); err == nil {
		t.Fatal("expected tampered public key validity proof error")
	}
	if _, err := service.generatePubkeyValidityProofWithReader(keypair, bytes.NewReader(nil)); err == nil {
		t.Fatal("expected random source error")
	}
}

func TestZeroCiphertextProofRustInteroperability(t *testing.T) {
	service := ConfidentialTransferService{}
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
	openingBytes := make([]byte, 32)
	openingBytes[0] = 5
	opening, err := service.PedersenOpeningFromBytes(openingBytes)
	if err != nil {
		t.Fatal(err)
	}
	ciphertext, err := service.EncryptElGamalWithOpening(keypair.PublicKey, 0, opening)
	if err != nil {
		t.Fatal(err)
	}
	random := make([]byte, 64)
	random[0] = 2
	data, err := service.generateZeroCiphertextProofWithReader(keypair, ciphertext, bytes.NewReader(random))
	if err != nil {
		t.Fatal(err)
	}
	if data.Discriminator != 1 || len(data.Context) != 96 || len(data.Proof) != 96 {
		t.Fatalf("proof data sizes = %d, %d, %d", data.Discriminator, len(data.Context), len(data.Proof))
	}
	if got, want := hex.EncodeToString(data.Context), "8c9240b456a9e6dc65c377a1048d745f94a08cdb7f44cbcd7b46f3404887113484dffada0b0ea52ecd8ad35f8e608eb6b1b5da75901afd5e378c9184bb6b927184dffada0b0ea52ecd8ad35f8e608eb6b1b5da75901afd5e378c9184bb6b9271"; got != want {
		t.Fatalf("context = %s, want %s", got, want)
	}
	if got, want := hex.EncodeToString(data.Proof), "10f21a8723d8943e2b37207a4815638fcc0b5efc9dc3346445ca985c6d5c220708d10de140c0e83fc1589a46875e63e618d37239be0f33b1a1033eefbd1aa35eabe18f407c5e33d4dffb4787115caeacd897dacafbfbd0b154807983746fef0c"; got != want {
		t.Fatalf("proof = %s, want %s", got, want)
	}
	if err := service.VerifyZeroCiphertextProof(data); err != nil {
		t.Fatal(err)
	}
	rustContext, err := hex.DecodeString("8c9240b456a9e6dc65c377a1048d745f94a08cdb7f44cbcd7b46f3404887113484dffada0b0ea52ecd8ad35f8e608eb6b1b5da75901afd5e378c9184bb6b927184dffada0b0ea52ecd8ad35f8e608eb6b1b5da75901afd5e378c9184bb6b9271")
	if err != nil {
		t.Fatal(err)
	}
	rustProof, err := hex.DecodeString("10f21a8723d8943e2b37207a4815638fcc0b5efc9dc3346445ca985c6d5c220708d10de140c0e83fc1589a46875e63e618d37239be0f33b1a1033eefbd1aa35eabe18f407c5e33d4dffb4787115caeacd897dacafbfbd0b154807983746fef0c")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyZeroCiphertextProof(ZKProofData{Discriminator: 1, Context: rustContext, Proof: rustProof}); err != nil {
		t.Fatalf("verify Rust proof: %v", err)
	}
	tampered := ZKProofData{Discriminator: data.Discriminator, Context: append([]byte(nil), data.Context...), Proof: append([]byte(nil), data.Proof...)}
	tampered.Proof[95] ^= 1
	if err := service.VerifyZeroCiphertextProof(tampered); err == nil {
		t.Fatal("expected tampered zero ciphertext proof error")
	}
	nonzero, err := service.EncryptElGamalWithOpening(keypair.PublicKey, 1, opening)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.generateZeroCiphertextProofWithReader(keypair, nonzero, bytes.NewReader(random)); err == nil {
		t.Fatal("expected nonzero ciphertext error")
	}
}

func TestConfidentialSigmaValidation(t *testing.T) {
	service := ConfidentialTransferService{}
	if err := service.VerifyPubkeyValidityProof(ZKProofData{Discriminator: 4}); err == nil {
		t.Fatal("expected invalid public key validity proof data error")
	}
	if err := service.VerifyZeroCiphertextProof(ZKProofData{Discriminator: 1}); err == nil {
		t.Fatal("expected invalid zero ciphertext proof data error")
	}
	identityPubkeyProof := ZKProofData{Discriminator: 4, Context: make([]byte, 32), Proof: make([]byte, 64)}
	if err := service.VerifyPubkeyValidityProof(identityPubkeyProof); err == nil {
		t.Fatal("expected identity public key error")
	}
	identityZeroProof := ZKProofData{Discriminator: 1, Context: make([]byte, 96), Proof: make([]byte, 96)}
	if err := service.VerifyZeroCiphertextProof(identityZeroProof); err == nil {
		t.Fatal("expected identity zero ciphertext context error")
	}
}
