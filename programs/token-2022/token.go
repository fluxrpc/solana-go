package token2022

import (
	solana "github.com/fluxrpc/solana-go"
	token "github.com/fluxrpc/solana-go/programs/token"
)

func useProgram[T interface{ SetProgramID(solana.PublicKey) }](inst T) T {
	inst.SetProgramID(ProgramID)
	return inst
}

func NewInitializeMintInstruction(decimals uint8, mintAuthority, freezeAuthority, mint, rent solana.PublicKey) *InitializeMint {
	return useProgram(token.NewInitializeMintInstruction(decimals, mintAuthority, freezeAuthority, mint, rent))
}
func NewInitializeAccountInstruction(account, mint, owner, rent solana.PublicKey) *InitializeAccount {
	return useProgram(token.NewInitializeAccountInstruction(account, mint, owner, rent))
}
func NewInitializeMultisigInstruction(m uint8, account, rent solana.PublicKey, signers []solana.PublicKey) *InitializeMultisig {
	return useProgram(token.NewInitializeMultisigInstruction(m, account, rent, signers))
}
func NewTransferInstruction(amount uint64, source, destination, owner solana.PublicKey, signers []solana.PublicKey) *Transfer {
	return useProgram(token.NewTransferInstruction(amount, source, destination, owner, signers))
}
func NewApproveInstruction(amount uint64, source, delegate, owner solana.PublicKey, signers []solana.PublicKey) *Approve {
	return useProgram(token.NewApproveInstruction(amount, source, delegate, owner, signers))
}
func NewRevokeInstruction(source, owner solana.PublicKey, signers []solana.PublicKey) *Revoke {
	return useProgram(token.NewRevokeInstruction(source, owner, signers))
}
func NewSetAuthorityInstruction(authorityType AuthorityType, newAuthority, subject, authority solana.PublicKey, signers []solana.PublicKey) *SetAuthority {
	return useProgram(token.NewSetAuthorityInstruction(authorityType, newAuthority, subject, authority, signers))
}
func NewMintToInstruction(amount uint64, mint, destination, authority solana.PublicKey, signers []solana.PublicKey) *MintTo {
	return useProgram(token.NewMintToInstruction(amount, mint, destination, authority, signers))
}
func NewBurnInstruction(amount uint64, source, mint, owner solana.PublicKey, signers []solana.PublicKey) *Burn {
	return useProgram(token.NewBurnInstruction(amount, source, mint, owner, signers))
}
func NewCloseAccountInstruction(account, destination, owner solana.PublicKey, signers []solana.PublicKey) *CloseAccount {
	return useProgram(token.NewCloseAccountInstruction(account, destination, owner, signers))
}
func NewFreezeAccountInstruction(account, mint, authority solana.PublicKey, signers []solana.PublicKey) *FreezeAccount {
	return useProgram(token.NewFreezeAccountInstruction(account, mint, authority, signers))
}
func NewThawAccountInstruction(account, mint, authority solana.PublicKey, signers []solana.PublicKey) *ThawAccount {
	return useProgram(token.NewThawAccountInstruction(account, mint, authority, signers))
}
func NewTransferCheckedInstruction(amount uint64, decimals uint8, source, mint, destination, owner solana.PublicKey, signers []solana.PublicKey) *TransferChecked {
	return useProgram(token.NewTransferCheckedInstruction(amount, decimals, source, mint, destination, owner, signers))
}
func NewApproveCheckedInstruction(amount uint64, decimals uint8, source, mint, delegate, owner solana.PublicKey, signers []solana.PublicKey) *ApproveChecked {
	return useProgram(token.NewApproveCheckedInstruction(amount, decimals, source, mint, delegate, owner, signers))
}
func NewMintToCheckedInstruction(amount uint64, decimals uint8, mint, destination, authority solana.PublicKey, signers []solana.PublicKey) *MintToChecked {
	return useProgram(token.NewMintToCheckedInstruction(amount, decimals, mint, destination, authority, signers))
}
func NewBurnCheckedInstruction(amount uint64, decimals uint8, source, mint, owner solana.PublicKey, signers []solana.PublicKey) *BurnChecked {
	return useProgram(token.NewBurnCheckedInstruction(amount, decimals, source, mint, owner, signers))
}
func NewInitializeAccount2Instruction(owner, account, mint, rent solana.PublicKey) *InitializeAccount2 {
	return useProgram(token.NewInitializeAccount2Instruction(owner, account, mint, rent))
}
func NewSyncNativeInstruction(account solana.PublicKey) *SyncNative {
	return useProgram(token.NewSyncNativeInstruction(account))
}
func NewInitializeAccount3Instruction(owner, account, mint solana.PublicKey) *InitializeAccount3 {
	return useProgram(token.NewInitializeAccount3Instruction(owner, account, mint))
}
func NewInitializeMultisig2Instruction(m uint8, account solana.PublicKey, signers []solana.PublicKey) *InitializeMultisig2 {
	return useProgram(token.NewInitializeMultisig2Instruction(m, account, signers))
}
func NewInitializeMint2Instruction(decimals uint8, mintAuthority, freezeAuthority, mint solana.PublicKey) *InitializeMint2 {
	return useProgram(token.NewInitializeMint2Instruction(decimals, mintAuthority, freezeAuthority, mint))
}
