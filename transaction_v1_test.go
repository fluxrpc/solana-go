package solana_go

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"strings"
	"testing"
)

func u32ptr(v uint32) *uint32 { return &v }
func u64ptr(v uint64) *uint64 { return &v }

func v1TestKeys() (payer, other, program PublicKey) {
	return builderKey(0x11), builderKey(0x12), builderKey(0xEE)
}

func v1TestTx() *TransactionV1 {
	payer, other, program := v1TestKeys()
	return &TransactionV1{
		Header: MessageHeader{
			NumRequiredSignatures:       1,
			NumReadonlySignedAccounts:   0,
			NumReadonlyUnsignedAccounts: 1,
		},
		Config: TransactionConfig{
			PriorityFeeLamports: u64ptr(1000),
			ComputeUnitLimit:    u32ptr(200_000),
		},
		LifetimeSpecifier: Hash(builderKey(0xAB)),
		AccountKeys:       []PublicKey{payer, other, program},
		Instructions: []CompiledInstruction{{
			ProgramIDIndex: 2,
			Accounts:       []uint16{0, 1},
			Data:           Base58{1, 2, 3},
		}},
		Signatures: []Signature{{0x51}},
	}
}

// TestTransactionV1GoldenBytes hand-assembles the SIMD-0385 wire layout
// and asserts MarshalBinary matches it byte for byte.
func TestTransactionV1GoldenBytes(t *testing.T) {
	tx := v1TestTx()
	payer, other, program := v1TestKeys()

	var want []byte
	want = append(want, 129)     // VersionByte
	want = append(want, 1, 0, 1) // LegacyHeader
	// TransactionConfigMask: bits 0,1 (priority fee) + bit 2 (CU limit).
	want = binary.LittleEndian.AppendUint32(want, 0b111)
	lifetime := builderKey(0xAB)
	want = append(want, lifetime[:]...) // LifetimeSpecifier
	want = append(want, 1)              // NumInstructions
	want = append(want, 3)              // NumAddresses
	want = append(want, payer[:]...)
	want = append(want, other[:]...)
	want = append(want, program[:]...)
	// ConfigValues in ascending bit order: 8-byte fee, 4-byte CU limit.
	want = binary.LittleEndian.AppendUint64(want, 1000)
	want = binary.LittleEndian.AppendUint32(want, 200_000)
	// InstructionHeaders: (ProgramAccountIndex, NumAccounts, DataBytes u16).
	want = append(want, 2, 2)
	want = binary.LittleEndian.AppendUint16(want, 3)
	// InstructionPayloads: account indexes then data.
	want = append(want, 0, 1, 1, 2, 3)
	// Signatures.
	sig := Signature{0x51}
	want = append(want, sig[:]...)

	got, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("wire mismatch:\ngot  %x\nwant %x", got, want)
	}
}

func TestTransactionV1RoundTrip(t *testing.T) {
	tx := v1TestTx()
	raw, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := TransactionV1FromBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(tx, decoded) {
		t.Fatalf("round trip mismatch:\ngot  %+v\nwant %+v", decoded, tx)
	}
	again, err := decoded.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, again) {
		t.Fatal("re-marshal not byte-identical")
	}

	// Base64 round trip.
	fromB64, err := TransactionV1FromBase64(tx.ToBase64())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(tx, fromB64) {
		t.Fatal("base64 round trip mismatch")
	}
}

func TestTransactionV1ConfigCombinations(t *testing.T) {
	cases := []TransactionConfig{
		{},
		{PriorityFeeLamports: u64ptr(1)},
		{ComputeUnitLimit: u32ptr(1_400_000)},
		{LoadedAccountsDataSizeLimit: u32ptr(1 << 20)},
		{HeapSize: u32ptr(64 * 1024)},
		{
			PriorityFeeLamports:         u64ptr(1 << 40),
			ComputeUnitLimit:            u32ptr(9),
			LoadedAccountsDataSizeLimit: u32ptr(7),
			HeapSize:                    u32ptr(256 * 1024),
		},
	}
	for i, config := range cases {
		tx := v1TestTx()
		tx.Config = config
		raw, err := tx.MarshalBinary()
		if err != nil {
			t.Fatalf("case %d: %v", i, err)
		}
		decoded, err := TransactionV1FromBytes(raw)
		if err != nil {
			t.Fatalf("case %d: %v", i, err)
		}
		if !reflect.DeepEqual(&config, &decoded.Config) {
			t.Fatalf("case %d: config mismatch %+v != %+v", i, decoded.Config, config)
		}
	}
}

