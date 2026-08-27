package benchcmp

import (
	"bytes"
	"testing"

	flux "github.com/fluxrpc/solana-go"
	foundation "github.com/solana-foundation/solana-go/v2"
)

// TestLoaderInstructionParity covers every loader-v2, loader-v3, and
// loader-v4 wire discriminator exposed by Foundation v2.
func TestLoaderInstructionParity(t *testing.T) {
	fixtures := newLoaderInstructionBenchmarks(t)
	if len(fixtures) != 17 {
		t.Fatalf("parity fixture count = %d, want 17", len(fixtures))
	}
	for _, fixture := range fixtures {
		t.Run(fixture.family+"/"+fixture.name, func(t *testing.T) {
			loaderParityInstruction(t, fixture.flux, fixture.foundation)
		})
	}
}

func loaderParityInstruction(t *testing.T, fluxInstruction flux.Instruction, foundationInstruction foundation.Instruction) {
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
