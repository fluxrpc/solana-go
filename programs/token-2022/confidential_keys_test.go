package token2022

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

type confidentialKDFSeedVector struct {
	Name             string `json:"name"`
	SeedHex          string `json:"seed_hex"`
	AEKeyHex         string `json:"ae_key_hex"`
	ElGamalSecretHex string `json:"elgamal_secret_hex"`
}

type confidentialKDFSignatureVector struct {
	Name             string `json:"name"`
	SignatureHex     string `json:"signature_hex"`
	AEKeyHex         string `json:"ae_key_hex"`
	ElGamalSecretHex string `json:"elgamal_secret_hex"`
}

type confidentialKDFSignerVector struct {
	Name             string `json:"name"`
	KeypairSecretHex string `json:"keypair_secret_hex"`
	PublicSeedHex    string `json:"public_seed_hex"`
	AEKeyHex         string `json:"ae_key_hex"`
	ElGamalSecretHex string `json:"elgamal_secret_hex"`
}

type confidentialKDFMnemonicVector struct {
	Name             string `json:"name"`
	Mnemonic         string `json:"mnemonic"`
	Passphrase       string `json:"passphrase"`
	AEKeyHex         string `json:"ae_key_hex"`
	ElGamalSecretHex string `json:"elgamal_secret_hex"`
}

type confidentialKDFVectors struct {
	FromSeed      []confidentialKDFSeedVector      `json:"from_seed"`
	FromSignature []confidentialKDFSignatureVector `json:"from_signature"`
	FromSigner    []confidentialKDFSignerVector    `json:"from_signer"`
	FromMnemonic  []confidentialKDFMnemonicVector  `json:"from_mnemonic"`
}

type confidentialDefaultSigner struct {
	err error
}

func (signer confidentialDefaultSigner) Sign([]byte) (solana.Signature, error) {
	return solana.Signature{}, signer.err
}

func TestConfidentialTransferKeyDerivationVectors(t *testing.T) {
	data, err := os.ReadFile("testdata/confidential_kdf.json")
	if err != nil {
		t.Fatal(err)
	}
	vectors := confidentialKDFVectors{}
	if err := json.Unmarshal(data, &vectors); err != nil {
		t.Fatal(err)
	}
	service := ConfidentialTransferService{}
	for _, vector := range vectors.FromSeed {
		seed, err := hex.DecodeString(vector.SeedHex)
		if err != nil {
			t.Fatalf("%s: %v", vector.Name, err)
		}
		aeKey, err := service.DeriveAEKeyFromSeedLegacy(seed)
		if err != nil || hex.EncodeToString(aeKey[:]) != vector.AEKeyHex {
			t.Fatalf("%s: AE key = %x, %v", vector.Name, aeKey, err)
		}
		secret, err := service.DeriveElGamalSecretKeyFromSeedLegacy(seed)
		if err != nil || hex.EncodeToString(secret[:]) != vector.ElGamalSecretHex {
			t.Fatalf("%s: ElGamal secret = %x, %v", vector.Name, secret, err)
		}
	}
	for _, vector := range vectors.FromSignature {
		raw, err := hex.DecodeString(vector.SignatureHex)
		if err != nil {
			t.Fatalf("%s: %v", vector.Name, err)
		}
		signature := solana.Signature{}
		copy(signature[:], raw)
		aeKey, err := service.DeriveAEKeyFromSignatureLegacy(signature)
		if err != nil || hex.EncodeToString(aeKey[:]) != vector.AEKeyHex {
			t.Fatalf("%s: AE key = %x, %v", vector.Name, aeKey, err)
		}
		secret, err := service.DeriveElGamalSecretKeyFromSignatureLegacy(signature)
		if err != nil || hex.EncodeToString(secret[:]) != vector.ElGamalSecretHex {
			t.Fatalf("%s: ElGamal secret = %x, %v", vector.Name, secret, err)
		}
	}
	for _, vector := range vectors.FromSigner {
		seed, err := hex.DecodeString(vector.KeypairSecretHex)
		if err != nil {
			t.Fatalf("%s: %v", vector.Name, err)
		}
		privateKey, err := solana.PrivateKeyFromSeed(seed)
		if err != nil {
			t.Fatalf("%s: %v", vector.Name, err)
		}
		publicSeed, err := hex.DecodeString(vector.PublicSeedHex)
		if err != nil {
			t.Fatalf("%s: %v", vector.Name, err)
		}
		aeKey, err := service.DeriveAEKeyFromSignerLegacy(privateKey, publicSeed)
		if err != nil || hex.EncodeToString(aeKey[:]) != vector.AEKeyHex {
			t.Fatalf("%s: AE key = %x, %v", vector.Name, aeKey, err)
		}
		secret, err := service.DeriveElGamalSecretKeyFromSignerLegacy(privateKey, publicSeed)
		if err != nil || hex.EncodeToString(secret[:]) != vector.ElGamalSecretHex {
			t.Fatalf("%s: ElGamal secret = %x, %v", vector.Name, secret, err)
		}
	}
	for _, vector := range vectors.FromMnemonic {
		aeKey, err := service.DeriveAEKeyFromMnemonicLegacy(vector.Mnemonic, vector.Passphrase)
		if err != nil || hex.EncodeToString(aeKey[:]) != vector.AEKeyHex {
			t.Fatalf("%s: AE key = %x, %v", vector.Name, aeKey, err)
		}
		secret, err := service.DeriveElGamalSecretKeyFromMnemonicLegacy(vector.Mnemonic, vector.Passphrase)
		if err != nil || hex.EncodeToString(secret[:]) != vector.ElGamalSecretHex {
			t.Fatalf("%s: ElGamal secret = %x, %v", vector.Name, secret, err)
		}
	}
}

