package token2022

import (
	"bytes"
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

func token2022Key(value byte) solana.PublicKey { var key solana.PublicKey; key[0] = value; return key }

func TestLegacyInstructionUsesToken2022Program(t *testing.T) {
	inst := NewTransferInstruction(1, token2022Key(1), token2022Key(2), token2022Key(3), nil)
	if inst.ProgramID() != ProgramID {
		t.Fatalf("program ID = %s", inst.ProgramID())
	}
	decoded, err := DecodeInstruction(inst.Accounts(), []byte{byte(TransferInstruction), 1, 0, 0, 0, 0, 0, 0, 0})
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Base == nil || decoded.Base.Transfer.ProgramID() != ProgramID {
		t.Fatalf("decoded = %+v", decoded)
	}
}

func TestTransferFeeAndMetadataFixtures(t *testing.T) {
	mint, config, withdraw := token2022Key(1), token2022Key(2), token2022Key(3)
	inst := NewInitializeTransferFeeConfig(mint, &config, &withdraw, 250, 1_000)
	data, err := inst.Data()
	if err != nil {
		t.Fatal(err)
	}
	if data[0] != byte(TransferFeeExtensionInstruction) || data[1] != 0 || len(data) != 78 {
		t.Fatalf("data = %x", data)
	}
	decoded, err := DecodeInstruction(inst.Accounts(), data)
	if err != nil {
		t.Fatal(err)
	}
	roundTrip, _ := decoded.InitializeTransferFeeConfig.Data()
	if !bytes.Equal(roundTrip, data) {
		t.Fatalf("round trip = %x, want %x", roundTrip, data)
	}
	none, _ := NewInitializeTransferFeeConfig(mint, nil, nil, 0, 0).Data()
	if len(none) != 78 || !bytes.Equal(none[2:68], make([]byte, 66)) {
		t.Fatalf("none options = %x", none)
	}

	metadata := NewInitializeMetadata(mint, token2022Key(4), token2022Key(5), &config, "name", "SYM", "uri")
	data, _ = metadata.Data()
	decoded, err = DecodeInstruction(metadata.Accounts(), data)
	if err != nil || decoded.InitializeMetadata == nil || decoded.InitializeMetadata.Name != "name" {
		t.Fatalf("decoded = %+v, %v", decoded, err)
	}
}

func TestDecodeExtensions(t *testing.T) {
	data := make([]byte, 166)
	data[165] = 2
	data = append(data, byte(ExtensionMemoTransfer), 0, 1, 0, 1)
	account, err := DecodeAccount(data)
	if err != nil {
		t.Fatal(err)
	}
	if got := account.Extensions.Find(ExtensionMemoTransfer); !bytes.Equal(got, []byte{1}) {
		t.Fatalf("extension = %x", got)
	}
}

func TestConfidentialInstructionFixtures(t *testing.T) {
	account := solana.AccountMeta{PublicKey: token2022Key(1), IsWritable: true, IsSigner: true}
	service := ConfidentialTransferService{}
	tests := []struct {
		name   string
		family InstructionType
		sizes  []int
	}{
		{name: "transfer", family: ConfidentialTransferExtensionInstruction, sizes: []int{65, 33, 45, 0, 1, 9, 47, 167, 44, 0, 0, 0, 0, 169, 0}},
		{name: "fee", family: ConfidentialTransferFeeExtensionInstruction, sizes: []int{64, 37, 38, 0, 0, 0}},
		{name: "mint-burn", family: ConfidentialMintBurnExtensionInstruction, sizes: []int{68, 33, 36, 167, 167, 0}},
	}
	for _, test := range tests {
		for sub, size := range test.sizes {
			payload := make([]byte, size)
			var inst solana.Instruction
			switch test.family {
			case ConfidentialTransferExtensionInstruction:
				inst = service.RawConfidentialTransferInstruction(uint8(sub), payload, account)
			case ConfidentialTransferFeeExtensionInstruction:
				inst = service.RawConfidentialTransferFeeInstruction(uint8(sub), payload, account)
			case ConfidentialMintBurnExtensionInstruction:
				inst = service.RawConfidentialMintBurnInstruction(uint8(sub), payload, account)
			}
			data, err := inst.Data()
			if err != nil {
				t.Fatalf("%s %d: %v", test.name, sub, err)
			}
			decoded, err := DecodeInstruction(inst.Accounts(), data)
			if err != nil {
				t.Fatalf("%s %d: %v", test.name, sub, err)
			}
			var roundTrip []byte
			switch test.family {
			case ConfidentialTransferExtensionInstruction:
				roundTrip, err = decoded.ConfidentialTransfer.Data()
			case ConfidentialTransferFeeExtensionInstruction:
				roundTrip, err = decoded.ConfidentialTransferFee.Data()
			case ConfidentialMintBurnExtensionInstruction:
				roundTrip, err = decoded.ConfidentialMintBurn.Data()
			}
			if err != nil || !bytes.Equal(roundTrip, data) {
				t.Fatalf("%s %d: round trip = %x, %v", test.name, sub, roundTrip, err)
			}
			if len(inst.Accounts()) != 1 || !inst.Accounts()[0].IsWritable || !inst.Accounts()[0].IsSigner {
				t.Fatalf("%s %d: accounts = %+v", test.name, sub, inst.Accounts())
			}
		}
	}
}

func FuzzDecodeInstruction(f *testing.F) {
	f.Add([]byte{byte(TransferFeeExtensionInstruction), 4})
	f.Fuzz(func(t *testing.T, data []byte) { _, _ = DecodeInstruction(nil, data) })
}
