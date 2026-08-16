package solana_go

import (
	"errors"
	"testing"
)

// The expected base64 fixtures below are ground truth generated with
// upstream gagliardetto/solana-go v1.22.0 NewTransaction for identical
// inputs (see benchcmp/txbuild_bench_test.go, which asserts byte-identity
// against upstream directly).
const (
	builderLegacyMessageB64 = "AQAEBgEGDxQdIiswOT5HTFVaY2hxdn+EjZKboKmut7zFytPYAgUMFx4hKDM6PURPVllga3J1fIeOkZijqq20v8bJ0NsDBA0WHyApMjs8RU5XWGFqc3R9ho+QmaKrrLW+x8jR2gQDChEYJy41PDtCSVBfZm10c3qBiJeepayrsrnAz9bd4ebv9P3Cy9DZ3qestbqDiJGWn2RtcntASU5XXCUqMzji5ez3/sHI09rdpK+2uYCLkpWcZ25xeENKTVRfJikwO6qtpL+2iYCbkpXs5/7xyMPa3dQvJjkwCwIFHBduYXhzAgQDAAECBAECAwQFAwADAQIJCA=="
	builderV0MessageB64     = "gAEAAgMBBg8UHSIrMDk+R0xVWmNocXZ/hI2Sm6Cprre8xcrT2OHm7/T9wsvQ2d6nrLW6g4iRlp9kbXJ7QElOV1wlKjM44uXs9/7ByNPa3aSvtrmAi5KVnGducXhDSk1UXyYpMDuqraS/tomAm5KV7Of+8cjD2t3ULyY5MAsCBRwXbmF4cwIBAwADBAQBAgMEAgMABQMCCQgB8Pf+5ezT2sHIz7a9pKuSmYCHjnV8Y2pRWF9GTTQ7IikBAQICAw=="
)

func builderKey(tag byte) (out PublicKey) {
	for i := range out {
		out[i] = tag ^ byte(i*7)
	}
	return out
}

var (
	builderAlice = builderKey(0x01)
	builderBob   = builderKey(0x02)
	builderCarol = builderKey(0x03)
	builderDave  = builderKey(0x04)
	builderProg1 = builderKey(0xE1)
	builderProg2 = builderKey(0xE2)
	builderTable = builderKey(0xF0)
	builderHash  = Hash(builderKey(0xAA))
)

func builderInstructions() []Instruction {
	return []Instruction{
		NewInstruction(builderProg1, AccountMetaSlice{
			{PublicKey: builderAlice, IsSigner: true, IsWritable: true},
			{PublicKey: builderBob, IsWritable: true},
			{PublicKey: builderCarol},
		}, []byte{1, 2, 3, 4}),
		NewInstruction(builderProg2, AccountMetaSlice{
			{PublicKey: builderAlice, IsSigner: true, IsWritable: true},
			{PublicKey: builderDave},
			{PublicKey: builderBob},
		}, []byte{9, 8}),
	}
}

func builderTables() map[PublicKey][]PublicKey {
	return map[PublicKey][]PublicKey{
		builderTable: {
			builderKey(0x77), // unrelated entry before the used ones
			builderBob,
			builderCarol,
			builderDave,
		},
	}
}

func TestNewTransactionLegacy(t *testing.T) {
	tx, err := NewTransaction(builderInstructions(), builderHash)
	if err != nil {
		t.Fatal(err)
	}
	if got := tx.Message.ToBase64(); got != builderLegacyMessageB64 {
		t.Fatalf("legacy message mismatch:\ngot  %s\nwant %s", got, builderLegacyMessageB64)
	}
	if tx.Message.GetVersion() != MessageVersionLegacy {
		t.Fatalf("version = %d, want legacy", tx.Message.GetVersion())
	}

	// Structure: fee payer (alice) first; bob writable (upgraded from ix2's
	// readonly use); header counts match.
	if tx.Message.AccountKeys[0] != builderAlice {
		t.Fatalf("fee payer not first: %s", tx.Message.AccountKeys[0])
	}
	h := tx.Message.Header
	if h.NumRequiredSignatures != 1 || h.NumReadonlySignedAccounts != 0 || h.NumReadonlyUnsignedAccounts != 4 {
		t.Fatalf("unexpected header %+v", h)
	}
	if len(tx.Message.AccountKeys) != 6 {
		t.Fatalf("len(AccountKeys) = %d, want 6", len(tx.Message.AccountKeys))
	}
}

