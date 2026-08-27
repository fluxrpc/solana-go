package vote

import (
	"bytes"
	"encoding/binary"
	"errors"
	"reflect"
	"testing"

	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

func testKey(value byte) (key solana.PublicKey) {
	for i := range key {
		key[i] = value
	}
	return key
}

func testHash(value byte) (hash solana.Hash) {
	for i := range hash {
		hash[i] = value
	}
	return hash
}

func testMeta(key solana.PublicKey, writable, signer bool) solana.AccountMeta {
	return solana.AccountMeta{PublicKey: key, IsWritable: writable, IsSigner: signer}
}

func mustTestInstruction[T solana.Instruction](instruction T, err error) T {
	if err != nil {
		panic(err)
	}
	return instruction
}

func testInstructions() []struct {
	typ  InstructionType
	inst solana.Instruction
	get  func(DecodedInstruction) any
} {
	k1, k2, k3, k4 := testKey(1), testKey(2), testKey(3), testKey(4)
	h1, h2, h3 := testHash(0x11), testHash(0x22), testHash(0x33)
	root, timestamp := uint64(99), int64(-1234)
	update := VoteStateUpdate{Lockouts: []Lockout{{100, 3}, {101, 2}, {105, 1}}, Root: &root, Hash: h1, Timestamp: &timestamp}
	sync := TowerSyncUpdate{Lockouts: update.Lockouts, Root: &root, Hash: h1, Timestamp: &timestamp, BlockID: h2}
	auth := VoteAuthorize{Kind: VoteAuthorizeVoter}
	seedArgs := VoteAuthorizeWithSeedArgs{AuthorizationType: auth, CurrentAuthorityDerivedKeyOwner: k2, CurrentAuthorityDerivedKeySeed: "seed", NewAuthority: k3}
	checkedSeedArgs := VoteAuthorizeCheckedWithSeedArgs{AuthorizationType: auth, CurrentAuthorityDerivedKeyOwner: k2, CurrentAuthorityDerivedKeySeed: "seed"}
	initV2 := VoteInitV2{NodePubkey: k1, AuthorizedVoter: k2, AuthorizedWithdrawer: k3, InflationRewardsCommissionBps: 101, BlockRevenueCommissionBps: 202}
	return []struct {
		typ  InstructionType
		inst solana.Instruction
		get  func(DecodedInstruction) any
	}{
		{InitializeAccountInstruction, NewInitializeAccountInstruction(k1, k2, k3, 7, k4), func(v DecodedInstruction) any { return v.InitializeAccount }},
		{AuthorizeInstruction, NewAuthorizeInstruction(k4, auth, k1, k2), func(v DecodedInstruction) any { return v.Authorize }},
		{VoteInstruction, NewVoteInstruction([]uint64{10, 11}, h1, &timestamp, k1, k2), func(v DecodedInstruction) any { return v.Vote }},
		{WithdrawInstruction, NewWithdrawInstruction(55, k1, k2, k3), func(v DecodedInstruction) any { return v.Withdraw }},
		{UpdateValidatorIdentityInstruction, NewUpdateValidatorIdentityInstruction(k1, k2, k3), func(v DecodedInstruction) any { return v.UpdateValidatorIdentity }},
		{UpdateCommissionInstruction, NewUpdateCommissionInstruction(9, k1, k2), func(v DecodedInstruction) any { return v.UpdateCommission }},
		{VoteSwitchInstruction, NewVoteSwitchInstruction([]uint64{10, 11}, h1, &timestamp, h2, k1, k2), func(v DecodedInstruction) any { return v.VoteSwitch }},
		{AuthorizeCheckedInstruction, NewAuthorizeCheckedInstruction(auth, k1, k2, k3), func(v DecodedInstruction) any { return v.AuthorizeChecked }},
		{UpdateVoteStateInstruction, NewUpdateVoteStateInstruction(update, k1, k2), func(v DecodedInstruction) any { return v.UpdateVoteState }},
		{UpdateVoteStateSwitchInstruction, NewUpdateVoteStateSwitchInstruction(update, h2, k1, k2), func(v DecodedInstruction) any { return v.UpdateVoteStateSwitch }},
		{AuthorizeWithSeedInstruction, mustTestInstruction(NewAuthorizeWithSeedInstruction(seedArgs, k1, k2)), func(v DecodedInstruction) any { return v.AuthorizeWithSeed }},
		{AuthorizeCheckedWithSeedInstruction, mustTestInstruction(NewAuthorizeCheckedWithSeedInstruction(checkedSeedArgs, k1, k2, k3)), func(v DecodedInstruction) any { return v.AuthorizeCheckedWithSeed }},
		{CompactUpdateVoteStateInstruction, NewCompactUpdateVoteStateInstruction(update, k1, k2), func(v DecodedInstruction) any { return v.CompactUpdateVoteState }},
		{CompactUpdateVoteStateSwitchInstruction, NewCompactUpdateVoteStateSwitchInstruction(update, h2, k1, k2), func(v DecodedInstruction) any { return v.CompactUpdateVoteStateSwitch }},
		{TowerSyncInstruction, NewTowerSyncInstruction(sync, k1, k2), func(v DecodedInstruction) any { return v.TowerSync }},
		{TowerSyncSwitchInstruction, NewTowerSyncSwitchInstruction(sync, h3, k1, k2), func(v DecodedInstruction) any { return v.TowerSyncSwitch }},
		{InitializeAccountV2Instruction, NewInitializeAccountV2Instruction(initV2, k4, k2, k3), func(v DecodedInstruction) any { return v.InitializeAccountV2 }},
		{UpdateCommissionCollectorInstruction, NewUpdateCommissionCollectorInstruction(CommissionKindBlockRevenue, k1, k2, k3), func(v DecodedInstruction) any { return v.UpdateCommissionCollector }},
		{UpdateCommissionBpsInstruction, NewUpdateCommissionBpsInstruction(250, CommissionKindInflationRewards, k1, k2), func(v DecodedInstruction) any { return v.UpdateCommissionBps }},
		{DepositDelegatorRewardsInstruction, NewDepositDelegatorRewardsInstruction(987, k1, k2), func(v DecodedInstruction) any { return v.DepositDelegatorRewards }},
	}
}

func TestAllInstructionsRoundTripAndTruncation(t *testing.T) {
	for _, tc := range testInstructions() {
		t.Run(tc.typ.String(), func(t *testing.T) {
			if tc.inst.ProgramID() != ProgramID {
				t.Fatalf("ProgramID = %s", tc.inst.ProgramID())
			}
			data, err := tc.inst.Data()
			if err != nil {
				t.Fatal(err)
			}
			if len(data) < 4 || InstructionType(binary.LittleEndian.Uint32(data)) != tc.typ {
				t.Fatalf("bad discriminator: %x", data)
			}

			decoded, err := DecodeInstruction(tc.inst.Accounts(), data)
			if err != nil {
				t.Fatal(err)
			}
			if decoded.Type != tc.typ {
				t.Fatalf("Type = %v", decoded.Type)
			}
			if got := tc.get(decoded); !reflect.DeepEqual(got, tc.inst) {
				t.Fatalf("round trip mismatch\n got: %#v\nwant: %#v", got, tc.inst)
			}
			if nonNilInstructionFields(decoded) != 1 {
				t.Fatalf("decoded union has %d populated fields", nonNilInstructionFields(decoded))
			}

			for length := 0; length < len(data); length++ {
				if _, err := DecodeInstruction(tc.inst.Accounts(), data[:length]); err == nil {
					t.Fatalf("prefix length %d unexpectedly decoded", length)
				}
			}
			withTrailing := append(append([]byte(nil), data...), 0xde, 0xad)
			if _, err := DecodeInstruction(tc.inst.Accounts(), withTrailing); err != nil {
				t.Fatalf("trailing bytes: %v", err)
			}
		})
	}
}

func nonNilInstructionFields(value DecodedInstruction) int {
	v := reflect.ValueOf(value)
	count := 0
	for i := 1; i < v.NumField(); i++ {
		if !v.Field(i).IsNil() {
			count++
		}
	}
	return count
}

func TestExactWireGoldens(t *testing.T) {
	k1, k2, k3 := testKey(1), testKey(2), testKey(3)
	h1, h2 := testHash(0xaa), testHash(0xbb)
	root := uint64(99)

	withdraw, _ := NewWithdrawInstruction(0x0102030405060708, k1, k2, k3).Data()
	wantWithdraw := []byte{3, 0, 0, 0, 8, 7, 6, 5, 4, 3, 2, 1}
	if !bytes.Equal(withdraw, wantWithdraw) {
		t.Fatalf("withdraw = %x", withdraw)
	}

	update := VoteStateUpdate{Lockouts: []Lockout{{100, 3}, {101, 2}, {105, 1}}, Root: &root, Hash: h1}
	compact, _ := NewCompactUpdateVoteStateInstruction(update, k1, k2).Data()
	wantCompact := []byte{12, 0, 0, 0, 99, 0, 0, 0, 0, 0, 0, 0, 3, 1, 3, 1, 2, 4, 1}
	wantCompact = append(wantCompact, bytes.Repeat([]byte{0xaa}, 32)...)
	wantCompact = append(wantCompact, 0)
	if !bytes.Equal(compact, wantCompact) {
		t.Fatalf("compact = %x\nwant    = %x", compact, wantCompact)
	}

	sync := TowerSyncUpdate{Lockouts: update.Lockouts, Root: &root, Hash: h1, BlockID: h2}
	tower, _ := NewTowerSyncInstruction(sync, k1, k2).Data()
	wantTower := append([]byte{14, 0, 0, 0}, wantCompact[4:]...)
	wantTower = append(wantTower, bytes.Repeat([]byte{0xbb}, 32)...)
	if !bytes.Equal(tower, wantTower) {
		t.Fatalf("tower = %x\nwant  = %x", tower, wantTower)
	}

	var blsKey [BLSPublicKeyCompressedSize]byte
	var proof [BLSProofOfPossessionCompressedSize]byte
	for i := range blsKey {
		blsKey[i] = 0x44
	}
	for i := range proof {
		proof[i] = 0x55
	}
	authorize, _ := NewAuthorizeInstruction(k3, VoteAuthorize{Kind: VoteAuthorizeVoterWithBLS, BLSPubkey: &blsKey, BLSProofOfPossession: &proof}, k1, k2).Data()
	wantAuthorize := []byte{1, 0, 0, 0}
	wantAuthorize = append(wantAuthorize, k3[:]...)
	wantAuthorize = append(wantAuthorize, 2, 0, 0, 0)
	wantAuthorize = append(wantAuthorize, bytes.Repeat([]byte{0x44}, 48)...)
	wantAuthorize = append(wantAuthorize, bytes.Repeat([]byte{0x55}, 96)...)
	if !bytes.Equal(authorize, wantAuthorize) {
		t.Fatalf("BLS authorize wire mismatch")
	}

	bps, _ := NewUpdateCommissionBpsInstruction(0x1234, CommissionKindBlockRevenue, k1, k2).Data()
	if want := []byte{18, 0, 0, 0, 0x34, 0x12, 1, 0, 0, 0}; !bytes.Equal(bps, want) {
		t.Fatalf("bps = %x", bps)
	}
	collector, _ := NewUpdateCommissionCollectorInstruction(CommissionKindBlockRevenue, k1, k2, k3).Data()
	if want := []byte{17, 0, 0, 0, 1, 0, 0, 0}; !bytes.Equal(collector, want) {
		t.Fatalf("collector = %x", collector)
	}
	deposit, _ := NewDepositDelegatorRewardsInstruction(0x0102030405060708, k1, k2).Data()
	if want := []byte{19, 0, 0, 0, 8, 7, 6, 5, 4, 3, 2, 1}; !bytes.Equal(deposit, want) {
		t.Fatalf("deposit = %x", deposit)
	}
}

func TestAccountMetas(t *testing.T) {
	k1, k2, k3, k4 := testKey(1), testKey(2), testKey(3), testKey(4)
	cases := []struct {
		name string
		inst solana.Instruction
		want []solana.AccountMeta
	}{
		{"initialize", NewInitializeAccountInstruction(k1, k2, k3, 0, k4), []solana.AccountMeta{testMeta(k4, true, false), testMeta(solana.SysVarRentPubkey, false, false), testMeta(solana.SysVarClockPubkey, false, false), testMeta(k1, false, true)}},
		{"vote", NewVoteInstruction(nil, solana.Hash{}, nil, k1, k2), []solana.AccountMeta{testMeta(k1, true, false), testMeta(solana.SysVarSlotHashesPubkey, false, false), testMeta(solana.SysVarClockPubkey, false, false), testMeta(k2, false, true)}},
		{"withdraw", NewWithdrawInstruction(1, k1, k2, k3), []solana.AccountMeta{testMeta(k1, true, false), testMeta(k2, true, false), testMeta(k3, false, true)}},
		{"checked", NewAuthorizeCheckedInstruction(VoteAuthorize{Kind: VoteAuthorizeWithdrawer}, k1, k2, k3), []solana.AccountMeta{testMeta(k1, true, false), testMeta(solana.SysVarClockPubkey, false, false), testMeta(k2, false, true), testMeta(k3, false, true)}},
		{"collector", NewUpdateCommissionCollectorInstruction(0, k1, k2, k3), []solana.AccountMeta{testMeta(k1, true, false), testMeta(k2, true, false), testMeta(k3, false, true)}},
		{"initialize-v2", NewInitializeAccountV2Instruction(VoteInitV2{NodePubkey: k1}, k4, k2, k3), []solana.AccountMeta{testMeta(k4, true, false), testMeta(k1, false, true), testMeta(k2, true, false), testMeta(k3, true, false)}},
		{"deposit", NewDepositDelegatorRewardsInstruction(1, k1, k2), []solana.AccountMeta{testMeta(k1, true, false), testMeta(k2, true, true)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			accounts := tc.inst.Accounts()
			if len(accounts) != len(tc.want) {
				t.Fatalf("len = %d", len(accounts))
			}
			for i, want := range tc.want {
				if accounts[i] == nil || *accounts[i] != want {
					t.Errorf("account %d = %#v, want %#v", i, accounts[i], want)
				}
			}
		})
	}
}

func TestDecodeRejectsInvalidData(t *testing.T) {
	if _, err := DecodeInstruction(nil, []byte{99, 0, 0, 0}); !errors.Is(err, ErrUnknownInstruction) {
		t.Fatalf("unknown error = %v", err)
	}
	if _, err := DecodeInstruction(nil, []byte{17, 0, 0, 0, 2, 0, 0, 0}); !errors.Is(err, ErrInvalidCommissionKind) {
		t.Fatalf("commission error = %v", err)
	}
	if _, err := DecodeInstruction(nil, []byte{17, 0, 0, 0, 0, 1, 0, 0}); !errors.Is(err, ErrInvalidCommissionKind) {
		t.Fatalf("narrowing commission error = %v", err)
	}
	if _, err := DecodeInstruction(nil, []byte{1, 0, 0, 0}); err == nil {
		t.Fatal("truncated authorize accepted")
	}

	tooMany := VoteStateUpdate{Lockouts: make([]Lockout, MaxLockoutHistory+1)}
	if _, err := NewUpdateVoteStateInstruction(tooMany, testKey(1), testKey(2)).Data(); !errors.Is(err, ErrTooManyLockouts) {
		t.Fatalf("lockout error = %v", err)
	}
	badConfirmation := VoteStateUpdate{Lockouts: []Lockout{{ConfirmationCount: 256}}}
	if _, err := NewCompactUpdateVoteStateInstruction(badConfirmation, testKey(1), testKey(2)).Data(); !errors.Is(err, ErrInvalidConfirmationCount) {
		t.Fatalf("confirmation error = %v", err)
	}
	if _, err := NewAuthorizeInstruction(testKey(3), VoteAuthorize{Kind: VoteAuthorizeVoterWithBLS}, testKey(1), testKey(2)).Data(); err == nil {
		t.Fatal("nil BLS material accepted")
	}
	longSeed := string(bytes.Repeat([]byte{'s'}, solana.MaxSeedLength+1))
	if _, err := NewAuthorizeWithSeedInstruction(VoteAuthorizeWithSeedArgs{CurrentAuthorityDerivedKeySeed: longSeed}, testKey(1), testKey(2)); !errors.Is(err, solana.ErrMaxSeedLengthExceeded) {
		t.Fatalf("seed constructor error = %v", err)
	}
	seeded := mustTestInstruction(NewAuthorizeCheckedWithSeedInstruction(VoteAuthorizeCheckedWithSeedArgs{CurrentAuthorityDerivedKeySeed: "ok"}, testKey(1), testKey(2), testKey(3)))
	seeded.CurrentAuthorityDerivedKeySeed = longSeed
	if _, err := seeded.Data(); !errors.Is(err, solana.ErrMaxSeedLengthExceeded) {
		t.Fatalf("mutated seed error = %v", err)
	}

	badSeed := []byte{11, 0, 0, 0, 0, 0, 0, 0}
	badSeed = append(badSeed, make([]byte, 32)...)
	badSeed = binary.LittleEndian.AppendUint64(badSeed, uint64(solana.MaxSeedLength+1))
	badSeed = append(badSeed, bytes.Repeat([]byte{'s'}, solana.MaxSeedLength+1)...)
	if _, err := DecodeInstruction(nil, badSeed); !errors.Is(err, solana.ErrMaxSeedLengthExceeded) {
		t.Fatalf("decoded seed error = %v", err)
	}
	badUTF8 := []byte{11, 0, 0, 0, 0, 0, 0, 0}
	badUTF8 = append(badUTF8, make([]byte, 32)...)
	badUTF8 = binary.LittleEndian.AppendUint64(badUTF8, 1)
	badUTF8 = append(badUTF8, 0xff)
	if _, err := DecodeInstruction(nil, badUTF8); !errors.Is(err, bin.ErrInvalidUTF8) {
		t.Fatalf("UTF-8 error = %v", err)
	}

	// A non-canonical compact-u16 count must not have an aliasing encoding.
	nonCanonical := append([]byte{12, 0, 0, 0}, bytes.Repeat([]byte{0xff}, 8)...)
	nonCanonical = append(nonCanonical, 0x80, 0x00)
	if _, err := DecodeInstruction(nil, nonCanonical); !errors.Is(err, bin.ErrNonCanonical) {
		t.Fatalf("noncanonical error = %v", err)
	}
}

func TestCompactNilRootAndOrder(t *testing.T) {
	update := VoteStateUpdate{Lockouts: []Lockout{{Slot: 100, ConfirmationCount: 3}, {Slot: 101, ConfirmationCount: 2}, {Slot: 105, ConfirmationCount: 1}}, Hash: testHash(1)}
	data, err := NewCompactUpdateVoteStateInstruction(update, testKey(1), testKey(2)).Data()
	if err != nil {
		t.Fatal(err)
	}
	// Root=None is u64::MAX on wire, but the first offset is relative to zero.
	if got, want := data[13:19], []byte{100, 3, 1, 2, 4, 1}; !bytes.Equal(got, want) {
		t.Fatalf("nil-root offsets = %v, want %v", got, want)
	}
	decoded, err := DecodeInstruction(nil, data)
	if err != nil {
		t.Fatal(err)
	}
	if got := decoded.CompactUpdateVoteState.Lockouts; !reflect.DeepEqual(got, update.Lockouts) {
		t.Fatalf("lockouts = %#v, want %#v", got, update.Lockouts)
	}
	towerData, err := NewTowerSyncInstruction(TowerSyncUpdate{Lockouts: update.Lockouts, Hash: update.Hash, BlockID: testHash(2)}, testKey(1), testKey(2)).Data()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := towerData[13:19], []byte{100, 3, 1, 2, 4, 1}; !bytes.Equal(got, want) {
		t.Fatalf("tower nil-root offsets = %v, want %v", got, want)
	}
	towerDecoded, err := DecodeInstruction(nil, towerData)
	if err != nil {
		t.Fatal(err)
	}
	if got := towerDecoded.TowerSync.Lockouts; !reflect.DeepEqual(got, update.Lockouts) {
		t.Fatalf("tower lockouts = %#v, want %#v", got, update.Lockouts)
	}

	bad := VoteStateUpdate{Lockouts: []Lockout{{Slot: 10}, {Slot: 9}}}
	if _, err := NewCompactUpdateVoteStateInstruction(bad, testKey(1), testKey(2)).Data(); !errors.Is(err, ErrInvalidLockoutOrder) {
		t.Fatalf("order error = %v", err)
	}
	root := uint64(10)
	belowRoot := VoteStateUpdate{Root: &root, Lockouts: []Lockout{{Slot: 9}}}
	if _, err := NewCompactUpdateVoteStateInstruction(belowRoot, testKey(1), testKey(2)).Data(); !errors.Is(err, ErrInvalidLockoutOrder) {
		t.Fatalf("root order error = %v", err)
	}
	equal := VoteStateUpdate{Root: &root, Lockouts: []Lockout{{Slot: 10}, {Slot: 10}}}
	if _, err := NewCompactUpdateVoteStateInstruction(equal, testKey(1), testKey(2)).Data(); err != nil {
		t.Fatalf("equal slots: %v", err)
	}
	if _, err := NewTowerSyncInstruction(TowerSyncUpdate{Lockouts: bad.Lockouts}, testKey(1), testKey(2)).Data(); !errors.Is(err, ErrInvalidLockoutOrder) {
		t.Fatalf("tower order error = %v", err)
	}
}

func FuzzDecodeInstruction(f *testing.F) {
	for _, tc := range testInstructions() {
		data, _ := tc.inst.Data()
		f.Add(data)
	}
	f.Fuzz(func(t *testing.T, data []byte) { _, _ = DecodeInstruction(nil, data) })
}
