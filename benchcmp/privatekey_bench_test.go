package benchcmp

import (
	"crypto/ed25519"
	"testing"

	flux "github.com/fluxrpc/solana-go"
	gagl "github.com/gagliardetto/solana-go"
)

// Deterministic ed25519 key (seed bytes 1..32) shared by both implementations,
// and a 200-byte message to sign.
var (
	privKeyBytes = func() []byte {
		var seed [ed25519.SeedSize]byte
		for i := range seed {
			seed[i] = byte(i + 1)
		}
		return []byte(ed25519.NewKeyFromSeed(seed[:]))
	}()

	fluxPrivKey = flux.PrivateKey(privKeyBytes)
	gaglPrivKey = gagl.PrivateKey(privKeyBytes)

	signPayload = func() []byte {
		b := make([]byte, 200)
		for i := range b {
			b[i] = byte(i)
		}
		return b
	}()
)

func BenchmarkPrivateKey_Sign(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkFluxSig, sinkErr = fluxPrivKey.Sign(signPayload)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkGaglSig, sinkErr = gaglPrivKey.Sign(signPayload)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
}

func BenchmarkPrivateKey_PublicKey(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkFluxPubKey = fluxPrivKey.PublicKey()
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkGaglPubKey = gaglPrivKey.PublicKey()
		}
	})
}
