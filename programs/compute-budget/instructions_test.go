package computebudget

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"

	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

func computeData(t *testing.T, instruction solana.Instruction) []byte {
	t.Helper()
	if instruction.ProgramID() != ProgramID {
		t.Fatalf("ProgramID = %s, want %s", instruction.ProgramID(), ProgramID)
	}
	if len(instruction.Accounts()) != 0 {
		t.Fatalf("accounts = %v, want none", instruction.Accounts())
	}
	data, err := instruction.Data()
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestComputeBudgetInstructions(t *testing.T) {
	tests := []struct {
		name        string
		instruction solana.Instruction
		want        []byte
		check       func(*testing.T, DecodedInstruction)
	}{
		{"Unused", NewUnusedInstruction(), []byte{0}, func(t *testing.T, got DecodedInstruction) {
			if got.Type != UnusedInstruction || got.Unused == nil {
				t.Fatalf("decoded = %#v", got)
			}
		}},
		{"RequestUnitsDeprecated", NewRequestUnitsDeprecatedInstruction(0x01020304, 0x05060708), []byte{0, 4, 3, 2, 1, 8, 7, 6, 5}, func(t *testing.T, got DecodedInstruction) {
			if got.RequestUnitsDeprecated == nil || got.RequestUnitsDeprecated.Units != 0x01020304 || got.RequestUnitsDeprecated.AdditionalFee != 0x05060708 {
				t.Fatalf("decoded = %#v", got)
			}
		}},
		{"RequestHeapFrame", NewRequestHeapFrameInstruction(0x01020304), []byte{1, 4, 3, 2, 1}, func(t *testing.T, got DecodedInstruction) {
			if got.RequestHeapFrame == nil || got.RequestHeapFrame.HeapSize != 0x01020304 {
				t.Fatalf("decoded = %#v", got)
			}
		}},
		{"SetComputeUnitLimit", NewSetComputeUnitLimitInstruction(0x05060708), []byte{2, 8, 7, 6, 5}, func(t *testing.T, got DecodedInstruction) {
			if got.SetComputeUnitLimit == nil || got.SetComputeUnitLimit.Units != 0x05060708 {
				t.Fatalf("decoded = %#v", got)
			}
		}},
		{"SetComputeUnitPrice", NewSetComputeUnitPriceInstruction(0x0102030405060708), []byte{3, 8, 7, 6, 5, 4, 3, 2, 1}, func(t *testing.T, got DecodedInstruction) {
			if got.SetComputeUnitPrice == nil || got.SetComputeUnitPrice.MicroLamports != 0x0102030405060708 {
				t.Fatalf("decoded = %#v", got)
			}
		}},
		{"SetLoadedAccountsDataSizeLimit", NewSetLoadedAccountsDataSizeLimitInstruction(0x11121314), []byte{4, 0x14, 0x13, 0x12, 0x11}, func(t *testing.T, got DecodedInstruction) {
			if got.SetLoadedAccountsDataSizeLimit == nil || got.SetLoadedAccountsDataSizeLimit.Bytes != 0x11121314 {
				t.Fatalf("decoded = %#v", got)
			}
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := computeData(t, test.instruction)
			if !bytes.Equal(data, test.want) {
				t.Fatalf("Data = %x, want %x", data, test.want)
			}
			decoded, err := DecodeInstruction(solana.AccountMetaSlice{solana.NewAccountMeta(solana.PublicKey{9}, true, true)}, data)
			if err != nil {
				t.Fatal(err)
			}
			test.check(t, decoded)

			// The current unit variant is complete at one byte; legacy tag-zero
			// prefixes cannot be distinguished from its accepted trailing data.
			if test.name != "Unused" && test.name != "RequestUnitsDeprecated" {
				for length := 0; length < len(data); length++ {
					if _, err := DecodeInstruction(nil, data[:length]); err == nil {
						t.Errorf("accepted %d of %d bytes", length, len(data))
					}
				}
			}
		})
	}
}

func TestComputeBudgetDecodeErrorsAndTrailingData(t *testing.T) {
	if _, err := DecodeInstruction(nil, nil); !errors.Is(err, bin.ErrUnexpectedEOF) {
		t.Fatalf("empty error = %v", err)
	}
	if _, err := DecodeInstruction(nil, []byte{99}); !errors.Is(err, ErrUnknownInstruction) {
		t.Fatalf("unknown error = %v", err)
	}

	data := append(computeData(t, NewSetComputeUnitLimitInstruction(7)), 0xaa, 0xbb)
	decoded, err := DecodeInstruction(nil, data)
	if err != nil || decoded.SetComputeUnitLimit == nil || decoded.SetComputeUnitLimit.Units != 7 {
		t.Fatalf("trailing decode = %#v, %v", decoded, err)
	}
}

func TestComputeBudgetInstructionTypeString(t *testing.T) {
	if got := SetComputeUnitPriceInstruction.String(); got != "SetComputeUnitPrice" {
		t.Fatalf("String = %q", got)
	}
	if got := InstructionType(99).String(); got != "InstructionType(99)" {
		t.Fatalf("unknown String = %q", got)
	}
}

func FuzzDecodeInstruction(f *testing.F) {
	for typ := UnusedInstruction; typ <= SetLoadedAccountsDataSizeLimitInstruction; typ++ {
		data := []byte{byte(typ)}
		if typ == SetComputeUnitPriceInstruction {
			data = append(data, make([]byte, 8)...)
		} else if typ != UnusedInstruction {
			data = append(data, make([]byte, 4)...)
		}
		f.Add(data)
	}
	f.Add([]byte{})
	f.Add([]byte{0xff})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeInstruction(nil, data)
	})
}

func TestComputeBudgetGoldenEndian(t *testing.T) {
	data := computeData(t, NewSetComputeUnitPriceInstruction(9))
	if binary.LittleEndian.Uint64(data[1:]) != 9 {
		t.Fatalf("price data = %x", data)
	}
}
