package benchcmp

import (
	"bytes"
	"testing"
)

func TestATAInstructionParity(t *testing.T) {
	fixtures := newATAInstructionBenchmarks(t)
	if len(fixtures) != 3 {
		t.Fatalf("parity fixture count = %d, want 3", len(fixtures))
	}
	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			fluxProgramID, gagliardettoProgramID := fixture.flux.ProgramID(), fixture.gagliardetto.ProgramID()
			if !bytes.Equal(fluxProgramID[:], gagliardettoProgramID[:]) {
				t.Errorf("ProgramID = %x, gagliardetto %x", fluxProgramID, gagliardettoProgramID)
			}
			if !bytes.Equal(fixture.fluxData, fixture.gagliardettoData) {
				t.Errorf("Data = %x, gagliardetto %x", fixture.fluxData, fixture.gagliardettoData)
			}
			fluxAccounts, gagliardettoAccounts := fixture.flux.Accounts(), fixture.gagliardetto.Accounts()
			if len(fluxAccounts) != len(gagliardettoAccounts) {
				t.Fatalf("account count = %d, gagliardetto %d", len(fluxAccounts), len(gagliardettoAccounts))
			}
			for index, fluxAccount := range fluxAccounts {
				gagliardettoAccount := gagliardettoAccounts[index]
				if !bytes.Equal(fluxAccount.PublicKey[:], gagliardettoAccount.PublicKey[:]) {
					t.Errorf("account %d public key = %x, gagliardetto %x", index, fluxAccount.PublicKey, gagliardettoAccount.PublicKey)
				}
				if fluxAccount.IsSigner != gagliardettoAccount.IsSigner {
					t.Errorf("account %d IsSigner = %v, gagliardetto %v", index, fluxAccount.IsSigner, gagliardettoAccount.IsSigner)
				}
				if fluxAccount.IsWritable != gagliardettoAccount.IsWritable {
					t.Errorf("account %d IsWritable = %v, gagliardetto %v", index, fluxAccount.IsWritable, gagliardettoAccount.IsWritable)
				}
			}
		})
	}
}
