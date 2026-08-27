package loaderv4

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
	program, authority, source := testKey(1), testKey(2), testKey(3)
	recipient, newAuthority, nextVersion := testKey(4), testKey(5), testKey(6)
	write := NewWriteInstruction(42, []byte{1, 2, 3}, program, authority)
	copyInstruction := NewCopyInstruction(10, 20, 30, program, authority, source)
	setLength := NewSetProgramLengthInstruction(100_000, program, authority, recipient)
	deploy := NewDeployInstruction(program, authority)
	retract := NewRetractInstruction(program, authority)
	transfer := NewTransferAuthorityInstruction(program, authority, newAuthority)
	finalize := NewFinalizeInstruction(program, authority, nextVersion)

	assertAccounts(t, write.Accounts(),
		solana.AccountMeta{PublicKey: program, IsWritable: true},
		solana.AccountMeta{PublicKey: authority, IsSigner: true},
	)
	assertAccounts(t, copyInstruction.Accounts(),
		solana.AccountMeta{PublicKey: program, IsWritable: true},
		solana.AccountMeta{PublicKey: authority, IsSigner: true},
		solana.AccountMeta{PublicKey: source},
	)
	assertAccounts(t, setLength.Accounts(),
		solana.AccountMeta{PublicKey: program, IsWritable: true},
		solana.AccountMeta{PublicKey: authority, IsSigner: true},
		solana.AccountMeta{PublicKey: recipient, IsWritable: true},
	)
	assertAccounts(t, deploy.Accounts(),
		solana.AccountMeta{PublicKey: program, IsWritable: true},
		solana.AccountMeta{PublicKey: authority, IsSigner: true},
	)
	assertAccounts(t, retract.Accounts(),
		solana.AccountMeta{PublicKey: program, IsWritable: true},
		solana.AccountMeta{PublicKey: authority, IsSigner: true},
	)
	assertAccounts(t, transfer.Accounts(),
		solana.AccountMeta{PublicKey: program, IsWritable: true},
		solana.AccountMeta{PublicKey: authority, IsSigner: true},
		solana.AccountMeta{PublicKey: newAuthority, IsSigner: true},
	)
	assertAccounts(t, finalize.Accounts(),
		solana.AccountMeta{PublicKey: program, IsWritable: true},
		solana.AccountMeta{PublicKey: authority, IsSigner: true},
		solana.AccountMeta{PublicKey: nextVersion},
	)

	tests := []struct {
		name string
		typ  InstructionType
		inst solana.Instruction
		want []byte
	}{
		{"Write", WriteInstruction, write, []byte{0, 0, 0, 0, 42, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3}},
		{"Copy", CopyInstruction, copyInstruction, []byte{1, 0, 0, 0, 10, 0, 0, 0, 20, 0, 0, 0, 30, 0, 0, 0}},
		{"SetProgramLength", SetProgramLengthInstruction, setLength, []byte{2, 0, 0, 0, 0xa0, 0x86, 1, 0}},
		{"Deploy", DeployInstruction, deploy, []byte{3, 0, 0, 0}},
		{"Retract", RetractInstruction, retract, []byte{4, 0, 0, 0}},
		{"TransferAuthority", TransferAuthorityInstruction, transfer, []byte{5, 0, 0, 0}},
		{"Finalize", FinalizeInstruction, finalize, []byte{6, 0, 0, 0}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.inst.ProgramID() != ProgramID {
				t.Errorf("ProgramID = %s, want %s", tc.inst.ProgramID(), ProgramID)
			}
			data, err := tc.inst.Data()
			if err != nil || !bytes.Equal(data, tc.want) {
				t.Fatalf("Data = %x, %v; want %x", data, err, tc.want)
			}
			decoded, err := DecodeInstruction(tc.inst.Accounts(), data)
			if err != nil || decoded.Type != tc.typ {
				t.Fatalf("decoded = %+v, %v", decoded, err)
			}
			var roundTrip solana.Instruction
			switch tc.typ {
			case WriteInstruction:
				if decoded.Write == nil || decoded.Write.Offset != 42 || !bytes.Equal(decoded.Write.Bytes, []byte{1, 2, 3}) {
					t.Fatalf("Write = %+v", decoded.Write)
				}
				roundTrip = decoded.Write
			case CopyInstruction:
				if decoded.Copy == nil || decoded.Copy.DestinationOffset != 10 || decoded.Copy.SourceOffset != 20 || decoded.Copy.Length != 30 {
					t.Fatalf("Copy = %+v", decoded.Copy)
				}
				roundTrip = decoded.Copy
			case SetProgramLengthInstruction:
				if decoded.SetProgramLength == nil || decoded.SetProgramLength.NewSize != 100_000 {
					t.Fatalf("SetProgramLength = %+v", decoded.SetProgramLength)
				}
				roundTrip = decoded.SetProgramLength
			case DeployInstruction:
				roundTrip = decoded.Deploy
			case RetractInstruction:
				roundTrip = decoded.Retract
			case TransferAuthorityInstruction:
				roundTrip = decoded.TransferAuthority
			case FinalizeInstruction:
				roundTrip = decoded.Finalize
			}
			got, err := roundTrip.Data()
			if err != nil || !bytes.Equal(got, tc.want) {
				t.Fatalf("round trip = %x, %v", got, err)
			}
			for n := 0; n < len(data); n++ {
				if _, err := DecodeInstruction(tc.inst.Accounts(), data[:n]); err == nil {
					t.Errorf("accepted truncation %d/%d", n, len(data))
				}
			}
		})
	}
}

func TestOptionalAccountLayouts(t *testing.T) {
	program, authority, source := testKey(11), testKey(12), testKey(13)
	assertAccounts(t, NewSetProgramLengthWithoutRecipientInstruction(1, program, authority).Accounts(),
		solana.AccountMeta{PublicKey: program, IsWritable: true},
		solana.AccountMeta{PublicKey: authority, IsSigner: true},
	)
	assertAccounts(t, NewDeployFromSourceInstruction(program, authority, source).Accounts(),
		solana.AccountMeta{PublicKey: program, IsWritable: true},
		solana.AccountMeta{PublicKey: authority, IsSigner: true},
		solana.AccountMeta{PublicKey: source, IsWritable: true},
	)
}

func TestMalformedWriteLength(t *testing.T) {
	data := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	if _, err := DecodeInstruction(nil, data); !errors.Is(err, binary.ErrUnexpectedEOF) {
		t.Fatalf("error = %v", err)
	}
}

func FuzzDecodeInstruction(f *testing.F) {
	for typ := WriteInstruction; typ <= FinalizeInstruction; typ++ {
		f.Add([]byte{byte(typ), 0, 0, 0})
	}
	f.Fuzz(func(t *testing.T, data []byte) { _, _ = DecodeInstruction(nil, data) })
}
