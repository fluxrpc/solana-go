package system

import (
	"bytes"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

func basicTestKey(fill byte) (key solana.PublicKey) {
	for index := range key {
		key[index] = fill
	}
	return key
}

func basicTestHex(t *testing.T, value string) []byte {
	t.Helper()
	data, err := hex.DecodeString(value)
	if err != nil {
		t.Fatalf("decode golden data: %v", err)
	}
	return data
}

func basicTestData(t *testing.T, instruction solana.Instruction) []byte {
	t.Helper()
	data, err := instruction.Data()
	if err != nil {
		t.Fatalf("Data: %v", err)
	}
	return data
}

func assertBasicAccounts(t *testing.T, got solana.AccountMetaSlice, want ...solana.AccountMeta) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("account count = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index] == nil {
			t.Fatalf("account %d is nil", index)
		}
		if *got[index] != want[index] {
			t.Errorf("account %d = %+v, want %+v", index, *got[index], want[index])
		}
	}
}

func assertBasicRoundTrip(t *testing.T, instruction solana.Instruction, want []byte) {
	t.Helper()
	got := basicTestData(t, instruction)
	if !bytes.Equal(got, want) {
		t.Fatalf("round-trip Data = %x, want %x", got, want)
	}
}

func assertBasicTruncation(t *testing.T, accounts solana.AccountMetaSlice, complete []byte) {
	t.Helper()
	for length := 0; length < len(complete); length++ {
		if _, err := DecodeInstruction(accounts, complete[:length]); err == nil {
			t.Errorf("DecodeInstruction accepted %d of %d bytes", length, len(complete))
		}
	}
}

func TestCreateAccountInstruction(t *testing.T) {
	owner := basicTestKey(0x11)
	funding := basicTestKey(0x31)
	created := basicTestKey(0x32)
	lamports := uint64(0x0102030405060708)
	space := uint64(0x1112131415161718)
	instruction := NewCreateAccountInstruction(lamports, space, owner, funding, created)

	if instruction.ProgramID() != ProgramID {
		t.Errorf("ProgramID = %s, want %s", instruction.ProgramID(), ProgramID)
	}
	assertBasicAccounts(t, instruction.Accounts(),
		solana.AccountMeta{PublicKey: funding, IsWritable: true, IsSigner: true},
		solana.AccountMeta{PublicKey: created, IsWritable: true, IsSigner: true},
	)

	wantData := basicTestHex(t,
		"00000000"+
			"0807060504030201"+
			"1817161514131211"+
			strings.Repeat("11", solana.PublicKeyLength),
	)
	data := basicTestData(t, instruction)
	if !bytes.Equal(data, wantData) {
		t.Fatalf("Data = %x, want %x", data, wantData)
	}

	decoded, err := DecodeInstruction(instruction.AccountMetaSlice, data)
	if err != nil {
		t.Fatalf("DecodeInstruction: %v", err)
	}
	if decoded.Type != CreateAccountInstruction || decoded.CreateAccount == nil {
		t.Fatalf("decoded instruction = %+v, want CreateAccount", decoded)
	}
	got := decoded.CreateAccount
	if got.Lamports != lamports || got.Space != space || got.Owner != owner {
		t.Errorf("decoded CreateAccount = %+v, want %+v", got, instruction)
	}
	assertBasicAccounts(t, got.Accounts(),
		solana.AccountMeta{PublicKey: funding, IsWritable: true, IsSigner: true},
		solana.AccountMeta{PublicKey: created, IsWritable: true, IsSigner: true},
	)
	assertBasicRoundTrip(t, got, wantData)
	assertBasicTruncation(t, instruction.AccountMetaSlice, wantData)
}

func TestAssignInstruction(t *testing.T) {
	owner := basicTestKey(0x12)
	assigned := basicTestKey(0x33)
	instruction := NewAssignInstruction(owner, assigned)

	assertBasicAccounts(t, instruction.Accounts(),
		solana.AccountMeta{PublicKey: assigned, IsWritable: true, IsSigner: true},
	)
	wantData := basicTestHex(t,
		"01000000"+strings.Repeat("12", solana.PublicKeyLength),
	)
	data := basicTestData(t, instruction)
	if !bytes.Equal(data, wantData) {
		t.Fatalf("Data = %x, want %x", data, wantData)
	}

	decoded, err := DecodeInstruction(instruction.AccountMetaSlice, data)
	if err != nil {
		t.Fatalf("DecodeInstruction: %v", err)
	}
	if decoded.Type != AssignInstruction || decoded.Assign == nil {
		t.Fatalf("decoded instruction = %+v, want Assign", decoded)
	}
	if decoded.Assign.Owner != owner {
		t.Errorf("decoded owner = %s, want %s", decoded.Assign.Owner, owner)
	}
	assertBasicRoundTrip(t, decoded.Assign, wantData)
	assertBasicTruncation(t, instruction.AccountMetaSlice, wantData)
}

