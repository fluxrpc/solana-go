package token

import (
	"bytes"
	"errors"
	"testing"

	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

func tokenKey(value byte) solana.PublicKey { var key solana.PublicKey; key[0] = value; return key }

func TestInstructionsRoundTrip(t *testing.T) {
	k1, k2, k3, k4 := tokenKey(1), tokenKey(2), tokenKey(3), tokenKey(4)
	instructions := []solana.Instruction{
		NewInitializeMintInstruction(6, k1, k2, k3, solana.SysVarRentPubkey),
		NewInitializeAccountInstruction(k1, k2, k3, solana.SysVarRentPubkey),
		NewInitializeMultisigInstruction(2, k1, solana.SysVarRentPubkey, []solana.PublicKey{k2, k3}),
		NewTransferInstruction(42, k1, k2, k3, nil),
		NewApproveInstruction(42, k1, k2, k3, nil),
		NewRevokeInstruction(k1, k2, nil),
		NewSetAuthorityInstruction(AuthorityAccountOwner, k1, k2, k3, nil),
		NewMintToInstruction(42, k1, k2, k3, nil),
		NewBurnInstruction(42, k1, k2, k3, nil),
		NewCloseAccountInstruction(k1, k2, k3, nil),
		NewFreezeAccountInstruction(k1, k2, k3, nil),
		NewThawAccountInstruction(k1, k2, k3, nil),
		NewTransferCheckedInstruction(42, 6, k1, k2, k3, k4, nil),
		NewApproveCheckedInstruction(42, 6, k1, k2, k3, k4, nil),
		NewMintToCheckedInstruction(42, 6, k1, k2, k3, nil),
		NewBurnCheckedInstruction(42, 6, k1, k2, k3, nil),
		NewInitializeAccount2Instruction(k1, k2, k3, solana.SysVarRentPubkey),
		NewSyncNativeInstruction(k1),
		NewInitializeAccount3Instruction(k1, k2, k3),
		NewInitializeMultisig2Instruction(2, k1, []solana.PublicKey{k2, k3}),
		NewInitializeMint2Instruction(6, k1, k2, k3),
	}
	for expectedType, inst := range instructions {
		data, err := inst.Data()
		if err != nil {
			t.Fatalf("%d Data: %v", expectedType, err)
		}
		if data[0] != byte(expectedType) {
			t.Fatalf("type = %d, want %d", data[0], expectedType)
		}
		decoded, err := DecodeInstruction(inst.Accounts(), data)
		if err != nil {
			t.Fatalf("%d decode: %v", expectedType, err)
		}
		var roundTrip solana.Instruction
		switch decoded.Type {
		case InitializeMintInstruction:
			roundTrip = decoded.InitializeMint
		case InitializeAccountInstruction:
			roundTrip = decoded.InitializeAccount
		case InitializeMultisigInstruction:
			roundTrip = decoded.InitializeMultisig
		case TransferInstruction:
			roundTrip = decoded.Transfer
		case ApproveInstruction:
			roundTrip = decoded.Approve
		case RevokeInstruction:
			roundTrip = decoded.Revoke
		case SetAuthorityInstruction:
			roundTrip = decoded.SetAuthority
		case MintToInstruction:
			roundTrip = decoded.MintTo
		case BurnInstruction:
			roundTrip = decoded.Burn
		case CloseAccountInstruction:
			roundTrip = decoded.CloseAccount
		case FreezeAccountInstruction:
			roundTrip = decoded.FreezeAccount
		case ThawAccountInstruction:
			roundTrip = decoded.ThawAccount
		case TransferCheckedInstruction:
			roundTrip = decoded.TransferChecked
		case ApproveCheckedInstruction:
			roundTrip = decoded.ApproveChecked
		case MintToCheckedInstruction:
			roundTrip = decoded.MintToChecked
		case BurnCheckedInstruction:
			roundTrip = decoded.BurnChecked
		case InitializeAccount2Instruction:
			roundTrip = decoded.InitializeAccount2
		case SyncNativeInstruction:
			roundTrip = decoded.SyncNative
		case InitializeAccount3Instruction:
			roundTrip = decoded.InitializeAccount3
		case InitializeMultisig2Instruction:
			roundTrip = decoded.InitializeMultisig2
		case InitializeMint2Instruction:
			roundTrip = decoded.InitializeMint2
		}
		got, err := roundTrip.Data()
		if err != nil || !bytes.Equal(got, data) {
			t.Fatalf("%d round trip = %x, %v; want %x", expectedType, got, err, data)
		}
	}
}

func TestInitializeMintFixtureAndAccounts(t *testing.T) {
	mintAuthority, freezeAuthority, mint := tokenKey(1), tokenKey(2), tokenKey(3)
	inst := NewInitializeMintInstruction(9, mintAuthority, freezeAuthority, mint, solana.SysVarRentPubkey)
	want := append([]byte{0, 9}, mintAuthority[:]...)
	want = append(want, 1)
	want = append(want, freezeAuthority[:]...)
	got, _ := inst.Data()
	if !bytes.Equal(got, want) {
		t.Fatalf("data = %x, want %x", got, want)
	}
	if len(inst.Accounts()) != 2 || !inst.Accounts()[0].IsWritable || inst.Accounts()[0].IsSigner {
		t.Fatalf("accounts = %+v", inst.Accounts())
	}
}

func TestDecodeInstructionErrors(t *testing.T) {
	if _, err := DecodeInstruction(nil, nil); !errors.Is(err, bin.ErrUnexpectedEOF) {
		t.Fatalf("empty error = %v", err)
	}
	if _, err := DecodeInstruction(nil, []byte{255}); !errors.Is(err, ErrUnknownInstruction) {
		t.Fatalf("unknown error = %v", err)
	}
	for size := 1; size < 9; size++ {
		if _, err := DecodeInstruction(nil, make([]byte, size)); !errors.Is(err, bin.ErrUnexpectedEOF) {
			t.Fatalf("size %d error = %v", size, err)
		}
	}
}

func TestDecodeTokenAccountFixture(t *testing.T) {
	mint, owner, delegate, closeAuthority := tokenKey(1), tokenKey(2), tokenKey(3), tokenKey(4)
	enc := bin.NewEncoder(make([]byte, 0, AccountSize))
	enc.WritePublicKey(mint)
	enc.WritePublicKey(owner)
	enc.WriteUint64(99)
	enc.WriteCOption(true)
	enc.WritePublicKey(delegate)
	enc.WriteUint8(uint8(Initialized))
	enc.WriteCOption(false)
	enc.WriteUint64(0)
	enc.WriteUint64(7)
	enc.WriteCOption(true)
	enc.WritePublicKey(closeAuthority)
	account, err := DecodeAccount(enc.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if account.Mint != mint || account.Owner != owner || account.Amount != 99 || account.Delegate == nil || *account.Delegate != delegate || account.CloseAuthority == nil || *account.CloseAuthority != closeAuthority {
		t.Fatalf("account = %+v", account)
	}
}

func FuzzDecodeInstruction(f *testing.F) {
	f.Add([]byte{byte(TransferInstruction), 1, 0, 0, 0, 0, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) { _, _ = DecodeInstruction(nil, data) })
}
