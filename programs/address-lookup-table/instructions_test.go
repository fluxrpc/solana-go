package addresslookuptable

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"

	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

func lookupKey(fill byte) (key solana.PublicKey) {
	for index := range key {
		key[index] = fill
	}
	return key
}

func lookupData(t *testing.T, instruction solana.Instruction) []byte {
	t.Helper()
	if instruction.ProgramID() != ProgramID {
		t.Fatalf("ProgramID = %s, want %s", instruction.ProgramID(), ProgramID)
	}
	data, err := instruction.Data()
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func assertLookupAccounts(t *testing.T, got solana.AccountMetaSlice, want ...solana.AccountMeta) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("account count = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index] == nil || *got[index] != want[index] {
			t.Errorf("account %d = %#v, want %#v", index, got[index], want[index])
		}
	}
}

func TestCreateLookupTableInstruction(t *testing.T) {
	authority, payer := lookupKey(1), lookupKey(2)
	instruction, table, err := NewCreateLookupTableInstruction(authority, payer, 10)
	if err != nil {
		t.Fatal(err)
	}
	wantTable, wantBump, err := DeriveLookupTableAddress(authority, 10)
	if err != nil {
		t.Fatal(err)
	}
	if table != wantTable || instruction.BumpSeed != wantBump {
		t.Fatalf("derived table/bump = %s/%d, want %s/%d", table, instruction.BumpSeed, wantTable, wantBump)
	}
	assertLookupAccounts(t, instruction.Accounts(),
		solana.AccountMeta{PublicKey: table, IsWritable: true},
		solana.AccountMeta{PublicKey: authority, IsSigner: true},
		solana.AccountMeta{PublicKey: payer, IsWritable: true, IsSigner: true},
		solana.AccountMeta{PublicKey: solana.SystemProgramID},
	)

	want := []byte{0, 0, 0, 0, 10, 0, 0, 0, 0, 0, 0, 0, instruction.BumpSeed}
	data := lookupData(t, instruction)
	if !bytes.Equal(data, want) {
		t.Fatalf("Data = %x, want %x", data, want)
	}
	decoded, err := DecodeInstruction(instruction.AccountMetaSlice, data)
	if err != nil || decoded.CreateLookupTable == nil {
		t.Fatalf("decoded = %#v, %v", decoded, err)
	}
	if decoded.CreateLookupTable.RecentSlot != 10 || decoded.CreateLookupTable.BumpSeed != instruction.BumpSeed {
		t.Fatalf("decoded create = %#v", decoded.CreateLookupTable)
	}
	if roundTrip := lookupData(t, decoded.CreateLookupTable); !bytes.Equal(roundTrip, want) {
		t.Fatalf("round trip = %x, want %x", roundTrip, want)
	}
	assertLookupTruncation(t, instruction.AccountMetaSlice, data)
}

func TestAddressLookupTableInstructions(t *testing.T) {
	lookupTable, authority := lookupKey(3), lookupKey(4)
	payer, recipient := lookupKey(5), lookupKey(6)
	addressOne, addressTwo := lookupKey(0x11), lookupKey(0x22)

	t.Run("Freeze", func(t *testing.T) {
		instruction := NewFreezeLookupTableInstruction(lookupTable, authority)
		assertLookupAccounts(t, instruction.Accounts(),
			solana.AccountMeta{PublicKey: lookupTable, IsWritable: true},
			solana.AccountMeta{PublicKey: authority, IsSigner: true},
		)
		assertLookupUnitRoundTrip(t, instruction, []byte{1, 0, 0, 0}, func(got DecodedInstruction) bool { return got.FreezeLookupTable != nil })
	})

	t.Run("Extend", func(t *testing.T) {
		instruction := NewExtendLookupTableInstruction(lookupTable, authority, payer, []solana.PublicKey{addressOne, addressTwo})
		assertLookupAccounts(t, instruction.Accounts(),
			solana.AccountMeta{PublicKey: lookupTable, IsWritable: true},
			solana.AccountMeta{PublicKey: authority, IsSigner: true},
			solana.AccountMeta{PublicKey: payer, IsWritable: true, IsSigner: true},
			solana.AccountMeta{PublicKey: solana.SystemProgramID},
		)
		want := make([]byte, 0, 76)
		want = append(want, 2, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0)
		want = append(want, addressOne[:]...)
		want = append(want, addressTwo[:]...)
		data := lookupData(t, instruction)
		if !bytes.Equal(data, want) {
			t.Fatalf("Data = %x, want %x", data, want)
		}
		decoded, err := DecodeInstruction(instruction.AccountMetaSlice, data)
		if err != nil || decoded.ExtendLookupTable == nil {
			t.Fatalf("decoded = %#v, %v", decoded, err)
		}
		if len(decoded.ExtendLookupTable.Addresses) != 2 || decoded.ExtendLookupTable.Addresses[0] != addressOne || decoded.ExtendLookupTable.Addresses[1] != addressTwo {
			t.Fatalf("decoded addresses = %v", decoded.ExtendLookupTable.Addresses)
		}
		if roundTrip := lookupData(t, decoded.ExtendLookupTable); !bytes.Equal(roundTrip, want) {
			t.Fatalf("round trip = %x, want %x", roundTrip, want)
		}
		assertLookupTruncation(t, instruction.AccountMetaSlice, data)
	})

	t.Run("ExtendWithoutPayer", func(t *testing.T) {
		instruction := NewExtendLookupTableInstructionWithoutPayer(lookupTable, authority, []solana.PublicKey{addressOne})
		assertLookupAccounts(t, instruction.Accounts(),
			solana.AccountMeta{PublicKey: lookupTable, IsWritable: true},
			solana.AccountMeta{PublicKey: authority, IsSigner: true},
		)
	})

	t.Run("Deactivate", func(t *testing.T) {
		instruction := NewDeactivateLookupTableInstruction(lookupTable, authority)
		assertLookupAccounts(t, instruction.Accounts(),
			solana.AccountMeta{PublicKey: lookupTable, IsWritable: true},
			solana.AccountMeta{PublicKey: authority, IsSigner: true},
		)
		assertLookupUnitRoundTrip(t, instruction, []byte{3, 0, 0, 0}, func(got DecodedInstruction) bool { return got.DeactivateLookupTable != nil })
	})

	t.Run("Close", func(t *testing.T) {
		instruction := NewCloseLookupTableInstruction(lookupTable, authority, recipient)
		assertLookupAccounts(t, instruction.Accounts(),
			solana.AccountMeta{PublicKey: lookupTable, IsWritable: true},
			solana.AccountMeta{PublicKey: authority, IsSigner: true},
			solana.AccountMeta{PublicKey: recipient, IsWritable: true},
		)
		assertLookupUnitRoundTrip(t, instruction, []byte{4, 0, 0, 0}, func(got DecodedInstruction) bool { return got.CloseLookupTable != nil })
	})
}

