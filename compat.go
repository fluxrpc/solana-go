package solana_go

// Meta initializes a new AccountMeta for pubKey.
//
// Deprecated: use pubKey.Meta() instead.
func Meta(pubKey PublicKey) *AccountMeta {
	return pubKey.Meta()
}

// IsAnyOfEncodingType reports whether candidate is one of the allowed values.
//
// Deprecated: use candidate.IsAnyOf() instead.
func IsAnyOfEncodingType(candidate EncodingType, allowed ...EncodingType) bool {
	return candidate.IsAnyOf(allowed...)
}

// IsOnCurve reports whether b is a valid compressed ed25519 curve point.
//
// Deprecated: convert b to a PublicKey and use PublicKey.IsOnCurve() instead.
func IsOnCurve(b []byte) bool {
	return isOnCurve(b)
}

// CreateWithSeed derives the address base+seed owned by owner.
//
// Deprecated: use base.CreateWithSeed() instead.
func CreateWithSeed(base PublicKey, seed string, owner PublicKey) (PublicKey, error) {
	return base.CreateWithSeed(seed, owner)
}

// CreateProgramAddress derives a program address from seeds and programID.
//
// Deprecated: use programID.CreateProgramAddress() instead.
func CreateProgramAddress(seeds [][]byte, programID PublicKey) (PublicKey, error) {
	return programID.CreateProgramAddress(seeds)
}

// FindProgramAddress finds the highest valid bump seed for seeds and programID.
//
// Deprecated: use programID.FindProgramAddress() instead.
func FindProgramAddress(seeds [][]byte, programID PublicKey) (PublicKey, uint8, error) {
	return programID.FindProgramAddress(seeds)
}

// FindAssociatedTokenAddress derives the associated token account of wallet
// for mint under the SPL Token program.
//
// Deprecated: use wallet.FindAssociatedTokenAddress() instead.
func FindAssociatedTokenAddress(wallet, mint PublicKey) (PublicKey, uint8, error) {
	return wallet.FindAssociatedTokenAddressWithProgram(mint, TokenProgramID)
}

// FindAssociatedTokenAddressWithProgram derives the associated token account
// of wallet for mint under tokenProgram.
//
// Deprecated: use wallet.FindAssociatedTokenAddressWithProgram() instead.
func FindAssociatedTokenAddressWithProgram(wallet, mint, tokenProgram PublicKey) (PublicKey, uint8, error) {
	return wallet.FindAssociatedTokenAddressWithProgram(mint, tokenProgram)
}

// FindTokenMetadataAddress derives the Metaplex token metadata address for mint.
//
// Deprecated: use mint.FindTokenMetadataAddress() instead.
func FindTokenMetadataAddress(mint PublicKey) (PublicKey, uint8, error) {
	return mint.FindTokenMetadataAddress()
}
