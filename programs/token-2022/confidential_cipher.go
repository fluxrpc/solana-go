package token2022

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"github.com/ericlagergren/siv"
)

func (service ConfidentialTransferService) EncryptAEAmount(key AEKey, amount uint64) (AeCiphertext, error) {
	nonce := [12]byte{}
	if _, err := rand.Read(nonce[:]); err != nil {
		return AeCiphertext{}, fmt.Errorf("encrypt confidential amount: nonce: %w", err)
	}
	return service.EncryptAEAmountWithNonce(key, amount, nonce)
}

func (ConfidentialTransferService) EncryptAEAmountWithNonce(key AEKey, amount uint64, nonce [12]byte) (AeCiphertext, error) {
	cipher, err := siv.NewGCM(key[:])
	if err != nil {
		return AeCiphertext{}, fmt.Errorf("encrypt confidential amount: %w", err)
	}
	plaintext := [8]byte{}
	binary.LittleEndian.PutUint64(plaintext[:], amount)
	sealed := cipher.Seal(nil, nonce[:], plaintext[:], nil)
	ciphertext := AeCiphertext{}
	copy(ciphertext[:12], nonce[:])
	copy(ciphertext[12:], sealed)
	return ciphertext, nil
}

func (service ConfidentialTransferService) DecryptAEAmount(key AEKey, ciphertext AeCiphertext) (uint64, error) {
	cipher, err := siv.NewGCM(key[:])
	if err != nil {
		return 0, service.decryptionError(fmt.Errorf("decrypt confidential amount: %w", err))
	}
	plaintext, err := cipher.Open(nil, ciphertext[:12], ciphertext[12:], nil)
	if err != nil {
		return 0, service.decryptionError(fmt.Errorf("decrypt confidential amount: authenticate: %w", err))
	}
	return binary.LittleEndian.Uint64(plaintext), nil
}