func TestNewTransactionV0Lookups(t *testing.T) {
	tx, err := NewTransaction(builderInstructions(), builderHash,
		TransactionAddressTables(builderTables()))
	if err != nil {
		t.Fatal(err)
	}
	if got := tx.Message.ToBase64(); got != builderV0MessageB64 {
		t.Fatalf("v0 message mismatch:\ngot  %s\nwant %s", got, builderV0MessageB64)
	}
	if tx.Message.GetVersion() != MessageVersionV0 {
		t.Fatalf("version = %d, want v0", tx.Message.GetVersion())
	}
	if len(tx.Message.AddressTableLookups) != 1 {
		t.Fatalf("lookups = %d, want 1", len(tx.Message.AddressTableLookups))
	}
	lookup := tx.Message.AddressTableLookups[0]
	if lookup.AccountKey != builderTable {
		t.Fatalf("lookup table key = %s", lookup.AccountKey)
	}
	// bob is writable; carol and dave readonly; positions 1..3 in the table.
	if len(lookup.WritableIndexes) != 1 || lookup.WritableIndexes[0] != 1 {
		t.Fatalf("writable indexes = %v", lookup.WritableIndexes)
	}
	if len(lookup.ReadonlyIndexes) != 2 {
		t.Fatalf("readonly indexes = %v", lookup.ReadonlyIndexes)
	}
	// Static keys: fee payer + programs only (signers and invoked programs
	// can never be table-loaded).
	if len(tx.Message.AccountKeys) != 3 {
		t.Fatalf("static keys = %d, want 3", len(tx.Message.AccountKeys))
	}

	// Binary round trip decodes back to the same lookups.
	raw, err := tx.Message.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	var decoded Message
	if err := decoded.UnmarshalBinary(raw); err != nil {
		t.Fatal(err)
	}
	if len(decoded.AddressTableLookups) != 1 || decoded.AddressTableLookups[0].AccountKey != builderTable {
		t.Fatalf("round-trip lookups mismatch: %+v", decoded.AddressTableLookups)
	}
}

func TestNewTransactionFeePayerOption(t *testing.T) {
	payer := builderKey(0x55)
	tx, err := NewTransaction(builderInstructions(), builderHash, TransactionPayer(payer))
	if err != nil {
		t.Fatal(err)
	}
	if tx.Message.AccountKeys[0] != payer {
		t.Fatalf("fee payer not first: %s", tx.Message.AccountKeys[0])
	}
	// Explicit payer joins alice as a second required signature.
	if tx.Message.Header.NumRequiredSignatures != 2 {
		t.Fatalf("NumRequiredSignatures = %d, want 2", tx.Message.Header.NumRequiredSignatures)
	}
}

func TestNewTransactionErrors(t *testing.T) {
	if _, err := NewTransaction(nil, builderHash); err == nil {
		t.Fatal("expected error for empty instructions")
	}

	// No signer anywhere and no explicit payer.
	noSigner := []Instruction{
		NewInstruction(builderProg1, AccountMetaSlice{{PublicKey: builderBob}}, nil),
	}
	if _, err := NewTransaction(noSigner, builderHash); err == nil {
		t.Fatal("expected error for undeterminable fee payer")
	}

	// Oversized lookup table.
	big := make([]PublicKey, 257)
	for i := range big {
		big[i] = builderKey(byte(i))
	}
	_, err := NewTransaction(builderInstructions(), builderHash,
		TransactionAddressTables(map[PublicKey][]PublicKey{builderTable: big}))
	if err == nil {
		t.Fatal("expected error for oversized lookup table")
	}
}

func TestNewTransactionInstructionDataError(t *testing.T) {
	failing := &failingDataInstruction{prog: builderProg1, meta: AccountMetaSlice{
		{PublicKey: builderAlice, IsSigner: true, IsWritable: true},
	}}
	if _, err := NewTransaction([]Instruction{failing}, builderHash); err == nil {
		t.Fatal("expected instruction data error to propagate")
	}
}

type failingDataInstruction struct {
	prog PublicKey
	meta AccountMetaSlice
}

func (in *failingDataInstruction) ProgramID() PublicKey     { return in.prog }
func (in *failingDataInstruction) Accounts() []*AccountMeta { return in.meta }
func (in *failingDataInstruction) Data() ([]byte, error)    { return nil, errors.New("boom") }

func TestNewTransactionDoesNotMutateCallerMetas(t *testing.T) {
	metas := AccountMetaSlice{
		{PublicKey: builderAlice, IsSigner: true, IsWritable: true},
		{PublicKey: builderBob}, // readonly here
	}
	other := NewInstruction(builderProg2, AccountMetaSlice{
		{PublicKey: builderBob, IsWritable: true}, // writable here -> upgrade in message
	}, nil)
	_, err := NewTransaction([]Instruction{NewInstruction(builderProg1, metas, nil), other}, builderHash)
	if err != nil {
		t.Fatal(err)
	}
	if metas[1].IsWritable {
		t.Fatal("caller's AccountMeta was mutated by NewTransaction")
	}
}

func TestTransactionBuilder(t *testing.T) {
	ixs := builderInstructions()
	tx, err := NewTransactionBuilder().
		AddInstruction(ixs[0]).
		AddInstruction(ixs[1]).
		SetRecentBlockHash(builderHash).
		SetFeePayer(builderAlice).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if got := tx.Message.ToBase64(); got != builderLegacyMessageB64 {
		t.Fatalf("builder message mismatch:\ngot  %s\nwant %s", got, builderLegacyMessageB64)
	}
}

func TestNewTransactionSignAndVerify(t *testing.T) {
	signer, err := NewRandomPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	pub := signer.PublicKey()
	tx, err := NewTransaction([]Instruction{
		NewInstruction(builderProg1, AccountMetaSlice{
			{PublicKey: pub, IsSigner: true, IsWritable: true},
			{PublicKey: builderBob, IsWritable: true},
		}, []byte{42}),
	}, builderHash)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Sign(func(key PublicKey) *PrivateKey {
		if key == pub {
			return &signer
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := tx.VerifySignatures(); err != nil {
		t.Fatal(err)
	}
}
