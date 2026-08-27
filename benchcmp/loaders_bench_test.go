package benchcmp

import (
	"bytes"
	"testing"

	flux "github.com/fluxrpc/solana-go"
	fluxv2 "github.com/fluxrpc/solana-go/programs/loader-v2"
	fluxv3 "github.com/fluxrpc/solana-go/programs/loader-v3"
	fluxv4 "github.com/fluxrpc/solana-go/programs/loader-v4"
	foundation "github.com/solana-foundation/solana-go/v2"
	foundationv2 "github.com/solana-foundation/solana-go/v2/programs/loader-v2"
	foundationv3 "github.com/solana-foundation/solana-go/v2/programs/loader-v3"
	foundationv4 "github.com/solana-foundation/solana-go/v2/programs/loader-v4"
)

var (
	sinkFluxLoaderInstruction       flux.Instruction
	sinkFoundationLoaderInstruction foundation.Instruction
	sinkFluxLoaderV2Decoded         fluxv2.DecodedInstruction
	sinkFluxLoaderV3Decoded         fluxv3.DecodedInstruction
	sinkFluxLoaderV4Decoded         fluxv4.DecodedInstruction
	sinkFoundationLoaderV2Decoded   *foundationv2.Instruction
	sinkFoundationLoaderV3Decoded   *foundationv3.Instruction
	sinkFoundationLoaderV4Decoded   *foundationv4.Instruction
	sinkLoaderData                  []byte
	sinkLoaderErr                   error
)

type loaderInstructionBenchmark struct {
	family           string
	name             string
	flux             flux.Instruction
	foundation       foundation.Instruction
	fluxData         []byte
	foundationData   []byte
	newFlux          func() flux.Instruction
	newFoundation    func() foundation.Instruction
	decodeFlux       func([]*flux.AccountMeta, []byte)
	decodeFoundation func([]*foundation.AccountMeta, []byte)
}

func loaderBenchmarkKey(tag byte) (key [32]byte) {
	for index := range key {
		key[index] = tag + byte(index)
	}
	return key
}

