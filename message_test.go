package solana_go

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"testing"
)

// Real mainnet-shaped fixtures taken from gagliardetto/solana-go tests:
// a legacy system-transfer transaction (transaction_test.go) and a v0
// transaction with an address table lookup (transaction_v0 tests).
const (
	legacyTxBase64 = "AfjEs3XhTc3hrxEvlnMPkm/cocvAUbFNbCl00qKnrFue6J53AhEqIFmcJJlJW3EDP5RmcMz+cNTTcZHW/WJYwAcBAAEDO8hh4VddzfcO5jbCt95jryl6y8ff65UcgukHNLWH+UQGgxCGGpgyfQVQV02EQYqm4QwzUt2qf9f1gVLM7rI4hwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA6ANIF55zOZWROWRkeh+lExxZBnKFqbvIxZDLE7EijjoBAgIAAQwCAAAAOTAAAAAAAAA="

	v0TxBase64 = "Alkhq/BfGdBeok4oBP21xAwT4oO/R5PvkKqbCTq4sHHRsto+uDQCFcdp8hXh1g5D3mTh8GAJW8xE+EDD27f9IweTkH2Afiu4h5aM+Xbo0mklc0/Vi1xawd7SZVbstXDLtWdoJaf4Zt+20F/SasURzw/P4dkD+Q6BjgUNHT+vg5gOgAIBAQgaJV0Ch/DG6XwNcizWbI7STLgSbIOrg0Dl67Oo30WU1uA/NIbYLPRmuLarIJ4J0CcN3IWEm4Gf8675KhnXef2LaDXzjFgWVSbAO2yyTF6dK1oO3gTExie957LXDwu6oJMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAVKU1qZKSEGTSTocWDaOHx8NbXdvJK7geQfqEBBBUSN1LfoiB9oYLDSHJL9rjAlchZhn+fd/23ACfq0oIGla54pt5JT0MdBTJhQI+z7dnVsisw2xWwW+vFSTs97l0tJPxmv9kxpXbHYZFenDpT2s6CT75/9QNFVTkHFLMK+UG6VlyFnQmYh1aMkGtq3c6TIOsk32S6XMUnN9DQgFGQq4lwEAwIAAgwCAAAAgJaYAAAAAAADAgAFDAIAAACAlpgAAAAAAAMCAAYMAgAAAICWmAAAAAAABAAMSGVsbG8gRmFiaW8hAX5s37FH6IeB4QeMYxD4LtpXf1DaupH/ro7W+kEQnofaAgECAQA="
)

// messageBytes strips the signature section of a base64 transaction fixture,
// returning just the encoded message.
func messageBytes(t testing.TB, txBase64 string) []byte {
	t.Helper()
	raw, err := base64.StdEncoding.DecodeString(txBase64)
	if err != nil {
		t.Fatal(err)
	}
	numSigs, n, err := decodeShortvecLen(raw)
	if err != nil {
		t.Fatal(err)
	}
	return raw[n+numSigs*SignatureLength:]
}

func TestMessageUnmarshalLegacy(t *testing.T) {
	data := messageBytes(t, legacyTxBase64)

	var msg Message
	if err := msg.UnmarshalBinary(data); err != nil {
		t.Fatal(err)
	}

	if msg.GetVersion() != MessageVersionLegacy {
		t.Fatalf("GetVersion() = %d, want legacy", msg.GetVersion())
	}
	wantHeader := MessageHeader{NumRequiredSignatures: 1, NumReadonlySignedAccounts: 0, NumReadonlyUnsignedAccounts: 1}
	if msg.Header != wantHeader {
		t.Fatalf("Header = %+v, want %+v", msg.Header, wantHeader)
	}
	if len(msg.AccountKeys) != 3 {
		t.Fatalf("len(AccountKeys) = %d, want 3", len(msg.AccountKeys))
	}
	if got, want := msg.AccountKeys[0].String(), "52NGrUqh6tSGhr59ajGxsH3VnAaoRdSdTbAaV9G3UW35"; got != want {
		t.Fatalf("AccountKeys[0] = %s, want %s", got, want)
	}
	if !msg.AccountKeys[2].IsZero() {
		t.Fatalf("AccountKeys[2] = %s, want system program (zero)", msg.AccountKeys[2])
	}
	if got, want := msg.RecentBlockhash.String(), "GcgVK9buRA7YepZh3zXuS399GJAESCisLnLDBCmR5Aoj"; got != want {
		t.Fatalf("RecentBlockhash = %s, want %s", got, want)
	}
	if len(msg.Instructions) != 1 {
		t.Fatalf("len(Instructions) = %d, want 1", len(msg.Instructions))
	}
	ins := msg.Instructions[0]
	if ins.ProgramIDIndex != 2 || !reflect.DeepEqual(ins.Accounts, []uint16{0, 1}) || len(ins.Data) != 12 {
		t.Fatalf("Instructions[0] = %+v", ins)
	}
	if msg.AddressTableLookups != nil {
		t.Fatalf("AddressTableLookups = %v, want nil", msg.AddressTableLookups)
	}
}

