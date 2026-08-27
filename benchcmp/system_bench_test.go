package benchcmp

import (
	"bytes"
	"testing"

	flux "github.com/fluxrpc/solana-go"
	fluxsystem "github.com/fluxrpc/solana-go/programs/system"
	foundation "github.com/solana-foundation/solana-go/v2"
	foundationsystem "github.com/solana-foundation/solana-go/v2/programs/system"
)

var (
	sinkFluxSystemInstruction       flux.Instruction
	sinkFoundationSystemInstruction foundation.Instruction
	sinkFluxSystemDecoded           fluxsystem.DecodedInstruction
	sinkFoundationSystemDecoded     *foundationsystem.Instruction
	sinkSystemData                  []byte
	sinkSystemErr                   error
)

type systemInstructionBenchmark struct {
	name               string
	flux               flux.Instruction
	foundation         foundation.Instruction
	data               []byte
	foundationData     []byte
	comparePayloadOnly bool
	newFlux            func() flux.Instruction
	newFoundation      func() foundation.Instruction
}

func systemBenchmarkPublicKey(tag byte) (key [32]byte) {
	for index := range key {
		key[index] = tag + byte(index)
	}
	return key
}

func newSystemInstructionBenchmarks(tb testing.TB) []systemInstructionBenchmark {
	tb.Helper()
	fluxKey := func(tag byte) flux.PublicKey {
		return flux.PublicKey(systemBenchmarkPublicKey(tag))
	}
	foundationKey := func(tag byte) foundation.PublicKey {
		return foundation.PublicKey(systemBenchmarkPublicKey(tag))
	}
	mustFlux := func(instruction flux.Instruction, err error) flux.Instruction {
		if err != nil {
			tb.Fatal(err)
		}
		return instruction
	}

	const (
		seed     = "system-benchmark-seed"
		lamports = uint64(1_234_567_890)
		space    = uint64(4_096)
	)

	benchmarks := []systemInstructionBenchmark{
		{
			name: "CreateAccount",
			newFlux: func() flux.Instruction {
				return fluxsystem.NewCreateAccountInstruction(lamports, space, fluxKey(1), fluxKey(2), fluxKey(3))
			},
			newFoundation: func() foundation.Instruction {
				return foundationsystem.NewCreateAccountInstruction(lamports, space, foundationKey(1), foundationKey(2), foundationKey(3)).Build()
			},
		},
		{
			name: "Assign",
			newFlux: func() flux.Instruction {
				return fluxsystem.NewAssignInstruction(fluxKey(4), fluxKey(5))
			},
			newFoundation: func() foundation.Instruction {
				return foundationsystem.NewAssignInstruction(foundationKey(4), foundationKey(5)).Build()
			},
		},
		{
			name: "Transfer",
			newFlux: func() flux.Instruction {
				return fluxsystem.NewTransferInstruction(lamports, fluxKey(6), fluxKey(7))
			},
			newFoundation: func() foundation.Instruction {
				return foundationsystem.NewTransferInstruction(lamports, foundationKey(6), foundationKey(7)).Build()
			},
		},
		{
			name: "CreateAccountWithSeed",
			newFlux: func() flux.Instruction {
				return mustFlux(fluxsystem.NewCreateAccountWithSeedInstruction(
					fluxKey(8), seed, lamports, space, fluxKey(9), fluxKey(10), fluxKey(11), fluxKey(12),
				))
			},
			newFoundation: func() foundation.Instruction {
				return foundationsystem.NewCreateAccountWithSeedInstruction(
					foundationKey(8), seed, lamports, space, foundationKey(9), foundationKey(10), foundationKey(11), foundationKey(12),
				).Build()
			},
		},
		{
			name: "AdvanceNonceAccount",
			newFlux: func() flux.Instruction {
				return fluxsystem.NewAdvanceNonceAccountInstruction(fluxKey(13), fluxKey(14), fluxKey(15))
			},
			newFoundation: func() foundation.Instruction {
				return foundationsystem.NewAdvanceNonceAccountInstruction(foundationKey(13), foundationKey(14), foundationKey(15)).Build()
			},
		},
		{
			name: "WithdrawNonceAccount",
			newFlux: func() flux.Instruction {
				return fluxsystem.NewWithdrawNonceAccountInstruction(
					lamports, fluxKey(16), fluxKey(17), fluxKey(18), fluxKey(19), fluxKey(20),
				)
			},
			newFoundation: func() foundation.Instruction {
				return foundationsystem.NewWithdrawNonceAccountInstruction(
					lamports, foundationKey(16), foundationKey(17), foundationKey(18), foundationKey(19), foundationKey(20),
				).Build()
			},
		},
		{
			name: "InitializeNonceAccount",
			newFlux: func() flux.Instruction {
				return fluxsystem.NewInitializeNonceAccountInstruction(fluxKey(21), fluxKey(22), fluxKey(23), fluxKey(24))
			},
			newFoundation: func() foundation.Instruction {
				return foundationsystem.NewInitializeNonceAccountInstruction(foundationKey(21), foundationKey(22), foundationKey(23), foundationKey(24)).Build()
			},
		},
		{
			name: "AuthorizeNonceAccount",
			newFlux: func() flux.Instruction {
				return fluxsystem.NewAuthorizeNonceAccountInstruction(fluxKey(25), fluxKey(26), fluxKey(27))
			},
			newFoundation: func() foundation.Instruction {
				return foundationsystem.NewAuthorizeNonceAccountInstruction(foundationKey(25), foundationKey(26), foundationKey(27)).Build()
			},
		},
		{
			name: "Allocate",
			newFlux: func() flux.Instruction {
				return fluxsystem.NewAllocateInstruction(space, fluxKey(28))
			},
			newFoundation: func() foundation.Instruction {
				return foundationsystem.NewAllocateInstruction(space, foundationKey(28)).Build()
			},
		},
		{
			name: "AllocateWithSeed",
			newFlux: func() flux.Instruction {
				return mustFlux(fluxsystem.NewAllocateWithSeedInstruction(
					fluxKey(29), seed, space, fluxKey(30), fluxKey(31), fluxKey(32),
				))
			},
			newFoundation: func() foundation.Instruction {
				return foundationsystem.NewAllocateWithSeedInstruction(
					foundationKey(29), seed, space, foundationKey(30), foundationKey(31), foundationKey(32),
				).Build()
			},
		},
		{
			name: "AssignWithSeed",
			newFlux: func() flux.Instruction {
				return mustFlux(fluxsystem.NewAssignWithSeedInstruction(
					fluxKey(33), seed, fluxKey(34), fluxKey(35), fluxKey(36),
				))
			},
			newFoundation: func() foundation.Instruction {
				return foundationsystem.NewAssignWithSeedInstruction(
					foundationKey(33), seed, foundationKey(34), foundationKey(35), foundationKey(36),
				).Build()
			},
		},
		{
			name: "TransferWithSeed",
			newFlux: func() flux.Instruction {
				return mustFlux(fluxsystem.NewTransferWithSeedInstruction(
					lamports, seed, fluxKey(37), fluxKey(38), fluxKey(39), fluxKey(40),
				))
			},
			newFoundation: func() foundation.Instruction {
				return foundationsystem.NewTransferWithSeedInstruction(
					lamports, seed, foundationKey(37), foundationKey(38), foundationKey(39), foundationKey(40),
				).Build()
			},
		},
		{
			name: "UpgradeNonceAccount",
			newFlux: func() flux.Instruction {
				return fluxsystem.NewUpgradeNonceAccountInstruction(fluxKey(41))
			},
			newFoundation: func() foundation.Instruction {
				return foundationsystem.NewUpgradeNonceAccountInstruction(foundationKey(41)).Build()
			},
		},
		{
			name: "CreateAccountAllowPrefund_SamePayloadVsCreateAccount",
			newFlux: func() flux.Instruction {
				return fluxsystem.NewCreateAccountAllowPrefundInstruction(
					lamports, space, fluxKey(42), fluxKey(43), fluxKey(44),
				)
			},
			// Foundation v2 predates discriminator 13. CreateAccount has the same
			// payload fields, so this measures codec cost only; it is not parity.
			newFoundation: func() foundation.Instruction {
				return foundationsystem.NewCreateAccountInstruction(
					lamports, space, foundationKey(42), foundationKey(44), foundationKey(43),
				).Build()
			},
			comparePayloadOnly: true,
		},
	}

	for index := range benchmarks {
		benchmarks[index].flux = benchmarks[index].newFlux()
		benchmarks[index].foundation = benchmarks[index].newFoundation()
		fluxData, err := benchmarks[index].flux.Data()
		if err != nil {
			tb.Fatalf("flux %s Data: %v", benchmarks[index].name, err)
		}
		foundationData, err := benchmarks[index].foundation.Data()
		if err != nil {
			tb.Fatalf("foundation %s Data: %v", benchmarks[index].name, err)
		}
		parity := bytes.Equal(fluxData, foundationData)
		if benchmarks[index].comparePayloadOnly {
			parity = len(fluxData) >= 4 && len(foundationData) >= 4 && bytes.Equal(fluxData[4:], foundationData[4:])
		}
		if !parity {
			tb.Fatalf(
				"%s instruction data differs: flux %x, foundation %x",
				benchmarks[index].name,
				fluxData,
				foundationData,
			)
		}
		benchmarks[index].data = fluxData
		benchmarks[index].foundationData = foundationData
	}

	return benchmarks
}

