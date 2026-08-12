package benchcmp

import (
	"testing"

	flux "github.com/fluxrpc/solana-go"
	gagl "github.com/gagliardetto/solana-go"
)

// Shared 96-byte payload: bytes 1..96.
var base58Payload = func() []byte {
	b := make([]byte, 96)
	for i := range b {
		b[i] = byte(i + 1)
	}
	return b
}()

var (
	fluxBase58Data = flux.Base58(base58Payload)
	gaglBase58Data = gagl.Base58(base58Payload)

	base58DataJSON = func() []byte {
		out, err := gagl.Base58(base58Payload).MarshalJSON()
		if err != nil {
			panic(err)
		}
		return out
	}()
)

var (
	sinkFluxBase58 flux.Base58
	sinkGaglBase58 gagl.Base58
)

func BenchmarkBase58Data_MarshalJSON(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBytes, sinkErr = fluxBase58Data.MarshalJSON()
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBytes, sinkErr = gaglBase58Data.MarshalJSON()
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
}

func BenchmarkBase58Data_UnmarshalJSON(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		var d flux.Base58
		for b.Loop() {
			sinkErr = d.UnmarshalJSON(base58DataJSON)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
		sinkFluxBase58 = d
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		var d gagl.Base58
		for b.Loop() {
			sinkErr = d.UnmarshalJSON(base58DataJSON)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
		sinkGaglBase58 = d
	})
}
