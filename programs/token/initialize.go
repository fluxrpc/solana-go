package token

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type InitializeMint struct {
	instruction
	Decimals        uint8
	MintAuthority   solana.PublicKey
	FreezeAuthority *solana.PublicKey
}

func NewInitializeMintInstruction(decimals uint8, mintAuthority solana.PublicKey, freezeAuthority solana.PublicKey, mint solana.PublicKey, rent solana.PublicKey) *InitializeMint {
	inst := newInitializeMint(decimals, mintAuthority, freezeAuthority)
	inst.AccountMetaSlice = solana.AccountMetaSlice{
		solana.NewAccountMeta(mint, true, false),
		solana.NewAccountMeta(rent, false, false),
	}
	return inst
}

func newInitializeMint(decimals uint8, mintAuthority solana.PublicKey, freezeAuthority solana.PublicKey) *InitializeMint {
	inst := &InitializeMint{Decimals: decimals, MintAuthority: mintAuthority}
	if !freezeAuthority.IsZero() {
		inst.FreezeAuthority = &freezeAuthority
	}
	return inst
}

func (inst *InitializeMint) tokenInstruction() *instruction { return &inst.instruction }
func (inst *InitializeMint) Data() ([]byte, error)          { return inst.data(InitializeMintInstruction) }
func (inst *InitializeMint) data(typ InstructionType) ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 67))
	enc.WriteUint8(uint8(typ))
	enc.WriteUint8(inst.Decimals)
	enc.WritePublicKey(inst.MintAuthority)
	enc.WriteOption(inst.FreezeAuthority != nil)
	if inst.FreezeAuthority != nil {
		enc.WritePublicKey(*inst.FreezeAuthority)
	}
	return enc.Bytes(), enc.Err()
}

type InitializeMint2 struct{ InitializeMint }

func NewInitializeMint2Instruction(decimals uint8, mintAuthority solana.PublicKey, freezeAuthority solana.PublicKey, mint solana.PublicKey) *InitializeMint2 {
	inst := &InitializeMint2{InitializeMint: *newInitializeMint(decimals, mintAuthority, freezeAuthority)}
	inst.AccountMetaSlice = solana.AccountMetaSlice{solana.NewAccountMeta(mint, true, false)}
	return inst
}

func (inst *InitializeMint2) tokenInstruction() *instruction { return &inst.instruction }
func (inst *InitializeMint2) Data() ([]byte, error)          { return inst.data(InitializeMint2Instruction) }

type InitializeAccount struct{ instruction }

func NewInitializeAccountInstruction(account, mint, owner, rent solana.PublicKey) *InitializeAccount {
	return &InitializeAccount{instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
		solana.NewAccountMeta(account, true, false),
		solana.NewAccountMeta(mint, false, false),
		solana.NewAccountMeta(owner, false, false),
		solana.NewAccountMeta(rent, false, false),
	}}}
}

func (inst *InitializeAccount) tokenInstruction() *instruction { return &inst.instruction }
func (inst *InitializeAccount) Data() ([]byte, error) {
	return []byte{byte(InitializeAccountInstruction)}, nil
}

type InitializeAccount2 struct {
	instruction
	Owner solana.PublicKey
}

func NewInitializeAccount2Instruction(owner, account, mint, rent solana.PublicKey) *InitializeAccount2 {
	return &InitializeAccount2{Owner: owner, instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
		solana.NewAccountMeta(account, true, false),
		solana.NewAccountMeta(mint, false, false),
		solana.NewAccountMeta(rent, false, false),
	}}}
}

func (inst *InitializeAccount2) tokenInstruction() *instruction { return &inst.instruction }
func (inst *InitializeAccount2) Data() ([]byte, error) {
	return publicKeyData(InitializeAccount2Instruction, inst.Owner)
}

type InitializeAccount3 struct {
	instruction
	Owner solana.PublicKey
}

func NewInitializeAccount3Instruction(owner, account, mint solana.PublicKey) *InitializeAccount3 {
	return &InitializeAccount3{Owner: owner, instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
		solana.NewAccountMeta(account, true, false),
		solana.NewAccountMeta(mint, false, false),
	}}}
}

func (inst *InitializeAccount3) tokenInstruction() *instruction { return &inst.instruction }
func (inst *InitializeAccount3) Data() ([]byte, error) {
	return publicKeyData(InitializeAccount3Instruction, inst.Owner)
}

type InitializeMultisig struct {
	instruction
	M uint8
}

func NewInitializeMultisigInstruction(m uint8, account, rent solana.PublicKey, signers []solana.PublicKey) *InitializeMultisig {
	accounts := make(solana.AccountMetaSlice, 0, 2+len(signers))
	accounts = append(accounts,
		solana.NewAccountMeta(account, true, false),
		solana.NewAccountMeta(rent, false, false),
	)
	for _, signer := range signers {
		accounts = append(accounts, solana.NewAccountMeta(signer, false, false))
	}
	return &InitializeMultisig{M: m, instruction: instruction{AccountMetaSlice: accounts}}
}

func (inst *InitializeMultisig) tokenInstruction() *instruction { return &inst.instruction }
func (inst *InitializeMultisig) Data() ([]byte, error) {
	return []byte{byte(InitializeMultisigInstruction), inst.M}, nil
}

type InitializeMultisig2 struct{ InitializeMultisig }

func NewInitializeMultisig2Instruction(m uint8, account solana.PublicKey, signers []solana.PublicKey) *InitializeMultisig2 {
	accounts := make(solana.AccountMetaSlice, 1, 1+len(signers))
	accounts[0] = solana.NewAccountMeta(account, true, false)
	for _, signer := range signers {
		accounts = append(accounts, solana.NewAccountMeta(signer, false, false))
	}
	return &InitializeMultisig2{InitializeMultisig: InitializeMultisig{M: m, instruction: instruction{AccountMetaSlice: accounts}}}
}

func (inst *InitializeMultisig2) tokenInstruction() *instruction { return &inst.instruction }
func (inst *InitializeMultisig2) Data() ([]byte, error) {
	return []byte{byte(InitializeMultisig2Instruction), inst.M}, nil
}

func publicKeyData(typ InstructionType, key solana.PublicKey) ([]byte, error) {
	data := make([]byte, 1, 33)
	data[0] = byte(typ)
	data = append(data, key[:]...)
	return data, nil
}