func TestConfidentialTransferKeyDerivationValidation(t *testing.T) {
	service := ConfidentialTransferService{}
	if _, err := service.DeriveAEKeyFromSeedLegacy(make([]byte, 15)); err == nil {
		t.Fatal("expected short AE seed error")
	}
	if _, err := service.DeriveAEKeyFromSeedLegacy(make([]byte, 65536)); err == nil {
		t.Fatal("expected long AE seed error")
	}
	if _, err := service.DeriveElGamalSecretKeyFromSeedLegacy(make([]byte, 31)); err == nil {
		t.Fatal("expected short ElGamal seed error")
	}
	if _, err := service.DeriveElGamalSecretKeyFromSeedLegacy(make([]byte, 65536)); err == nil {
		t.Fatal("expected long ElGamal seed error")
	}
	if _, err := service.DeriveAEKeyFromSignatureLegacy(solana.Signature{}); err == nil {
		t.Fatal("expected default AE signature error")
	}
	if _, err := service.DeriveElGamalSecretKeyFromSignatureLegacy(solana.Signature{}); err == nil {
		t.Fatal("expected default ElGamal signature error")
	}
	if _, err := service.DeriveAEKeyFromSignerLegacy(confidentialDefaultSigner{}, nil); err == nil {
		t.Fatal("expected default signature error")
	}
	if _, err := service.DeriveElGamalSecretKeyFromSignerLegacy(confidentialDefaultSigner{}, nil); err == nil {
		t.Fatal("expected default signature error")
	}
	signerErr := errors.New("signer unavailable")
	if _, err := service.DeriveAEKeyFromSignerLegacy(confidentialDefaultSigner{err: signerErr}, nil); !errors.Is(err, signerErr) {
		t.Fatalf("AE signer error = %v", err)
	}
	if _, err := service.DeriveElGamalSecretKeyFromSignerLegacy(confidentialDefaultSigner{err: signerErr}, nil); !errors.Is(err, signerErr) {
		t.Fatalf("ElGamal signer error = %v", err)
	}
}

func TestConfidentialTransferElGamalPubkeyVector(t *testing.T) {
	secretBytes, err := hex.DecodeString("97e676b07bfa0d85006fc9c443428e90083a83a61116e65ccc2ba760ea66ab04")
	if err != nil {
		t.Fatal(err)
	}
	secret := ElGamalSecretKey{}
	copy(secret[:], secretBytes)
	publicKey, err := (ConfidentialTransferService{}).DeriveElGamalPubkey(secret)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := hex.EncodeToString(publicKey[:]), "662efdf69d5836866356685d5049c9a72beca7a6a3d56900550273d38b26fe03"; got != want {
		t.Fatalf("public key = %s, want %s", got, want)
	}
}