func TestSystemInstructionAllocationsBeatFoundation(t *testing.T) {
	for _, test := range newSystemInstructionBenchmarks(t) {
		t.Run(test.name, func(t *testing.T) {
			fluxConstructorAllocs := testing.AllocsPerRun(1_000, func() {
				sinkFluxSystemInstruction = test.newFlux()
				sinkSystemData, sinkSystemErr = sinkFluxSystemInstruction.Data()
			})
			if sinkSystemErr != nil {
				t.Fatal(sinkSystemErr)
			}
			foundationConstructorAllocs := testing.AllocsPerRun(1_000, func() {
				sinkFoundationSystemInstruction = test.newFoundation()
				sinkSystemData, sinkSystemErr = sinkFoundationSystemInstruction.Data()
			})
			if sinkSystemErr != nil {
				t.Fatal(sinkSystemErr)
			}
			if fluxConstructorAllocs >= foundationConstructorAllocs {
				t.Fatalf(
					"constructor+Data allocations: flux %.0f, foundation %.0f",
					fluxConstructorAllocs,
					foundationConstructorAllocs,
				)
			}

			fluxDataAllocs := testing.AllocsPerRun(1_000, func() {
				sinkSystemData, sinkSystemErr = test.flux.Data()
			})
			if sinkSystemErr != nil {
				t.Fatal(sinkSystemErr)
			}
			foundationDataAllocs := testing.AllocsPerRun(1_000, func() {
				sinkSystemData, sinkSystemErr = test.foundation.Data()
			})
			if sinkSystemErr != nil {
				t.Fatal(sinkSystemErr)
			}
			if fluxDataAllocs >= foundationDataAllocs {
				t.Fatalf("Data allocations: flux %.0f, foundation %.0f", fluxDataAllocs, foundationDataAllocs)
			}

			fluxAccounts := test.flux.Accounts()
			foundationAccounts := test.foundation.Accounts()
			fluxDecodeAllocs := testing.AllocsPerRun(1_000, func() {
				sinkFluxSystemDecoded, sinkSystemErr = fluxsystem.DecodeInstruction(fluxAccounts, test.data)
			})
			if sinkSystemErr != nil {
				t.Fatal(sinkSystemErr)
			}
			foundationDecodeAllocs := testing.AllocsPerRun(1_000, func() {
				sinkFoundationSystemDecoded, sinkSystemErr = foundationsystem.DecodeInstruction(foundationAccounts, test.foundationData)
			})
			if sinkSystemErr != nil {
				t.Fatal(sinkSystemErr)
			}
			if fluxDecodeAllocs >= foundationDecodeAllocs {
				t.Fatalf("DecodeInstruction allocations: flux %.0f, foundation %.0f", fluxDecodeAllocs, foundationDecodeAllocs)
			}
		})
	}
}