func TestTransactionV1SanitizationErrors(t *testing.T) {
	mutate := func(f func(*TransactionV1)) *TransactionV1 {
		tx := v1TestTx()
		f(tx)
		return tx
	}
	dupKey := builderKey(0x11)
	cases := map[string]*TransactionV1{
		"no signers":         mutate(func(tx *TransactionV1) { tx.Header.NumRequiredSignatures = 0 }),
		"readonly fee payer": mutate(func(tx *TransactionV1) { tx.Header.NumReadonlySignedAccounts = 1 }),
		"too many sigs": mutate(func(tx *TransactionV1) {
			tx.Header.NumRequiredSignatures = 13
			tx.Signatures = make([]Signature, 13)
			for len(tx.AccountKeys) < 14 {
				tx.AccountKeys = append(tx.AccountKeys, builderKey(byte(0x20+len(tx.AccountKeys))))
			}
		}),
		"too few addresses":  mutate(func(tx *TransactionV1) { tx.Header.NumReadonlyUnsignedAccounts = 60 }),
		"duplicate address":  mutate(func(tx *TransactionV1) { tx.AccountKeys[1] = dupKey }),
		"program index oob":  mutate(func(tx *TransactionV1) { tx.Instructions[0].ProgramIDIndex = 3 }),
		"account index oob":  mutate(func(tx *TransactionV1) { tx.Instructions[0].Accounts[0] = 9 }),
		"oversized ix data":  mutate(func(tx *TransactionV1) { tx.Instructions[0].Data = make(Base58, 0x10000) }),
		"heap too small":     mutate(func(tx *TransactionV1) { tx.Config.HeapSize = u32ptr(31 * 1024) }),
		"heap too large":     mutate(func(tx *TransactionV1) { tx.Config.HeapSize = u32ptr(257 * 1024) }),
		"heap not multiple":  mutate(func(tx *TransactionV1) { tx.Config.HeapSize = u32ptr(32*1024 + 1) }),
		"sig count mismatch": mutate(func(tx *TransactionV1) { tx.Signatures = nil }),
		"too many addresses": mutate(func(tx *TransactionV1) {
			for len(tx.AccountKeys) < 65 {
				tx.AccountKeys = append(tx.AccountKeys, builderKey(byte(0x20+len(tx.AccountKeys))))
			}
		}),
		"too many instructions": mutate(func(tx *TransactionV1) {
			for len(tx.Instructions) < 65 {
				tx.Instructions = append(tx.Instructions, CompiledInstruction{ProgramIDIndex: 2})
			}
		}),
	}
	for name, tx := range cases {
		if _, err := tx.MarshalBinary(); err == nil {
			t.Errorf("%s: expected marshal error", name)
		}
	}
}

func TestTransactionV1DecodeErrors(t *testing.T) {
	valid, err := v1TestTx().MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	corrupt := func(f func([]byte) []byte) []byte {
		c := bytes.Clone(valid)
		return f(c)
	}
	cases := map[string][]byte{
		"empty":         {},
		"short":         valid[:30],
		"wrong version": corrupt(func(b []byte) []byte { b[0] = 0x80; return b }),
		"unknown mask bits": corrupt(func(b []byte) []byte {
			binary.LittleEndian.PutUint32(b[4:8], 1<<7)
			return b
		}),
		"single priority bit": corrupt(func(b []byte) []byte {
			binary.LittleEndian.PutUint32(b[4:8], 0b101) // bit 0 without bit 1
			return b
		}),
		"trailing data":       append(bytes.Clone(valid), 0x00),
		"truncated addresses": valid[:60],
		"truncated sigs":      valid[:len(valid)-1],
		"oversized": corrupt(func(b []byte) []byte {
			return append(b, make([]byte, TransactionV1MaxSize)...)
		}),
	}
	for name, data := range cases {
		if _, err := TransactionV1FromBytes(data); err == nil {
			t.Errorf("%s: expected decode error", name)
		}
	}
}

