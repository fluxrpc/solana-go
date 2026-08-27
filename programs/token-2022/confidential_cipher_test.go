package token2022

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestConfidentialTransferAECipherRFC8452Vector(t *testing.T) {
	key := AEKey{1}
	nonce := [12]byte{3}
	ciphertext, err := (ConfidentialTransferService{}).EncryptAEAmountWithNonce(key, 1, nonce)
	if err != nil {
		t.Fatal(err)
	}
	want, err := hex.DecodeString("030000000000000000000000b5d839330ac7b786578782fff6013b815b287c22493a364c")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(ciphertext[:], want) {
		t.Fatalf("ciphertext = %x, want %x", ciphertext, want)
	}
	amount, err := (ConfidentialTransferService{}).DecryptAEAmount(key, ciphertext)
	if err != nil || amount != 1 {
		t.Fatalf("amount = %d, %v", amount, err)
	}
}

func TestConfidentialTransferAECipherAuthenticates(t *testing.T) {
	service := ConfidentialTransferService{}
	key := AEKey{1, 2, 3, 4}
	ciphertext, err := service.EncryptAEAmountWithNonce(key, 55, [12]byte{9, 8, 7})
	if err != nil {
		t.Fatal(err)
	}
	tamperedNonce := ciphertext
	tamperedNonce[0] ^= 1
	if _, err := service.DecryptAEAmount(key, tamperedNonce); err == nil {
		t.Fatal("expected nonce authentication failure")
	}
	tamperedCiphertext := ciphertext
	tamperedCiphertext[12] ^= 1
	if _, err := service.DecryptAEAmount(key, tamperedCiphertext); err == nil {
		t.Fatal("expected ciphertext authentication failure")
	}
	tamperedTag := ciphertext
	tamperedTag[35] ^= 1
	if _, err := service.DecryptAEAmount(key, tamperedTag); err == nil {
		t.Fatal("expected tag authentication failure")
	}
	wrongKey := key
	wrongKey[0] ^= 1
	if _, err := service.DecryptAEAmount(wrongKey, ciphertext); err == nil {
		t.Fatal("expected key authentication failure")
	}
}

func TestConfidentialTransferAECipherRandomNonce(t *testing.T) {
	service := ConfidentialTransferService{}
	key := AEKey{4, 3, 2, 1}
	ciphertext, err := service.EncryptAEAmount(key, 1<<63+7)
	if err != nil {
		t.Fatal(err)
	}
	amount, err := service.DecryptAEAmount(key, ciphertext)
	if err != nil || amount != 1<<63+7 {
		t.Fatalf("amount = %d, %v", amount, err)
	}
}
