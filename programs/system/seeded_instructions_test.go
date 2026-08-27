package system

import (
	"bytes"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

func seededTestKey(fill byte) (key solana.PublicKey) {
	for index := range key {
		key[index] = fill
	}
	return key
}

func seededTestHex(t *testing.T, value string) []byte {
	t.Helper()
	decoded, err := hex.DecodeString(value)
	if err != nil {
		t.Fatalf("decode test fixture: %v", err)
	}
	return decoded
}

func seededTestAccounts(t *testing.T, got solana.AccountMetaSlice, expected ...solana.AccountMeta) {
	t.Helper()
	if len(got) != len(expected) {
		t.Fatalf("account count = %d, want %d", len(got), len(expected))
	}
	for index, want := range expected {
		if got[index] == nil {
			t.Fatalf("account %d is nil", index)
		}
		if *got[index] != want {
			t.Errorf("account %d = %+v, want %+v", index, *got[index], want)
		}
	}
}

func TestAllocateInstruction(t *testing.T) {
	newAccount := seededTestKey(0x31)
	inst := NewAllocateInstruction(0x0102030405060708, newAccount)

	seededTestAccounts(t, inst.Accounts(), solana.AccountMeta{
		PublicKey: newAccount, IsWritable: true, IsSigner: true,
	})

	wantData := seededTestHex(t, "080000000807060504030201")
	gotData, err := inst.Data()
	if err != nil {
		t.Fatalf("Data: %v", err)
	}
	if !bytes.Equal(gotData, wantData) {
		t.Fatalf("Data = %x, want %x", gotData, wantData)
	}

	decoded, err := DecodeInstruction(inst.AccountMetaSlice, gotData)
	if err != nil {
		t.Fatalf("DecodeInstruction: %v", err)
	}
	if decoded.Type != AllocateInstruction || decoded.Allocate == nil {
		t.Fatalf("decoded instruction = %+v", decoded)
	}
	if decoded.Allocate.Space != inst.Space {
		t.Errorf("decoded space = %d, want %d", decoded.Allocate.Space, inst.Space)
	}
	seededTestRoundTrip(t, decoded.Allocate, wantData)
	seededTestTruncation(t, inst.AccountMetaSlice, wantData)
}

func TestAllocateWithSeedInstruction(t *testing.T) {
	base := seededTestKey(0x11)
	owner := seededTestKey(0x22)
	allocatedAccount := seededTestKey(0x32)
	baseAccount := seededTestKey(0x33)
	inst, err := NewAllocateWithSeedInstruction(
		base, "abc", 0x0102030405060708, owner, allocatedAccount, baseAccount,
	)
	if err != nil {
		t.Fatalf("NewAllocateWithSeedInstruction: %v", err)
	}

	seededTestAccounts(t, inst.Accounts(),
		solana.AccountMeta{PublicKey: allocatedAccount, IsWritable: true},
		solana.AccountMeta{PublicKey: baseAccount, IsSigner: true},
	)

	wantData := seededTestHex(t,
		"09000000"+
			strings.Repeat("11", solana.PublicKeyLength)+
			"0300000000000000"+"616263"+
			"0807060504030201"+
			strings.Repeat("22", solana.PublicKeyLength),
	)
	gotData, err := inst.Data()
	if err != nil {
		t.Fatalf("Data: %v", err)
	}
	if !bytes.Equal(gotData, wantData) {
		t.Fatalf("Data = %x, want %x", gotData, wantData)
	}

	decoded, err := DecodeInstruction(inst.AccountMetaSlice, gotData)
	if err != nil {
		t.Fatalf("DecodeInstruction: %v", err)
	}
	if decoded.Type != AllocateWithSeedInstruction || decoded.AllocateWithSeed == nil {
		t.Fatalf("decoded instruction = %+v", decoded)
	}
	got := decoded.AllocateWithSeed
	if got.Base != base || got.Seed != "abc" || got.Space != inst.Space || got.Owner != owner {
		t.Errorf("decoded AllocateWithSeed = %+v, want %+v", got, inst)
	}
	seededTestRoundTrip(t, got, wantData)
	seededTestTruncation(t, inst.AccountMetaSlice, wantData)
}

func TestAssignWithSeedInstruction(t *testing.T) {
	base := seededTestKey(0x11)
	owner := seededTestKey(0x22)
	assignedAccount := seededTestKey(0x34)
	baseAccount := seededTestKey(0x35)
	inst, err := NewAssignWithSeedInstruction(base, "abc", owner, assignedAccount, baseAccount)
	if err != nil {
		t.Fatalf("NewAssignWithSeedInstruction: %v", err)
	}

	seededTestAccounts(t, inst.Accounts(),
		solana.AccountMeta{PublicKey: assignedAccount, IsWritable: true},
		solana.AccountMeta{PublicKey: baseAccount, IsSigner: true},
	)

	wantData := seededTestHex(t,
		"0a000000"+
			strings.Repeat("11", solana.PublicKeyLength)+
			"0300000000000000"+"616263"+
			strings.Repeat("22", solana.PublicKeyLength),
	)
	gotData, err := inst.Data()
	if err != nil {
		t.Fatalf("Data: %v", err)
	}
	if !bytes.Equal(gotData, wantData) {
		t.Fatalf("Data = %x, want %x", gotData, wantData)
	}

	decoded, err := DecodeInstruction(inst.AccountMetaSlice, gotData)
	if err != nil {
		t.Fatalf("DecodeInstruction: %v", err)
	}
	if decoded.Type != AssignWithSeedInstruction || decoded.AssignWithSeed == nil {
		t.Fatalf("decoded instruction = %+v", decoded)
	}
	got := decoded.AssignWithSeed
	if got.Base != base || got.Seed != "abc" || got.Owner != owner {
		t.Errorf("decoded AssignWithSeed = %+v, want %+v", got, inst)
	}
	seededTestRoundTrip(t, got, wantData)
	seededTestTruncation(t, inst.AccountMetaSlice, wantData)
}

func TestTransferWithSeedInstruction(t *testing.T) {
	owner := seededTestKey(0x22)
	fundingAccount := seededTestKey(0x36)
	baseAccount := seededTestKey(0x37)
	recipientAccount := seededTestKey(0x38)
	inst, err := NewTransferWithSeedInstruction(
		0x0102030405060708, "abc", owner, fundingAccount, baseAccount, recipientAccount,
	)
	if err != nil {
		t.Fatalf("NewTransferWithSeedInstruction: %v", err)
	}

	seededTestAccounts(t, inst.Accounts(),
		solana.AccountMeta{PublicKey: fundingAccount, IsWritable: true},
		solana.AccountMeta{PublicKey: baseAccount, IsSigner: true},
		solana.AccountMeta{PublicKey: recipientAccount, IsWritable: true},
	)

	wantData := seededTestHex(t,
		"0b000000"+
			"0807060504030201"+
			"0300000000000000"+"616263"+
			strings.Repeat("22", solana.PublicKeyLength),
	)
	gotData, err := inst.Data()
	if err != nil {
		t.Fatalf("Data: %v", err)
	}
	if !bytes.Equal(gotData, wantData) {
		t.Fatalf("Data = %x, want %x", gotData, wantData)
	}

	decoded, err := DecodeInstruction(inst.AccountMetaSlice, gotData)
	if err != nil {
		t.Fatalf("DecodeInstruction: %v", err)
	}
	if decoded.Type != TransferWithSeedInstruction || decoded.TransferWithSeed == nil {
		t.Fatalf("decoded instruction = %+v", decoded)
	}
	got := decoded.TransferWithSeed
	if got.Lamports != inst.Lamports || got.FromSeed != "abc" || got.FromOwner != owner {
		t.Errorf("decoded TransferWithSeed = %+v, want %+v", got, inst)
	}
	seededTestRoundTrip(t, got, wantData)
	seededTestTruncation(t, inst.AccountMetaSlice, wantData)
}

func TestSeededInstructionSeedLength(t *testing.T) {
	key := seededTestKey(0x44)
	maximum := strings.Repeat("s", solana.MaxSeedLength)
	tooLong := maximum + "s"

	if _, err := NewAllocateWithSeedInstruction(key, maximum, 1, key, key, key); err != nil {
		t.Errorf("AllocateWithSeed maximum seed: %v", err)
	}
	if _, err := NewAssignWithSeedInstruction(key, maximum, key, key, key); err != nil {
		t.Errorf("AssignWithSeed maximum seed: %v", err)
	}
	if _, err := NewTransferWithSeedInstruction(1, maximum, key, key, key, key); err != nil {
		t.Errorf("TransferWithSeed maximum seed: %v", err)
	}

	if _, err := NewAllocateWithSeedInstruction(key, tooLong, 1, key, key, key); !errors.Is(err, solana.ErrMaxSeedLengthExceeded) {
		t.Errorf("AllocateWithSeed oversized seed error = %v", err)
	}
	if _, err := NewAssignWithSeedInstruction(key, tooLong, key, key, key); !errors.Is(err, solana.ErrMaxSeedLengthExceeded) {
		t.Errorf("AssignWithSeed oversized seed error = %v", err)
	}
	if _, err := NewTransferWithSeedInstruction(1, tooLong, key, key, key, key); !errors.Is(err, solana.ErrMaxSeedLengthExceeded) {
		t.Errorf("TransferWithSeed oversized seed error = %v", err)
	}

	allocate := &AllocateWithSeed{Seed: tooLong}
	if _, err := allocate.Data(); !errors.Is(err, solana.ErrMaxSeedLengthExceeded) {
		t.Errorf("AllocateWithSeed.Data oversized seed error = %v", err)
	}
	assign := &AssignWithSeed{Seed: tooLong}
	if _, err := assign.Data(); !errors.Is(err, solana.ErrMaxSeedLengthExceeded) {
		t.Errorf("AssignWithSeed.Data oversized seed error = %v", err)
	}
	transfer := &TransferWithSeed{FromSeed: tooLong}
	if _, err := transfer.Data(); !errors.Is(err, solana.ErrMaxSeedLengthExceeded) {
		t.Errorf("TransferWithSeed.Data oversized seed error = %v", err)
	}
}

func seededTestRoundTrip(t *testing.T, inst solana.Instruction, want []byte) {
	t.Helper()
	data, err := inst.Data()
	if err != nil {
		t.Fatalf("round-trip Data: %v", err)
	}
	if !bytes.Equal(data, want) {
		t.Fatalf("round-trip Data = %x, want %x", data, want)
	}
}

func seededTestTruncation(t *testing.T, accounts solana.AccountMetaSlice, complete []byte) {
	t.Helper()
	for length := 0; length < len(complete); length++ {
		if _, err := DecodeInstruction(accounts, complete[:length]); err == nil {
			t.Fatalf("DecodeInstruction accepted truncated data length %d of %d", length, len(complete))
		}
	}
}
