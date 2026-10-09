package benchcmp

import (
	"testing"

	flux "github.com/fluxrpc/solana-go"
	fluxata "github.com/fluxrpc/solana-go/programs/associated-token-account"
	gagliardetto "github.com/gagliardetto/solana-go"
	gagliardettoata "github.com/gagliardetto/solana-go/programs/associated-token-account"
)

var (
	sinkFluxATAInstruction         flux.Instruction
	sinkGagliardettoATAInstruction gagliardetto.Instruction
	sinkFluxATADecoded             fluxata.DecodedInstruction
	sinkGagliardettoATADecoded     *gagliardettoata.Instruction
	sinkATAData                    []byte
	sinkATAErr                     error
)

type ataInstructionBenchmark struct {
	name             string
	flux             flux.Instruction
	gagliardetto     gagliardetto.Instruction
	fluxData         []byte
	gagliardettoData []byte
	newFlux          func() flux.Instruction
	newGagliardetto  func() gagliardetto.Instruction
}

func ataBenchmarkKey(tag byte) (key [32]byte) {
	for index := range key {
		key[index] = tag + byte(index)
	}
	return key
}

func newATAInstructionBenchmarks(tb testing.TB) []ataInstructionBenchmark {
	tb.Helper()
	fluxKey := func(tag byte) flux.PublicKey { return flux.PublicKey(ataBenchmarkKey(tag)) }
	gagliardettoKey := func(tag byte) gagliardetto.PublicKey { return gagliardetto.PublicKey(ataBenchmarkKey(tag)) }

	fluxATA := func(wallet, mint flux.PublicKey) flux.PublicKey {
		ata, _, err := wallet.FindAssociatedTokenAddressWithProgram(mint, flux.TokenProgramID)
		if err != nil {
			tb.Fatal(err)
		}
		return ata
	}

	benchmarks := []ataInstructionBenchmark{
		{
			name: "Create",
			newFlux: func() flux.Instruction {
				return fluxata.NewCreateInstruction(fluxKey(1), fluxATA(fluxKey(2), fluxKey(3)), fluxKey(2), fluxKey(3), flux.TokenProgramID)
			},
			newGagliardetto: func() gagliardetto.Instruction {
				return gagliardettoata.NewCreateInstruction(gagliardettoKey(1), gagliardettoKey(2), gagliardettoKey(3)).Build()
			},
		},
		{
			name: "CreateIdempotent",
			newFlux: func() flux.Instruction {
				return fluxata.NewCreateIdempotentInstruction(fluxKey(1), fluxATA(fluxKey(2), fluxKey(3)), fluxKey(2), fluxKey(3), flux.TokenProgramID)
			},
			newGagliardetto: func() gagliardetto.Instruction {
				return gagliardettoata.NewCreateIdempotentInstruction(gagliardettoKey(1), gagliardettoKey(2), gagliardettoKey(3)).Build()
			},
		},
		{
			name: "RecoverNested",
			newFlux: func() flux.Instruction {
				return fluxata.NewRecoverNestedInstruction(
					fluxATA(fluxATA(fluxKey(2), fluxKey(3)), fluxKey(4)), fluxKey(4),
					fluxATA(fluxKey(2), fluxKey(4)), fluxATA(fluxKey(2), fluxKey(3)), fluxKey(3),
					fluxKey(2), flux.TokenProgramID,
				)
			},
			newGagliardetto: func() gagliardetto.Instruction {
				return gagliardettoata.NewRecoverNestedInstruction(gagliardettoKey(2), gagliardettoKey(4), gagliardettoKey(3)).Build()
			},
		},
	}
	for index := range benchmarks {
		benchmarks[index].flux = benchmarks[index].newFlux()
		benchmarks[index].gagliardetto = benchmarks[index].newGagliardetto()
		var err error
		if benchmarks[index].fluxData, err = benchmarks[index].flux.Data(); err != nil {
			tb.Fatal(err)
		}
		if benchmarks[index].gagliardettoData, err = benchmarks[index].gagliardetto.Data(); err != nil {
			tb.Fatal(err)
		}
	}
	return benchmarks
}

func TestATAInstructionAllocationsBeatGagliardetto(t *testing.T) {
	for _, test := range newATAInstructionBenchmarks(t) {
		t.Run(test.name, func(t *testing.T) {
			fluxConstructor := testing.AllocsPerRun(1_000, func() {
				sinkFluxATAInstruction = test.newFlux()
				sinkATAData, sinkATAErr = sinkFluxATAInstruction.Data()
			})
			gagliardettoConstructor := testing.AllocsPerRun(1_000, func() {
				sinkGagliardettoATAInstruction = test.newGagliardetto()
				sinkATAData, sinkATAErr = sinkGagliardettoATAInstruction.Data()
			})
			if fluxConstructor >= gagliardettoConstructor {
				t.Errorf("constructor+Data allocations: flux %.0f, gagliardetto %.0f", fluxConstructor, gagliardettoConstructor)
			}

			fluxData := testing.AllocsPerRun(1_000, func() { sinkATAData, sinkATAErr = test.flux.Data() })
			gagliardettoData := testing.AllocsPerRun(1_000, func() { sinkATAData, sinkATAErr = test.gagliardetto.Data() })
			if fluxData >= gagliardettoData {
				t.Errorf("Data allocations: flux %.0f, gagliardetto %.0f", fluxData, gagliardettoData)
			}

			fluxDecode := testing.AllocsPerRun(1_000, func() {
				sinkFluxATADecoded, sinkATAErr = fluxata.DecodeInstruction(test.flux.Accounts(), test.fluxData)
			})
			gagliardettoDecode := testing.AllocsPerRun(1_000, func() {
				sinkGagliardettoATADecoded, sinkATAErr = gagliardettoata.DecodeInstruction(test.gagliardetto.Accounts(), test.gagliardettoData)
			})
			if fluxDecode >= gagliardettoDecode {
				t.Errorf("Decode allocations: flux %.0f, gagliardetto %.0f", fluxDecode, gagliardettoDecode)
			}
		})
	}
}

func BenchmarkATAConstructorAndData(b *testing.B) {
	for _, test := range newATAInstructionBenchmarks(b) {
		b.Run(test.name, func(b *testing.B) {
			b.Run("Flux", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFluxATAInstruction = test.newFlux()
					sinkATAData, sinkATAErr = sinkFluxATAInstruction.Data()
				}
			})
			b.Run("Gagliardetto", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkGagliardettoATAInstruction = test.newGagliardetto()
					sinkATAData, sinkATAErr = sinkGagliardettoATAInstruction.Data()
				}
			})
		})
	}
}

func BenchmarkATAData(b *testing.B) {
	for _, test := range newATAInstructionBenchmarks(b) {
		b.Run(test.name, func(b *testing.B) {
			b.Run("Flux", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkATAData, sinkATAErr = test.flux.Data()
				}
			})
			b.Run("Gagliardetto", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkATAData, sinkATAErr = test.gagliardetto.Data()
				}
			})
		})
	}
}

func BenchmarkATADecode(b *testing.B) {
	for _, test := range newATAInstructionBenchmarks(b) {
		b.Run(test.name, func(b *testing.B) {
			b.Run("Flux", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFluxATADecoded, sinkATAErr = fluxata.DecodeInstruction(test.flux.Accounts(), test.fluxData)
				}
			})
			b.Run("Gagliardetto", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkGagliardettoATADecoded, sinkATAErr = gagliardettoata.DecodeInstruction(test.gagliardetto.Accounts(), test.gagliardettoData)
				}
			})
		})
	}
}
