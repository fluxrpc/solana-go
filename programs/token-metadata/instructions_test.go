package tokenmetadata

import (
	"bytes"
	"testing"

	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

func metadataKey(value byte) solana.PublicKey { var key solana.PublicKey; key[0] = value; return key }

func TestCreateMetadataV2RoundTrip(t *testing.T) {
	creators := []Creator{{Address: metadataKey(8), Verified: true, Share: 100}}
	args := CreateMetadataAccountArgsV2{Data: DataV2{
		Data:       Data{Name: "Asset", Symbol: "AST", URI: "https://example.test/a.json", SellerFeeBasisPoints: 500, Creators: &creators},
		Collection: &Collection{Verified: false, Key: metadataKey(9)},
		Uses:       &Uses{UseMethod: UseMethodSingle, Remaining: 1, Total: 1},
	}, IsMutable: true}
	inst := NewCreateMetadataAccountV2Instruction(args, metadataKey(1), metadataKey(2), metadataKey(3), metadataKey(4), metadataKey(5), solana.SystemProgramID, solana.SysVarRentPubkey)
	data, err := inst.Data()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeInstruction(inst.Accounts(), data)
	if err != nil {
		t.Fatal(err)
	}
	roundTrip, _ := decoded.CreateMetadataAccountV2.Data()
	if !bytes.Equal(roundTrip, data) {
		t.Fatalf("round trip = %x, want %x", roundTrip, data)
	}
	accounts := inst.Accounts()
	if len(accounts) != 7 || !accounts[0].IsWritable || !accounts[2].IsSigner || !accounts[3].IsWritable || !accounts[3].IsSigner {
		t.Fatalf("accounts = %+v", accounts)
	}
}

func TestDecodeMetadata(t *testing.T) {
	creators := []Creator{{Address: metadataKey(4), Verified: true, Share: 100}}
	data := Data{Name: "Asset", Symbol: "AST", URI: "uri", SellerFeeBasisPoints: 500, Creators: &creators}
	enc := binary.NewEncoder(make([]byte, 0, 256))
	enc.WriteUint8(uint8(KeyMetadataV1))
	enc.WritePublicKey(metadataKey(1))
	enc.WritePublicKey(metadataKey(2))
	data.encode(enc)
	enc.WriteBool(false)
	enc.WriteBool(true)
	enc.WriteOption(true)
	enc.WriteUint8(7)
	metadata, err := DecodeMetadata(enc.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if metadata.Data.Name != "Asset" || metadata.EditionNonce == nil || *metadata.EditionNonce != 7 {
		t.Fatalf("metadata = %+v", metadata)
	}
}

func FuzzDecodeInstruction(f *testing.F) {
	f.Add([]byte{byte(PuffMetadataInstruction)})
	f.Fuzz(func(t *testing.T, data []byte) { _, _ = DecodeInstruction(nil, data) })
}
