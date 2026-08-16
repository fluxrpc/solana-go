package benchcmp

import (
	"testing"

	flux "github.com/fluxrpc/solana-go"
	gagl "github.com/gagliardetto/solana-go"
)

const testMnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

// TestPrivateKeyFromMnemonic_Parity asserts both implementations derive
// identical keys for the default Solana path, custom paths, and raw
// seed-at-path derivation.
func TestPrivateKeyFromMnemonic_Parity(t *testing.T) {
	for _, passphrase := range []string{"", "TREZOR"} {
		fluxKey, err := flux.PrivateKeyFromMnemonic(testMnemonic, passphrase)
		if err != nil {
			t.Fatal(err)
		}
		gaglKey, err := gagl.PrivateKeyFromMnemonic(testMnemonic, passphrase)
		if err != nil {
			t.Fatal(err)
		}
		if fluxKey.String() != gaglKey.String() {
			t.Fatalf("passphrase %q: flux %s != gagl %s", passphrase, fluxKey, gaglKey)
		}
		t.Logf("passphrase %q -> pubkey %s", passphrase, fluxKey.PublicKey())
	}

	for _, path := range []string{"m", "m/0'", "m/44'/501'/0'/0'", "m/44'/501'/7'/0'", "m/44h/501h/0h"} {
		fluxKey, err := flux.PrivateKeyFromMnemonicAtPath(testMnemonic, "", path)
		if err != nil {
			t.Fatal(err)
		}
		gaglKey, err := gagl.PrivateKeyFromMnemonicAtPath(testMnemonic, "", path)
		if err != nil {
			t.Fatal(err)
		}
		if fluxKey.String() != gaglKey.String() {
			t.Fatalf("path %q: flux %s != gagl %s", path, fluxKey, gaglKey)
		}
		t.Logf("path %q -> pubkey %s", path, fluxKey.PublicKey())
	}

	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i)
	}
	fluxKey, err := flux.PrivateKeyFromSeedAtPath(seed, "m/1'/2'")
	if err != nil {
		t.Fatal(err)
	}
	gaglKey, err := gagl.PrivateKeyFromSeedAtPath(seed, "m/1'/2'")
	if err != nil {
		t.Fatal(err)
	}
	if fluxKey.String() != gaglKey.String() {
		t.Fatalf("seed-at-path: flux %s != gagl %s", fluxKey, gaglKey)
	}
	t.Logf("seed-at-path m/1'/2' -> pubkey %s", fluxKey.PublicKey())
}

func BenchmarkPrivateKey_FromMnemonic(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkFluxPriv, sinkErr = flux.PrivateKeyFromMnemonic(testMnemonic, "")
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkGaglPriv, sinkErr = gagl.PrivateKeyFromMnemonic(testMnemonic, "")
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
}
