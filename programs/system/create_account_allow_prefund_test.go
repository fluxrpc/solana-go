package system

import (
	"bytes"
	"strings"
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

func TestCreateAccountAllowPrefundInstruction(t *testing.T) {
	owner := basicTestKey(0x21)
	newAccount := basicTestKey(0x31)
	fundingAccount := basicTestKey(0x32)
	lamports := uint64(0x0102030405060708)
	space := uint64(0x1112131415161718)
	instruction := NewCreateAccountAllowPrefundInstruction(
		lamports,
		space,
		owner,
		newAccount,
		fundingAccount,
	)

	if instruction.ProgramID() != ProgramID {
		t.Errorf("ProgramID = %s, want %s", instruction.ProgramID(), ProgramID)
	}
	assertBasicAccounts(t, instruction.Accounts(),
		solana.AccountMeta{PublicKey: newAccount, IsWritable: true, IsSigner: true},
		solana.AccountMeta{PublicKey: fundingAccount, IsWritable: true, IsSigner: true},
	)

	wantData := basicTestHex(t,
		"0d000000"+
			"0807060504030201"+
			"1817161514131211"+
			strings.Repeat("21", solana.PublicKeyLength),
	)
	data := basicTestData(t, instruction)
	if !bytes.Equal(data, wantData) {
		t.Fatalf("Data = %x, want %x", data, wantData)
	}

	decoded, err := DecodeInstruction(instruction.AccountMetaSlice, data)
	if err != nil {
		t.Fatalf("DecodeInstruction: %v", err)
	}
	if decoded.Type != CreateAccountAllowPrefundInstruction ||
		decoded.CreateAccountAllowPrefund == nil {
		t.Fatalf("decoded instruction = %+v, want CreateAccountAllowPrefund", decoded)
	}
	got := decoded.CreateAccountAllowPrefund
	if got.Lamports != lamports || got.Space != space || got.Owner != owner {
		t.Errorf("decoded CreateAccountAllowPrefund = %+v, want %+v", got, instruction)
	}
	assertBasicAccounts(t, got.Accounts(),
		solana.AccountMeta{PublicKey: newAccount, IsWritable: true, IsSigner: true},
		solana.AccountMeta{PublicKey: fundingAccount, IsWritable: true, IsSigner: true},
	)
	assertBasicRoundTrip(t, got, wantData)
	assertBasicTruncation(t, instruction.AccountMetaSlice, wantData)
}

func TestCreateAccountAllowPrefundWithoutFundingInstruction(t *testing.T) {
	owner := basicTestKey(0x22)
	newAccount := basicTestKey(0x33)
	space := uint64(0x2122232425262728)
	instruction := NewCreateAccountAllowPrefundWithoutFundingInstruction(
		space,
		owner,
		newAccount,
	)

	if instruction.Lamports != 0 {
		t.Fatalf("Lamports = %d, want 0", instruction.Lamports)
	}
	assertBasicAccounts(t, instruction.Accounts(),
		solana.AccountMeta{PublicKey: newAccount, IsWritable: true, IsSigner: true},
	)

	wantData := basicTestHex(t,
		"0d000000"+
			"0000000000000000"+
			"2827262524232221"+
			strings.Repeat("22", solana.PublicKeyLength),
	)
	data := basicTestData(t, instruction)
	if !bytes.Equal(data, wantData) {
		t.Fatalf("Data = %x, want %x", data, wantData)
	}

	decoded, err := DecodeInstruction(instruction.AccountMetaSlice, data)
	if err != nil {
		t.Fatalf("DecodeInstruction: %v", err)
	}
	if decoded.Type != CreateAccountAllowPrefundInstruction ||
		decoded.CreateAccountAllowPrefund == nil {
		t.Fatalf("decoded instruction = %+v, want CreateAccountAllowPrefund", decoded)
	}
	got := decoded.CreateAccountAllowPrefund
	if got.Lamports != 0 || got.Space != space || got.Owner != owner {
		t.Errorf("decoded CreateAccountAllowPrefund = %+v, want %+v", got, instruction)
	}
	assertBasicAccounts(t, got.Accounts(),
		solana.AccountMeta{PublicKey: newAccount, IsWritable: true, IsSigner: true},
	)
	assertBasicRoundTrip(t, got, wantData)
}
