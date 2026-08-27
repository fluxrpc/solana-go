package benchcmp

import (
	"bytes"
	"testing"

	flux "github.com/fluxrpc/solana-go"
	fluxvote "github.com/fluxrpc/solana-go/programs/vote"
	foundation "github.com/solana-foundation/solana-go/v2"
)

func TestVoteInstructionParity(t *testing.T) {
	fixtures := newVoteInstructionBenchmarks(t)
	if len(fixtures) != 18 {
		t.Fatalf("parity fixture count = %d, want 18", len(fixtures))
	}
	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			voteParityInstruction(t, fixture.flux, fixture.foundation, fixture.skipAccountParity)
		})
	}
}

func voteParityInstruction(t *testing.T, fluxInstruction flux.Instruction, foundationInstruction foundation.Instruction, skipAccounts bool) {
	t.Helper()
	fluxProgramID := fluxInstruction.ProgramID()
	foundationProgramID := foundationInstruction.ProgramID()
	if !bytes.Equal(fluxProgramID[:], foundationProgramID[:]) {
		t.Errorf("ProgramID = %x, foundation %x", fluxProgramID, foundationProgramID)
	}
	fluxData, err := fluxInstruction.Data()
	if err != nil {
		t.Fatalf("flux Data: %v", err)
	}
	foundationData, err := foundationInstruction.Data()
	if err != nil {
		t.Fatalf("foundation Data: %v", err)
	}
	if !bytes.Equal(fluxData, foundationData) {
		t.Errorf("Data = %x, foundation %x", fluxData, foundationData)
	}
	if skipAccounts {
		return
	}

	fluxAccounts := fluxInstruction.Accounts()
	foundationAccounts := foundationInstruction.Accounts()
	if len(fluxAccounts) != len(foundationAccounts) {
		t.Fatalf("account count = %d, foundation %d", len(fluxAccounts), len(foundationAccounts))
	}
	for index, fluxAccount := range fluxAccounts {
		foundationAccount := foundationAccounts[index]
		if fluxAccount == nil || foundationAccount == nil {
			if fluxAccount != nil || foundationAccount != nil {
				t.Errorf("account %d nil mismatch", index)
			}
			continue
		}
		if !bytes.Equal(fluxAccount.PublicKey[:], foundationAccount.PublicKey[:]) {
			t.Errorf("account %d public key = %x, foundation %x", index, fluxAccount.PublicKey, foundationAccount.PublicKey)
		}
		if fluxAccount.IsSigner != foundationAccount.IsSigner {
			t.Errorf("account %d IsSigner = %v, foundation %v", index, fluxAccount.IsSigner, foundationAccount.IsSigner)
		}
		if fluxAccount.IsWritable != foundationAccount.IsWritable {
			t.Errorf("account %d IsWritable = %v, foundation %v", index, fluxAccount.IsWritable, foundationAccount.IsWritable)
		}
	}
}

// Variants 17 and 18 are checked against the current official Vote interface
// wire goldens. Foundation encodes their CommissionKind as a stale u8 rather
// than the current u32, so both are absent from comparative timings.
func TestVoteCurrentOnlyInstructionGoldens(t *testing.T) {
	key := func(tag byte) flux.PublicKey { return flux.PublicKey(voteBenchmarkKey(tag)) }
	collectorData, err := fluxvote.NewUpdateCommissionCollectorInstruction(
		fluxvote.CommissionKindBlockRevenue, key(1), key(2), key(3),
	).Data()
	if err != nil {
		t.Fatal(err)
	}
	if want := []byte{17, 0, 0, 0, 1, 0, 0, 0}; !bytes.Equal(collectorData, want) {
		t.Errorf("UpdateCommissionCollector = %x, want %x", collectorData, want)
	}

	bpsData, err := fluxvote.NewUpdateCommissionBpsInstruction(
		0x1234, fluxvote.CommissionKindBlockRevenue, key(1), key(2),
	).Data()
	if err != nil {
		t.Fatal(err)
	}
	if want := []byte{18, 0, 0, 0, 0x34, 0x12, 1, 0, 0, 0}; !bytes.Equal(bpsData, want) {
		t.Errorf("UpdateCommissionBps = %x, want %x", bpsData, want)
	}

}