func TestTransactionV1SignAndVerify(t *testing.T) {
	payerKey, err := NewRandomPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	coKey, err := NewRandomPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	payer, co := payerKey.PublicKey(), coKey.PublicKey()

	tx, err := NewTransactionV1([]Instruction{
		NewInstruction(builderProg1, AccountMetaSlice{
			{PublicKey: payer, IsSigner: true, IsWritable: true},
			{PublicKey: co, IsSigner: true, IsWritable: true},
			{PublicKey: builderKey(0x33), IsWritable: true},
		}, []byte{9}),
	}, Hash(builderKey(0xAB)), TransactionConfig{ComputeUnitLimit: u32ptr(150_000)},
		TransactionPayer(payer))
	if err != nil {
		t.Fatal(err)
	}
	if tx.Header.NumRequiredSignatures != 2 {
		t.Fatalf("required signatures = %d", tx.Header.NumRequiredSignatures)
	}

	// Sign with an incomplete getter fails while a slot is still empty.
	if _, err := tx.Sign(func(k PublicKey) *PrivateKey {
		if k == payer {
			return &payerKey
		}
		return nil
	}); err == nil {
		t.Fatal("Sign should fail while a signature is missing")
	}
	// Partial sign with only the co-signer, then complete with the payer.
	tx.Signatures = nil
	if _, err := tx.PartialSign(func(k PublicKey) *PrivateKey {
		if k == co {
			return &coKey
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Sign(func(k PublicKey) *PrivateKey {
		switch k {
		case payer:
			return &payerKey
		case co:
			return &coKey
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := tx.VerifySignatures(); err != nil {
		t.Fatal(err)
	}

	// Round trip retains valid signatures.
	raw, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := TransactionV1FromBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := decoded.VerifySignatures(); err != nil {
		t.Fatal(err)
	}

	// Tampering breaks verification.
	decoded.Instructions[0].Data = Base58{8}
	if err := decoded.VerifySignatures(); err == nil {
		t.Fatal("expected verification failure after tampering")
	}
}

func TestNewTransactionV1(t *testing.T) {
	tx, err := NewTransactionV1(builderInstructions(), Hash(builderKey(0xAB)),
		TransactionConfig{PriorityFeeLamports: u64ptr(5000)})
	if err != nil {
		t.Fatal(err)
	}
	// Same compilation as the legacy builder: alice payer first, 6 keys.
	if len(tx.AccountKeys) != 6 || tx.AccountKeys[0] != builderAlice {
		t.Fatalf("keys = %v", tx.AccountKeys)
	}
	if *tx.Config.PriorityFeeLamports != 5000 {
		t.Fatal("config not carried")
	}

	// ALTs are not part of the v1 format.
	if _, err := NewTransactionV1(builderInstructions(), Hash(builderKey(0xAB)),
		TransactionConfig{}, TransactionAddressTables(builderTables())); err == nil ||
		!strings.Contains(err.Error(), "lookup tables") {
		t.Fatalf("expected lookup-table rejection, got %v", err)
	}
}

func TestTransactionV1SizeLimit(t *testing.T) {
	payer := builderKey(0x11)
	build := func(dataLen int) (*TransactionV1, error) {
		return NewTransactionV1([]Instruction{
			NewInstruction(builderProg1, AccountMetaSlice{
				{PublicKey: payer, IsSigner: true, IsWritable: true},
			}, make([]byte, dataLen)),
		}, Hash(builderKey(0xAB)), TransactionConfig{})
	}

	// Well past the legacy 1232-byte cap but under 4096: must build,
	// marshal and round-trip.
	tx, err := build(3800)
	if err != nil {
		t.Fatal(err)
	}
	tx.Signatures = make([]Signature, 1)
	raw, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) <= 1232 || len(raw) > TransactionV1MaxSize {
		t.Fatalf("size = %d", len(raw))
	}
	if _, err := TransactionV1FromBytes(raw); err != nil {
		t.Fatal(err)
	}

	// Past 4096: rejected at build time.
	if _, err := build(4100); err == nil {
		t.Fatal("expected size error")
	}
}