func TestMessageUnmarshalV0(t *testing.T) {
	data := messageBytes(t, v0TxBase64)

	var msg Message
	if err := msg.UnmarshalBinary(data); err != nil {
		t.Fatal(err)
	}

	if msg.GetVersion() != MessageVersionV0 {
		t.Fatalf("GetVersion() = %d, want v0", msg.GetVersion())
	}
	wantHeader := MessageHeader{NumRequiredSignatures: 2, NumReadonlySignedAccounts: 1, NumReadonlyUnsignedAccounts: 1}
	if msg.Header != wantHeader {
		t.Fatalf("Header = %+v, want %+v", msg.Header, wantHeader)
	}
	if len(msg.AccountKeys) != 8 {
		t.Fatalf("len(AccountKeys) = %d, want 8", len(msg.AccountKeys))
	}
	if len(msg.Instructions) != 4 {
		t.Fatalf("len(Instructions) = %d, want 4", len(msg.Instructions))
	}
	last := msg.Instructions[3]
	if last.ProgramIDIndex != 4 || len(last.Accounts) != 0 || string(last.Data) != "Hello Fabio!" {
		t.Fatalf("Instructions[3] = %+v", last)
	}
	if len(msg.AddressTableLookups) != 1 {
		t.Fatalf("len(AddressTableLookups) = %d, want 1", len(msg.AddressTableLookups))
	}
	lookup := msg.AddressTableLookups[0]
	if got, want := lookup.AccountKey.String(), "9WWfC3y4uCNofr2qEFHSVUXkCxW99JiYkMWmSZvVt8j3"; got != want {
		t.Fatalf("lookup.AccountKey = %s, want %s", got, want)
	}
	if !reflect.DeepEqual(lookup.WritableIndexes, Uint8SliceAsNum{1, 2}) ||
		!reflect.DeepEqual(lookup.ReadonlyIndexes, Uint8SliceAsNum{0}) {
		t.Fatalf("lookup indexes = %+v", lookup)
	}
	if msg.AddressTableLookups.NumLookups() != 3 || msg.AddressTableLookups.NumWritableLookups() != 2 {
		t.Fatalf("NumLookups() = %d, NumWritableLookups() = %d", msg.AddressTableLookups.NumLookups(), msg.AddressTableLookups.NumWritableLookups())
	}
}

func TestMessageBinaryRoundTrip(t *testing.T) {
	for name, fixture := range map[string]string{"legacy": legacyTxBase64, "v0": v0TxBase64} {
		t.Run(name, func(t *testing.T) {
			data := messageBytes(t, fixture)

			var msg Message
			if err := msg.UnmarshalBinary(data); err != nil {
				t.Fatal(err)
			}
			out, err := msg.MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(out, data) {
				t.Fatalf("binary round trip mismatch:\ngot  %x\nwant %x", out, data)
			}
			if got, want := msg.ToBase64(), base64.StdEncoding.EncodeToString(data); got != want {
				t.Fatalf("ToBase64() = %s, want %s", got, want)
			}
		})
	}
}

func TestMessageJSONRoundTrip(t *testing.T) {
	for name, fixture := range map[string]string{"legacy": legacyTxBase64, "v0": v0TxBase64} {
		t.Run(name, func(t *testing.T) {
			var want Message
			if err := want.UnmarshalBinary(messageBytes(t, fixture)); err != nil {
				t.Fatal(err)
			}

			data, err := json.Marshal(want)
			if err != nil {
				t.Fatal(err)
			}
			// The addressTableLookups key is what marks a message as versioned.
			if got := bytes.Contains(data, []byte(`"addressTableLookups"`)); got != (name == "v0") {
				t.Fatalf("addressTableLookups present = %v in %s JSON: %s", got, name, data)
			}

			var got Message
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatal(err)
			}
			if got.GetVersion() != want.GetVersion() {
				t.Fatalf("version = %d, want %d", got.GetVersion(), want.GetVersion())
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("JSON round trip mismatch:\ngot  %+v\nwant %+v", got, want)
			}
		})
	}
}

func TestMessageUnmarshalBinaryRejectsTruncatedInput(t *testing.T) {
	for name, fixture := range map[string]string{"legacy": legacyTxBase64, "v0": v0TxBase64} {
		t.Run(name, func(t *testing.T) {
			data := messageBytes(t, fixture)
			for size := 0; size < len(data); size++ {
				var msg Message
				if err := msg.UnmarshalBinary(data[:size]); err == nil {
					t.Fatalf("UnmarshalBinary accepted truncated input of %d/%d bytes", size, len(data))
				}
			}
		})
	}
}

