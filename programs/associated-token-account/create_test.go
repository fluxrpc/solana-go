package associatedtokenaccount

import (
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

func TestCreateBuild(t *testing.T) {
	payer := solana.NewWallet().PublicKey()
	wallet := solana.NewWallet().PublicKey()
	mint := solana.NewWallet().PublicKey()

	for _, tokenProgram := range []solana.PublicKey{solana.TokenProgramID, solana.Token2022ProgramID} {
		ix := NewCreateInstruction(payer, wallet, mint, tokenProgram).Build()

		if ix.ProgramID() != solana.SPLAssociatedTokenAccountProgramID {
			t.Fatalf("program id = %s", ix.ProgramID())
		}
		data, err := ix.Data()
		if err != nil || len(data) != 1 || data[0] != 1 {
			t.Fatalf("data = %v, err = %v", data, err)
		}

		ata, _, err := wallet.FindAssociatedTokenAddressWithProgram(mint, tokenProgram)
		if err != nil {
			t.Fatal(err)
		}
		want := []struct {
			key              solana.PublicKey
			signer, writable bool
		}{
			{payer, true, true},
			{ata, false, true},
			{wallet, false, false},
			{mint, false, false},
			{solana.SystemProgramID, false, false},
			{tokenProgram, false, false},
		}
		got := ix.Accounts()
		if len(got) != len(want) {
			t.Fatalf("accounts = %d, want %d", len(got), len(want))
		}
		for i, w := range want {
			if got[i].PublicKey != w.key || got[i].IsSigner != w.signer || got[i].IsWritable != w.writable {
				t.Fatalf("account %d = %+v, want %+v", i, got[i], w)
			}
		}
	}
}

func TestBuilderDefaultsToTokenProgram(t *testing.T) {
	if NewCreateInstructionBuilder().TokenProgram != solana.TokenProgramID {
		t.Fatal("default token program should be SPL Token")
	}
}

func TestValidate(t *testing.T) {
	if _, err := NewCreateInstructionBuilder().ValidateAndBuild(); err == nil {
		t.Fatal("expected error for empty fields")
	}
	k := solana.NewWallet().PublicKey()
	if _, err := NewCreateInstruction(k, k, k, solana.TokenProgramID).ValidateAndBuild(); err != nil {
		t.Fatal(err)
	}
}
