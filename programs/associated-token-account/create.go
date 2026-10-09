package associatedtokenaccount

import (
	"errors"
	"fmt"

	solana "github.com/fluxrpc/solana-go"
)

// Create builds the instruction that creates an associated token account for
// Wallet and Mint. It emits the CreateIdempotent variant (data = [1]), so it
// succeeds if the account already exists.
//
// Accounts:
//
//	[0] WRITE, SIGNER  Payer                  funding account
//	[1] WRITE          AssociatedTokenAccount address to be created
//	[2]                Wallet                 owner of the new account
//	[3]                Mint                   token mint
//	[4]                SystemProgram
//	[5]                TokenProgram
type Create struct {
	Payer        solana.PublicKey
	Wallet       solana.PublicKey
	Mint         solana.PublicKey
	TokenProgram solana.PublicKey
}

// NewCreateInstructionBuilder returns a Create using the SPL Token program.
func NewCreateInstructionBuilder() *Create {
	return &Create{TokenProgram: solana.TokenProgramID}
}

func NewCreateInstruction(payer, wallet, mint, tokenProgram solana.PublicKey) *Create {
	return NewCreateInstructionBuilder().
		SetPayer(payer).
		SetWallet(wallet).
		SetMint(mint).
		SetTokenProgram(tokenProgram)
}

func (inst *Create) SetPayer(payer solana.PublicKey) *Create {
	inst.Payer = payer
	return inst
}

func (inst *Create) SetWallet(wallet solana.PublicKey) *Create {
	inst.Wallet = wallet
	return inst
}

func (inst *Create) SetMint(mint solana.PublicKey) *Create {
	inst.Mint = mint
	return inst
}

func (inst *Create) SetTokenProgram(tokenProgram solana.PublicKey) *Create {
	inst.TokenProgram = tokenProgram
	return inst
}

// Instruction is a built associated-token-account instruction.
type Instruction struct {
	accounts []*solana.AccountMeta
}

func (inst *Instruction) ProgramID() solana.PublicKey { return ProgramID }

func (inst *Instruction) Accounts() []*solana.AccountMeta { return inst.accounts }

func (inst *Instruction) Data() ([]byte, error) { return []byte{1}, nil }

// Build derives the associated token address and returns the instruction.
func (inst Create) Build() *Instruction {
	ata, _, _ := inst.Wallet.FindAssociatedTokenAddressWithProgram(inst.Mint, inst.TokenProgram)

	return &Instruction{accounts: []*solana.AccountMeta{
		{PublicKey: inst.Payer, IsSigner: true, IsWritable: true},
		{PublicKey: ata, IsWritable: true},
		{PublicKey: inst.Wallet},
		{PublicKey: inst.Mint},
		{PublicKey: solana.SystemProgramID},
		{PublicKey: inst.TokenProgram},
	}}
}

// ValidateAndBuild validates the fields and builds the instruction.
func (inst Create) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *Create) Validate() error {
	if inst.Payer.IsZero() {
		return errors.New("Payer not set")
	}
	if inst.Wallet.IsZero() {
		return errors.New("Wallet not set")
	}
	if inst.Mint.IsZero() {
		return errors.New("Mint not set")
	}
	if _, _, err := inst.Wallet.FindAssociatedTokenAddressWithProgram(inst.Mint, inst.TokenProgram); err != nil {
		return fmt.Errorf("error while FindAssociatedTokenAddress: %w", err)
	}
	return nil
}
