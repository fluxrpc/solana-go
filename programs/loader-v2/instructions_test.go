package loaderv2

import (
	"bytes"
	"errors"
	"testing"

	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

func testKey(fill byte) (key solana.PublicKey) {
	for i := range key {
		key[i] = fill
	}
	return key
}

func assertAccounts(t *testing.T, got solana.AccountMetaSlice, want ...solana.AccountMeta) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("account count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] == nil || *got[i] != want[i] {
			t.Errorf("account %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestInstructions(t *testing.T) {
	program := testKey(0x11)
	write := NewWriteInstruction(42, []byte{1, 2, 3}, program)
	finalize := NewFinalizeInstruction(program)
	tests := []struct {
		name     string
		typ      InstructionType
		inst     solana.Instruction
		accounts solana.AccountMetaSlice
		want     []byte
	}{
		{
			name: "Write", typ: WriteInstruction, inst: write, accounts: write.AccountMetaSlice,
			want: []byte{0, 0, 0, 0, 42, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3},
		},
		{
			name: "Finalize", typ: FinalizeInstruction, inst: finalize, accounts: finalize.AccountMetaSlice,
			want: []byte{1, 0, 0, 0},
		},
	}

	assertAccounts(t, write.Accounts(), solana.AccountMeta{PublicKey: program, IsWritable: true, IsSigner: true})
	assertAccounts(t, finalize.Accounts(),
		solana.AccountMeta{PublicKey: program, IsWritable: true, IsSigner: true},
		solana.AccountMeta{PublicKey: solana.SysVarRentPubkey},
	)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.inst.ProgramID() != ProgramID {
				t.Errorf("ProgramID = %s, want %s", tc.inst.ProgramID(), ProgramID)
			}
			data, err := tc.inst.Data()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(data, tc.want) {
				t.Fatalf("Data = %x, want %x", data, tc.want)
			}

			decoded, err := DecodeInstruction(tc.accounts, data)
			if err != nil {
				t.Fatal(err)
			}
			if decoded.Type != tc.typ {
				t.Fatalf("Type = %s, want %s", decoded.Type, tc.typ)
			}
			var roundTrip solana.Instruction
			switch tc.typ {
			case WriteInstruction:
				if decoded.Write == nil || decoded.Write.Offset != 42 || !bytes.Equal(decoded.Write.Bytes, []byte{1, 2, 3}) {
					t.Fatalf("decoded Write = %+v", decoded.Write)
				}
				roundTrip = decoded.Write
			case FinalizeInstruction:
				if decoded.Finalize == nil {
					t.Fatal("decoded Finalize is nil")
				}
				roundTrip = decoded.Finalize
			}
			got, err := roundTrip.Data()
			if err != nil || !bytes.Equal(got, tc.want) {
				t.Fatalf("round trip = %x, %v", got, err)
			}
			for n := 0; n < len(data); n++ {
				if _, err := DecodeInstruction(tc.accounts, data[:n]); err == nil {
					t.Errorf("accepted truncation %d/%d", n, len(data))
				}
			}
		})
	}
}

func TestMalformedWriteLength(t *testing.T) {
	data := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	if _, err := DecodeInstruction(nil, data); !errors.Is(err, binary.ErrUnexpectedEOF) {
		t.Fatalf("error = %v", err)
	}
}

func FuzzDecodeInstruction(f *testing.F) {
	f.Add([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	f.Add([]byte{1, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) { _, _ = DecodeInstruction(nil, data) })
}
