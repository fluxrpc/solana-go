package token2022

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestCiphertextCiphertextEqualityProofRustInteroperability(t *testing.T) {
	service := ConfidentialTransferService{}
	firstSecretBytes := make([]byte, 32)
	firstSecretBytes[0] = 1
	firstSecret, err := service.ElGamalSecretKeyFromBytes(firstSecretBytes)
	if err != nil {
		t.Fatal(err)
	}
	firstKeypair, err := service.ElGamalKeypairFromSecret(firstSecret)
	if err != nil {
		t.Fatal(err)
	}
	secondSecretBytes := make([]byte, 32)
	secondSecretBytes[0] = 3
	secondSecret, err := service.ElGamalSecretKeyFromBytes(secondSecretBytes)
	if err != nil {
		t.Fatal(err)
	}
	secondKeypair, err := service.ElGamalKeypairFromSecret(secondSecret)
	if err != nil {
		t.Fatal(err)
	}
	firstOpeningBytes := make([]byte, 32)
	firstOpeningBytes[0] = 2
	firstOpening, err := service.PedersenOpeningFromBytes(firstOpeningBytes)
	if err != nil {
		t.Fatal(err)
	}
	secondOpeningBytes := make([]byte, 32)
	secondOpeningBytes[0] = 5
	secondOpening, err := service.PedersenOpeningFromBytes(secondOpeningBytes)
	if err != nil {
		t.Fatal(err)
	}
	firstCiphertext, err := service.EncryptElGamalWithOpening(firstKeypair.PublicKey, 42, firstOpening)
	if err != nil {
		t.Fatal(err)
	}
	secondCiphertext, err := service.EncryptElGamalWithOpening(secondKeypair.PublicKey, 42, secondOpening)
	if err != nil {
		t.Fatal(err)
	}
	random := make([]byte, 192)
	random[0] = 8
	random[64] = 9
	random[128] = 10
	data, err := service.generateCiphertextCiphertextEqualityProofWithReader(firstKeypair, secondKeypair.PublicKey, firstCiphertext, secondCiphertext, secondOpening, 42, bytes.NewReader(random))
	if err != nil {
		t.Fatal(err)
	}
	if data.Discriminator != 2 || len(data.Context) != 192 || len(data.Proof) != 224 {
		t.Fatalf("proof data sizes = %d, %d, %d", data.Discriminator, len(data.Context), len(data.Proof))
	}
	if got, want := hex.EncodeToString(data.Context), "8c9240b456a9e6dc65c377a1048d745f94a08cdb7f44cbcd7b46f34048871134c29d170ab8a5b42a3520878501a87a27f9b5653fca8b0c59fc2786cf26e37824f06bddcc8c6fc05e28e513c93c8745048b70923dae7cd67e02376100b909e87410f21a8723d8943e2b37207a4815638fcc0b5efc9dc3346445ca985c6d5c2207422d138fc131895156f91e3e7e11de68e4c370da80e54cbc2bd1b5e3285c3e213e34d08d94993e15383c5e26aa984b3096bdd41012b951fdaf74796b71fb8930"; got != want {
		t.Fatalf("context = %s, want %s", got, want)
	}
	if got, want := hex.EncodeToString(data.Proof), "0c9022fbdb3f718a13d10f169e72dbc4b646f8341cc634f5f888517f7a36ec4bc4c88745cc0abdca52ec5425bf84125d849a9b87783f8c8b0912dcb5f92cab3ea4d4516715408ec537223c0631e312f14b0e860bff4e05300317309e2c5ec018143f2682b709ce56756559998653b3ade40cc295d04f2dc748868a6d7bd8a44917dc9e5ff27ffb748afe982dacce8711df1e86378b291fefa5fb950a6bea050f64d19c86bfe475c60fde60a754d74fb29510011cd7d01c3b39499bbc8f75f806a1fc426a52f39fe85a851e58e2212b045b9a9e15b8cf9bab3deaed3417941d0b"; got != want {
		t.Fatalf("proof = %s, want %s", got, want)
	}
	if err := service.VerifyCiphertextCiphertextEqualityProof(data); err != nil {
		t.Fatal(err)
	}
	randomData, err := service.GenerateCiphertextCiphertextEqualityProof(firstKeypair, secondKeypair.PublicKey, firstCiphertext, secondCiphertext, secondOpening, 42)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyCiphertextCiphertextEqualityProof(randomData); err != nil {
		t.Fatalf("verify random proof: %v", err)
	}
	rustContext, err := hex.DecodeString("8c9240b456a9e6dc65c377a1048d745f94a08cdb7f44cbcd7b46f34048871134c29d170ab8a5b42a3520878501a87a27f9b5653fca8b0c59fc2786cf26e37824f06bddcc8c6fc05e28e513c93c8745048b70923dae7cd67e02376100b909e87410f21a8723d8943e2b37207a4815638fcc0b5efc9dc3346445ca985c6d5c2207422d138fc131895156f91e3e7e11de68e4c370da80e54cbc2bd1b5e3285c3e213e34d08d94993e15383c5e26aa984b3096bdd41012b951fdaf74796b71fb8930")
	if err != nil {
		t.Fatal(err)
	}
	rustProof, err := hex.DecodeString("0c9022fbdb3f718a13d10f169e72dbc4b646f8341cc634f5f888517f7a36ec4bc4c88745cc0abdca52ec5425bf84125d849a9b87783f8c8b0912dcb5f92cab3ea4d4516715408ec537223c0631e312f14b0e860bff4e05300317309e2c5ec018143f2682b709ce56756559998653b3ade40cc295d04f2dc748868a6d7bd8a44917dc9e5ff27ffb748afe982dacce8711df1e86378b291fefa5fb950a6bea050f64d19c86bfe475c60fde60a754d74fb29510011cd7d01c3b39499bbc8f75f806a1fc426a52f39fe85a851e58e2212b045b9a9e15b8cf9bab3deaed3417941d0b")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyCiphertextCiphertextEqualityProof(ZKProofData{Discriminator: 2, Context: rustContext, Proof: rustProof}); err != nil {
		t.Fatalf("verify Rust proof: %v", err)
	}
	tampered := ZKProofData{Discriminator: data.Discriminator, Context: append([]byte(nil), data.Context...), Proof: append([]byte(nil), data.Proof...)}
	tampered.Proof[223] ^= 1
	if err := service.VerifyCiphertextCiphertextEqualityProof(tampered); err == nil {
		t.Fatal("expected tampered ciphertext equality proof error")
	}
	if _, err := service.generateCiphertextCiphertextEqualityProofWithReader(firstKeypair, secondKeypair.PublicKey, firstCiphertext, secondCiphertext, secondOpening, 43, bytes.NewReader(random)); err == nil {
		t.Fatal("expected first ciphertext consistency error")
	}
	wrongSecondCiphertext, err := service.EncryptElGamalWithOpening(secondKeypair.PublicKey, 43, secondOpening)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.generateCiphertextCiphertextEqualityProofWithReader(firstKeypair, secondKeypair.PublicKey, firstCiphertext, wrongSecondCiphertext, secondOpening, 42, bytes.NewReader(random)); err == nil {
		t.Fatal("expected second ciphertext consistency error")
	}
	if _, err := service.generateCiphertextCiphertextEqualityProofWithReader(firstKeypair, secondKeypair.PublicKey, firstCiphertext, secondCiphertext, secondOpening, 42, bytes.NewReader(random[:128])); err == nil {
		t.Fatal("expected random source error")
	}
}

