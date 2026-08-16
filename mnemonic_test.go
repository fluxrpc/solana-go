package solana_go

import (
	"encoding/hex"
	"testing"
)

const testMnemonic12 = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

func TestMnemonicToSeedTrezorVector(t *testing.T) {
	// Official BIP-39 test vector (entropy 0x00*16, passphrase "TREZOR").
	seed := MnemonicToSeed(testMnemonic12, "TREZOR")
	want := "c55257c360c07c72029aebc1b53c05ed0362ada38ead3e3e9efa3708e53495531f09a6987599d18264c1e1c92f2cf141630c7a3c4ab7c81b2f001698e7463b04"
	if got := hex.EncodeToString(seed); got != want {
		t.Fatalf("seed = %s, want %s", got, want)
	}
}

func TestIsMnemonicValid(t *testing.T) {
	if !IsMnemonicValid(testMnemonic12) {
		t.Fatal("valid mnemonic rejected")
	}
	// 24-word vector (entropy 0x00*32).
	long := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art"
	if !IsMnemonicValid(long) {
		t.Fatal("valid 24-word mnemonic rejected")
	}

	invalid := []string{
		"",
		"abandon",
		"abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon", // bad checksum
		"abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon zzzzz",   // unknown word
		"abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about abandon",   // reordered
	}
	for _, m := range invalid {
		if IsMnemonicValid(m) {
			t.Errorf("invalid mnemonic accepted: %q", m)
		}
	}
}

// Ground truth: derived with gagliardetto/solana-go v1.22.0 for identical
// inputs (see benchcmp/mnemonic_bench_test.go for the live parity test).
func TestPrivateKeyFromMnemonicVectors(t *testing.T) {
	cases := []struct {
		passphrase string
		path       string
		wantPubkey string
	}{
		{"", SolanaDerivationPath, "HAgk14JpMQLgt6rVgv7cBQFJWFto5Dqxi472uT3DKpqk"},
		{"TREZOR", SolanaDerivationPath, "7zSmbu6gKkb6HB7UDPtHYjwCWuBHU1D4TpNZFm4sndQe"},
		{"", "m", "GiAiwXD6TQUqsf4pvRLtPPmVma7oGmRusLvo58SBNn1e"},
		{"", "m/0'", "D1Rwdcr6RtByD4izfmxvYfTon6TjaiaCSrt7RiegN5ux"},
		{"", "m/44'/501'/7'/0'", "9h1cLBiraaUqM1CdJTaVaew1oQtgQUW24FZ8YdnLLgJY"},
		{"", "m/44h/501h/0h", "GjJyeC1r2RgkuoCWMyPYkCWSGSGLcz266EaAkLA27AhL"},
	}
	for _, tc := range cases {
		key, err := PrivateKeyFromMnemonicAtPath(testMnemonic12, tc.passphrase, tc.path)
		if err != nil {
			t.Fatalf("path %q: %v", tc.path, err)
		}
		if err := key.Validate(); err != nil {
			t.Fatalf("path %q: %v", tc.path, err)
		}
		if got := key.PublicKey().String(); got != tc.wantPubkey {
			t.Errorf("path %q passphrase %q: pubkey %s, want %s", tc.path, tc.passphrase, got, tc.wantPubkey)
		}
	}

	// Default-path helper matches the explicit path.
	key, err := PrivateKeyFromMnemonic(testMnemonic12, "")
	if err != nil {
		t.Fatal(err)
	}
	if key.PublicKey().String() != cases[0].wantPubkey {
		t.Fatal("PrivateKeyFromMnemonic does not use the default Solana path")
	}
}

func TestPrivateKeyFromSeedAtPath(t *testing.T) {
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i)
	}
	key, err := PrivateKeyFromSeedAtPath(seed, "m/1'/2'")
	if err != nil {
		t.Fatal(err)
	}
	// Ground truth from upstream (see benchcmp parity test).
	if got := key.PublicKey().String(); got != "2WEU9zdmRFhkRsgD6dm9izCcJfPtDdzjqGbrGXMqJhTF" {
		t.Fatalf("pubkey = %s", got)
	}

	if _, err := PrivateKeyFromSeedAtPath(make([]byte, 8), "m"); err == nil {
		t.Fatal("expected error for short seed")
	}
	if _, err := PrivateKeyFromSeedAtPath(seed, "m/0"); err == nil {
		t.Fatal("expected error for non-hardened segment")
	}
	if _, err := PrivateKeyFromSeedAtPath(seed, "m//0'"); err == nil {
		t.Fatal("expected error for empty segment")
	}
	if _, err := PrivateKeyFromSeedAtPath(seed, "m/4294967296'"); err == nil {
		t.Fatal("expected error for out-of-range index")
	}
	if _, err := PrivateKeyFromSeedAtPath(seed, "m/abc'"); err == nil {
		t.Fatal("expected error for non-numeric segment")
	}
}

func TestPrivateKeyFromMnemonicRejectsInvalid(t *testing.T) {
	if _, err := PrivateKeyFromMnemonic("not a mnemonic", ""); err == nil {
		t.Fatal("expected error for invalid mnemonic")
	}
}