func TestMessageUnmarshalBinaryRejectsBadInput(t *testing.T) {
	tests := map[string][]byte{
		"empty":               nil,
		"unsupported version": {0x81, 1, 0, 1},
		"absurd key count":    {1, 0, 1, 0xff, 0xff, 0x03}, // claims 65535 keys
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			var msg Message
			if err := msg.UnmarshalBinary(data); err == nil {
				t.Fatal("UnmarshalBinary unexpectedly succeeded")
			}
		})
	}
}

func TestMessageSetVersion(t *testing.T) {
	var msg Message
	if _, err := msg.SetVersion(MessageVersionV0); err != nil || msg.GetVersion() != MessageVersionV0 {
		t.Fatalf("SetVersion(v0): %v, version %d", err, msg.GetVersion())
	}
	if _, err := msg.SetVersion(MessageVersion(42)); err == nil {
		t.Fatal("SetVersion accepted an invalid version")
	}

	msg = Message{}
	msg.AddAddressTableLookup(MessageAddressTableLookup{})
	if msg.GetVersion() != MessageVersionV0 {
		t.Fatal("AddAddressTableLookup did not switch version to v0")
	}
}

func TestMessageSigners(t *testing.T) {
	var msg Message
	if err := msg.UnmarshalBinary(messageBytes(t, legacyTxBase64)); err != nil {
		t.Fatal(err)
	}

	signers := msg.Signers()
	if len(signers) != 1 || signers[0] != msg.AccountKeys[0] {
		t.Fatalf("Signers() = %v", signers)
	}
	if !msg.IsSigner(msg.AccountKeys[0]) {
		t.Fatal("IsSigner(fee payer) = false")
	}
	if msg.IsSigner(msg.AccountKeys[1]) || msg.IsSigner(PublicKey{0xaa}) {
		t.Fatal("IsSigner accepted a non-signer")
	}
}

func TestUint8SliceAsNumJSON(t *testing.T) {
	data, err := json.Marshal(Uint8SliceAsNum{0, 1, 255})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "[0,1,255]" {
		t.Fatalf("MarshalJSON() = %s, want [0,1,255]", data)
	}

	var got Uint8SliceAsNum
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, Uint8SliceAsNum{0, 1, 255}) {
		t.Fatalf("round trip = %v", got)
	}

	if err := json.Unmarshal([]byte("[256]"), &got); err == nil {
		t.Fatal("UnmarshalJSON accepted a value beyond uint8 range")
	}
}

func TestShortvecRoundTrip(t *testing.T) {
	for _, n := range []int{0, 1, 0x7f, 0x80, 0x3fff, 0x4000, 0xffff} {
		encoded := appendShortvecLen(nil, n)
		if len(encoded) != shortvecLen(n) {
			t.Fatalf("shortvecLen(%d) = %d, encoded %d bytes", n, shortvecLen(n), len(encoded))
		}
		got, read, err := decodeShortvecLen(encoded)
		if err != nil {
			t.Fatalf("decodeShortvecLen(%x): %v", encoded, err)
		}
		if got != n || read != len(encoded) {
			t.Fatalf("decodeShortvecLen(%x) = (%d, %d), want (%d, %d)", encoded, got, read, n, len(encoded))
		}
	}
}

func TestShortvecRejectsBadInput(t *testing.T) {
	tests := map[string][]byte{
		"empty":            nil,
		"truncated":        {0x80},
		"beyond uint16":    {0xff, 0xff, 0x7f},
		"3-byte overflow":  {0x80, 0x80, 0x80},
		"value over limit": appendShortvecLen(nil, 0x10000),
	}
	for name, data := range tests {
		if _, _, err := decodeShortvecLen(data); err == nil {
			t.Errorf("decodeShortvecLen(%s %x) unexpectedly succeeded", name, data)
		}
	}
}

var (
	benchmarkMessage       Message
	benchmarkMessageBinary []byte
	benchmarkMessageJSON   []byte
)

func benchmarkLegacyMessage(b *testing.B) *Message {
	b.Helper()
	var msg Message
	if err := msg.UnmarshalBinary(messageBytes(b, legacyTxBase64)); err != nil {
		b.Fatal(err)
	}
	return &msg
}

func BenchmarkMessageMarshalBinary(b *testing.B) {
	msg := benchmarkLegacyMessage(b)
	b.ReportAllocs()
	for b.Loop() {
		var err error
		benchmarkMessageBinary, err = msg.MarshalBinary()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMessageUnmarshalBinary(b *testing.B) {
	data := messageBytes(b, legacyTxBase64)
	b.ReportAllocs()
	for b.Loop() {
		if err := benchmarkMessage.UnmarshalBinary(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMessageMarshalJSON(b *testing.B) {
	msg := benchmarkLegacyMessage(b)
	b.ReportAllocs()
	for b.Loop() {
		var err error
		benchmarkMessageJSON, err = json.Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMessageUnmarshalJSON(b *testing.B) {
	data, err := json.Marshal(benchmarkLegacyMessage(b))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		if err := benchmarkMessage.UnmarshalJSON(data); err != nil {
			b.Fatal(err)
		}
	}
}
