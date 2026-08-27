package token2022

import (
	"strings"
	"testing"
)

func TestConfidentialVerifyZKProofData(t *testing.T) {
	service := ConfidentialTransferService{}
	secret := ElGamalSecretKey{1}
	keypair, err := service.ElGamalKeypairFromSecret(secret)
	if err != nil {
		t.Fatal(err)
	}
	proof, err := service.GeneratePubkeyValidityProof(keypair)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyZKProofData(proof); err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyZKProofData(ZKProofData{}); err == nil || !strings.Contains(err.Error(), "unsupported discriminator 0") {
		t.Fatalf("unsupported proof error = %v", err)
	}
}
