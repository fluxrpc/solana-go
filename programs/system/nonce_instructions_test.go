package system

import (
	"bytes"
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

func nonceTestPublicKey(first byte) (key solana.PublicKey) {
	for index := range key {
		key[index] = first + byte(index)
	}
	return key
}

func assertNonceAccounts(
	t *testing.T,
	got []*solana.AccountMeta,
	want []solana.AccountMeta,
) {
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

func nonceInstructionData(t *testing.T, instruction solana.Instruction) []byte {
	t.Helper()
	data, err := instruction.Data()
	if err != nil {
		t.Fatalf("Data: %v", err)
	}
	return data
}

func TestAdvanceNonceAccountInstruction(t *testing.T) {
	nonce := nonceTestPublicKey(1)
	recentBlockhashes := nonceTestPublicKey(2)
	authority := nonceTestPublicKey(3)
	instruction := NewAdvanceNonceAccountInstruction(nonce, recentBlockhashes, authority)

	if instruction.ProgramID() != ProgramID {
		t.Errorf("ProgramID = %s, want %s", instruction.ProgramID(), ProgramID)
	}
	assertNonceAccounts(t, instruction.Accounts(), []solana.AccountMeta{
		{PublicKey: nonce, IsWritable: true},
		{PublicKey: recentBlockhashes},
		{PublicKey: authority, IsSigner: true},
	})

	data := nonceInstructionData(t, instruction)
	wantData := []byte{4, 0, 0, 0}
	if !bytes.Equal(data, wantData) {
		t.Fatalf("Data = %x, want %x", data, wantData)
	}

	decoded, err := DecodeInstruction(instruction.Accounts(), data)
	if err != nil {
		t.Fatalf("DecodeInstruction: %v", err)
	}
	if decoded.Type != AdvanceNonceAccountInstruction || decoded.AdvanceNonceAccount == nil {
		t.Fatalf("decoded = %+v, want AdvanceNonceAccount", decoded)
	}
	if decoded.CreateAccount != nil || decoded.Assign != nil || decoded.Transfer != nil ||
		decoded.CreateAccountWithSeed != nil || decoded.WithdrawNonceAccount != nil ||
		decoded.InitializeNonceAccount != nil || decoded.AuthorizeNonceAccount != nil ||
		decoded.Allocate != nil || decoded.AllocateWithSeed != nil || decoded.AssignWithSeed != nil ||
		decoded.TransferWithSeed != nil || decoded.UpgradeNonceAccount != nil {
		t.Fatal("decoded envelope populated more than AdvanceNonceAccount")
	}
	assertNonceAccounts(t, decoded.AdvanceNonceAccount.Accounts(), []solana.AccountMeta{
		{PublicKey: nonce, IsWritable: true},
		{PublicKey: recentBlockhashes},
		{PublicKey: authority, IsSigner: true},
	})
	if roundTrip := nonceInstructionData(t, decoded.AdvanceNonceAccount); !bytes.Equal(roundTrip, wantData) {
		t.Errorf("round-trip Data = %x, want %x", roundTrip, wantData)
	}
}

func TestWithdrawNonceAccountInstruction(t *testing.T) {
	lamports := uint64(0x0807060504030201)
	nonce := nonceTestPublicKey(4)
	recipient := nonceTestPublicKey(5)
	recentBlockhashes := nonceTestPublicKey(6)
	rent := nonceTestPublicKey(7)
	authority := nonceTestPublicKey(8)
	instruction := NewWithdrawNonceAccountInstruction(
		lamports,
		nonce,
		recipient,
		recentBlockhashes,
		rent,
		authority,
	)

	assertNonceAccounts(t, instruction.Accounts(), []solana.AccountMeta{
		{PublicKey: nonce, IsWritable: true},
		{PublicKey: recipient, IsWritable: true},
		{PublicKey: recentBlockhashes},
		{PublicKey: rent},
		{PublicKey: authority, IsSigner: true},
	})

	data := nonceInstructionData(t, instruction)
	wantData := []byte{5, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8}
	if !bytes.Equal(data, wantData) {
		t.Fatalf("Data = %x, want %x", data, wantData)
	}

	decoded, err := DecodeInstruction(instruction.Accounts(), data)
	if err != nil {
		t.Fatalf("DecodeInstruction: %v", err)
	}
	if decoded.Type != WithdrawNonceAccountInstruction || decoded.WithdrawNonceAccount == nil {
		t.Fatalf("decoded = %+v, want WithdrawNonceAccount", decoded)
	}
	if decoded.WithdrawNonceAccount.Lamports != lamports {
		t.Errorf("Lamports = %d, want %d", decoded.WithdrawNonceAccount.Lamports, lamports)
	}
	if decoded.AdvanceNonceAccount != nil || decoded.InitializeNonceAccount != nil ||
		decoded.AuthorizeNonceAccount != nil || decoded.UpgradeNonceAccount != nil {
		t.Fatal("decoded envelope populated more than WithdrawNonceAccount")
	}
	if roundTrip := nonceInstructionData(t, decoded.WithdrawNonceAccount); !bytes.Equal(roundTrip, wantData) {
		t.Errorf("round-trip Data = %x, want %x", roundTrip, wantData)
	}

	assertTruncatedNonceInstruction(t, instruction.Accounts(), data)
}

func TestInitializeNonceAccountInstruction(t *testing.T) {
	authorized := nonceTestPublicKey(9)
	nonce := nonceTestPublicKey(10)
	recentBlockhashes := nonceTestPublicKey(11)
	rent := nonceTestPublicKey(12)
	instruction := NewInitializeNonceAccountInstruction(authorized, nonce, recentBlockhashes, rent)

	assertNonceAccounts(t, instruction.Accounts(), []solana.AccountMeta{
		{PublicKey: nonce, IsWritable: true},
		{PublicKey: recentBlockhashes},
		{PublicKey: rent},
	})

	wantData := make([]byte, 4+solana.PublicKeyLength)
	wantData[0] = 6
	copy(wantData[4:], authorized[:])
	data := nonceInstructionData(t, instruction)
	if !bytes.Equal(data, wantData) {
		t.Fatalf("Data = %x, want %x", data, wantData)
	}

	decoded, err := DecodeInstruction(instruction.Accounts(), data)
	if err != nil {
		t.Fatalf("DecodeInstruction: %v", err)
	}
	if decoded.Type != InitializeNonceAccountInstruction || decoded.InitializeNonceAccount == nil {
		t.Fatalf("decoded = %+v, want InitializeNonceAccount", decoded)
	}
	if decoded.InitializeNonceAccount.Authorized != authorized {
		t.Errorf("Authorized = %s, want %s", decoded.InitializeNonceAccount.Authorized, authorized)
	}
	if decoded.AdvanceNonceAccount != nil || decoded.WithdrawNonceAccount != nil ||
		decoded.AuthorizeNonceAccount != nil || decoded.UpgradeNonceAccount != nil {
		t.Fatal("decoded envelope populated more than InitializeNonceAccount")
	}
	if roundTrip := nonceInstructionData(t, decoded.InitializeNonceAccount); !bytes.Equal(roundTrip, wantData) {
		t.Errorf("round-trip Data = %x, want %x", roundTrip, wantData)
	}

	assertTruncatedNonceInstruction(t, instruction.Accounts(), data)
}

func TestAuthorizeNonceAccountInstruction(t *testing.T) {
	authorized := nonceTestPublicKey(13)
	nonce := nonceTestPublicKey(14)
	authority := nonceTestPublicKey(15)
	instruction := NewAuthorizeNonceAccountInstruction(authorized, nonce, authority)

	assertNonceAccounts(t, instruction.Accounts(), []solana.AccountMeta{
		{PublicKey: nonce, IsWritable: true},
		{PublicKey: authority, IsSigner: true},
	})

	wantData := make([]byte, 4+solana.PublicKeyLength)
	wantData[0] = 7
	copy(wantData[4:], authorized[:])
	data := nonceInstructionData(t, instruction)
	if !bytes.Equal(data, wantData) {
		t.Fatalf("Data = %x, want %x", data, wantData)
	}

	decoded, err := DecodeInstruction(instruction.Accounts(), data)
	if err != nil {
		t.Fatalf("DecodeInstruction: %v", err)
	}
	if decoded.Type != AuthorizeNonceAccountInstruction || decoded.AuthorizeNonceAccount == nil {
		t.Fatalf("decoded = %+v, want AuthorizeNonceAccount", decoded)
	}
	if decoded.AuthorizeNonceAccount.Authorized != authorized {
		t.Errorf("Authorized = %s, want %s", decoded.AuthorizeNonceAccount.Authorized, authorized)
	}
	if decoded.AdvanceNonceAccount != nil || decoded.WithdrawNonceAccount != nil ||
		decoded.InitializeNonceAccount != nil || decoded.UpgradeNonceAccount != nil {
		t.Fatal("decoded envelope populated more than AuthorizeNonceAccount")
	}
	if roundTrip := nonceInstructionData(t, decoded.AuthorizeNonceAccount); !bytes.Equal(roundTrip, wantData) {
		t.Errorf("round-trip Data = %x, want %x", roundTrip, wantData)
	}

	assertTruncatedNonceInstruction(t, instruction.Accounts(), data)
}

func TestUpgradeNonceAccountInstruction(t *testing.T) {
	nonce := nonceTestPublicKey(16)
	instruction := NewUpgradeNonceAccountInstruction(nonce)

	assertNonceAccounts(t, instruction.Accounts(), []solana.AccountMeta{
		{PublicKey: nonce, IsWritable: true},
	})

	data := nonceInstructionData(t, instruction)
	wantData := []byte{12, 0, 0, 0}
	if !bytes.Equal(data, wantData) {
		t.Fatalf("Data = %x, want %x", data, wantData)
	}

	decoded, err := DecodeInstruction(instruction.Accounts(), data)
	if err != nil {
		t.Fatalf("DecodeInstruction: %v", err)
	}
	if decoded.Type != UpgradeNonceAccountInstruction || decoded.UpgradeNonceAccount == nil {
		t.Fatalf("decoded = %+v, want UpgradeNonceAccount", decoded)
	}
	if decoded.AdvanceNonceAccount != nil || decoded.WithdrawNonceAccount != nil ||
		decoded.InitializeNonceAccount != nil || decoded.AuthorizeNonceAccount != nil {
		t.Fatal("decoded envelope populated more than UpgradeNonceAccount")
	}
	if roundTrip := nonceInstructionData(t, decoded.UpgradeNonceAccount); !bytes.Equal(roundTrip, wantData) {
		t.Errorf("round-trip Data = %x, want %x", roundTrip, wantData)
	}
}

func assertTruncatedNonceInstruction(t *testing.T, accounts solana.AccountMetaSlice, data []byte) {
	t.Helper()
	for length := 0; length < len(data); length++ {
		if _, err := DecodeInstruction(accounts, data[:length]); err == nil {
			t.Errorf("DecodeInstruction accepted %d of %d bytes", length, len(data))
		}
	}
}
