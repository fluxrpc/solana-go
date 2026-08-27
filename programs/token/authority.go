package token

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type Revoke struct{ instruction }

func NewRevokeInstruction(source, owner solana.PublicKey, signers []solana.PublicKey) *Revoke {
	accounts := solana.AccountMetaSlice{solana.NewAccountMeta(source, true, false)}
	return &Revoke{instruction{AccountMetaSlice: appendAuthority(accounts, owner, signers)}}
}
func (inst *Revoke) tokenInstruction() *instruction { return &inst.instruction }
func (inst *Revoke) Data() ([]byte, error)          { return []byte{byte(RevokeInstruction)}, nil }

type SetAuthority struct {
	instruction
	AuthorityType AuthorityType
	NewAuthority  *solana.PublicKey
}

func NewSetAuthorityInstruction(authorityType AuthorityType, newAuthority, subject, authority solana.PublicKey, signers []solana.PublicKey) *SetAuthority {
	inst := &SetAuthority{AuthorityType: authorityType}
	if !newAuthority.IsZero() {
		inst.NewAuthority = &newAuthority
	}
	accounts := solana.AccountMetaSlice{solana.NewAccountMeta(subject, true, false)}
	inst.AccountMetaSlice = appendAuthority(accounts, authority, signers)
	return inst
}
func (inst *SetAuthority) tokenInstruction() *instruction { return &inst.instruction }
func (inst *SetAuthority) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 35))
	enc.WriteUint8(uint8(SetAuthorityInstruction))
	enc.WriteUint8(uint8(inst.AuthorityType))
	enc.WriteOption(inst.NewAuthority != nil)
	if inst.NewAuthority != nil {
		enc.WritePublicKey(*inst.NewAuthority)
	}
	return enc.Bytes(), enc.Err()
}

type CloseAccount struct{ instruction }

func NewCloseAccountInstruction(account, destination, owner solana.PublicKey, signers []solana.PublicKey) *CloseAccount {
	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(account, true, false),
		solana.NewAccountMeta(destination, true, false),
	}
	return &CloseAccount{instruction{AccountMetaSlice: appendAuthority(accounts, owner, signers)}}
}
func (inst *CloseAccount) tokenInstruction() *instruction { return &inst.instruction }
func (inst *CloseAccount) Data() ([]byte, error)          { return []byte{byte(CloseAccountInstruction)}, nil }

type FreezeAccount struct{ instruction }

func NewFreezeAccountInstruction(account, mint, authority solana.PublicKey, signers []solana.PublicKey) *FreezeAccount {
	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(account, true, false),
		solana.NewAccountMeta(mint, false, false),
	}
	return &FreezeAccount{instruction{AccountMetaSlice: appendAuthority(accounts, authority, signers)}}
}
func (inst *FreezeAccount) tokenInstruction() *instruction { return &inst.instruction }
func (inst *FreezeAccount) Data() ([]byte, error)          { return []byte{byte(FreezeAccountInstruction)}, nil }

type ThawAccount struct{ instruction }

func NewThawAccountInstruction(account, mint, authority solana.PublicKey, signers []solana.PublicKey) *ThawAccount {
	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(account, true, false),
		solana.NewAccountMeta(mint, false, false),
	}
	return &ThawAccount{instruction{AccountMetaSlice: appendAuthority(accounts, authority, signers)}}
}
func (inst *ThawAccount) tokenInstruction() *instruction { return &inst.instruction }
func (inst *ThawAccount) Data() ([]byte, error)          { return []byte{byte(ThawAccountInstruction)}, nil }

type SyncNative struct{ instruction }

func NewSyncNativeInstruction(account solana.PublicKey) *SyncNative {
	return &SyncNative{instruction{AccountMetaSlice: solana.AccountMetaSlice{solana.NewAccountMeta(account, true, false)}}}
}
func (inst *SyncNative) tokenInstruction() *instruction { return &inst.instruction }
func (inst *SyncNative) Data() ([]byte, error)          { return []byte{byte(SyncNativeInstruction)}, nil }
