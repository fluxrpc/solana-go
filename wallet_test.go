package solana_go

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewWallet(t *testing.T) {
	w := NewWallet()
	if err := w.PrivateKey.Validate(); err != nil {
		t.Fatal(err)
	}
	if w.PublicKey() != w.PrivateKey.PublicKey() {
		t.Fatal("PublicKey mismatch")
	}
	if NewWallet().PublicKey() == w.PublicKey() {
		t.Fatal("two random wallets share a key")
	}
	if w.PrivateKeyFor(w.PublicKey()) != &w.PrivateKey || w.PrivateKeyFor(PublicKey{}) != nil {
		t.Fatal("PrivateKeyFor mismatch")
	}
}

func TestWalletFromPrivateKeyBase58(t *testing.T) {
	w := NewWallet()
	restored, err := WalletFromPrivateKeyBase58(w.PrivateKey.String())
	if err != nil {
		t.Fatal(err)
	}
	if restored.PublicKey() != w.PublicKey() {
		t.Fatal("restored wallet mismatch")
	}
	if _, err := WalletFromPrivateKeyBase58("nope"); err == nil {
		t.Fatal("expected error for invalid base58")
	}
}

func TestWalletFromSolanaKeygenFile(t *testing.T) {
	key := testSeedKey(t)
	path := filepath.Join(t.TempDir(), "id.json")
	if err := os.WriteFile(path, keygenFileContent(key), 0o600); err != nil {
		t.Fatal(err)
	}
	w, err := WalletFromSolanaKeygenFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if w.PublicKey() != key.PublicKey() {
		t.Fatal("keygen wallet mismatch")
	}
	if _, err := WalletFromSolanaKeygenFile(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestNewWalletFromMnemonic(t *testing.T) {
	w, err := NewWalletFromMnemonic(testMnemonic12, "")
	if err != nil {
		t.Fatal(err)
	}
	// Ground truth shared with TestPrivateKeyFromMnemonicVectors.
	if got := w.PublicKey().String(); got != "HAgk14JpMQLgt6rVgv7cBQFJWFto5Dqxi472uT3DKpqk" {
		t.Fatalf("pubkey = %s", got)
	}
	if _, err := NewWalletFromMnemonic("not a mnemonic", ""); err == nil {
		t.Fatal("expected error for invalid mnemonic")
	}
}
