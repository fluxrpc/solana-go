package tokenmetadata

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type InstructionType uint8

func (typ InstructionType) String() string {
	names := [...]string{
		"CreateMetadataAccount", "UpdateMetadataAccount", "DeprecatedCreateMasterEdition",
		"DeprecatedMintNewEditionFromMasterEditionViaPrintingToken", "UpdatePrimarySaleHappenedViaToken",
		"DeprecatedSetReservationList", "DeprecatedCreateReservationList", "SignMetadata",
		"DeprecatedMintPrintingTokensViaToken", "DeprecatedMintPrintingTokens", "CreateMasterEdition",
		"MintNewEditionFromMasterEditionViaToken", "ConvertMasterEditionV1ToV2",
		"MintNewEditionFromMasterEditionViaVaultProxy", "PuffMetadata", "UpdateMetadataAccountV2",
		"CreateMetadataAccountV2", "CreateMasterEditionV3", "VerifyCollection", "Utilize",
		"ApproveUseAuthority", "RevokeUseAuthority", "UnverifyCollection", "ApproveCollectionAuthority",
		"RevokeCollectionAuthority",
	}
	if int(typ) < len(names) {
		return names[typ]
	}
	return fmt.Sprintf("InstructionType(%d)", uint8(typ))
}

type Key uint8

const (
	KeyUninitialized Key = iota
	KeyEditionV1
	KeyMasterEditionV1
	KeyReservationListV1
	KeyMetadataV1
	KeyReservationListV2
	KeyMasterEditionV2
	KeyEditionMarker
	KeyUseAuthorityRecord
	KeyCollectionAuthorityRecord
)

type Creator struct {
	Address  solana.PublicKey
	Verified bool
	Share    uint8
}

type Data struct {
	Name                 string
	Symbol               string
	URI                  string
	SellerFeeBasisPoints uint16
	Creators             *[]Creator
}

type UseMethod uint8

const (
	UseMethodBurn UseMethod = iota
	UseMethodMultiple
	UseMethodSingle
)

type Uses struct {
	UseMethod UseMethod
	Remaining uint64
	Total     uint64
}

type Collection struct {
	Verified bool
	Key      solana.PublicKey
}

type DataV2 struct {
	Data
	Collection *Collection
	Uses       *Uses
}

type TokenStandard uint8

const (
	TokenStandardNonFungible TokenStandard = iota
	TokenStandardFungibleAsset
	TokenStandardFungible
	TokenStandardNonFungibleEdition
)

type CreateMetadataAccountArgs struct {
	Data      Data
	IsMutable bool
}
type CreateMetadataAccountArgsV2 struct {
	Data      DataV2
	IsMutable bool
}
type UpdateMetadataAccountArgs struct {
	Data                *Data
	UpdateAuthority     *solana.PublicKey
	PrimarySaleHappened *bool
}
type UpdateMetadataAccountArgsV2 struct {
	Data                *DataV2
	UpdateAuthority     *solana.PublicKey
	PrimarySaleHappened *bool
	IsMutable           *bool
}
type CreateMasterEditionArgs struct{ MaxSupply *uint64 }

func (data *Data) encode(enc *binary.Encoder) {
	enc.WriteBorshString(data.Name)
	enc.WriteBorshString(data.Symbol)
	enc.WriteBorshString(data.URI)
	enc.WriteUint16(data.SellerFeeBasisPoints)
	enc.WriteOption(data.Creators != nil)
	if data.Creators == nil {
		return
	}
	enc.WriteUint32(uint32(len(*data.Creators)))
	for _, creator := range *data.Creators {
		enc.WritePublicKey(creator.Address)
		enc.WriteBool(creator.Verified)
		enc.WriteUint8(creator.Share)
	}
}

func (data *Data) decode(dec *binary.Decoder) {
	data.Name = dec.ReadBorshString()
	data.Symbol = dec.ReadBorshString()
	data.URI = dec.ReadBorshString()
	data.SellerFeeBasisPoints = dec.ReadUint16()
	if !dec.ReadOption() {
		return
	}
	creators := make([]Creator, dec.ReadUint32())
	for i := range creators {
		creators[i] = Creator{Address: dec.ReadPublicKey(), Verified: dec.ReadBool(), Share: dec.ReadUint8()}
	}
	data.Creators = &creators
}

func (data *DataV2) encode(enc *binary.Encoder) {
	data.Data.encode(enc)
	enc.WriteOption(data.Collection != nil)
	if data.Collection != nil {
		enc.WriteBool(data.Collection.Verified)
		enc.WritePublicKey(data.Collection.Key)
	}
	enc.WriteOption(data.Uses != nil)
	if data.Uses != nil {
		enc.WriteUint8(uint8(data.Uses.UseMethod))
		enc.WriteUint64(data.Uses.Remaining)
		enc.WriteUint64(data.Uses.Total)
	}
}

func (data *DataV2) decode(dec *binary.Decoder) {
	data.Data.decode(dec)
	if dec.ReadOption() {
		data.Collection = &Collection{Verified: dec.ReadBool(), Key: dec.ReadPublicKey()}
	}
	if dec.ReadOption() {
		data.Uses = &Uses{UseMethod: UseMethod(dec.ReadUint8()), Remaining: dec.ReadUint64(), Total: dec.ReadUint64()}
	}
}

func writeOptionalPublicKey(enc *binary.Encoder, value *solana.PublicKey) {
	enc.WriteOption(value != nil)
	if value != nil {
		enc.WritePublicKey(*value)
	}
}

func writeOptionalBool(enc *binary.Encoder, value *bool) {
	enc.WriteOption(value != nil)
	if value != nil {
		enc.WriteBool(*value)
	}
}
