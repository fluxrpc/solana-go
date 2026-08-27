package stake

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"

	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

func testKey(value byte) solana.PublicKey {
	var key solana.PublicKey
	key[0] = value
	return key
}

func le32(value uint32) []byte {
	var out [4]byte
	binary.LittleEndian.PutUint32(out[:], value)
	return out[:]
}
func le64(value uint64) []byte {
	var out [8]byte
	binary.LittleEndian.PutUint64(out[:], value)
	return out[:]
}
func pubkeyBytes(value byte) []byte {
	out := make([]byte, solana.PublicKeyLength)
	out[0] = value
	return out
}

func join(parts ...[]byte) []byte {
	var size int
	for _, part := range parts {
		size += len(part)
	}
	out := make([]byte, 0, size)
	for _, part := range parts {
		out = append(out, part...)
	}
	return out
}

func instructionFixtures(t testing.TB) []struct {
	name        string
	instruction solana.Instruction
	want        []byte
} {
	t.Helper()
	timestamp := int64(-7)
	epoch := uint64(9)
	custodian := testKey(7)
	seeded, err := NewAuthorizeWithSeedInstruction(AuthorizeWithSeedArgs{
		NewAuthorized: testKey(4), StakeAuthorize: StakeAuthorizeWithdrawer, AuthoritySeed: "seed", AuthorityOwner: testKey(5),
	}, testKey(1), testKey(2), &custodian)
	if err != nil {
		t.Fatal(err)
	}
	checkedSeeded, err := NewAuthorizeCheckedWithSeedInstruction(AuthorizeCheckedWithSeedArgs{
		StakeAuthorize: StakeAuthorizeStaker, AuthoritySeed: "x", AuthorityOwner: testKey(5),
	}, testKey(1), testKey(2), testKey(4), &custodian)
	if err != nil {
		t.Fatal(err)
	}

	return []struct {
		name        string
		instruction solana.Instruction
		want        []byte
	}{
		{"Initialize", NewInitializeInstruction(Authorized{testKey(1), testKey(2)}, Lockup{-7, 9, testKey(3)}, testKey(8)),
			join(le32(0), pubkeyBytes(1), pubkeyBytes(2), le64(uint64(timestamp)), le64(9), pubkeyBytes(3))},
		{"Authorize", NewAuthorizeInstruction(testKey(4), StakeAuthorizeWithdrawer, testKey(1), testKey(2), &custodian), join(le32(1), pubkeyBytes(4), le32(1))},
		{"DelegateStake", NewDelegateStakeInstruction(testKey(2), testKey(3), testKey(1)), le32(2)},
		{"Split", NewSplitInstruction(42, testKey(1), testKey(2), testKey(3)), join(le32(3), le64(42))},
		{"Withdraw", NewWithdrawInstruction(43, testKey(1), testKey(2), testKey(3), &custodian), join(le32(4), le64(43))},
		{"Deactivate", NewDeactivateInstruction(testKey(1), testKey(2)), le32(5)},
		{"SetLockup", NewSetLockupInstruction(LockupArgs{&timestamp, &epoch, &custodian}, testKey(1), testKey(2)),
			join(le32(6), []byte{1}, le64(uint64(timestamp)), []byte{1}, le64(epoch), []byte{1}, pubkeyBytes(7))},
		{"Merge", NewMergeInstruction(testKey(1), testKey(2), testKey(3)), le32(7)},
		{"AuthorizeWithSeed", seeded, join(le32(8), pubkeyBytes(4), le32(1), le64(4), []byte("seed"), pubkeyBytes(5))},
		{"InitializeChecked", NewInitializeCheckedInstruction(Authorized{testKey(2), testKey(3)}, testKey(1)), le32(9)},
		{"AuthorizeChecked", NewAuthorizeCheckedInstruction(StakeAuthorizeWithdrawer, testKey(1), testKey(2), testKey(3), &custodian), join(le32(10), le32(1))},
		{"AuthorizeCheckedWithSeed", checkedSeeded, join(le32(11), le32(0), le64(1), []byte("x"), pubkeyBytes(5))},
		{"SetLockupChecked", NewSetLockupCheckedInstruction(LockupCheckedArgs{&timestamp, &epoch}, testKey(1), testKey(2), &custodian),
			join(le32(12), []byte{1}, le64(uint64(timestamp)), []byte{1}, le64(epoch))},
		{"GetMinimumDelegation", NewGetMinimumDelegationInstruction(), le32(13)},
		{"DeactivateDelinquent", NewDeactivateDelinquentInstruction(testKey(1), testKey(2), testKey(3)), le32(14)},
		{"Redelegate", NewRedelegateInstruction(testKey(1), testKey(2), testKey(3), testKey(4)), le32(15)},
		{"MoveStake", NewMoveStakeInstruction(44, testKey(1), testKey(2), testKey(3)), join(le32(16), le64(44))},
		{"MoveLamports", NewMoveLamportsInstruction(45, testKey(1), testKey(2), testKey(3)), join(le32(17), le64(45))},
	}
}

