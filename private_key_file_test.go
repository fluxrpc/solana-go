package solana_go

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func keygenFileContent(k PrivateKey) []byte {
	parts := make([]string, len(k))
	for i, b := range k {
		parts[i] = fmt.Sprint(b)
	}
	return []byte("[" + strings.Join(parts, ",") + "]")
}

func testSeedKey(t *testing.T) PrivateKey {
	t.Helper()
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i + 1)
	}
	key, err := PrivateKeyFromSeed(seed)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func TestPrivateKeyFromSeed(t *testing.T) {
	key := testSeedKey(t)
	if err := key.Validate(); err != nil {
		t.Fatal(err)
	}
	// Deterministic: same seed, same key.
	again, err := PrivateKeyFromSeed(key[:32])
	if err != nil {
		t.Fatal(err)
	}
	if key.String() != again.String() {
		t.Fatal("PrivateKeyFromSeed is not deterministic")
	}
	if _, err := PrivateKeyFromSeed(make([]byte, 31)); err == nil {
		t.Fatal("expected error for short seed")
	}
}

func TestPrivateKeyFromSolanaKeygenFileBytes(t *testing.T) {
	key := testSeedKey(t)
	parsed, err := PrivateKeyFromSolanaKeygenFileBytes(keygenFileContent(key))
	if err != nil {
		t.Fatal(err)
	}
	if parsed.String() != key.String() {
		t.Fatal("keygen file round trip mismatch")
	}

	// Whitespace-tolerant, as produced by various tools.
	spaced := []byte("[ " + strings.TrimSuffix(strings.TrimPrefix(string(keygenFileContent(key)), "["), "]") + " ]\n")
	spaced = []byte(strings.ReplaceAll(string(spaced), ",", ", "))
	parsed, err = PrivateKeyFromSolanaKeygenFileBytes(spaced)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.String() != key.String() {
		t.Fatal("whitespace keygen file round trip mismatch")
	}
}

func TestPrivateKeyFromSolanaKeygenFileBytesErrors(t *testing.T) {
	key := testSeedKey(t)
	good := string(keygenFileContent(key))

	cases := map[string]string{
		"not an array":     `"abc"`,
		"empty":            ``,
		"empty array":      `[]`,
		"short array":      `[1,2,3]`,
		"value too big":    `[300` + good[2:],
		"missing value":    `[,` + good[1:],
		"unexpected char":  `[1;2]`,
		"unterminated":     good[:len(good)-1],
		"trailing garbage": good + "x",
	}
	for name, content := range cases {
		if _, err := PrivateKeyFromSolanaKeygenFileBytes([]byte(content)); err == nil {
			t.Errorf("%s: expected error for %q", name, content)
		}
	}

	// Corrupted public-key half fails validation.
	corrupt := make(PrivateKey, len(key))
	copy(corrupt, key)
	corrupt[63] ^= 0xFF
	if _, err := PrivateKeyFromSolanaKeygenFileBytes(keygenFileContent(corrupt)); err == nil {
		t.Error("expected validation error for corrupted public key half")
	}
}

func TestPrivateKeyFromSolanaKeygenFile(t *testing.T) {
	key := testSeedKey(t)
	path := filepath.Join(t.TempDir(), "id.json")
	if err := os.WriteFile(path, keygenFileContent(key), 0o600); err != nil {
		t.Fatal(err)
	}
	parsed, err := PrivateKeyFromSolanaKeygenFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.String() != key.String() {
		t.Fatal("keygen file mismatch")
	}
	if _, err := PrivateKeyFromSolanaKeygenFile(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestPrivateKeyValidate(t *testing.T) {
	key := testSeedKey(t)
	if err := key.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := PrivateKey(key[:40]).Validate(); err == nil {
		t.Fatal("expected size error")
	}
	corrupt := make(PrivateKey, len(key))
	copy(corrupt, key)
	corrupt[40] ^= 0x01
	if err := corrupt.Validate(); err == nil {
		t.Fatal("expected mismatch error")
	}
}