func BenchmarkSystemInstructions_ConstructorData(b *testing.B) {
	for _, benchmark := range newSystemInstructionBenchmarks(b) {
		b.Run(benchmark.name, func(b *testing.B) {
			b.Run("Flux", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFluxSystemInstruction = benchmark.newFlux()
					sinkSystemData, sinkSystemErr = sinkFluxSystemInstruction.Data()
				}
				if sinkSystemErr != nil {
					b.Fatal(sinkSystemErr)
				}
			})
			b.Run("Foundation", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFoundationSystemInstruction = benchmark.newFoundation()
					sinkSystemData, sinkSystemErr = sinkFoundationSystemInstruction.Data()
				}
				if sinkSystemErr != nil {
					b.Fatal(sinkSystemErr)
				}
			})
		})
	}
}

func BenchmarkSystemInstructions_Data(b *testing.B) {
	for _, benchmark := range newSystemInstructionBenchmarks(b) {
		b.Run(benchmark.name, func(b *testing.B) {
			b.Run("Flux", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkSystemData, sinkSystemErr = benchmark.flux.Data()
				}
				if sinkSystemErr != nil {
					b.Fatal(sinkSystemErr)
				}
			})
			b.Run("Foundation", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkSystemData, sinkSystemErr = benchmark.foundation.Data()
				}
				if sinkSystemErr != nil {
					b.Fatal(sinkSystemErr)
				}
			})
		})
	}
}

func BenchmarkSystemInstructions_DecodeInstruction(b *testing.B) {
	for _, benchmark := range newSystemInstructionBenchmarks(b) {
		fluxAccounts := benchmark.flux.Accounts()
		foundationAccounts := benchmark.foundation.Accounts()
		b.Run(benchmark.name, func(b *testing.B) {
			b.Run("Flux", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFluxSystemDecoded, sinkSystemErr = fluxsystem.DecodeInstruction(fluxAccounts, benchmark.data)
				}
				if sinkSystemErr != nil {
					b.Fatal(sinkSystemErr)
				}
			})
			b.Run("Foundation", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFoundationSystemDecoded, sinkSystemErr = foundationsystem.DecodeInstruction(foundationAccounts, benchmark.foundationData)
				}
				if sinkSystemErr != nil {
					b.Fatal(sinkSystemErr)
				}
			})
		})
	}
}
