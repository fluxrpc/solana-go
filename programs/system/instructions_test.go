package system

import (
	"encoding/binary"
	"errors"
	"testing"

	solana "github.com/fluxrpc/solana-go"
	fluxbin "github.com/fluxrpc/solana-go/binary"
)

func TestInstructionTypeString(t *testing.T) {
	if got := TransferInstruction.String(); got != "Transfer" {
		t.Fatalf("TransferInstruction.String() = %q", got)
	}
	if got := InstructionType(99).String(); got != "InstructionType(99)" {
		t.Fatalf("unknown InstructionType.String() = %q", got)
	}
	if got := CreateAccountAllowPrefundInstruction.String(); got != "CreateAccountAllowPrefund" {
		t.Fatalf("CreateAccountAllowPrefundInstruction.String() = %q", got)
	}
}

func TestDecodeInstructionErrors(t *testing.T) {
	for n := 0; n < 4; n++ {
		_, err := DecodeInstruction(nil, make([]byte, n))
		if !errors.Is(err, fluxbin.ErrUnexpectedEOF) {
			t.Fatalf("discriminator length %d: error = %v", n, err)
		}
	}

	unknown := make([]byte, 4)
	binary.LittleEndian.PutUint32(unknown, 99)
	if _, err := DecodeInstruction(nil, unknown); !errors.Is(err, ErrUnknownInstruction) {
		t.Fatalf("unknown instruction error = %v", err)
	}

}

func TestDecodeInstructionAllowsTrailingData(t *testing.T) {
	data, err := NewTransferInstruction(7, solana.PublicKey{1}, solana.PublicKey{2}).Data()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeInstruction(nil, append(data, 0xff))
	if err != nil {
		t.Fatalf("DecodeInstruction() error = %v", err)
	}
	if decoded.Transfer == nil || decoded.Transfer.Lamports != 7 {
		t.Fatalf("decoded = %#v", decoded)
	}
}

func TestDecodeInstructionAliasesAccounts(t *testing.T) {
	from := solana.NewAccountMeta(solana.PublicKey{1}, true, true)
	to := solana.NewAccountMeta(solana.PublicKey{2}, true, false)
	accounts := solana.AccountMetaSlice{from, to}
	data, err := NewTransferInstruction(42, from.PublicKey, to.PublicKey).Data()
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := DecodeInstruction(accounts, data)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Type != TransferInstruction || decoded.Transfer == nil {
		t.Fatalf("decoded = %#v", decoded)
	}
	if decoded.Transfer.AccountMetaSlice[0] != from || decoded.Transfer.AccountMetaSlice[1] != to {
		t.Fatal("decoded accounts do not alias the supplied account metadata")
	}
}

func FuzzDecodeInstruction(f *testing.F) {
	for typ := CreateAccountInstruction; typ <= CreateAccountAllowPrefundInstruction; typ++ {
		data := make([]byte, 4)
		binary.LittleEndian.PutUint32(data, uint32(typ))
		f.Add(data)
	}
	f.Add([]byte{})
	f.Add([]byte{0xff, 0xff, 0xff, 0xff})

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeInstruction(nil, data)
	})
}
