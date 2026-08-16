package benchcmp

import (
	"strconv"
	"testing"

	flux "github.com/fluxrpc/solana-go"
	gagl "github.com/gagliardetto/solana-go"
)

// Deterministic solana-keygen file content shared by both sides.
var keygenFileContent = func() []byte {
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i + 1)
	}
	key, err := flux.PrivateKeyFromSeed(seed)
	if err != nil {
		panic(err)
	}
	out := []byte{'['}
	for i, b := range key {
		if i > 0 {
			out = append(out, ',')
		}
		out = strconv.AppendUint(out, uint64(b), 10)
	}
	return append(out, ']')
}()

var (
	sinkFluxPriv flux.PrivateKey
	sinkGaglPriv gagl.PrivateKey
)

// TestPrivateKeyFromKeygenFile_Parity asserts both implementations parse the
// same keygen file to the same key.
func TestPrivateKeyFromKeygenFile_Parity(t *testing.T) {
	fluxKey, err := flux.PrivateKeyFromSolanaKeygenFileBytes(keygenFileContent)
	if err != nil {
		t.Fatal(err)
	}
	gaglKey, err := gagl.PrivateKeyFromSolanaKeygenFileBytes(keygenFileContent)
	if err != nil {
		t.Fatal(err)
	}
	if fluxKey.String() != gaglKey.String() {
		t.Fatalf("key mismatch: flux %s gagl %s", fluxKey, gaglKey)
	}
}

func BenchmarkPrivateKey_FromKeygenFile(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkFluxPriv, sinkErr = flux.PrivateKeyFromSolanaKeygenFileBytes(keygenFileContent)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkGaglPriv, sinkErr = gagl.PrivateKeyFromSolanaKeygenFileBytes(keygenFileContent)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
}
