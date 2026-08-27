package tokenmetadata

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type Metadata struct {
	Key                 Key
	UpdateAuthority     solana.PublicKey
	Mint                solana.PublicKey
	Data                Data
	PrimarySaleHappened bool
	IsMutable           bool
	EditionNonce        *uint8
	TokenStandard       *TokenStandard
	Collection          *Collection
	Uses                *Uses
}

func DecodeMetadata(data []byte) (Metadata, error) {
	dec := bin.NewDecoder(data)
	metadata := Metadata{Key: Key(dec.ReadUint8()), UpdateAuthority: dec.ReadPublicKey(), Mint: dec.ReadPublicKey()}
	metadata.Data.decode(dec)
	metadata.PrimarySaleHappened = dec.ReadBool()
	metadata.IsMutable = dec.ReadBool()
	if dec.Remaining() > 0 && dec.ReadOption() {
		value := dec.ReadUint8()
		metadata.EditionNonce = &value
	}
	if dec.Remaining() > 0 && dec.ReadOption() {
		value := TokenStandard(dec.ReadUint8())
		metadata.TokenStandard = &value
	}
	if dec.Remaining() > 0 && dec.ReadOption() {
		metadata.Collection = &Collection{Verified: dec.ReadBool(), Key: dec.ReadPublicKey()}
	}
	if dec.Remaining() > 0 && dec.ReadOption() {
		metadata.Uses = &Uses{UseMethod: UseMethod(dec.ReadUint8()), Remaining: dec.ReadUint64(), Total: dec.ReadUint64()}
	}
	if err := dec.Err(); err != nil {
		return Metadata{}, fmt.Errorf("decode token metadata: %w", err)
	}
	return metadata, nil
}

type MasterEditionV2 struct {
	Key       Key
	Supply    uint64
	MaxSupply *uint64
}

func DecodeMasterEditionV2(data []byte) (MasterEditionV2, error) {
	dec := bin.NewDecoder(data)
	edition := MasterEditionV2{Key: Key(dec.ReadUint8()), Supply: dec.ReadUint64()}
	if dec.ReadOption() {
		value := dec.ReadUint64()
		edition.MaxSupply = &value
	}
	if err := dec.Err(); err != nil {
		return MasterEditionV2{}, fmt.Errorf("decode master edition v2: %w", err)
	}
	return edition, nil
}

type Edition struct {
	Key     Key
	Parent  solana.PublicKey
	Edition uint64
}

func DecodeEdition(data []byte) (Edition, error) {
	dec := bin.NewDecoder(data)
	edition := Edition{Key: Key(dec.ReadUint8()), Parent: dec.ReadPublicKey(), Edition: dec.ReadUint64()}
	if err := dec.Err(); err != nil {
		return Edition{}, fmt.Errorf("decode edition: %w", err)
	}
	return edition, nil
}

type EditionMarker struct {
	Key    Key
	Ledger [31]byte
}

func DecodeEditionMarker(data []byte) (EditionMarker, error) {
	dec := bin.NewDecoder(data)
	marker := EditionMarker{Key: Key(dec.ReadUint8())}
	copy(marker.Ledger[:], dec.ReadBytes(len(marker.Ledger)))
	if err := dec.Err(); err != nil {
		return EditionMarker{}, fmt.Errorf("decode edition marker: %w", err)
	}
	return marker, nil
}
