// Package tokenmetadata implements Metaplex Token Metadata codecs.
package tokenmetadata

import solana "github.com/fluxrpc/solana-go"

var ProgramID = solana.TokenMetadataProgramID

// FindMasterEditionAddress derives the Metaplex master edition PDA for mint.
func FindMasterEditionAddress(mint solana.PublicKey) (solana.PublicKey, uint8, error) {
	return ProgramID.FindProgramAddress(
		[][]byte{[]byte("metadata"), ProgramID[:], mint[:], []byte("edition")},
	)
}
