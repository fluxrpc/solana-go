package loaderv3

import (
	"bytes"
	"errors"
	"testing"

	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

func testKey(fill byte) (key solana.PublicKey) {
	for i := range key {
		key[i] = fill
	}
	return key
}

func assertAccounts(t *testing.T, got solana.AccountMetaSlice, want ...solana.AccountMeta) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("account count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] == nil || *got[i] != want[i] {
			t.Errorf("account %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestInstructions(t *testing.T) {
	buffer, authority := testKey(1), testKey(2)
	payer, programData, program := testKey(3), testKey(4), testKey(5)
	spill, newAuthority := testKey(6), testKey(7)

	initialize := NewInitializeBufferInstruction(buffer, authority)
	write := NewWriteInstruction(42, []byte{1, 2, 3}, buffer, authority)
	deploy := NewDeployWithMaxDataLenInstruction(1_000_000, true, payer, programData, program, buffer, authority)
	upgrade := NewUpgradeInstruction(true, programData, program, buffer, spill, authority)
	setAuthority := NewSetAuthorityInstruction(buffer, authority, newAuthority)
	close := NewCloseInstruction(false, buffer, payer, authority)
	extend := NewExtendProgramInstruction(10_240, programData, program)
	checked := NewSetAuthorityCheckedInstruction(buffer, authority, newAuthority)

	assertAccounts(t, initialize.Accounts(),
		solana.AccountMeta{PublicKey: buffer, IsWritable: true},
		solana.AccountMeta{PublicKey: authority},
	)
	assertAccounts(t, write.Accounts(),
		solana.AccountMeta{PublicKey: buffer, IsWritable: true},
		solana.AccountMeta{PublicKey: authority, IsSigner: true},
	)
	assertAccounts(t, deploy.Accounts(),
		solana.AccountMeta{PublicKey: payer, IsWritable: true, IsSigner: true},
		solana.AccountMeta{PublicKey: programData, IsWritable: true},
		solana.AccountMeta{PublicKey: program, IsWritable: true},
		solana.AccountMeta{PublicKey: buffer, IsWritable: true},
		solana.AccountMeta{PublicKey: solana.SysVarRentPubkey},
		solana.AccountMeta{PublicKey: solana.SysVarClockPubkey},
		solana.AccountMeta{PublicKey: solana.SystemProgramID},
		solana.AccountMeta{PublicKey: authority, IsSigner: true},
	)
	assertAccounts(t, upgrade.Accounts(),
		solana.AccountMeta{PublicKey: programData, IsWritable: true},
		solana.AccountMeta{PublicKey: program, IsWritable: true},
		solana.AccountMeta{PublicKey: buffer, IsWritable: true},
		solana.AccountMeta{PublicKey: spill, IsWritable: true},
		solana.AccountMeta{PublicKey: solana.SysVarRentPubkey},
		solana.AccountMeta{PublicKey: solana.SysVarClockPubkey},
		solana.AccountMeta{PublicKey: authority, IsSigner: true},
	)
	assertAccounts(t, setAuthority.Accounts(),
		solana.AccountMeta{PublicKey: buffer, IsWritable: true},
		solana.AccountMeta{PublicKey: authority, IsSigner: true},
		solana.AccountMeta{PublicKey: newAuthority},
	)
	assertAccounts(t, close.Accounts(),
		solana.AccountMeta{PublicKey: buffer, IsWritable: true},
		solana.AccountMeta{PublicKey: payer, IsWritable: true},
		solana.AccountMeta{PublicKey: authority, IsSigner: true},
	)
	assertAccounts(t, extend.Accounts(),
		solana.AccountMeta{PublicKey: programData, IsWritable: true},
		solana.AccountMeta{PublicKey: program, IsWritable: true},
	)
	assertAccounts(t, checked.Accounts(),
		solana.AccountMeta{PublicKey: buffer, IsWritable: true},
		solana.AccountMeta{PublicKey: authority, IsSigner: true},
		solana.AccountMeta{PublicKey: newAuthority, IsSigner: true},
	)

	tests := []struct {
		name       string
		typ        InstructionType
		inst       solana.Instruction
		want       []byte
		legacyTail bool
	}{
		{"InitializeBuffer", InitializeBufferInstruction, initialize, []byte{0, 0, 0, 0}, false},
		{"Write", WriteInstruction, write, []byte{1, 0, 0, 0, 42, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3}, false},
		{"DeployWithMaxDataLen", DeployWithMaxDataLenInstruction, deploy, []byte{2, 0, 0, 0, 0x40, 0x42, 0x0f, 0, 0, 0, 0, 0, 1}, true},
		{"Upgrade", UpgradeInstruction, upgrade, []byte{3, 0, 0, 0, 1}, true},
		{"SetAuthority", SetAuthorityInstruction, setAuthority, []byte{4, 0, 0, 0}, false},
		{"Close", CloseInstruction, close, []byte{5, 0, 0, 0, 0}, true},
		{"ExtendProgram", ExtendProgramInstruction, extend, []byte{6, 0, 0, 0, 0, 0x28, 0, 0}, false},
		{"SetAuthorityChecked", SetAuthorityCheckedInstruction, checked, []byte{7, 0, 0, 0}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := tc.inst.Data()
			if err != nil || !bytes.Equal(data, tc.want) {
				t.Fatalf("Data = %x, %v; want %x", data, err, tc.want)
			}
			decoded, err := DecodeInstruction(tc.inst.Accounts(), data)
			if err != nil || decoded.Type != tc.typ {
				t.Fatalf("decoded = %+v, %v", decoded, err)
			}
			var roundTrip solana.Instruction
			switch tc.typ {
			case InitializeBufferInstruction:
				roundTrip = decoded.InitializeBuffer
			case WriteInstruction:
				if decoded.Write == nil || decoded.Write.Offset != 42 || !bytes.Equal(decoded.Write.Bytes, []byte{1, 2, 3}) {
					t.Fatalf("Write = %+v", decoded.Write)
				}
				roundTrip = decoded.Write
			case DeployWithMaxDataLenInstruction:
				if decoded.DeployWithMaxDataLen == nil || decoded.DeployWithMaxDataLen.MaxDataLen != 1_000_000 || !decoded.DeployWithMaxDataLen.CloseBuffer {
					t.Fatalf("DeployWithMaxDataLen = %+v", decoded.DeployWithMaxDataLen)
				}
				roundTrip = decoded.DeployWithMaxDataLen
			case UpgradeInstruction:
				if decoded.Upgrade == nil || !decoded.Upgrade.CloseBuffer {
					t.Fatalf("Upgrade = %+v", decoded.Upgrade)
				}
				roundTrip = decoded.Upgrade
			case SetAuthorityInstruction:
				roundTrip = decoded.SetAuthority
			case CloseInstruction:
				if decoded.Close == nil || decoded.Close.Tombstone {
					t.Fatalf("Close = %+v", decoded.Close)
				}
				roundTrip = decoded.Close
			case ExtendProgramInstruction:
				if decoded.ExtendProgram == nil || decoded.ExtendProgram.AdditionalBytes != 10_240 {
					t.Fatalf("ExtendProgram = %+v", decoded.ExtendProgram)
				}
				roundTrip = decoded.ExtendProgram
			case SetAuthorityCheckedInstruction:
				roundTrip = decoded.SetAuthorityChecked
			}
			got, err := roundTrip.Data()
			if err != nil || !bytes.Equal(got, tc.want) {
				t.Fatalf("round trip = %x, %v", got, err)
			}
			for n := 0; n < len(data); n++ {
				if tc.legacyTail && n == len(data)-1 {
					continue
				}
				if _, err := DecodeInstruction(tc.inst.Accounts(), data[:n]); err == nil {
					t.Errorf("accepted truncation %d/%d", n, len(data))
				}
			}
		})
	}
}

func TestOptionalAccountLayouts(t *testing.T) {
	a, b, c, d := testKey(11), testKey(12), testKey(13), testKey(14)
	assertAccounts(t, NewInitializeImmutableBufferInstruction(a).Accounts(),
		solana.AccountMeta{PublicKey: a, IsWritable: true},
	)
	assertAccounts(t, NewRemoveAuthorityInstruction(a, b).Accounts(),
		solana.AccountMeta{PublicKey: a, IsWritable: true},
		solana.AccountMeta{PublicKey: b, IsSigner: true},
	)
	assertAccounts(t, NewCloseUninitializedInstruction(false, a, b).Accounts(),
		solana.AccountMeta{PublicKey: a, IsWritable: true},
		solana.AccountMeta{PublicKey: b, IsWritable: true},
	)
	assertAccounts(t, NewCloseProgramInstruction(true, a, b, c, d).Accounts(),
		solana.AccountMeta{PublicKey: a, IsWritable: true},
		solana.AccountMeta{PublicKey: b, IsWritable: true},
		solana.AccountMeta{PublicKey: c, IsSigner: true},
		solana.AccountMeta{PublicKey: d, IsWritable: true},
	)
	assertAccounts(t, NewExtendProgramWithPayerInstruction(10_240, a, b, c).Accounts(),
		solana.AccountMeta{PublicKey: a, IsWritable: true},
		solana.AccountMeta{PublicKey: b, IsWritable: true},
		solana.AccountMeta{PublicKey: solana.SystemProgramID},
		solana.AccountMeta{PublicKey: c, IsWritable: true, IsSigner: true},
	)
}

func TestMalformedWriteLength(t *testing.T) {
	data := []byte{1, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	if _, err := DecodeInstruction(nil, data); !errors.Is(err, binary.ErrUnexpectedEOF) {
		t.Fatalf("error = %v", err)
	}
}

func TestOptionalTrailingBoolCompatibility(t *testing.T) {
	deploy, err := DecodeInstruction(nil, []byte{2, 0, 0, 0, 42, 0, 0, 0, 0, 0, 0, 0})
	if err != nil || deploy.DeployWithMaxDataLen == nil || !deploy.DeployWithMaxDataLen.CloseBuffer {
		t.Fatalf("legacy DeployWithMaxDataLen = %+v, %v", deploy, err)
	}
	upgrade, err := DecodeInstruction(nil, []byte{3, 0, 0, 0})
	if err != nil || upgrade.Upgrade == nil || !upgrade.Upgrade.CloseBuffer {
		t.Fatalf("legacy Upgrade = %+v, %v", upgrade, err)
	}
	closeInstruction, err := DecodeInstruction(nil, []byte{5, 0, 0, 0})
	if err != nil || closeInstruction.Close == nil || closeInstruction.Close.Tombstone {
		t.Fatalf("legacy Close = %+v, %v", closeInstruction, err)
	}

	invalid := [][]byte{
		{2, 0, 0, 0, 42, 0, 0, 0, 0, 0, 0, 0, 2},
		{3, 0, 0, 0, 2},
		{5, 0, 0, 0, 2},
	}
	for index, data := range invalid {
		if _, err := DecodeInstruction(nil, data); !errors.Is(err, binary.ErrInvalidTag) {
			t.Errorf("invalid bool case %d error = %v", index, err)
		}
	}
}

func FuzzDecodeInstruction(f *testing.F) {
	for typ := InitializeBufferInstruction; typ <= SetAuthorityCheckedInstruction; typ++ {
		f.Add([]byte{byte(typ), 0, 0, 0})
	}
	f.Fuzz(func(t *testing.T, data []byte) { _, _ = DecodeInstruction(nil, data) })
}