func newLoaderInstructionBenchmarks(tb testing.TB) []loaderInstructionBenchmark {
	tb.Helper()
	fluxKey := func(tag byte) flux.PublicKey { return flux.PublicKey(loaderBenchmarkKey(tag)) }
	foundationKey := func(tag byte) foundation.PublicKey { return foundation.PublicKey(loaderBenchmarkKey(tag)) }
	data := []byte("loader benchmark payload")

	program := foundationKey(3)
	foundationProgramData := foundationv3.MustGetProgramDataAddress(program)
	fluxProgramData := flux.PublicKey(foundationProgramData)
	newAuthorityFoundation := foundationKey(7)

	decodeV2Flux := func(accounts []*flux.AccountMeta, data []byte) {
		sinkFluxLoaderV2Decoded, sinkLoaderErr = fluxv2.DecodeInstruction(accounts, data)
	}
	decodeV2Foundation := func(accounts []*foundation.AccountMeta, data []byte) {
		sinkFoundationLoaderV2Decoded, sinkLoaderErr = foundationv2.DecodeInstruction(accounts, data)
	}
	decodeV3Flux := func(accounts []*flux.AccountMeta, data []byte) {
		sinkFluxLoaderV3Decoded, sinkLoaderErr = fluxv3.DecodeInstruction(accounts, data)
	}
	decodeV3Foundation := func(accounts []*foundation.AccountMeta, data []byte) {
		sinkFoundationLoaderV3Decoded, sinkLoaderErr = foundationv3.DecodeInstruction(accounts, data)
	}
	decodeV4Flux := func(accounts []*flux.AccountMeta, data []byte) {
		sinkFluxLoaderV4Decoded, sinkLoaderErr = fluxv4.DecodeInstruction(accounts, data)
	}
	decodeV4Foundation := func(accounts []*foundation.AccountMeta, data []byte) {
		sinkFoundationLoaderV4Decoded, sinkLoaderErr = foundationv4.DecodeInstruction(accounts, data)
	}

	benchmarks := []loaderInstructionBenchmark{
		{
			family: "LoaderV2", name: "Write",
			newFlux: func() flux.Instruction { return fluxv2.NewWriteInstruction(42, data, fluxKey(1)) },
			newFoundation: func() foundation.Instruction {
				return foundationv2.NewWriteInstruction(42, data, foundationKey(1)).Build()
			},
			decodeFlux: decodeV2Flux, decodeFoundation: decodeV2Foundation,
		},
		{
			family: "LoaderV2", name: "Finalize",
			newFlux:       func() flux.Instruction { return fluxv2.NewFinalizeInstruction(fluxKey(1)) },
			newFoundation: func() foundation.Instruction { return foundationv2.NewFinalizeInstruction(foundationKey(1)).Build() },
			decodeFlux:    decodeV2Flux, decodeFoundation: decodeV2Foundation,
		},
		{
			family: "LoaderV3", name: "InitializeBuffer",
			newFlux: func() flux.Instruction { return fluxv3.NewInitializeBufferInstruction(fluxKey(1), fluxKey(2)) },
			newFoundation: func() foundation.Instruction {
				return foundationv3.NewInitializeBufferInstruction(foundationKey(1), foundationKey(2)).Build()
			},
			decodeFlux: decodeV3Flux, decodeFoundation: decodeV3Foundation,
		},
		{
			family: "LoaderV3", name: "Write",
			newFlux: func() flux.Instruction { return fluxv3.NewWriteInstruction(42, data, fluxKey(1), fluxKey(2)) },
			newFoundation: func() foundation.Instruction {
				return foundationv3.NewWriteInstruction(foundationKey(1), foundationKey(2), 42, data).Build()
			},
			decodeFlux: decodeV3Flux, decodeFoundation: decodeV3Foundation,
		},
		{
			family: "LoaderV3", name: "DeployWithMaxDataLen",
			newFlux: func() flux.Instruction {
				return fluxv3.NewDeployWithMaxDataLenInstruction(1_000_000, true, fluxKey(8), fluxKey(9), fluxKey(3), fluxKey(1), fluxKey(2))
			},
			newFoundation: func() foundation.Instruction {
				return foundationv3.NewDeployWithMaxDataLenInstruction(foundationKey(8), foundationKey(9), foundationKey(3), foundationKey(1), foundationKey(2), 1_000_000, true).Build()
			},
			decodeFlux: decodeV3Flux, decodeFoundation: decodeV3Foundation,
		},
		{
			family: "LoaderV3", name: "Upgrade",
			newFlux: func() flux.Instruction {
				return fluxv3.NewUpgradeInstruction(true, fluxProgramData, fluxKey(3), fluxKey(1), fluxKey(6), fluxKey(2))
			},
			newFoundation: func() foundation.Instruction {
				return foundationv3.NewUpgradeInstruction(foundationKey(3), foundationKey(1), foundationKey(2), foundationKey(6), true).Build()
			},
			decodeFlux: decodeV3Flux, decodeFoundation: decodeV3Foundation,
		},
		{
			family: "LoaderV3", name: "SetAuthority",
			newFlux: func() flux.Instruction { return fluxv3.NewSetAuthorityInstruction(fluxKey(1), fluxKey(2), fluxKey(7)) },
			newFoundation: func() foundation.Instruction {
				return foundationv3.NewSetBufferAuthorityInstruction(foundationKey(1), foundationKey(2), foundationKey(7)).Build()
			},
			decodeFlux: decodeV3Flux, decodeFoundation: decodeV3Foundation,
		},
		{
			family: "LoaderV3", name: "Close",
			newFlux: func() flux.Instruction { return fluxv3.NewCloseInstruction(true, fluxKey(1), fluxKey(6), fluxKey(2)) },
			newFoundation: func() foundation.Instruction {
				return foundationv3.NewCloseInstruction(foundationKey(1), foundationKey(6), foundationKey(2), true).Build()
			},
			decodeFlux: decodeV3Flux, decodeFoundation: decodeV3Foundation,
		},
		{
			family: "LoaderV3", name: "ExtendProgram",
			newFlux: func() flux.Instruction {
				return fluxv3.NewExtendProgramInstruction(10_240, fluxProgramData, fluxKey(3))
			},
			newFoundation: func() foundation.Instruction {
				return foundationv3.NewExtendProgramInstruction(foundationKey(3), nil, 10_240).Build()
			},
			decodeFlux: decodeV3Flux, decodeFoundation: decodeV3Foundation,
		},
		{
			family: "LoaderV3", name: "SetAuthorityChecked",
			newFlux: func() flux.Instruction {
				return fluxv3.NewSetAuthorityCheckedInstruction(fluxKey(1), fluxKey(2), fluxKey(7))
			},
			newFoundation: func() foundation.Instruction {
				return foundationv3.NewSetBufferAuthorityCheckedInstruction(foundationKey(1), foundationKey(2), foundationKey(7)).Build()
			},
			decodeFlux: decodeV3Flux, decodeFoundation: decodeV3Foundation,
		},
		{
			family: "LoaderV4", name: "Write",
			newFlux: func() flux.Instruction { return fluxv4.NewWriteInstruction(42, data, fluxKey(3), fluxKey(2)) },
			newFoundation: func() foundation.Instruction {
				return foundationv4.NewWriteInstruction(foundationKey(3), foundationKey(2), 42, data).Build()
			},
			decodeFlux: decodeV4Flux, decodeFoundation: decodeV4Foundation,
		},
		{
			family: "LoaderV4", name: "Copy",
			newFlux: func() flux.Instruction {
				return fluxv4.NewCopyInstruction(10, 20, 30, fluxKey(3), fluxKey(2), fluxKey(1))
			},
			newFoundation: func() foundation.Instruction {
				return foundationv4.NewCopyInstruction(foundationKey(3), foundationKey(2), foundationKey(1), 10, 20, 30).Build()
			},
			decodeFlux: decodeV4Flux, decodeFoundation: decodeV4Foundation,
		},
		{
			family: "LoaderV4", name: "SetProgramLength",
			newFlux: func() flux.Instruction {
				return fluxv4.NewSetProgramLengthInstruction(1_000_000, fluxKey(3), fluxKey(2), fluxKey(6))
			},
			newFoundation: func() foundation.Instruction {
				return foundationv4.NewSetProgramLengthInstruction(foundationKey(3), foundationKey(2), foundationKey(6), 1_000_000).Build()
			},
			decodeFlux: decodeV4Flux, decodeFoundation: decodeV4Foundation,
		},
		{
			family: "LoaderV4", name: "Deploy",
			newFlux: func() flux.Instruction { return fluxv4.NewDeployInstruction(fluxKey(3), fluxKey(2)) },
			newFoundation: func() foundation.Instruction {
				return foundationv4.NewDeployInstruction(foundationKey(3), foundationKey(2)).Build()
			},
			decodeFlux: decodeV4Flux, decodeFoundation: decodeV4Foundation,
		},
		{
			family: "LoaderV4", name: "Retract",
			newFlux: func() flux.Instruction { return fluxv4.NewRetractInstruction(fluxKey(3), fluxKey(2)) },
			newFoundation: func() foundation.Instruction {
				return foundationv4.NewRetractInstruction(foundationKey(3), foundationKey(2)).Build()
			},
			decodeFlux: decodeV4Flux, decodeFoundation: decodeV4Foundation,
		},
		{
			family: "LoaderV4", name: "TransferAuthority",
			newFlux: func() flux.Instruction {
				return fluxv4.NewTransferAuthorityInstruction(fluxKey(3), fluxKey(2), fluxKey(7))
			},
			newFoundation: func() foundation.Instruction {
				return foundationv4.NewTransferAuthorityInstruction(foundationKey(3), foundationKey(2), newAuthorityFoundation).Build()
			},
			decodeFlux: decodeV4Flux, decodeFoundation: decodeV4Foundation,
		},
		{
			family: "LoaderV4", name: "Finalize",
			newFlux: func() flux.Instruction { return fluxv4.NewFinalizeInstruction(fluxKey(3), fluxKey(2), fluxKey(4)) },
			newFoundation: func() foundation.Instruction {
				return foundationv4.NewFinalizeInstruction(foundationKey(3), foundationKey(2), foundationKey(4)).Build()
			},
			decodeFlux: decodeV4Flux, decodeFoundation: decodeV4Foundation,
		},
	}

	for index := range benchmarks {
		entry := &benchmarks[index]
		entry.flux = entry.newFlux()
		entry.foundation = entry.newFoundation()
		var err error
		entry.fluxData, err = entry.flux.Data()
		if err != nil {
			tb.Fatalf("flux %s/%s Data: %v", entry.family, entry.name, err)
		}
		entry.foundationData, err = entry.foundation.Data()
		if err != nil {
			tb.Fatalf("foundation %s/%s Data: %v", entry.family, entry.name, err)
		}
		if !bytes.Equal(entry.fluxData, entry.foundationData) {
			tb.Fatalf("%s/%s data differs: flux %x, foundation %x", entry.family, entry.name, entry.fluxData, entry.foundationData)
		}
	}
	return benchmarks
}