func TestInstructionGoldenDataAndTypedRoundTrip(t *testing.T) {
	for _, fixture := range instructionFixtures(t) {
		t.Run(fixture.name, func(t *testing.T) {
			if fixture.instruction.ProgramID() != ProgramID {
				t.Fatalf("ProgramID = %s", fixture.instruction.ProgramID())
			}
			got, err := fixture.instruction.Data()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, fixture.want) {
				t.Fatalf("Data = %x\nwant   %x", got, fixture.want)
			}

			decoded, err := DecodeInstruction(fixture.instruction.Accounts(), got)
			if err != nil {
				t.Fatal(err)
			}
			roundTrip, err := decodedData(decoded)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(roundTrip, fixture.want) {
				t.Fatalf("decoded Data = %x, want %x", roundTrip, fixture.want)
			}
			if len(fixture.instruction.Accounts()) != 0 && &decodedAccounts(decoded)[0] != &fixture.instruction.Accounts()[0] {
				t.Fatal("decoded accounts do not retain supplied slice")
			}
		})
	}
}

func decodedData(decoded DecodedInstruction) ([]byte, error) {
	switch decoded.Type {
	case InitializeInstruction:
		return decoded.Initialize.Data()
	case AuthorizeInstruction:
		return decoded.Authorize.Data()
	case DelegateStakeInstruction:
		return decoded.DelegateStake.Data()
	case SplitInstruction:
		return decoded.Split.Data()
	case WithdrawInstruction:
		return decoded.Withdraw.Data()
	case DeactivateInstruction:
		return decoded.Deactivate.Data()
	case SetLockupInstruction:
		return decoded.SetLockup.Data()
	case MergeInstruction:
		return decoded.Merge.Data()
	case AuthorizeWithSeedInstruction:
		return decoded.AuthorizeWithSeed.Data()
	case InitializeCheckedInstruction:
		return decoded.InitializeChecked.Data()
	case AuthorizeCheckedInstruction:
		return decoded.AuthorizeChecked.Data()
	case AuthorizeCheckedWithSeedInstruction:
		return decoded.AuthorizeCheckedWithSeed.Data()
	case SetLockupCheckedInstruction:
		return decoded.SetLockupChecked.Data()
	case GetMinimumDelegationInstruction:
		return decoded.GetMinimumDelegation.Data()
	case DeactivateDelinquentInstruction:
		return decoded.DeactivateDelinquent.Data()
	case RedelegateInstruction:
		return decoded.Redelegate.Data()
	case MoveStakeInstruction:
		return decoded.MoveStake.Data()
	case MoveLamportsInstruction:
		return decoded.MoveLamports.Data()
	default:
		return nil, ErrUnknownInstruction
	}
}

func decodedAccounts(decoded DecodedInstruction) solana.AccountMetaSlice {
	switch decoded.Type {
	case InitializeInstruction:
		return decoded.Initialize.AccountMetaSlice
	case AuthorizeInstruction:
		return decoded.Authorize.AccountMetaSlice
	case DelegateStakeInstruction:
		return decoded.DelegateStake.AccountMetaSlice
	case SplitInstruction:
		return decoded.Split.AccountMetaSlice
	case WithdrawInstruction:
		return decoded.Withdraw.AccountMetaSlice
	case DeactivateInstruction:
		return decoded.Deactivate.AccountMetaSlice
	case SetLockupInstruction:
		return decoded.SetLockup.AccountMetaSlice
	case MergeInstruction:
		return decoded.Merge.AccountMetaSlice
	case AuthorizeWithSeedInstruction:
		return decoded.AuthorizeWithSeed.AccountMetaSlice
	case InitializeCheckedInstruction:
		return decoded.InitializeChecked.AccountMetaSlice
	case AuthorizeCheckedInstruction:
		return decoded.AuthorizeChecked.AccountMetaSlice
	case AuthorizeCheckedWithSeedInstruction:
		return decoded.AuthorizeCheckedWithSeed.AccountMetaSlice
	case SetLockupCheckedInstruction:
		return decoded.SetLockupChecked.AccountMetaSlice
	case GetMinimumDelegationInstruction:
		return decoded.GetMinimumDelegation.AccountMetaSlice
	case DeactivateDelinquentInstruction:
		return decoded.DeactivateDelinquent.AccountMetaSlice
	case RedelegateInstruction:
		return decoded.Redelegate.AccountMetaSlice
	case MoveStakeInstruction:
		return decoded.MoveStake.AccountMetaSlice
	case MoveLamportsInstruction:
		return decoded.MoveLamports.AccountMetaSlice
	default:
		return nil
	}
}

func TestDecodeInstructionRejectsEveryTruncation(t *testing.T) {
	for _, fixture := range instructionFixtures(t) {
		t.Run(fixture.name, func(t *testing.T) {
			for cut := 0; cut < len(fixture.want); cut++ {
				if _, err := DecodeInstruction(nil, fixture.want[:cut]); err == nil {
					t.Fatalf("cut %d decoded successfully", cut)
				}
			}
		})
	}
}

