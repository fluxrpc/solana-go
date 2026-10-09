package associatedtokenaccount

import (
	"bytes"
	"errors"
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

func ataKey(value byte) solana.PublicKey { var key solana.PublicKey; key[0] = value; return key }

func TestInstructionsRoundTrip(t *testing.T) {
	k1, k2, k3, k4, k5, k6, k7 := ataKey(1), ataKey(2), ataKey(3), ataKey(4), ataKey(5), ataKey(6), ataKey(7)
	tests := []struct {
		name        string
		instruction solana.Instruction
		typ         InstructionType
		accounts    int
	}{
		{"Create", NewCreateInstruction(k1, k2, k3, k4, k5), CreateInstruction, 6},
		{"CreateIdempotent", NewCreateIdempotentInstruction(k1, k2, k3, k4, k5), CreateIdempotentInstruction, 6},
		{"RecoverNested", NewRecoverNestedInstruction(k1, k2, k3, k4, k5, k6, k7), RecoverNestedInstruction, 7},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.instruction.ProgramID() != ProgramID {
				t.Fatalf("ProgramID = %s, want %s", test.instruction.ProgramID(), ProgramID)
			}
			if got := len(test.instruction.Accounts()); got != test.accounts {
				t.Fatalf("accounts = %d, want %d", got, test.accounts)
			}
			data, err := test.instruction.Data()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(data, []byte{uint8(test.typ)}) {
				t.Fatalf("data = %v, want [%d]", data, test.typ)
			}

			got, err := DecodeInstruction(test.instruction.Accounts(), data)
			if err != nil {
				t.Fatal(err)
			}
			if got.Type != test.typ {
				t.Fatalf("type = %s, want %s", got.Type, test.typ)
			}
		})
	}
}

func TestCreateAccounts(t *testing.T) {
	payer, ata, wallet, mint, tokenProgram := ataKey(1), ataKey(2), ataKey(3), ataKey(4), solana.Token2022ProgramID
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
	for _, inst := range []solana.Instruction{
		NewCreateInstruction(payer, ata, wallet, mint, tokenProgram),
		NewCreateIdempotentInstruction(payer, ata, wallet, mint, tokenProgram),
	} {
		got := inst.Accounts()
		for i, w := range want {
			if got[i].PublicKey != w.key || got[i].IsSigner != w.signer || got[i].IsWritable != w.writable {
				t.Fatalf("account %d = %+v, want %+v", i, got[i], w)
			}
		}
	}
}

func TestRecoverNestedAccounts(t *testing.T) {
	keys := []solana.PublicKey{ataKey(1), ataKey(2), ataKey(3), ataKey(4), ataKey(5), ataKey(6), ataKey(7)}
	got := NewRecoverNestedInstruction(keys[0], keys[1], keys[2], keys[3], keys[4], keys[5], keys[6]).Accounts()
	flags := []struct{ signer, writable bool }{
		{false, true}, {false, false}, {false, true}, {false, false}, {false, false}, {true, true}, {false, false},
	}
	for i, f := range flags {
		if got[i].PublicKey != keys[i] || got[i].IsSigner != f.signer || got[i].IsWritable != f.writable {
			t.Fatalf("account %d = %+v, want key %s signer=%v writable=%v", i, got[i], keys[i], f.signer, f.writable)
		}
	}
}

func TestDecodeLegacyEmptyDataIsCreate(t *testing.T) {
	got, err := DecodeInstruction(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != CreateInstruction || got.Create == nil {
		t.Fatalf("decoded = %#v", got)
	}
}

func TestDecodeInstructionErrors(t *testing.T) {
	if _, err := DecodeInstruction(nil, []byte{3}); !errors.Is(err, ErrUnknownInstruction) {
		t.Fatalf("err = %v, want ErrUnknownInstruction", err)
	}
}
