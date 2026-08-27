package benchcmp

import (
	"bytes"
	"testing"

	flux "github.com/fluxrpc/solana-go"
	fluxsystem "github.com/fluxrpc/solana-go/programs/system"
	foundation "github.com/solana-foundation/solana-go/v2"
	foundationsystem "github.com/solana-foundation/solana-go/v2/programs/system"
)

func systemParityFluxKey(fill byte) (key flux.PublicKey) {
	for index := range key {
		key[index] = fill
	}
	return key
}

func systemParityFoundationKey(key flux.PublicKey) (foundation foundation.PublicKey) {
	copy(foundation[:], key[:])
	return foundation
}

func TestSystemInstructionParity(t *testing.T) {
	owner := systemParityFluxKey(0x11)
	funding := systemParityFluxKey(0x22)
	target := systemParityFluxKey(0x33)
	base := systemParityFluxKey(0x44)
	nonce := systemParityFluxKey(0x55)
	recipient := systemParityFluxKey(0x66)
	recentBlockhashes := systemParityFluxKey(0x77)
	rent := systemParityFluxKey(0x88)
	authority := systemParityFluxKey(0x99)
	newAuthority := systemParityFluxKey(0xaa)

	foundationOwner := systemParityFoundationKey(owner)
	foundationFunding := systemParityFoundationKey(funding)
	foundationTarget := systemParityFoundationKey(target)
	foundationBase := systemParityFoundationKey(base)
	foundationNonce := systemParityFoundationKey(nonce)
	foundationRecipient := systemParityFoundationKey(recipient)
	foundationRecentBlockhashes := systemParityFoundationKey(recentBlockhashes)
	foundationRent := systemParityFoundationKey(rent)
	foundationAuthority := systemParityFoundationKey(authority)
	foundationNewAuthority := systemParityFoundationKey(newAuthority)

	const (
		lamports = uint64(0x0102030405060708)
		space    = uint64(0x1112131415161718)
		seed     = "system-parity-seed"
	)

	fluxCreateWithSeed, err := fluxsystem.NewCreateAccountWithSeedInstruction(
		base, seed, lamports, space, owner, funding, target, base,
	)
	if err != nil {
		t.Fatalf("create flux CreateAccountWithSeed: %v", err)
	}
	fluxAllocateWithSeed, err := fluxsystem.NewAllocateWithSeedInstruction(
		base, seed, space, owner, target, base,
	)
	if err != nil {
		t.Fatalf("create flux AllocateWithSeed: %v", err)
	}
	fluxAssignWithSeed, err := fluxsystem.NewAssignWithSeedInstruction(
		base, seed, owner, target, base,
	)
	if err != nil {
		t.Fatalf("create flux AssignWithSeed: %v", err)
	}
	fluxTransferWithSeed, err := fluxsystem.NewTransferWithSeedInstruction(
		lamports, seed, owner, funding, base, recipient,
	)
	if err != nil {
		t.Fatalf("create flux TransferWithSeed: %v", err)
	}

	tests := []struct {
		name       string
		flux       flux.Instruction
		foundation foundation.Instruction
	}{
		{
			name:       "CreateAccount",
			flux:       fluxsystem.NewCreateAccountInstruction(lamports, space, owner, funding, target),
			foundation: foundationsystem.NewCreateAccountInstruction(lamports, space, foundationOwner, foundationFunding, foundationTarget).Build(),
		},
		{
			name:       "Assign",
			flux:       fluxsystem.NewAssignInstruction(owner, target),
			foundation: foundationsystem.NewAssignInstruction(foundationOwner, foundationTarget).Build(),
		},
		{
			name:       "Transfer",
			flux:       fluxsystem.NewTransferInstruction(lamports, funding, recipient),
			foundation: foundationsystem.NewTransferInstruction(lamports, foundationFunding, foundationRecipient).Build(),
		},
		{
			name: "CreateAccountWithSeed",
			flux: fluxCreateWithSeed,
			foundation: foundationsystem.NewCreateAccountWithSeedInstruction(
				foundationBase, seed, lamports, space, foundationOwner,
				foundationFunding, foundationTarget, foundationBase,
			).Build(),
		},
		{
			name: "AdvanceNonceAccount",
			flux: fluxsystem.NewAdvanceNonceAccountInstruction(
				nonce, recentBlockhashes, authority,
			),
			foundation: foundationsystem.NewAdvanceNonceAccountInstruction(
				foundationNonce, foundationRecentBlockhashes, foundationAuthority,
			).Build(),
		},
		{
			name: "WithdrawNonceAccount",
			flux: fluxsystem.NewWithdrawNonceAccountInstruction(
				lamports, nonce, recipient, recentBlockhashes, rent, authority,
			),
			foundation: foundationsystem.NewWithdrawNonceAccountInstruction(
				lamports, foundationNonce, foundationRecipient,
				foundationRecentBlockhashes, foundationRent, foundationAuthority,
			).Build(),
		},
		{
			name: "InitializeNonceAccount",
			flux: fluxsystem.NewInitializeNonceAccountInstruction(
				authority, nonce, recentBlockhashes, rent,
			),
			foundation: foundationsystem.NewInitializeNonceAccountInstruction(
				foundationAuthority, foundationNonce, foundationRecentBlockhashes, foundationRent,
			).Build(),
		},
		{
			name: "AuthorizeNonceAccount",
			flux: fluxsystem.NewAuthorizeNonceAccountInstruction(
				newAuthority, nonce, authority,
			),
			foundation: foundationsystem.NewAuthorizeNonceAccountInstruction(
				foundationNewAuthority, foundationNonce, foundationAuthority,
			).Build(),
		},
		{
			name:       "Allocate",
			flux:       fluxsystem.NewAllocateInstruction(space, target),
			foundation: foundationsystem.NewAllocateInstruction(space, foundationTarget).Build(),
		},
		{
			name: "AllocateWithSeed",
			flux: fluxAllocateWithSeed,
			foundation: foundationsystem.NewAllocateWithSeedInstruction(
				foundationBase, seed, space, foundationOwner, foundationTarget, foundationBase,
			).Build(),
		},
		{
			name: "AssignWithSeed",
			flux: fluxAssignWithSeed,
			foundation: foundationsystem.NewAssignWithSeedInstruction(
				foundationBase, seed, foundationOwner, foundationTarget, foundationBase,
			).Build(),
		},
		{
			name: "TransferWithSeed",
			flux: fluxTransferWithSeed,
			foundation: foundationsystem.NewTransferWithSeedInstruction(
				lamports, seed, foundationOwner,
				foundationFunding, foundationBase, foundationRecipient,
			).Build(),
		},
		{
			name:       "UpgradeNonceAccount",
			flux:       fluxsystem.NewUpgradeNonceAccountInstruction(nonce),
			foundation: foundationsystem.NewUpgradeNonceAccountInstruction(foundationNonce).Build(),
		},
	}

	if len(tests) != 13 {
		t.Fatalf("parity fixture count = %d, want 13", len(tests))
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			systemParityInstruction(t, test.flux, test.foundation)
		})
	}
}

