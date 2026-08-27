package benchcmp

import (
	"bytes"
	"testing"

	flux "github.com/fluxrpc/solana-go"
	fluxstake "github.com/fluxrpc/solana-go/programs/stake"
	foundation "github.com/solana-foundation/solana-go/v2"
	foundationstake "github.com/solana-foundation/solana-go/v2/programs/stake"
)

var (
	sinkFluxStakeInstruction       flux.Instruction
	sinkFoundationStakeInstruction foundation.Instruction
	sinkFluxStakeDecoded           fluxstake.DecodedInstruction
	sinkFoundationStakeDecoded     *foundationstake.Instruction
	sinkStakeData                  []byte
	sinkStakeErr                   error
)

type stakeInstructionBenchmark struct {
	name           string
	flux           flux.Instruction
	foundation     foundation.Instruction
	fluxData       []byte
	foundationData []byte
	newFlux        func() flux.Instruction
	newFoundation  func() foundation.Instruction
}

func stakeBenchmarkKey(tag byte) (key [32]byte) {
	for index := range key {
		key[index] = tag + byte(index)
	}
	return key
}

func newStakeInstructionBenchmarks(tb testing.TB) []stakeInstructionBenchmark {
	tb.Helper()
	fluxKey := func(tag byte) flux.PublicKey { return flux.PublicKey(stakeBenchmarkKey(tag)) }
	foundationKey := func(tag byte) foundation.PublicKey { return foundation.PublicKey(stakeBenchmarkKey(tag)) }
	mustFlux := func(instruction flux.Instruction, err error) flux.Instruction {
		if err != nil {
			tb.Fatal(err)
		}
		return instruction
	}

	const (
		lamports = uint64(1_234_567_890)
		seed     = "stake-benchmark-seed"
	)
	benchmarks := []stakeInstructionBenchmark{
		{
			name: "Initialize",
			newFlux: func() flux.Instruction {
				return fluxstake.NewInitializeInstruction(
					fluxstake.Authorized{Staker: fluxKey(1), Withdrawer: fluxKey(2)},
					fluxstake.Lockup{Custodian: flux.SystemProgramID}, fluxKey(3),
				)
			},
			newFoundation: func() foundation.Instruction {
				return foundationstake.NewInitializeInstruction(foundationKey(1), foundationKey(2), foundationKey(3)).Build()
			},
		},
		{
			name: "Authorize",
			newFlux: func() flux.Instruction {
				return fluxstake.NewAuthorizeInstruction(fluxKey(4), fluxstake.StakeAuthorizeWithdrawer, fluxKey(3), fluxKey(2), nil)
			},
			newFoundation: func() foundation.Instruction {
				return foundationstake.NewAuthorizeInstruction(foundationKey(4), foundationstake.StakeAuthorizeWithdrawer, foundationKey(3), foundationKey(2)).Build()
			},
		},
		{
			name: "DelegateStake",
			newFlux: func() flux.Instruction {
				return fluxstake.NewDelegateStakeInstruction(fluxKey(5), fluxKey(1), fluxKey(3))
			},
			newFoundation: func() foundation.Instruction {
				return foundationstake.NewDelegateStakeInstruction(foundationKey(5), foundationKey(1), foundationKey(3)).Build()
			},
		},
		{
			name: "Split",
			newFlux: func() flux.Instruction {
				return fluxstake.NewSplitInstruction(lamports, fluxKey(3), fluxKey(6), fluxKey(1))
			},
			newFoundation: func() foundation.Instruction {
				return foundationstake.NewSplitInstruction(lamports, foundationKey(3), foundationKey(6), foundationKey(1)).Build()
			},
		},
		{
			name: "Withdraw",
			newFlux: func() flux.Instruction {
				return fluxstake.NewWithdrawInstruction(lamports, fluxKey(3), fluxKey(7), fluxKey(2), nil)
			},
			newFoundation: func() foundation.Instruction {
				return foundationstake.NewWithdrawInstruction(lamports, foundationKey(3), foundationKey(7), foundationKey(2)).Build()
			},
		},
		{
			name:    "Deactivate",
			newFlux: func() flux.Instruction { return fluxstake.NewDeactivateInstruction(fluxKey(3), fluxKey(1)) },
			newFoundation: func() foundation.Instruction {
				return foundationstake.NewDeactivateInstruction(foundationKey(3), foundationKey(1)).Build()
			},
		},
		{
			name: "SetLockup",
			newFlux: func() flux.Instruction {
				return fluxstake.NewSetLockupInstruction(fluxstake.LockupArgs{}, fluxKey(3), fluxKey(8))
			},
			newFoundation: func() foundation.Instruction {
				return foundationstake.NewSetLockupInstruction(foundationKey(3), foundationKey(8)).Build()
			},
		},
		{
			name:    "Merge",
			newFlux: func() flux.Instruction { return fluxstake.NewMergeInstruction(fluxKey(3), fluxKey(6), fluxKey(1)) },
			newFoundation: func() foundation.Instruction {
				return foundationstake.NewMergeInstruction(foundationKey(3), foundationKey(6), foundationKey(1)).Build()
			},
		},
		{
			name: "AuthorizeWithSeed",
			newFlux: func() flux.Instruction {
				return mustFlux(fluxstake.NewAuthorizeWithSeedInstruction(fluxstake.AuthorizeWithSeedArgs{
					NewAuthorized: fluxKey(4), StakeAuthorize: fluxstake.StakeAuthorizeStaker, AuthoritySeed: seed, AuthorityOwner: fluxKey(9),
				}, fluxKey(3), fluxKey(10), nil))
			},
			newFoundation: func() foundation.Instruction {
				return foundationstake.NewAuthorizeWithSeedInstruction(
					foundationKey(4), foundationstake.StakeAuthorizeStaker, seed, foundationKey(9), foundationKey(3), foundationKey(10),
				).Build()
			},
		},
		{
			name: "InitializeChecked",
			newFlux: func() flux.Instruction {
				return fluxstake.NewInitializeCheckedInstruction(fluxstake.Authorized{Staker: fluxKey(1), Withdrawer: fluxKey(2)}, fluxKey(3))
			},
			newFoundation: func() foundation.Instruction {
				return foundationstake.NewInitializeCheckedInstruction(foundationKey(3), foundationKey(1), foundationKey(2)).Build()
			},
		},
		{
			name: "AuthorizeChecked",
			newFlux: func() flux.Instruction {
				return fluxstake.NewAuthorizeCheckedInstruction(fluxstake.StakeAuthorizeWithdrawer, fluxKey(3), fluxKey(2), fluxKey(4), nil)
			},
			newFoundation: func() foundation.Instruction {
				return foundationstake.NewAuthorizeCheckedInstruction(foundationstake.StakeAuthorizeWithdrawer, foundationKey(3), foundationKey(2), foundationKey(4)).Build()
			},
		},
		{
			name: "AuthorizeCheckedWithSeed",
			newFlux: func() flux.Instruction {
				return mustFlux(fluxstake.NewAuthorizeCheckedWithSeedInstruction(fluxstake.AuthorizeCheckedWithSeedArgs{
					StakeAuthorize: fluxstake.StakeAuthorizeWithdrawer, AuthoritySeed: seed, AuthorityOwner: fluxKey(9),
				}, fluxKey(3), fluxKey(10), fluxKey(4), nil))
			},
			newFoundation: func() foundation.Instruction {
				return foundationstake.NewAuthorizeCheckedWithSeedInstruction(
					foundationstake.StakeAuthorizeWithdrawer, seed, foundationKey(9), foundationKey(3), foundationKey(10), foundationKey(4),
				).Build()
			},
		},
		{
			name: "SetLockupChecked",
			newFlux: func() flux.Instruction {
				return fluxstake.NewSetLockupCheckedInstruction(fluxstake.LockupCheckedArgs{}, fluxKey(3), fluxKey(8), nil)
			},
			newFoundation: func() foundation.Instruction {
				return foundationstake.NewSetLockupCheckedInstruction(foundationKey(3), foundationKey(8)).Build()
			},
		},
		{
			name:          "GetMinimumDelegation",
			newFlux:       func() flux.Instruction { return fluxstake.NewGetMinimumDelegationInstruction() },
			newFoundation: func() foundation.Instruction { return foundationstake.NewGetMinimumDelegationInstruction().Build() },
		},
		{
			name: "DeactivateDelinquent",
			newFlux: func() flux.Instruction {
				return fluxstake.NewDeactivateDelinquentInstruction(fluxKey(3), fluxKey(5), fluxKey(11))
			},
			newFoundation: func() foundation.Instruction {
				return foundationstake.NewDeactivateDelinquentInstruction(foundationKey(3), foundationKey(5), foundationKey(11)).Build()
			},
		},
		{
			name: "Redelegate",
			newFlux: func() flux.Instruction {
				return fluxstake.NewRedelegateInstruction(fluxKey(3), fluxKey(6), fluxKey(5), fluxKey(1))
			},
			newFoundation: func() foundation.Instruction {
				return foundationstake.NewRedelegateInstruction(foundationKey(3), foundationKey(6), foundationKey(5), foundationKey(1)).Build()
			},
		},
		{
			name: "MoveStake",
			newFlux: func() flux.Instruction {
				return fluxstake.NewMoveStakeInstruction(lamports, fluxKey(3), fluxKey(6), fluxKey(1))
			},
			newFoundation: func() foundation.Instruction {
				return foundationstake.NewMoveStakeInstruction(lamports, foundationKey(3), foundationKey(6), foundationKey(1)).Build()
			},
		},
		{
			name: "MoveLamports",
			newFlux: func() flux.Instruction {
				return fluxstake.NewMoveLamportsInstruction(lamports, fluxKey(3), fluxKey(6), fluxKey(1))
			},
			newFoundation: func() foundation.Instruction {
				return foundationstake.NewMoveLamportsInstruction(lamports, foundationKey(3), foundationKey(6), foundationKey(1)).Build()
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

func TestStakeInstructionAllocationsBeatFoundation(t *testing.T) {
	for _, test := range newStakeInstructionBenchmarks(t) {
		t.Run(test.name, func(t *testing.T) {
			fluxConstructor := testing.AllocsPerRun(1_000, func() {
				sinkFluxStakeInstruction = test.newFlux()
				sinkStakeData, sinkStakeErr = sinkFluxStakeInstruction.Data()
			})
			if sinkStakeErr != nil {
				t.Fatal(sinkStakeErr)
			}
			foundationConstructor := testing.AllocsPerRun(1_000, func() {
				sinkFoundationStakeInstruction = test.newFoundation()
				sinkStakeData, sinkStakeErr = sinkFoundationStakeInstruction.Data()
			})
			if sinkStakeErr != nil {
				t.Fatal(sinkStakeErr)
			}
			if fluxConstructor >= foundationConstructor {
				t.Errorf("constructor+Data allocations: flux %.0f, foundation %.0f", fluxConstructor, foundationConstructor)
			}

			fluxData := testing.AllocsPerRun(1_000, func() { sinkStakeData, sinkStakeErr = test.flux.Data() })
			if sinkStakeErr != nil {
				t.Fatal(sinkStakeErr)
			}
			foundationData := testing.AllocsPerRun(1_000, func() { sinkStakeData, sinkStakeErr = test.foundation.Data() })
			if sinkStakeErr != nil {
				t.Fatal(sinkStakeErr)
			}
			if fluxData >= foundationData {
				t.Errorf("Data allocations: flux %.0f, foundation %.0f", fluxData, foundationData)
			}

			fluxDecode := testing.AllocsPerRun(1_000, func() {
				sinkFluxStakeDecoded, sinkStakeErr = fluxstake.DecodeInstruction(test.flux.Accounts(), test.fluxData)
			})
			if sinkStakeErr != nil {
				t.Fatal(sinkStakeErr)
			}
			foundationDecode := testing.AllocsPerRun(1_000, func() {
				sinkFoundationStakeDecoded, sinkStakeErr = foundationstake.DecodeInstruction(test.foundation.Accounts(), test.foundationData)
			})
			if sinkStakeErr != nil {
				t.Fatal(sinkStakeErr)
			}
			if fluxDecode >= foundationDecode {
				t.Errorf("Decode allocations: flux %.0f, foundation %.0f", fluxDecode, foundationDecode)
			}
		})
	}
}

func BenchmarkStakeConstructorAndData(b *testing.B) {
	for _, test := range newStakeInstructionBenchmarks(b) {
		b.Run(test.name, func(b *testing.B) {
			b.Run("Flux", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFluxStakeInstruction = test.newFlux()
					sinkStakeData, sinkStakeErr = sinkFluxStakeInstruction.Data()
				}
			})
			b.Run("Foundation", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFoundationStakeInstruction = test.newFoundation()
					sinkStakeData, sinkStakeErr = sinkFoundationStakeInstruction.Data()
				}
			})
		})
	}
}

func BenchmarkStakeData(b *testing.B) {
	for _, test := range newStakeInstructionBenchmarks(b) {
		b.Run(test.name, func(b *testing.B) {
			b.Run("Flux", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkStakeData, sinkStakeErr = test.flux.Data()
				}
			})
			b.Run("Foundation", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkStakeData, sinkStakeErr = test.foundation.Data()
				}
			})
		})
	}
}

func BenchmarkStakeDecode(b *testing.B) {
	for _, test := range newStakeInstructionBenchmarks(b) {
		b.Run(test.name, func(b *testing.B) {
			b.Run("Flux", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFluxStakeDecoded, sinkStakeErr = fluxstake.DecodeInstruction(test.flux.Accounts(), test.fluxData)
				}
			})
			b.Run("Foundation", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFoundationStakeDecoded, sinkStakeErr = foundationstake.DecodeInstruction(test.foundation.Accounts(), test.foundationData)
				}
			})
		})
	}
}