func TestTransferInstruction(t *testing.T) {
	funding := basicTestKey(0x34)
	recipient := basicTestKey(0x35)
	lamports := uint64(0x0102030405060708)
	instruction := NewTransferInstruction(lamports, funding, recipient)

	assertBasicAccounts(t, instruction.Accounts(),
		solana.AccountMeta{PublicKey: funding, IsWritable: true, IsSigner: true},
		solana.AccountMeta{PublicKey: recipient, IsWritable: true},
	)
	wantData := basicTestHex(t, "020000000807060504030201")
	data := basicTestData(t, instruction)
	if !bytes.Equal(data, wantData) {
		t.Fatalf("Data = %x, want %x", data, wantData)
	}

	decoded, err := DecodeInstruction(instruction.AccountMetaSlice, data)
	if err != nil {
		t.Fatalf("DecodeInstruction: %v", err)
	}
	if decoded.Type != TransferInstruction || decoded.Transfer == nil {
		t.Fatalf("decoded instruction = %+v, want Transfer", decoded)
	}
	if decoded.Transfer.Lamports != lamports {
		t.Errorf("decoded lamports = %d, want %d", decoded.Transfer.Lamports, lamports)
	}
	assertBasicRoundTrip(t, decoded.Transfer, wantData)
	assertBasicTruncation(t, instruction.AccountMetaSlice, wantData)
}

func TestCreateAccountWithSeedInstruction(t *testing.T) {
	base := basicTestKey(0x13)
	owner := basicTestKey(0x14)
	funding := basicTestKey(0x36)
	created := basicTestKey(0x37)
	baseAccount := basicTestKey(0x38)
	lamports := uint64(0x0102030405060708)
	space := uint64(0x1112131415161718)
	instruction, err := NewCreateAccountWithSeedInstruction(
		base,
		"abc",
		lamports,
		space,
		owner,
		funding,
		created,
		baseAccount,
	)
	if err != nil {
		t.Fatalf("NewCreateAccountWithSeedInstruction: %v", err)
	}

	assertBasicAccounts(t, instruction.Accounts(),
		solana.AccountMeta{PublicKey: funding, IsWritable: true, IsSigner: true},
		solana.AccountMeta{PublicKey: created, IsWritable: true},
		solana.AccountMeta{PublicKey: baseAccount, IsSigner: true},
	)
	wantData := basicTestHex(t,
		"03000000"+
			strings.Repeat("13", solana.PublicKeyLength)+
			"0300000000000000"+"616263"+
			"0807060504030201"+
			"1817161514131211"+
			strings.Repeat("14", solana.PublicKeyLength),
	)
	data := basicTestData(t, instruction)
	if !bytes.Equal(data, wantData) {
		t.Fatalf("Data = %x, want %x", data, wantData)
	}
	if prefix := data[4+solana.PublicKeyLength : 4+solana.PublicKeyLength+8]; !bytes.Equal(prefix, []byte{3, 0, 0, 0, 0, 0, 0, 0}) {
		t.Errorf("seed length prefix = %x, want uint64 little-endian 3", prefix)
	}

	decoded, err := DecodeInstruction(instruction.AccountMetaSlice, data)
	if err != nil {
		t.Fatalf("DecodeInstruction: %v", err)
	}
	if decoded.Type != CreateAccountWithSeedInstruction || decoded.CreateAccountWithSeed == nil {
		t.Fatalf("decoded instruction = %+v, want CreateAccountWithSeed", decoded)
	}
	got := decoded.CreateAccountWithSeed
	if got.Base != base || got.Seed != "abc" || got.Lamports != lamports ||
		got.Space != space || got.Owner != owner {
		t.Errorf("decoded CreateAccountWithSeed = %+v, want %+v", got, instruction)
	}
	assertBasicRoundTrip(t, got, wantData)
	assertBasicTruncation(t, instruction.AccountMetaSlice, wantData)
}

func TestCreateAccountWithSeedSeedLength(t *testing.T) {
	key := basicTestKey(0x41)
	maximum := strings.Repeat("s", solana.MaxSeedLength)
	tooLong := maximum + "s"

	if _, err := NewCreateAccountWithSeedInstruction(key, maximum, 1, 2, key, key, key, key); err != nil {
		t.Errorf("maximum-length seed: %v", err)
	}
	if _, err := NewCreateAccountWithSeedInstruction(key, tooLong, 1, 2, key, key, key, key); !errors.Is(err, solana.ErrMaxSeedLengthExceeded) {
		t.Errorf("oversized seed error = %v", err)
	}
	if _, err := NewCreateAccountWithSeedInstruction(
		key, strings.Repeat("é", solana.MaxSeedLength/2+1), 1, 2, key, key, key, key,
	); !errors.Is(err, solana.ErrMaxSeedLengthExceeded) {
		t.Errorf("oversized UTF-8 seed error = %v", err)
	}

	mutated := &CreateAccountWithSeed{Seed: tooLong}
	if _, err := mutated.Data(); !errors.Is(err, solana.ErrMaxSeedLengthExceeded) {
		t.Errorf("Data oversized seed error = %v", err)
	}
}