func TestDecodeInstructionAllowsTrailingData(t *testing.T) {
	data := append(append([]byte(nil), le32(uint32(MoveStakeInstruction))...), le64(11)...)
	data = append(data, 0xaa, 0xbb)
	decoded, err := DecodeInstruction(nil, data)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.MoveStake == nil || decoded.MoveStake.Lamports != 11 {
		t.Fatalf("decoded = %+v", decoded)
	}
}

func TestDecodeInstructionMalformedValues(t *testing.T) {
	t.Run("unknown instruction", func(t *testing.T) {
		_, err := DecodeInstruction(nil, le32(18))
		if !errors.Is(err, ErrUnknownInstruction) {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("authority enum", func(t *testing.T) {
		data := join(le32(uint32(AuthorizeCheckedInstruction)), le32(2))
		_, err := DecodeInstruction(nil, data)
		if !errors.Is(err, ErrInvalidStakeAuthorize) {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("option tag", func(t *testing.T) {
		data := join(le32(uint32(SetLockupInstruction)), []byte{2})
		_, err := DecodeInstruction(nil, data)
		if !errors.Is(err, bin.ErrInvalidTag) {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("string length overflow", func(t *testing.T) {
		data := join(le32(uint32(AuthorizeCheckedWithSeedInstruction)), le32(0), le64(^uint64(0)))
		_, err := DecodeInstruction(nil, data)
		if !errors.Is(err, bin.ErrOverflow) {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("seed limit", func(t *testing.T) {
		data := join(le32(uint32(AuthorizeCheckedWithSeedInstruction)), le32(0), le64(33), bytes.Repeat([]byte{'s'}, 33), pubkeyBytes(5))
		_, err := DecodeInstruction(nil, data)
		if !errors.Is(err, solana.ErrMaxSeedLengthExceeded) {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("invalid UTF-8 seed", func(t *testing.T) {
		data := join(le32(uint32(AuthorizeCheckedWithSeedInstruction)), le32(0), le64(1), []byte{0xff}, pubkeyBytes(5))
		_, err := DecodeInstruction(nil, data)
		if !errors.Is(err, bin.ErrInvalidUTF8) {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestInstructionMutationValidation(t *testing.T) {
	bad := NewAuthorizeCheckedInstruction(StakeAuthorize(9), testKey(1), testKey(2), testKey(3), nil)
	if _, err := bad.Data(); !errors.Is(err, ErrInvalidStakeAuthorize) {
		t.Fatalf("error = %v", err)
	}
	good, err := NewAuthorizeWithSeedInstruction(AuthorizeWithSeedArgs{StakeAuthorize: StakeAuthorizeStaker, AuthoritySeed: "ok"}, testKey(1), testKey(2), nil)
	if err != nil {
		t.Fatal(err)
	}
	good.Args.AuthoritySeed = string(bytes.Repeat([]byte{'s'}, 33))
	if _, err := good.Data(); !errors.Is(err, solana.ErrMaxSeedLengthExceeded) {
		t.Fatalf("error = %v", err)
	}
	good.Args.AuthoritySeed = string([]byte{0xff})
	if _, err := good.Data(); !errors.Is(err, bin.ErrInvalidUTF8) {
		t.Fatalf("error = %v", err)
	}
	if _, err := NewAuthorizeCheckedWithSeedInstruction(
		AuthorizeCheckedWithSeedArgs{StakeAuthorize: StakeAuthorizeStaker, AuthoritySeed: string(bytes.Repeat([]byte{'s'}, 33))},
		testKey(1), testKey(2), testKey(3), nil,
	); !errors.Is(err, solana.ErrMaxSeedLengthExceeded) {
		t.Fatalf("constructor error = %v", err)
	}
}

func TestNoneOptionGoldenData(t *testing.T) {
	lockup := NewSetLockupInstruction(LockupArgs{}, testKey(1), testKey(2))
	data, err := lockup.Data()
	if err != nil {
		t.Fatal(err)
	}
	if want := join(le32(uint32(SetLockupInstruction)), []byte{0, 0, 0}); !bytes.Equal(data, want) {
		t.Fatalf("SetLockup Data = %x, want %x", data, want)
	}
	checked := NewSetLockupCheckedInstruction(LockupCheckedArgs{}, testKey(1), testKey(2), nil)
	data, err = checked.Data()
	if err != nil {
		t.Fatal(err)
	}
	if want := join(le32(uint32(SetLockupCheckedInstruction)), []byte{0, 0}); !bytes.Equal(data, want) {
		t.Fatalf("SetLockupChecked Data = %x, want %x", data, want)
	}
}

func FuzzDecodeInstruction(f *testing.F) {
	for _, fixture := range instructionFixtures(f) {
		f.Add(fixture.want)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		decoded, err := DecodeInstruction(nil, data)
		if err != nil {
			return
		}
		if _, err := decodedData(decoded); err != nil {
			t.Fatalf("successful decode did not re-encode: %v", err)
		}
	})
}