func systemParityInstruction(t *testing.T, fluxInstruction flux.Instruction, foundationInstruction foundation.Instruction) {
	t.Helper()

	fluxProgramID := fluxInstruction.ProgramID()
	foundationProgramID := foundationInstruction.ProgramID()
	if !bytes.Equal(fluxProgramID[:], foundationProgramID[:]) {
		t.Errorf("ProgramID = %x, foundation %x", fluxProgramID, foundationProgramID)
	}

	fluxData, err := fluxInstruction.Data()
	if err != nil {
		t.Fatalf("flux Data: %v", err)
	}
	foundationData, err := foundationInstruction.Data()
	if err != nil {
		t.Fatalf("foundation Data: %v", err)
	}
	if !bytes.Equal(fluxData, foundationData) {
		t.Errorf("Data = %x, foundation %x", fluxData, foundationData)
	}

	fluxAccounts := fluxInstruction.Accounts()
	foundationAccounts := foundationInstruction.Accounts()
	if len(fluxAccounts) != len(foundationAccounts) {
		t.Fatalf("account count = %d, foundation %d", len(fluxAccounts), len(foundationAccounts))
	}
	for index, fluxAccount := range fluxAccounts {
		foundationAccount := foundationAccounts[index]
		if fluxAccount == nil || foundationAccount == nil {
			if fluxAccount != nil || foundationAccount != nil {
				t.Errorf("account %d nil mismatch", index)
			}
			continue
		}
		if !bytes.Equal(fluxAccount.PublicKey[:], foundationAccount.PublicKey[:]) {
			t.Errorf(
				"account %d public key = %x, foundation %x",
				index, fluxAccount.PublicKey, foundationAccount.PublicKey,
			)
		}
		if fluxAccount.IsSigner != foundationAccount.IsSigner {
			t.Errorf(
				"account %d IsSigner = %v, foundation %v",
				index, fluxAccount.IsSigner, foundationAccount.IsSigner,
			)
		}
		if fluxAccount.IsWritable != foundationAccount.IsWritable {
			t.Errorf(
				"account %d IsWritable = %v, foundation %v",
				index, fluxAccount.IsWritable, foundationAccount.IsWritable,
			)
		}
	}
}
