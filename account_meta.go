package solana_go

// AccountMeta describes how an account is used by an instruction.
type AccountMeta struct {
	PublicKey  PublicKey `json:"publicKey"`
	IsWritable bool      `json:"isWritable"`
	IsSigner   bool      `json:"isSigner"`
}

// Meta initializes a new AccountMeta for this public key.
func (p PublicKey) Meta() *AccountMeta {
	return &AccountMeta{PublicKey: p}
}

// WRITE sets IsWritable to true.
func (meta *AccountMeta) WRITE() *AccountMeta {
	meta.IsWritable = true
	return meta
}

// SIGNER sets IsSigner to true.
func (meta *AccountMeta) SIGNER() *AccountMeta {
	meta.IsSigner = true
	return meta
}

// NewAccountMeta initializes an AccountMeta with the given roles.
func NewAccountMeta(pubKey PublicKey, writable bool, signer bool) *AccountMeta {
	return &AccountMeta{
		PublicKey:  pubKey,
		IsWritable: writable,
		IsSigner:   signer,
	}
}

// AccountMetaSlice is a list of AccountMeta with convenience helpers.
type AccountMetaSlice []*AccountMeta

// Append adds an account to the slice.
func (slice *AccountMetaSlice) Append(account *AccountMeta) {
	*slice = append(*slice, account)
}

// Get returns the AccountMeta at the desired index, or nil if the index is
// out of range.
func (slice AccountMetaSlice) Get(index int) *AccountMeta {
	if index >= 0 && index < len(slice) {
		return slice[index]
	}
	return nil
}

// GetSigners returns the accounts that are signers.
func (slice AccountMetaSlice) GetSigners() []*AccountMeta {
	signers := make([]*AccountMeta, 0, len(slice))
	for _, ac := range slice {
		if ac.IsSigner {
			signers = append(signers, ac)
		}
	}
	return signers
}

// GetKeys returns the public keys of all AccountMeta.
func (slice AccountMetaSlice) GetKeys() []PublicKey {
	keys := make([]PublicKey, len(slice))
	for i, ac := range slice {
		keys[i] = ac.PublicKey
	}
	return keys
}

// Len returns the number of accounts in the slice.
func (slice AccountMetaSlice) Len() int {
	return len(slice)
}
