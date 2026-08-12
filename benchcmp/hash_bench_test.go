package benchcmp

import (
	"testing"

	flux "github.com/fluxrpc/solana-go"
	gagl "github.com/gagliardetto/solana-go"
)

// Shared 32-byte test hash: bytes 1..32 (same as the public key fixture).
var (
	fluxHash = flux.Hash(pubKeyBytes)
	gaglHash = gagl.Hash(pubKeyBytes)

	hashBase58 = gagl.Hash(pubKeyBytes).String()
)

var (
	sinkFluxHash flux.Hash
	sinkGaglHash gagl.Hash
)

func BenchmarkHash_String(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkString = fluxHash.String()
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkString = gaglHash.String()
		}
	})
}

func BenchmarkHash_FromBase58(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkFluxHash, sinkErr = flux.HashFromBase58(hashBase58)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkGaglHash, sinkErr = gagl.HashFromBase58(hashBase58)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
}