func TestCiphertextCommitmentEqualityProofRustInteroperability(t *testing.T) {
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
	ciphertextOpeningBytes := make([]byte, 32)
	ciphertextOpeningBytes[0] = 2
	ciphertextOpening, err := service.PedersenOpeningFromBytes(ciphertextOpeningBytes)
	if err != nil {
		t.Fatal(err)
	}
	commitmentOpeningBytes := make([]byte, 32)
	commitmentOpeningBytes[0] = 7
	commitmentOpening, err := service.PedersenOpeningFromBytes(commitmentOpeningBytes)
	if err != nil {
		t.Fatal(err)
	}
	ciphertext, err := service.EncryptElGamalWithOpening(keypair.PublicKey, 42, ciphertextOpening)
	if err != nil {
		t.Fatal(err)
	}
	commitment, err := service.CommitPedersen(42, commitmentOpening)
	if err != nil {
		t.Fatal(err)
	}
	random := make([]byte, 192)
	random[0] = 8
	random[64] = 9
	random[128] = 10
	data, err := service.generateCiphertextCommitmentEqualityProofWithReader(keypair, ciphertext, commitment, commitmentOpening, 42, bytes.NewReader(random))
	if err != nil {
		t.Fatal(err)
	}
	if data.Discriminator != 3 || len(data.Context) != 128 || len(data.Proof) != 192 {
		t.Fatalf("proof data sizes = %d, %d, %d", data.Discriminator, len(data.Context), len(data.Proof))
	}
	if got, want := hex.EncodeToString(data.Context), "8c9240b456a9e6dc65c377a1048d745f94a08cdb7f44cbcd7b46f34048871134f06bddcc8c6fc05e28e513c93c8745048b70923dae7cd67e02376100b909e87410f21a8723d8943e2b37207a4815638fcc0b5efc9dc3346445ca985c6d5c2207a69ed12fb9c42f06a8c6ff8b535a781b613f46c7944d013c078eb0b5f3745c44"; got != want {
		t.Fatalf("context = %s, want %s", got, want)
	}
	if got, want := hex.EncodeToString(data.Proof), "0c9022fbdb3f718a13d10f169e72dbc4b646f8341cc634f5f888517f7a36ec4bc4c88745cc0abdca52ec5425bf84125d849a9b87783f8c8b0912dcb5f92cab3ea4d4516715408ec537223c0631e312f14b0e860bff4e05300317309e2c5ec01869f36e4b44c934940217a5a8a13b5fb836856df27c7b8a7309a596584a691905ea2ab8a7ddfcb9d789cf826437184d30f9daf7c57f42b8f48c13b78832462b06d7ff1c56a9ba4c5d65679456aeaddce07ea4fea06a60c92842831e6c08e1b103"; got != want {
		t.Fatalf("proof = %s, want %s", got, want)
	}
	if err := service.VerifyCiphertextCommitmentEqualityProof(data); err != nil {
		t.Fatal(err)
	}
	randomData, err := service.GenerateCiphertextCommitmentEqualityProof(keypair, ciphertext, commitment, commitmentOpening, 42)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyCiphertextCommitmentEqualityProof(randomData); err != nil {
		t.Fatalf("verify random proof: %v", err)
	}
	rustContext, err := hex.DecodeString("8c9240b456a9e6dc65c377a1048d745f94a08cdb7f44cbcd7b46f34048871134f06bddcc8c6fc05e28e513c93c8745048b70923dae7cd67e02376100b909e87410f21a8723d8943e2b37207a4815638fcc0b5efc9dc3346445ca985c6d5c2207a69ed12fb9c42f06a8c6ff8b535a781b613f46c7944d013c078eb0b5f3745c44")
	if err != nil {
		t.Fatal(err)
	}
	rustProof, err := hex.DecodeString("0c9022fbdb3f718a13d10f169e72dbc4b646f8341cc634f5f888517f7a36ec4bc4c88745cc0abdca52ec5425bf84125d849a9b87783f8c8b0912dcb5f92cab3ea4d4516715408ec537223c0631e312f14b0e860bff4e05300317309e2c5ec01869f36e4b44c934940217a5a8a13b5fb836856df27c7b8a7309a596584a691905ea2ab8a7ddfcb9d789cf826437184d30f9daf7c57f42b8f48c13b78832462b06d7ff1c56a9ba4c5d65679456aeaddce07ea4fea06a60c92842831e6c08e1b103")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyCiphertextCommitmentEqualityProof(ZKProofData{Discriminator: 3, Context: rustContext, Proof: rustProof}); err != nil {
		t.Fatalf("verify Rust proof: %v", err)
	}
	tampered := ZKProofData{Discriminator: data.Discriminator, Context: append([]byte(nil), data.Context...), Proof: append([]byte(nil), data.Proof...)}
	tampered.Context[127] ^= 1
	if err := service.VerifyCiphertextCommitmentEqualityProof(tampered); err == nil {
		t.Fatal("expected tampered ciphertext commitment equality proof error")
	}
	wrongCommitment, err := service.CommitPedersen(43, commitmentOpening)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.generateCiphertextCommitmentEqualityProofWithReader(keypair, ciphertext, wrongCommitment, commitmentOpening, 42, bytes.NewReader(random)); err == nil {
		t.Fatal("expected commitment consistency error")
	}
	if _, err := service.generateCiphertextCommitmentEqualityProofWithReader(keypair, ciphertext, commitment, commitmentOpening, 43, bytes.NewReader(random)); err == nil {
		t.Fatal("expected ciphertext consistency error")
	}
}

func TestConfidentialEqualityProofValidation(t *testing.T) {
	service := ConfidentialTransferService{}
	if err := service.VerifyCiphertextCiphertextEqualityProof(ZKProofData{Discriminator: 2}); err == nil {
		t.Fatal("expected invalid ciphertext equality proof data error")
	}
	if err := service.VerifyCiphertextCommitmentEqualityProof(ZKProofData{Discriminator: 3}); err == nil {
		t.Fatal("expected invalid ciphertext commitment equality proof data error")
	}
	identityCiphertextData := ZKProofData{Discriminator: 2, Context: make([]byte, 192), Proof: make([]byte, 224)}
	if err := service.VerifyCiphertextCiphertextEqualityProof(identityCiphertextData); err == nil {
		t.Fatal("expected identity ciphertext equality context error")
	}
	identityCommitmentData := ZKProofData{Discriminator: 3, Context: make([]byte, 128), Proof: make([]byte, 192)}
	if err := service.VerifyCiphertextCommitmentEqualityProof(identityCommitmentData); err == nil {
		t.Fatal("expected identity ciphertext commitment equality context error")
	}
}