func TestLoaderInstructionAllocationsBeatFoundation(t *testing.T) {
	for _, test := range newLoaderInstructionBenchmarks(t) {
		t.Run(test.family+"/"+test.name, func(t *testing.T) {
			fluxConstructor := testing.AllocsPerRun(1_000, func() {
				sinkFluxLoaderInstruction = test.newFlux()
				sinkLoaderData, sinkLoaderErr = sinkFluxLoaderInstruction.Data()
			})
			if sinkLoaderErr != nil {
				t.Fatal(sinkLoaderErr)
			}
			foundationConstructor := testing.AllocsPerRun(1_000, func() {
				sinkFoundationLoaderInstruction = test.newFoundation()
				sinkLoaderData, sinkLoaderErr = sinkFoundationLoaderInstruction.Data()
			})
			if sinkLoaderErr != nil {
				t.Fatal(sinkLoaderErr)
			}
			if fluxConstructor >= foundationConstructor {
				t.Errorf("constructor+Data allocations: flux %.0f, foundation %.0f", fluxConstructor, foundationConstructor)
			}

			fluxData := testing.AllocsPerRun(1_000, func() { sinkLoaderData, sinkLoaderErr = test.flux.Data() })
			foundationData := testing.AllocsPerRun(1_000, func() { sinkLoaderData, sinkLoaderErr = test.foundation.Data() })
			if fluxData >= foundationData {
				t.Errorf("Data allocations: flux %.0f, foundation %.0f", fluxData, foundationData)
			}

			fluxDecode := testing.AllocsPerRun(1_000, func() { test.decodeFlux(test.flux.Accounts(), test.fluxData) })
			if sinkLoaderErr != nil {
				t.Fatal(sinkLoaderErr)
			}
			foundationDecode := testing.AllocsPerRun(1_000, func() { test.decodeFoundation(test.foundation.Accounts(), test.foundationData) })
			if sinkLoaderErr != nil {
				t.Fatal(sinkLoaderErr)
			}
			if fluxDecode >= foundationDecode {
				t.Errorf("Decode allocations: flux %.0f, foundation %.0f", fluxDecode, foundationDecode)
			}
		})
	}
}

