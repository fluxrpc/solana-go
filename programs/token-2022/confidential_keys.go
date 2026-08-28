package token2022

import (
	"crypto/rand"
	"crypto/sha3"
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	"github.com/oasisprotocol/curve25519-voi/curve"
	"github.com/oasisprotocol/curve25519-voi/curve/scalar"
)

type AEKey [16]byte

type ElGamalSecretKey [32]byte

type ConfidentialTransferSigner interface {
	Sign(message []byte) (solana.Signature, error)
}

func (ConfidentialTransferService) GenerateAEKey() (AEKey, error) {
	key := AEKey{}
	if _, err := rand.Read(key[:]); err != nil {
		return AEKey{}, fmt.Errorf("generate AE key: %w", err)
	}
	return key, nil
}

func (ConfidentialTransferService) DeriveAEKeyFromSeedLegacy(seed []byte) (AEKey, error) {
	if len(seed) < 16 {
		return AEKey{}, fmt.Errorf("derive AE key: seed must contain at least 16 bytes")
	}
	if len(seed) > 65535 {
		return AEKey{}, fmt.Errorf("derive AE key: seed exceeds 65535 bytes")
	}
	hash := sha3.Sum512(seed)
	key := AEKey{}
	copy(key[:], hash[:len(key)])
	return key, nil
}

func (service ConfidentialTransferService) DeriveAEKeyFromSignatureLegacy(signature solana.Signature) (AEKey, error) {
	if signature == (solana.Signature{}) {
		return AEKey{}, fmt.Errorf("derive legacy AE key: default signature")
	}
	seed := sha3.Sum512(signature[:])
	return service.DeriveAEKeyFromSeedLegacy(seed[:])
}

func (service ConfidentialTransferService) DeriveAEKeyFromSignerLegacy(signer ConfidentialTransferSigner, publicSeed []byte) (AEKey, error) {
	message := make([]byte, 0, len("AeKey")+len(publicSeed))
	message = append(message, "AeKey"...)
	message = append(message, publicSeed...)
	signature, err := signer.Sign(message)
	if err != nil {
		return AEKey{}, fmt.Errorf("derive AE key: sign public seed: %w", err)
	}
	if signature == (solana.Signature{}) {
		return AEKey{}, fmt.Errorf("derive AE key: default signature")
	}
	return service.DeriveAEKeyFromSignatureLegacy(signature)
}

func (service ConfidentialTransferService) DeriveAEKeyFromMnemonicLegacy(mnemonic, passphrase string) (AEKey, error) {
	return service.DeriveAEKeyFromSeedLegacy(solana.MnemonicToSeed(mnemonic, passphrase))
}

func (ConfidentialTransferService) DeriveElGamalSecretKeyFromSeedLegacy(seed []byte) (ElGamalSecretKey, error) {
	if len(seed) < 32 {
		return ElGamalSecretKey{}, fmt.Errorf("derive ElGamal secret key: seed must contain at least 32 bytes")
	}
	if len(seed) > 65535 {
		return ElGamalSecretKey{}, fmt.Errorf("derive ElGamal secret key: seed exceeds 65535 bytes")
	}
	hash := sha3.Sum512(seed)
	value, err := scalar.NewFromBytesModOrderWide(hash[:])
	if err != nil {
		return ElGamalSecretKey{}, fmt.Errorf("derive ElGamal secret key: %w", err)
	}
	key := ElGamalSecretKey{}
	if err := value.ToBytes(key[:]); err != nil {
		return ElGamalSecretKey{}, fmt.Errorf("derive ElGamal secret key: %w", err)
	}
	return key, nil
}

func (service ConfidentialTransferService) DeriveElGamalSecretKeyFromSignatureLegacy(signature solana.Signature) (ElGamalSecretKey, error) {
	if signature == (solana.Signature{}) {
		return ElGamalSecretKey{}, fmt.Errorf("derive legacy ElGamal secret key: default signature")
	}
	seed := sha3.Sum512(signature[:])
	return service.DeriveElGamalSecretKeyFromSeedLegacy(seed[:])
}

func (service ConfidentialTransferService) DeriveElGamalSecretKeyFromSignerLegacy(signer ConfidentialTransferSigner, publicSeed []byte) (ElGamalSecretKey, error) {
	message := make([]byte, 0, len("ElGamalSecretKey")+len(publicSeed))
	message = append(message, "ElGamalSecretKey"...)
	message = append(message, publicSeed...)
	signature, err := signer.Sign(message)
	if err != nil {
		return ElGamalSecretKey{}, fmt.Errorf("derive ElGamal secret key: sign public seed: %w", err)
	}
	if signature == (solana.Signature{}) {
		return ElGamalSecretKey{}, fmt.Errorf("derive ElGamal secret key: default signature")
	}
	return service.DeriveElGamalSecretKeyFromSignatureLegacy(signature)
}

func (service ConfidentialTransferService) DeriveElGamalSecretKeyFromMnemonicLegacy(mnemonic, passphrase string) (ElGamalSecretKey, error) {
	return service.DeriveElGamalSecretKeyFromSeedLegacy(solana.MnemonicToSeed(mnemonic, passphrase))
}

func (ConfidentialTransferService) DeriveElGamalPubkey(secret ElGamalSecretKey) (ElGamalPubkey, error) {
	secretScalar, err := scalar.NewFromCanonicalBytes(secret[:])
	if err != nil {
		return ElGamalPubkey{}, fmt.Errorf("derive ElGamal public key: %w", err)
	}
	if secretScalar.Equal(scalar.New()) == 1 {
		return ElGamalPubkey{}, fmt.Errorf("derive ElGamal public key: zero secret key")
	}
	hash := sha3.Sum512(curve.RISTRETTO_BASEPOINT_COMPRESSED[:])
	basepoint, err := curve.NewRistrettoPoint().SetUniformBytes(hash[:])
	if err != nil {
		return ElGamalPubkey{}, fmt.Errorf("derive ElGamal public key: %w", err)
	}
	point := curve.NewRistrettoPoint().Mul(basepoint, scalar.New().Invert(secretScalar))
	compressed := curve.NewCompressedRistretto().SetRistrettoPoint(point)
	publicKey := ElGamalPubkey{}
	copy(publicKey[:], compressed[:])
	return publicKey, nil
}
