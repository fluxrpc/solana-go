package benchcmp

import (
	"bytes"
	"testing"

	flux "github.com/fluxrpc/solana-go"
	fluxalt "github.com/fluxrpc/solana-go/programs/address-lookup-table"
	foundation "github.com/solana-foundation/solana-go/v2"
	foundationalt "github.com/solana-foundation/solana-go/v2/programs/address-lookup-table"
)

var (
	sinkFluxALTInstruction       flux.Instruction
	sinkFoundationALTInstruction foundation.Instruction
	sinkFluxALTDecoded           fluxalt.DecodedInstruction
	sinkFoundationALTDecoded     *foundationalt.Instruction
	sinkALTData                  []byte
	sinkALTErr                   error
)

type altInstructionBenchmark struct {
	name           string
	flux           flux.Instruction
	foundation     foundation.Instruction
	fluxData       []byte
	foundationData []byte
	newFlux        func() flux.Instruction
	newFoundation  func() foundation.Instruction
}

func altBenchmarkKey(tag byte) (key [32]byte) {
	for index := range key {
		key[index] = tag + byte(index)
	}
	return key
}

func newALTInstructionBenchmarks(tb testing.TB) []altInstructionBenchmark {
	tb.Helper()
	fluxKey := func(tag byte) flux.PublicKey { return flux.PublicKey(altBenchmarkKey(tag)) }
	foundationKey := func(tag byte) foundation.PublicKey { return foundation.PublicKey(altBenchmarkKey(tag)) }

	fluxAddresses := []flux.PublicKey{fluxKey(10), fluxKey(11), fluxKey(12), fluxKey(13)}
	foundationAddresses := []foundation.PublicKey{
		foundationKey(10), foundationKey(11), foundationKey(12), foundationKey(13),
	}
	const recentSlot = uint64(123_456_789)

	benchmarks := []altInstructionBenchmark{
		{
			name: "CreateLookupTable",
			newFlux: func() flux.Instruction {
				instruction, _, err := fluxalt.NewCreateLookupTableInstruction(fluxKey(1), fluxKey(2), recentSlot)
				if err != nil {
					tb.Fatal(err)
				}
				return instruction
			},
			newFoundation: func() foundation.Instruction {
				instruction, _, err := foundationalt.NewCreateLookupTableInstruction(foundationKey(1), foundationKey(2), recentSlot)
				if err != nil {
					tb.Fatal(err)
				}
				return instruction.Build()
			},
		},
		{
			name:    "FreezeLookupTable",
			newFlux: func() flux.Instruction { return fluxalt.NewFreezeLookupTableInstruction(fluxKey(3), fluxKey(1)) },
			newFoundation: func() foundation.Instruction {
				return foundationalt.NewFreezeLookupTableInstruction(foundationKey(3), foundationKey(1)).Build()
			},
		},
		{
			name: "ExtendLookupTable",
			newFlux: func() flux.Instruction {
				return fluxalt.NewExtendLookupTableInstruction(fluxKey(3), fluxKey(1), fluxKey(2), fluxAddresses)
			},
			newFoundation: func() foundation.Instruction {
				return foundationalt.NewExtendLookupTableInstruction(foundationKey(3), foundationKey(1), foundationKey(2), foundationAddresses).Build()
			},
		},
		{
			name:    "DeactivateLookupTable",
			newFlux: func() flux.Instruction { return fluxalt.NewDeactivateLookupTableInstruction(fluxKey(3), fluxKey(1)) },
			newFoundation: func() foundation.Instruction {
				return foundationalt.NewDeactivateLookupTableInstruction(foundationKey(3), foundationKey(1)).Build()
			},
		},
		{
			name: "CloseLookupTable",
			newFlux: func() flux.Instruction {
				return fluxalt.NewCloseLookupTableInstruction(fluxKey(3), fluxKey(1), fluxKey(4))
			},
			newFoundation: func() foundation.Instruction {
				return foundationalt.NewCloseLookupTableInstruction(foundationKey(3), foundationKey(1), foundationKey(4)).Build()
			},
		},
	}

	for index := range benchmarks {
		entry := &benchmarks[index]
		entry.flux = entry.newFlux()
		entry.foundation = entry.newFoundation()
		var err error
		entry.fluxData, err = entry.flux.Data()
		if err != nil {
			tb.Fatalf("flux %s Data: %v", entry.name, err)
		}
		entry.foundationData, err = entry.foundation.Data()
		if err != nil {
			tb.Fatalf("foundation %s Data: %v", entry.name, err)
		}
		if !bytes.Equal(entry.fluxData, entry.foundationData) {
			tb.Fatalf("%s data differs: flux %x, foundation %x", entry.name, entry.fluxData, entry.foundationData)
		}
	}
	return benchmarks
}

