package token2022

import (
	"crypto/hkdf"
	"crypto/sha512"
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	"github.com/oasisprotocol/curve25519-voi/curve/scalar"
)

type ConfidentialKeys struct {
	AEKey            AEKey
	ElGamalSecretKey ElGamalSecretKey
	ElGamalPubkey    ElGamalPubkey
}

func (service ConfidentialTransferService) DeriveConfidentialKeysFromIKM(ikm []byte) (ConfidentialKeys, error) {
	if len(ikm) < 32 {
		return ConfidentialKeys{}, fmt.Errorf("derive confidential keys: input must contain at least 32 bytes")
	}
	if len(ikm) > 65535 {
		return ConfidentialKeys{}, fmt.Errorf("derive confidential keys: input exceeds 65535 bytes")
	}
	aeBytes, err := hkdf.Key(sha512.New, ikm, []byte("solana-conf-bal/v1"), "ae", 16)
	if err != nil {
		return ConfidentialKeys{}, fmt.Errorf("derive confidential AE key: %w", err)
	}
	secretBytes, err := hkdf.Key(sha512.New, ikm, []byte("solana-conf-bal/v1"), "elgamal", 64)
	if err != nil {
		return ConfidentialKeys{}, fmt.Errorf("derive confidential ElGamal key: %w", err)
	}
	secretScalar, err := scalar.NewFromBytesModOrderWide(secretBytes)
	if err != nil {
		return ConfidentialKeys{}, fmt.Errorf("derive confidential ElGamal key: %w", err)
	}
	keys := ConfidentialKeys{}
	copy(keys.AEKey[:], aeBytes)
	if err := secretScalar.ToBytes(keys.ElGamalSecretKey[:]); err != nil {
		return ConfidentialKeys{}, fmt.Errorf("derive confidential ElGamal key: %w", err)
	}
	keys.ElGamalPubkey, err = service.DeriveElGamalPubkey(keys.ElGamalSecretKey)
	return keys, err
}

func (service ConfidentialTransferService) DeriveConfidentialKeysFromSignature(signature solana.Signature) (ConfidentialKeys, error) {
	if signature == (solana.Signature{}) {
		return ConfidentialKeys{}, fmt.Errorf("derive confidential keys: default signature")
	}
	return service.DeriveConfidentialKeysFromIKM(signature[:])
}

func (service ConfidentialTransferService) DeriveConfidentialKeys(signer ConfidentialTransferSigner, publicSeed []byte) (ConfidentialKeys, error) {
	message := make([]byte, 0, len("solana-conf-bal/v1")+len(publicSeed))
	message = append(message, "solana-conf-bal/v1"...)
	message = append(message, publicSeed...)
	signature, err := signer.Sign(message)
	if err != nil {
		return ConfidentialKeys{}, fmt.Errorf("derive confidential keys: sign public seed: %w", err)
	}
	return service.DeriveConfidentialKeysFromSignature(signature)
}

func (ConfidentialTransferService) PDAWalletPublicSeed(programID, wallet, mint, tokenAccount solana.PublicKey) [128]byte {
	seed := [128]byte{}
	copy(seed[0:32], programID[:])
	copy(seed[32:64], wallet[:])
	copy(seed[64:96], mint[:])
	copy(seed[96:128], tokenAccount[:])
	return seed
}
