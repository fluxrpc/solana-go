package benchcmp

import (
	"bytes"
	"testing"

	flux "github.com/fluxrpc/solana-go"
	foundation "github.com/solana-foundation/solana-go/v2"
)

// TestStakeInstructionParity compares every variant exposed by both the
// current Flux client and Foundation v2. Official interface 4.4.0 and the
// pinned Foundation module both currently expose discriminators 0 through 17.
func TestStakeInstructionParity(t *testing.T) {
	fixtures := newStakeInstructionBenchmarks(t)
	if len(fixtures) != 18 {
		t.Fatalf("parity fixture count = %d, want 18", len(fixtures))
	}
	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) { stakeParityInstruction(t, fixture.flux, fixture.foundation) })
	}
}

func stakeParityInstruction(t *testing.T, fluxInstruction flux.Instruction, foundationInstruction foundation.Instruction) {
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
