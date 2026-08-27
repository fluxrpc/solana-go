package stake

import (
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

type accountSpec struct {
	key      solana.PublicKey
	writable bool
	signer   bool
}

func TestCanonicalAccountMetas(t *testing.T) {
	fixtures := instructionFixtures(t)
	want := map[string][]accountSpec{
		"Initialize": {
			{testKey(8), true, false}, {solana.SysVarRentPubkey, false, false},
		},
		"Authorize": {
			{testKey(1), true, false}, {solana.SysVarClockPubkey, false, false}, {testKey(2), false, true}, {testKey(7), false, true},
		},
		"DelegateStake": {
			{testKey(1), true, false}, {testKey(2), false, false}, {solana.SysVarClockPubkey, false, false},
			{solana.SysVarStakeHistoryPubkey, false, false}, {solana.SysVarStakeConfigPubkey, false, false}, {testKey(3), false, true},
		},
		"Split": {
			{testKey(1), true, false}, {testKey(2), true, false}, {testKey(3), false, true},
		},
		"Withdraw": {
			{testKey(1), true, false}, {testKey(2), true, false}, {solana.SysVarClockPubkey, false, false},
			{solana.SysVarStakeHistoryPubkey, false, false}, {testKey(3), false, true}, {testKey(7), false, true},
		},
		"Deactivate": {
			{testKey(1), true, false}, {solana.SysVarClockPubkey, false, false}, {testKey(2), false, true},
		},
		"SetLockup": {
			{testKey(1), true, false}, {testKey(2), false, true},
		},
		"Merge": {
			{testKey(1), true, false}, {testKey(2), true, false}, {solana.SysVarClockPubkey, false, false},
			{solana.SysVarStakeHistoryPubkey, false, false}, {testKey(3), false, true},
		},
		"AuthorizeWithSeed": {
			{testKey(1), true, false}, {testKey(2), false, true}, {solana.SysVarClockPubkey, false, false}, {testKey(7), false, true},
		},
		"InitializeChecked": {
			{testKey(1), true, false}, {solana.SysVarRentPubkey, false, false}, {testKey(2), false, false}, {testKey(3), false, true},
		},
		"AuthorizeChecked": {
			{testKey(1), true, false}, {solana.SysVarClockPubkey, false, false}, {testKey(2), false, true},
			{testKey(3), false, true}, {testKey(7), false, true},
		},
		"AuthorizeCheckedWithSeed": {
			{testKey(1), true, false}, {testKey(2), false, true}, {solana.SysVarClockPubkey, false, false},
			{testKey(4), false, true}, {testKey(7), false, true},
		},
		"SetLockupChecked": {
			{testKey(1), true, false}, {testKey(2), false, true}, {testKey(7), false, true},
		},
		"GetMinimumDelegation": nil,
		"DeactivateDelinquent": {
			{testKey(1), true, false}, {testKey(2), false, false}, {testKey(3), false, false},
		},
		"Redelegate": {
			{testKey(1), true, false}, {testKey(2), true, false}, {testKey(3), false, false},
			{solana.SysVarStakeConfigPubkey, false, false}, {testKey(4), false, true},
		},
		"MoveStake": {
			{testKey(1), true, false}, {testKey(2), true, false}, {testKey(3), false, true},
		},
		"MoveLamports": {
			{testKey(1), true, false}, {testKey(2), true, false}, {testKey(3), false, true},
		},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			accounts := fixture.instruction.Accounts()
			expected := want[fixture.name]
			if len(accounts) != len(expected) {
				t.Fatalf("account count = %d, want %d", len(accounts), len(expected))
			}
			for i, spec := range expected {
				meta := accounts[i]
				if meta.PublicKey != spec.key || meta.IsWritable != spec.writable || meta.IsSigner != spec.signer {
					t.Errorf("account[%d] = {%s writable=%t signer=%t}, want {%s writable=%t signer=%t}", i, meta.PublicKey, meta.IsWritable, meta.IsSigner, spec.key, spec.writable, spec.signer)
				}
			}
		})
	}
}

func TestOptionalAccountsAreOmitted(t *testing.T) {
	if got := len(NewAuthorizeInstruction(testKey(1), StakeAuthorizeStaker, testKey(2), testKey(3), nil).Accounts()); got != 3 {
		t.Fatalf("Authorize accounts = %d", got)
	}
	if got := len(NewWithdrawInstruction(1, testKey(1), testKey(2), testKey(3), nil).Accounts()); got != 5 {
		t.Fatalf("Withdraw accounts = %d", got)
	}
	if got := len(NewSetLockupCheckedInstruction(LockupCheckedArgs{}, testKey(1), testKey(2), nil).Accounts()); got != 2 {
		t.Fatalf("SetLockupChecked accounts = %d", got)
	}
}
