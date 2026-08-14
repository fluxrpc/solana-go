package benchcmp

import (
	"testing"

	flux "github.com/fluxrpc/solana-go"
	gagl "github.com/gagliardetto/solana-go"
)

var (
	fluxPdaWallet = flux.MustPublicKeyFromBase58("G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo")
	fluxPdaMint   = flux.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")
	gaglPdaWallet = gagl.MustPublicKeyFromBase58("G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo")
	gaglPdaMint   = gagl.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")

	sinkFluxKey flux.PublicKey
	sinkGaglKey gagl.PublicKey
	sinkBump    uint8
	sinkBool    bool
)

func BenchmarkPda_FindProgramAddress(b *testing.B) {
	fluxSeeds := [][]byte{fluxPdaWallet[:], flux.TokenProgramID[:], fluxPdaMint[:]}
	gaglSeeds := [][]byte{gaglPdaWallet[:], gagl.TokenProgramID[:], gaglPdaMint[:]}

	// Both implementations must agree before we compare their speed.
	fk, fb, err1 := flux.FindProgramAddress(fluxSeeds, flux.SPLAssociatedTokenAccountProgramID)
	gk, gb, err2 := gagl.FindProgramAddress(gaglSeeds, gagl.SPLAssociatedTokenAccountProgramID)
	if err1 != nil || err2 != nil || fk.String() != gk.String() || fb != gb {
		b.Fatalf("parity mismatch: %s/%d vs %s/%d (%v, %v)", fk, fb, gk, gb, err1, err2)
	}

	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var err error
			sinkFluxKey, sinkBump, err = flux.FindProgramAddress(fluxSeeds, flux.SPLAssociatedTokenAccountProgramID)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var err error
			sinkGaglKey, sinkBump, err = gagl.FindProgramAddress(gaglSeeds, gagl.SPLAssociatedTokenAccountProgramID)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkPda_CreateProgramAddress(b *testing.B) {
	fluxSeeds := [][]byte{[]byte("seed-one"), {255}}
	gaglSeeds := [][]byte{[]byte("seed-one"), {255}}

	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var err error
			sinkFluxKey, err = flux.CreateProgramAddress(fluxSeeds, flux.TokenProgramID)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var err error
			sinkGaglKey, err = gagl.CreateProgramAddress(gaglSeeds, gagl.TokenProgramID)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkPda_FindAssociatedTokenAddress(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var err error
			sinkFluxKey, sinkBump, err = flux.FindAssociatedTokenAddress(fluxPdaWallet, fluxPdaMint)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var err error
			sinkGaglKey, sinkBump, err = gagl.FindAssociatedTokenAddress(gaglPdaWallet, gaglPdaMint)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkPda_IsOnCurve(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBool = fluxPdaWallet.IsOnCurve()
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBool = gaglPdaWallet.IsOnCurve()
		}
	})
}
