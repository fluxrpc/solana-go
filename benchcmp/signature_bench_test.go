package benchcmp

import (
	"testing"

	flux "github.com/fluxrpc/solana-go"
	gagl "github.com/gagliardetto/solana-go"
)

// Shared 64-byte test signature: bytes 1..64.
var sigBytes = func() [64]byte {
	var b [64]byte
	for i := range b {
		b[i] = byte(i + 1)
	}
	return b
}()

var (
	fluxSig = flux.Signature(sigBytes)
	gaglSig = gagl.Signature(sigBytes)

	sigBase58 = gagl.Signature(sigBytes).String()
	sigJSON   = []byte(`"` + sigBase58 + `"`)
)

var (
	sinkFluxSig flux.Signature
	sinkGaglSig gagl.Signature
)

func BenchmarkSignature_String(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkString = fluxSig.String()
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkString = gaglSig.String()
		}
	})
}

func BenchmarkSignature_FromBase58(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkFluxSig, sinkErr = flux.SignatureFromBase58(sigBase58)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkGaglSig, sinkErr = gagl.SignatureFromBase58(sigBase58)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
}

func BenchmarkSignature_MarshalJSON(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBytes, sinkErr = fluxSig.MarshalJSON()
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBytes, sinkErr = gaglSig.MarshalJSON()
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
}

func BenchmarkSignature_UnmarshalJSON(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		var s flux.Signature
		for b.Loop() {
			sinkErr = s.UnmarshalJSON(sigJSON)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
		sinkFluxSig = s
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		var s gagl.Signature
		for b.Loop() {
			sinkErr = s.UnmarshalJSON(sigJSON)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
		sinkGaglSig = s
	})
}

// A real signed message pair for Verify: sign with a fixed keypair via each
// implementation's own types.
func BenchmarkSignature_Verify(b *testing.B) {
	msg := make([]byte, 200)
	for i := range msg {
		msg[i] = byte(i)
	}

	fluxKey, err := flux.NewRandomPrivateKey()
	if err != nil {
		b.Fatal(err)
	}
	fluxSigned, err := fluxKey.Sign(msg)
	if err != nil {
		b.Fatal(err)
	}
	fluxPub := fluxKey.PublicKey()

	gaglKey := gagl.PrivateKey(fluxKey)
	gaglSigned, err := gaglKey.Sign(msg)
	if err != nil {
		b.Fatal(err)
	}
	gaglPub := gaglKey.PublicKey()

	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if !fluxSigned.Verify(fluxPub, msg) {
				b.Fatal("verify failed")
			}
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if !gaglSigned.Verify(gaglPub, msg) {
				b.Fatal("verify failed")
			}
		}
	})
}