func TestALTInstructionAllocationsBeatFoundation(t *testing.T) {
	for _, test := range newALTInstructionBenchmarks(t) {
		t.Run(test.name, func(t *testing.T) {
			fluxConstructor := testing.AllocsPerRun(1_000, func() {
				sinkFluxALTInstruction = test.newFlux()
				sinkALTData, sinkALTErr = sinkFluxALTInstruction.Data()
			})
			if sinkALTErr != nil {
				t.Fatal(sinkALTErr)
			}
			foundationConstructor := testing.AllocsPerRun(1_000, func() {
				sinkFoundationALTInstruction = test.newFoundation()
				sinkALTData, sinkALTErr = sinkFoundationALTInstruction.Data()
			})
			if sinkALTErr != nil {
				t.Fatal(sinkALTErr)
			}
			if fluxConstructor >= foundationConstructor {
				t.Errorf("constructor+Data allocations: flux %.0f, foundation %.0f", fluxConstructor, foundationConstructor)
			}

			fluxData := testing.AllocsPerRun(1_000, func() { sinkALTData, sinkALTErr = test.flux.Data() })
			foundationData := testing.AllocsPerRun(1_000, func() { sinkALTData, sinkALTErr = test.foundation.Data() })
			if fluxData >= foundationData {
				t.Errorf("Data allocations: flux %.0f, foundation %.0f", fluxData, foundationData)
			}

			fluxDecode := testing.AllocsPerRun(1_000, func() {
				sinkFluxALTDecoded, sinkALTErr = fluxalt.DecodeInstruction(test.flux.Accounts(), test.fluxData)
			})
			foundationDecode := testing.AllocsPerRun(1_000, func() {
				sinkFoundationALTDecoded, sinkALTErr = foundationalt.DecodeInstruction(test.foundation.Accounts(), test.foundationData)
			})
			if fluxDecode >= foundationDecode {
				t.Errorf("Decode allocations: flux %.0f, foundation %.0f", fluxDecode, foundationDecode)
			}
		})
	}
}

func BenchmarkALTConstructorAndData(b *testing.B) {
	for _, test := range newALTInstructionBenchmarks(b) {
		b.Run(test.name, func(b *testing.B) {
			b.Run("Flux", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFluxALTInstruction = test.newFlux()
					sinkALTData, sinkALTErr = sinkFluxALTInstruction.Data()
				}
			})
			b.Run("Foundation", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFoundationALTInstruction = test.newFoundation()
					sinkALTData, sinkALTErr = sinkFoundationALTInstruction.Data()
				}
			})
		})
	}
}

func BenchmarkALTData(b *testing.B) {
	for _, test := range newALTInstructionBenchmarks(b) {
		b.Run(test.name, func(b *testing.B) {
			b.Run("Flux", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkALTData, sinkALTErr = test.flux.Data()
				}
			})
			b.Run("Foundation", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkALTData, sinkALTErr = test.foundation.Data()
				}
			})
		})
	}
}

func BenchmarkALTDecode(b *testing.B) {
	for _, test := range newALTInstructionBenchmarks(b) {
		b.Run(test.name, func(b *testing.B) {
			b.Run("Flux", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFluxALTDecoded, sinkALTErr = fluxalt.DecodeInstruction(test.flux.Accounts(), test.fluxData)
				}
			})
			b.Run("Foundation", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFoundationALTDecoded, sinkALTErr = foundationalt.DecodeInstruction(test.foundation.Accounts(), test.foundationData)
				}
			})
		})
	}
}
