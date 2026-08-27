package benchcmp

import (
	"bytes"
	"testing"

	flux "github.com/fluxrpc/solana-go"
	fluxcompute "github.com/fluxrpc/solana-go/programs/compute-budget"
	foundation "github.com/solana-foundation/solana-go/v2"
	foundationcompute "github.com/solana-foundation/solana-go/v2/programs/compute-budget"
)

var (
	sinkFluxComputeInstruction       flux.Instruction
	sinkFoundationComputeInstruction foundation.Instruction
	sinkFluxComputeDecoded           fluxcompute.DecodedInstruction
	sinkFoundationComputeDecoded     *foundationcompute.Instruction
	sinkComputeData                  []byte
	sinkComputeErr                   error
)

type computeInstructionBenchmark struct {
	name           string
	flux           flux.Instruction
	foundation     foundation.Instruction
	fluxData       []byte
	foundationData []byte
	newFlux        func() flux.Instruction
	newFoundation  func() foundation.Instruction
}

func newComputeInstructionBenchmarks(tb testing.TB) []computeInstructionBenchmark {
	tb.Helper()
	benchmarks := []computeInstructionBenchmark{
		{
			name:    "RequestUnitsDeprecated",
			newFlux: func() flux.Instruction { return fluxcompute.NewRequestUnitsDeprecatedInstruction(1_400_000, 37) },
			newFoundation: func() foundation.Instruction {
				return foundationcompute.NewRequestUnitsDeprecatedInstruction(1_400_000, 37).Build()
			},
		},
		{
			name:    "RequestHeapFrame",
			newFlux: func() flux.Instruction { return fluxcompute.NewRequestHeapFrameInstruction(256 * 1024) },
			newFoundation: func() foundation.Instruction {
				return foundationcompute.NewRequestHeapFrameInstruction(256 * 1024).Build()
			},
		},
		{
			name:    "SetComputeUnitLimit",
			newFlux: func() flux.Instruction { return fluxcompute.NewSetComputeUnitLimitInstruction(1_000_000) },
			newFoundation: func() foundation.Instruction {
				return foundationcompute.NewSetComputeUnitLimitInstruction(1_000_000).Build()
			},
		},
		{
			name:    "SetComputeUnitPrice",
			newFlux: func() flux.Instruction { return fluxcompute.NewSetComputeUnitPriceInstruction(1_234_567) },
			newFoundation: func() foundation.Instruction {
				return foundationcompute.NewSetComputeUnitPriceInstruction(1_234_567).Build()
			},
		},
		{
			name: "SetLoadedAccountsDataSizeLimit",
			newFlux: func() flux.Instruction {
				return fluxcompute.NewSetLoadedAccountsDataSizeLimitInstruction(64 * 1024 * 1024)
			},
			newFoundation: func() foundation.Instruction {
				return foundationcompute.NewSetLoadedAccountsDataSizeLimitInstruction(64 * 1024 * 1024).Build()
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

func TestComputeInstructionAllocationsBeatFoundation(t *testing.T) {
	for _, test := range newComputeInstructionBenchmarks(t) {
		t.Run(test.name, func(t *testing.T) {
			fluxConstructor := testing.AllocsPerRun(1_000, func() {
				sinkFluxComputeInstruction = test.newFlux()
				sinkComputeData, sinkComputeErr = sinkFluxComputeInstruction.Data()
			})
			foundationConstructor := testing.AllocsPerRun(1_000, func() {
				sinkFoundationComputeInstruction = test.newFoundation()
				sinkComputeData, sinkComputeErr = sinkFoundationComputeInstruction.Data()
			})
			if fluxConstructor >= foundationConstructor {
				t.Errorf("constructor+Data allocations: flux %.0f, foundation %.0f", fluxConstructor, foundationConstructor)
			}

			fluxData := testing.AllocsPerRun(1_000, func() { sinkComputeData, sinkComputeErr = test.flux.Data() })
			foundationData := testing.AllocsPerRun(1_000, func() { sinkComputeData, sinkComputeErr = test.foundation.Data() })
			if fluxData >= foundationData {
				t.Errorf("Data allocations: flux %.0f, foundation %.0f", fluxData, foundationData)
			}

			fluxDecode := testing.AllocsPerRun(1_000, func() {
				sinkFluxComputeDecoded, sinkComputeErr = fluxcompute.DecodeInstruction(nil, test.fluxData)
			})
			foundationDecode := testing.AllocsPerRun(1_000, func() {
				sinkFoundationComputeDecoded, sinkComputeErr = foundationcompute.DecodeInstruction(nil, test.foundationData)
			})
			if fluxDecode >= foundationDecode {
				t.Errorf("Decode allocations: flux %.0f, foundation %.0f", fluxDecode, foundationDecode)
			}
		})
	}
}

func BenchmarkComputeConstructorAndData(b *testing.B) {
	for _, test := range newComputeInstructionBenchmarks(b) {
		b.Run(test.name, func(b *testing.B) {
			b.Run("Flux", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFluxComputeInstruction = test.newFlux()
					sinkComputeData, sinkComputeErr = sinkFluxComputeInstruction.Data()
				}
			})
			b.Run("Foundation", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFoundationComputeInstruction = test.newFoundation()
					sinkComputeData, sinkComputeErr = sinkFoundationComputeInstruction.Data()
				}
			})
		})
	}
}

func BenchmarkComputeData(b *testing.B) {
	for _, test := range newComputeInstructionBenchmarks(b) {
		b.Run(test.name, func(b *testing.B) {
			b.Run("Flux", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkComputeData, sinkComputeErr = test.flux.Data()
				}
			})
			b.Run("Foundation", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkComputeData, sinkComputeErr = test.foundation.Data()
				}
			})
		})
	}
}

func BenchmarkComputeDecode(b *testing.B) {
	for _, test := range newComputeInstructionBenchmarks(b) {
		b.Run(test.name, func(b *testing.B) {
			b.Run("Flux", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFluxComputeDecoded, sinkComputeErr = fluxcompute.DecodeInstruction(nil, test.fluxData)
				}
			})
			b.Run("Foundation", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFoundationComputeDecoded, sinkComputeErr = foundationcompute.DecodeInstruction(nil, test.foundationData)
				}
			})
		})
	}
}
