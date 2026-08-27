package benchcmp

import (
	"bytes"
	"testing"

	flux "github.com/fluxrpc/solana-go"
	fluxvote "github.com/fluxrpc/solana-go/programs/vote"
	foundation "github.com/solana-foundation/solana-go/v2"
	foundationvote "github.com/solana-foundation/solana-go/v2/programs/vote"
)

var (
	sinkFluxVoteInstruction       flux.Instruction
	sinkFoundationVoteInstruction foundation.Instruction
	sinkFluxVoteDecoded           fluxvote.DecodedInstruction
	sinkFoundationVoteDecoded     *foundationvote.Instruction
	sinkVoteData                  []byte
	sinkVoteErr                   error
)

type voteInstructionBenchmark struct {
	name                 string
	flux                 flux.Instruction
	foundation           foundation.Instruction
	fluxData             []byte
	foundationData       []byte
	newFlux              func() flux.Instruction
	newFoundation        func() foundation.Instruction
	skipAccountParity    bool
	allocationComparable bool
}

func voteBenchmarkKey(tag byte) (key [32]byte) {
	for index := range key {
		key[index] = tag + byte(index)
	}
	return key
}

func voteBenchmarkHash(tag byte) (hash [32]byte) {
	for index := range hash {
		hash[index] = tag ^ byte(index)
	}
	return hash
}

