package token2022

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestCurrentConfidentialKeyDerivationVector(t *testing.T) {
	service := ConfidentialTransferService{}
	keys, err := service.DeriveConfidentialKeysFromIKM(bytes.Repeat([]byte{0x42}, 64))
	if err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(keys.AEKey[:]); got != "1b29778a3493dab218c24e87144be43d" {
		t.Fatalf("AE key = %s", got)
	}
	if got := hex.EncodeToString(keys.ElGamalSecretKey[:]); got != "716d2431fa74fc1c48ffb9b9972b0be2f0b29ab755081f2ea35d3c74f442190e" {
		t.Fatalf("ElGamal secret = %s", got)
	}
}

func TestPDAWalletPublicSeed(t *testing.T) {
	service := ConfidentialTransferService{}
	seed := service.PDAWalletPublicSeed(token2022Key(1), token2022Key(2), token2022Key(3), token2022Key(4))
	if seed[0] != 1 || seed[32] != 2 || seed[64] != 3 || seed[96] != 4 {
		t.Fatalf("seed = %x", seed)
	}
}