func BenchmarkLoaderConstructorAndData(b *testing.B) {
	for _, test := range newLoaderInstructionBenchmarks(b) {
		b.Run(test.family+"/"+test.name, func(b *testing.B) {
			b.Run("Flux", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFluxLoaderInstruction = test.newFlux()
					sinkLoaderData, sinkLoaderErr = sinkFluxLoaderInstruction.Data()
				}
			})
			b.Run("Foundation", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFoundationLoaderInstruction = test.newFoundation()
					sinkLoaderData, sinkLoaderErr = sinkFoundationLoaderInstruction.Data()
				}
			})
		})
	}
}

func BenchmarkLoaderData(b *testing.B) {
	for _, test := range newLoaderInstructionBenchmarks(b) {
		b.Run(test.family+"/"+test.name, func(b *testing.B) {
			b.Run("Flux", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkLoaderData, sinkLoaderErr = test.flux.Data()
				}
			})
			b.Run("Foundation", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkLoaderData, sinkLoaderErr = test.foundation.Data()
				}
			})
		})
	}
}

func BenchmarkLoaderDecode(b *testing.B) {
	for _, test := range newLoaderInstructionBenchmarks(b) {
		b.Run(test.family+"/"+test.name, func(b *testing.B) {
			b.Run("Flux", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					test.decodeFlux(test.flux.Accounts(), test.fluxData)
				}
			})
			b.Run("Foundation", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					test.decodeFoundation(test.foundation.Accounts(), test.foundationData)
				}
			})
		})
	}
}