func newVoteInstructionBenchmarks(tb testing.TB) []voteInstructionBenchmark {
	tb.Helper()
	fluxKey := func(tag byte) flux.PublicKey { return flux.PublicKey(voteBenchmarkKey(tag)) }
	foundationKey := func(tag byte) foundation.PublicKey { return foundation.PublicKey(voteBenchmarkKey(tag)) }
	fluxHash := func(tag byte) flux.Hash { return flux.Hash(voteBenchmarkHash(tag)) }
	foundationHash := func(tag byte) foundation.Hash { return foundation.Hash(voteBenchmarkHash(tag)) }
	mustFlux := func(instruction flux.Instruction, err error) flux.Instruction {
		if err != nil {
			tb.Fatal(err)
		}
		return instruction
	}

	const seed = "vote-benchmark-seed"
	root, timestamp := uint64(99), int64(-1_234_567)
	slots := []uint64{100, 101, 105}
	fluxLockouts := []fluxvote.Lockout{{Slot: 100, ConfirmationCount: 3}, {Slot: 101, ConfirmationCount: 2}, {Slot: 105, ConfirmationCount: 1}}
	foundationLockouts := []foundationvote.Lockout{{Slot: 100, ConfirmationCount: 3}, {Slot: 101, ConfirmationCount: 2}, {Slot: 105, ConfirmationCount: 1}}
	fluxUpdate := fluxvote.VoteStateUpdate{Lockouts: fluxLockouts, Root: &root, Hash: fluxHash(0x31), Timestamp: &timestamp}
	foundationUpdate := foundationvote.VoteStateUpdate{Lockouts: foundationLockouts, Root: &root, Hash: foundationHash(0x31), Timestamp: &timestamp}
	fluxSync := fluxvote.TowerSyncUpdate{Lockouts: fluxLockouts, Root: &root, Hash: fluxHash(0x31), Timestamp: &timestamp, BlockID: fluxHash(0x32)}
	foundationSync := foundationvote.TowerSyncUpdate{Lockouts: foundationLockouts, Root: &root, Hash: foundationHash(0x31), Timestamp: &timestamp, BlockID: foundationHash(0x32)}
	fluxAuthorization := fluxvote.VoteAuthorize{Kind: fluxvote.VoteAuthorizeVoter}
	foundationAuthorization := foundationvote.VoteAuthorize{Kind: foundationvote.VoteAuthorizeVoter}
	fluxSeedArgs := fluxvote.VoteAuthorizeWithSeedArgs{
		AuthorizationType: fluxAuthorization, CurrentAuthorityDerivedKeyOwner: fluxKey(7),
		CurrentAuthorityDerivedKeySeed: seed, NewAuthority: fluxKey(8),
	}
	foundationSeedArgs := foundationvote.VoteAuthorizeWithSeedArgs{
		AuthorizationType: foundationAuthorization, CurrentAuthorityDerivedKeyOwner: foundationKey(7),
		CurrentAuthorityDerivedKeySeed: seed, NewAuthority: foundationKey(8),
	}
	fluxCheckedSeedArgs := fluxvote.VoteAuthorizeCheckedWithSeedArgs{
		AuthorizationType: fluxAuthorization, CurrentAuthorityDerivedKeyOwner: fluxKey(7), CurrentAuthorityDerivedKeySeed: seed,
	}
	foundationCheckedSeedArgs := foundationvote.VoteAuthorizeCheckedWithSeedArgs{
		AuthorizationType: foundationAuthorization, CurrentAuthorityDerivedKeyOwner: foundationKey(7), CurrentAuthorityDerivedKeySeed: seed,
	}

	var fluxBLSKey [fluxvote.BLSPublicKeyCompressedSize]byte
	var fluxBLSProof [fluxvote.BLSProofOfPossessionCompressedSize]byte
	var foundationBLSKey [foundationvote.BLS_PUBLIC_KEY_COMPRESSED_SIZE]byte
	var foundationBLSProof [foundationvote.BLS_PROOF_OF_POSSESSION_COMPRESSED_SIZE]byte
	for index := range fluxBLSKey {
		fluxBLSKey[index] = 0x44 + byte(index)
		foundationBLSKey[index] = fluxBLSKey[index]
	}
	for index := range fluxBLSProof {
		fluxBLSProof[index] = 0x55 ^ byte(index)
		foundationBLSProof[index] = fluxBLSProof[index]
	}
	fluxInitV2 := fluxvote.VoteInitV2{
		NodePubkey: fluxKey(1), AuthorizedVoter: fluxKey(2), AuthorizedVoterBLSPubkey: fluxBLSKey,
		AuthorizedVoterBLSProofOfPossession: fluxBLSProof, AuthorizedWithdrawer: fluxKey(3),
		InflationRewardsCommissionBps: 500, BlockRevenueCommissionBps: 1_000,
	}
	foundationInitV2 := foundationvote.VoteInitV2{
		NodePubkey: foundationKey(1), AuthorizedVoter: foundationKey(2), AuthorizedVoterBLSPubkey: foundationBLSKey,
		AuthorizedVoterBLSProofOfPossession: foundationBLSProof, AuthorizedWithdrawer: foundationKey(3),
		InflationRewardsCommissionBps: 500, BlockRevenueCommissionBps: 1_000,
	}

	benchmarks := []voteInstructionBenchmark{
		{
			name: "InitializeAccount", allocationComparable: true,
			newFlux: func() flux.Instruction {
				return fluxvote.NewInitializeAccountInstruction(fluxKey(1), fluxKey(2), fluxKey(3), 7, fluxKey(4))
			},
			newFoundation: func() foundation.Instruction {
				return foundationvote.NewInitializeAccountInstruction(foundationKey(1), foundationKey(2), foundationKey(3), 7, foundationKey(4)).Build()
			},
		},
		{
			name: "Authorize", allocationComparable: true,
			newFlux: func() flux.Instruction {
				return fluxvote.NewAuthorizeInstruction(fluxKey(8), fluxAuthorization, fluxKey(4), fluxKey(2))
			},
			newFoundation: func() foundation.Instruction {
				return foundationvote.NewAuthorizeInstruction(foundationKey(8), foundationvote.VoteAuthorizeVoter, foundationKey(4), foundationKey(2)).Build()
			},
		},
		{
			name: "Vote", allocationComparable: true,
			newFlux: func() flux.Instruction {
				return fluxvote.NewVoteInstruction(slots, fluxHash(0x31), &timestamp, fluxKey(4), fluxKey(2))
			},
			newFoundation: func() foundation.Instruction {
				return foundationvote.NewVoteInstruction(slots, foundationHash(0x31), &timestamp, foundationKey(4), foundationKey(2)).Build()
			},
		},
		{
			name: "Withdraw", allocationComparable: true,
			newFlux: func() flux.Instruction {
				return fluxvote.NewWithdrawInstruction(1_234_567_890, fluxKey(4), fluxKey(5), fluxKey(3))
			},
			newFoundation: func() foundation.Instruction {
				return foundationvote.NewWithdrawInstruction(1_234_567_890, foundationKey(4), foundationKey(5), foundationKey(3)).Build()
			},
		},
		{
			name: "UpdateValidatorIdentity", allocationComparable: true,
			newFlux: func() flux.Instruction {
				return fluxvote.NewUpdateValidatorIdentityInstruction(fluxKey(4), fluxKey(1), fluxKey(3))
			},
			newFoundation: func() foundation.Instruction {
				return foundationvote.NewUpdateValidatorIdentityInstruction(foundationKey(4), foundationKey(1), foundationKey(3)).Build()
			},
		},
		{
			name: "UpdateCommission", allocationComparable: true,
			newFlux: func() flux.Instruction { return fluxvote.NewUpdateCommissionInstruction(9, fluxKey(4), fluxKey(3)) },
			newFoundation: func() foundation.Instruction {
				return foundationvote.NewUpdateCommissionInstruction(9, foundationKey(4), foundationKey(3)).Build()
			},
		},
		{
			name: "VoteSwitch", allocationComparable: true,
			newFlux: func() flux.Instruction {
				return fluxvote.NewVoteSwitchInstruction(slots, fluxHash(0x31), &timestamp, fluxHash(0x33), fluxKey(4), fluxKey(2))
			},
			newFoundation: func() foundation.Instruction {
				return foundationvote.NewVoteSwitchInstruction(slots, foundationHash(0x31), &timestamp, foundationHash(0x33), foundationKey(4), foundationKey(2)).Build()
			},
		},
		{
			name: "AuthorizeChecked", allocationComparable: true,
			newFlux: func() flux.Instruction {
				return fluxvote.NewAuthorizeCheckedInstruction(fluxAuthorization, fluxKey(4), fluxKey(2), fluxKey(8))
			},
			newFoundation: func() foundation.Instruction {
				return foundationvote.NewAuthorizeCheckedInstruction(foundationvote.VoteAuthorizeVoter, foundationKey(4), foundationKey(2), foundationKey(8)).Build()
			},
		},
		{
			name: "UpdateVoteState", allocationComparable: true,
			newFlux: func() flux.Instruction {
				return fluxvote.NewUpdateVoteStateInstruction(fluxUpdate, fluxKey(4), fluxKey(2))
			},
			newFoundation: func() foundation.Instruction {
				return foundationvote.NewUpdateVoteStateInstruction(foundationUpdate, foundationKey(4), foundationKey(2)).Build()
			},
		},
		{
			name: "UpdateVoteStateSwitch", allocationComparable: true,
			newFlux: func() flux.Instruction {
				return fluxvote.NewUpdateVoteStateSwitchInstruction(fluxUpdate, fluxHash(0x33), fluxKey(4), fluxKey(2))
			},
			newFoundation: func() foundation.Instruction {
				return foundationvote.NewUpdateVoteStateSwitchInstruction(foundationUpdate, foundationHash(0x33), foundationKey(4), foundationKey(2)).Build()
			},
		},
		{
			name: "AuthorizeWithSeed", allocationComparable: true,
			newFlux: func() flux.Instruction {
				return mustFlux(fluxvote.NewAuthorizeWithSeedInstruction(fluxSeedArgs, fluxKey(4), fluxKey(6)))
			},
			newFoundation: func() foundation.Instruction {
				return foundationvote.NewAuthorizeWithSeedInstruction(foundationSeedArgs, foundationKey(4), foundationKey(6)).Build()
			},
		},
		{
			name: "AuthorizeCheckedWithSeed", allocationComparable: true,
			newFlux: func() flux.Instruction {
				return mustFlux(fluxvote.NewAuthorizeCheckedWithSeedInstruction(fluxCheckedSeedArgs, fluxKey(4), fluxKey(6), fluxKey(8)))
			},
			newFoundation: func() foundation.Instruction {
				return foundationvote.NewAuthorizeCheckedWithSeedInstruction(foundationCheckedSeedArgs, foundationKey(4), foundationKey(6), foundationKey(8)).Build()
			},
		},
		{
			name: "CompactUpdateVoteState", allocationComparable: true,
			newFlux: func() flux.Instruction {
				return fluxvote.NewCompactUpdateVoteStateInstruction(fluxUpdate, fluxKey(4), fluxKey(2))
			},
			newFoundation: func() foundation.Instruction {
				return foundationvote.NewCompactUpdateVoteStateInstruction(foundationUpdate, foundationKey(4), foundationKey(2)).Build()
			},
		},
		{
			name: "CompactUpdateVoteStateSwitch", allocationComparable: true,
			newFlux: func() flux.Instruction {
				return fluxvote.NewCompactUpdateVoteStateSwitchInstruction(fluxUpdate, fluxHash(0x33), fluxKey(4), fluxKey(2))
			},
			newFoundation: func() foundation.Instruction {
				return foundationvote.NewCompactUpdateVoteStateSwitchInstruction(foundationUpdate, foundationHash(0x33), foundationKey(4), foundationKey(2)).Build()
			},
		},
		{
			name: "TowerSync", allocationComparable: true,
			newFlux: func() flux.Instruction { return fluxvote.NewTowerSyncInstruction(fluxSync, fluxKey(4), fluxKey(2)) },
			newFoundation: func() foundation.Instruction {
				return foundationvote.NewTowerSyncInstruction(foundationSync, foundationKey(4), foundationKey(2)).Build()
			},
		},
		{
			name: "TowerSyncSwitch", allocationComparable: true,
			newFlux: func() flux.Instruction {
				return fluxvote.NewTowerSyncSwitchInstruction(fluxSync, fluxHash(0x33), fluxKey(4), fluxKey(2))
			},
			newFoundation: func() foundation.Instruction {
				return foundationvote.NewTowerSyncSwitchInstruction(foundationSync, foundationHash(0x33), foundationKey(4), foundationKey(2)).Build()
			},
		},
		{
			name: "InitializeAccountV2_StaleFoundationAccounts", skipAccountParity: true,
			newFlux: func() flux.Instruction {
				return fluxvote.NewInitializeAccountV2Instruction(fluxInitV2, fluxKey(4), fluxKey(5), fluxKey(6))
			},
			newFoundation: func() foundation.Instruction {
				return foundationvote.NewInitializeAccountV2Instruction(foundationInitV2, foundationKey(4)).Build()
			},
		},
		{
			name: "DepositDelegatorRewards", allocationComparable: true,
			newFlux: func() flux.Instruction {
				return fluxvote.NewDepositDelegatorRewardsInstruction(987_654_321, fluxKey(4), fluxKey(5))
			},
			newFoundation: func() foundation.Instruction {
				return foundationvote.NewDepositDelegatorRewardsInstruction(987_654_321, foundationKey(4), foundationKey(5)).Build()
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

func TestVoteInstructionAllocationsBeatFoundation(t *testing.T) {
	for _, test := range newVoteInstructionBenchmarks(t) {
		if !test.allocationComparable {
			continue
		}
		t.Run(test.name, func(t *testing.T) {
			fluxConstructor := testing.AllocsPerRun(1_000, func() {
				sinkFluxVoteInstruction = test.newFlux()
				sinkVoteData, sinkVoteErr = sinkFluxVoteInstruction.Data()
			})
			foundationConstructor := testing.AllocsPerRun(1_000, func() {
				sinkFoundationVoteInstruction = test.newFoundation()
				sinkVoteData, sinkVoteErr = sinkFoundationVoteInstruction.Data()
			})
			if fluxConstructor >= foundationConstructor {
				t.Errorf("constructor+Data allocations: flux %.0f, foundation %.0f", fluxConstructor, foundationConstructor)
			}

			fluxData := testing.AllocsPerRun(1_000, func() { sinkVoteData, sinkVoteErr = test.flux.Data() })
			foundationData := testing.AllocsPerRun(1_000, func() { sinkVoteData, sinkVoteErr = test.foundation.Data() })
			if fluxData >= foundationData {
				t.Errorf("Data allocations: flux %.0f, foundation %.0f", fluxData, foundationData)
			}

			fluxDecode := testing.AllocsPerRun(1_000, func() {
				sinkFluxVoteDecoded, sinkVoteErr = fluxvote.DecodeInstruction(test.flux.Accounts(), test.fluxData)
			})
			foundationDecode := testing.AllocsPerRun(1_000, func() {
				sinkFoundationVoteDecoded, sinkVoteErr = foundationvote.DecodeInstruction(test.foundation.Accounts(), test.foundationData)
			})
			if fluxDecode >= foundationDecode {
				t.Errorf("Decode allocations: flux %.0f, foundation %.0f", fluxDecode, foundationDecode)
			}
		})
	}
}

func BenchmarkVoteConstructorAndData(b *testing.B) {
	for _, test := range newVoteInstructionBenchmarks(b) {
		b.Run(test.name, func(b *testing.B) {
			b.Run("Flux", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFluxVoteInstruction = test.newFlux()
					sinkVoteData, sinkVoteErr = sinkFluxVoteInstruction.Data()
				}
			})
			b.Run("Foundation", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFoundationVoteInstruction = test.newFoundation()
					sinkVoteData, sinkVoteErr = sinkFoundationVoteInstruction.Data()
				}
			})
		})
	}
}

func BenchmarkVoteData(b *testing.B) {
	for _, test := range newVoteInstructionBenchmarks(b) {
		b.Run(test.name, func(b *testing.B) {
			b.Run("Flux", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkVoteData, sinkVoteErr = test.flux.Data()
				}
			})
			b.Run("Foundation", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkVoteData, sinkVoteErr = test.foundation.Data()
				}
			})
		})
	}
}

func BenchmarkVoteDecode(b *testing.B) {
	for _, test := range newVoteInstructionBenchmarks(b) {
		b.Run(test.name, func(b *testing.B) {
			b.Run("Flux", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFluxVoteDecoded, sinkVoteErr = fluxvote.DecodeInstruction(test.flux.Accounts(), test.fluxData)
				}
			})
			b.Run("Foundation", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFoundationVoteDecoded, sinkVoteErr = foundationvote.DecodeInstruction(test.foundation.Accounts(), test.foundationData)
				}
			})
		})
	}
}
