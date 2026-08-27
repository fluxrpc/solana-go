package addresslookuptable

import (
	"encoding/binary"
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

// CreateLookupTable creates an empty address lookup table.
type CreateLookupTable struct {
	RecentSlot uint64
	BumpSeed   uint8
	// [0] Lookup table: writable.
	// [1] Authority: signer.
	// [2] Payer: writable, signer.
	// [3] System Program.
	instruction
}

// NewCreateLookupTableInstruction derives the table address and creates its
// initialization instruction.
func NewCreateLookupTableInstruction(
	authority solana.PublicKey,
	payer solana.PublicKey,
	recentSlot uint64,
) (*CreateLookupTable, solana.PublicKey, error) {
	lookupTable, bumpSeed, err := DeriveLookupTableAddress(authority, recentSlot)
	if err != nil {
		return nil, solana.PublicKey{}, fmt.Errorf("derive lookup table address: %w", err)
	}
	return &CreateLookupTable{
		RecentSlot: recentSlot,
		BumpSeed:   bumpSeed,
		instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
			solana.NewAccountMeta(lookupTable, true, false),
			solana.NewAccountMeta(authority, false, true),
			solana.NewAccountMeta(payer, true, true),
			solana.NewAccountMeta(solana.SystemProgramID, false, false),
		}},
	}, lookupTable, nil
}

// DeriveLookupTableAddress derives a lookup table PDA from its authority and
// recent slot.
func DeriveLookupTableAddress(authority solana.PublicKey, recentSlot uint64) (solana.PublicKey, uint8, error) {
	var slot [8]byte
	binary.LittleEndian.PutUint64(slot[:], recentSlot)
	return ProgramID.FindProgramAddress([][]byte{authority[:], slot[:]})
}

func (inst *CreateLookupTable) Data() ([]byte, error) {
	enc := bin.NewEncoder(make([]byte, 0, 13))
	enc.WriteUint32(uint32(CreateLookupTableInstruction))
	enc.WriteUint64(inst.RecentSlot)
	enc.WriteUint8(inst.BumpSeed)
	return enc.Bytes(), enc.Err()
}
