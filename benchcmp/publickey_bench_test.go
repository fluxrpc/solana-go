package benchcmp

import (
	"testing"

	flux "github.com/fluxrpc/solana-go"
	gagl "github.com/gagliardetto/solana-go"
)

// Shared 32-byte test key: bytes 1..32.
var pubKeyBytes = func() [32]byte {
	var b [32]byte
	for i := range b {
		b[i] = byte(i + 1)
	}
	return b
}()

var (
	fluxPubKey = flux.PublicKey(pubKeyBytes)
	gaglPubKey = gagl.PublicKey(pubKeyBytes)

	pubKeyBase58 = gagl.PublicKey(pubKeyBytes).String()
	pubKeyJSON   = []byte(`"` + pubKeyBase58 + `"`)
)

// Package-level sinks to prevent dead-code elimination.
var (
	sinkString string
	sinkBytes  []byte
	sinkErr    error

	sinkFluxPubKey flux.PublicKey
	sinkGaglPubKey gagl.PublicKey
)

func BenchmarkPublicKey_String(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkString = fluxPubKey.String()
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkString = gaglPubKey.String()
		}
	})
}

func BenchmarkPublicKey_FromBase58(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkFluxPubKey, sinkErr = flux.PublicKeyFromBase58(pubKeyBase58)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkGaglPubKey, sinkErr = gagl.PublicKeyFromBase58(pubKeyBase58)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
}

func BenchmarkPublicKey_MarshalJSON(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBytes, sinkErr = fluxPubKey.MarshalJSON()
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBytes, sinkErr = gaglPubKey.MarshalJSON()
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
}

func BenchmarkPublicKey_UnmarshalJSON(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		var k flux.PublicKey
		for b.Loop() {
			sinkErr = k.UnmarshalJSON(pubKeyJSON)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
		sinkFluxPubKey = k
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		var k gagl.PublicKey
		for b.Loop() {
			sinkErr = k.UnmarshalJSON(pubKeyJSON)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
		sinkGaglPubKey = k
	})
}

func BenchmarkPublicKey_MarshalText(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBytes, sinkErr = fluxPubKey.MarshalText()
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBytes, sinkErr = gaglPubKey.MarshalText()
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
}

func BenchmarkPublicKey_UnmarshalText(b *testing.B) {
	text := []byte(pubKeyBase58)
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		var k flux.PublicKey
		for b.Loop() {
			sinkErr = k.UnmarshalText(text)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
		sinkFluxPubKey = k
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		var k gagl.PublicKey
		for b.Loop() {
			sinkErr = k.UnmarshalText(text)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
		sinkGaglPubKey = k
	})
}