func assertLookupUnitRoundTrip(t *testing.T, instruction solana.Instruction, want []byte, check func(DecodedInstruction) bool) {
	t.Helper()
	data := lookupData(t, instruction)
	if !bytes.Equal(data, want) {
		t.Fatalf("Data = %x, want %x", data, want)
	}
	decoded, err := DecodeInstruction(instruction.Accounts(), data)
	if err != nil || !check(decoded) {
		t.Fatalf("decoded = %#v, %v", decoded, err)
	}
	assertLookupTruncation(t, instruction.Accounts(), data)
}

func assertLookupTruncation(t *testing.T, accounts solana.AccountMetaSlice, data []byte) {
	t.Helper()
	for length := 0; length < len(data); length++ {
		if _, err := DecodeInstruction(accounts, data[:length]); err == nil {
			t.Errorf("accepted %d of %d bytes", length, len(data))
		}
	}
}

func TestExtendLookupTableMalformedLengths(t *testing.T) {
	data := make([]byte, 12)
	binary.LittleEndian.PutUint32(data, uint32(ExtendLookupTableInstruction))

	binary.LittleEndian.PutUint64(data[4:], LookupTableMaxAddresses+1)
	if _, err := DecodeInstruction(nil, data); !errors.Is(err, ErrTooManyAddresses) {
		t.Fatalf("oversized count error = %v", err)
	}
	binary.LittleEndian.PutUint64(data[4:], 2)
	data = append(data, make([]byte, solana.PublicKeyLength)...)
	if _, err := DecodeInstruction(nil, data); !errors.Is(err, bin.ErrUnexpectedEOF) {
		t.Fatalf("short addresses error = %v", err)
	}

	tooMany := NewExtendLookupTableInstructionWithoutPayer(
		lookupKey(1), lookupKey(2), make([]solana.PublicKey, LookupTableMaxAddresses+1),
	)
	if _, err := tooMany.Data(); !errors.Is(err, ErrTooManyAddresses) {
		t.Fatalf("encode oversized error = %v", err)
	}
}

func TestAddressLookupDecodeErrorsTrailingAndAliases(t *testing.T) {
	for length := 0; length < 4; length++ {
		if _, err := DecodeInstruction(nil, make([]byte, length)); !errors.Is(err, bin.ErrUnexpectedEOF) {
			t.Fatalf("discriminator length %d error = %v", length, err)
		}
	}
	unknown := make([]byte, 4)
	binary.LittleEndian.PutUint32(unknown, 99)
	if _, err := DecodeInstruction(nil, unknown); !errors.Is(err, ErrUnknownInstruction) {
		t.Fatalf("unknown error = %v", err)
	}

	instruction := NewFreezeLookupTableInstruction(lookupKey(1), lookupKey(2))
	data := append(lookupData(t, instruction), 0xaa)
	decoded, err := DecodeInstruction(instruction.AccountMetaSlice, data)
	if err != nil || decoded.FreezeLookupTable == nil {
		t.Fatalf("trailing decode = %#v, %v", decoded, err)
	}
	if decoded.FreezeLookupTable.AccountMetaSlice[0] != instruction.AccountMetaSlice[0] {
		t.Fatal("decoded accounts do not alias supplied accounts")
	}
}

func TestAddressLookupInstructionTypeString(t *testing.T) {
	if got := ExtendLookupTableInstruction.String(); got != "ExtendLookupTable" {
		t.Fatalf("String = %q", got)
	}
	if got := InstructionType(99).String(); got != "InstructionType(99)" {
		t.Fatalf("unknown String = %q", got)
	}
}

func FuzzDecodeInstruction(f *testing.F) {
	for typ := CreateLookupTableInstruction; typ <= CloseLookupTableInstruction; typ++ {
		data := make([]byte, 4)
		binary.LittleEndian.PutUint32(data, uint32(typ))
		f.Add(data)
	}
	f.Add([]byte{})
	f.Add([]byte{0xff, 0xff, 0xff, 0xff})
	f.Add([]byte{2, 0, 0, 0, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeInstruction(nil, data)
	})
}
