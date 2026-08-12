package solana_go

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"strings"
	"testing"
)

func testSignature() Signature {
	var sig Signature
	for i := range sig {
		sig[i] = byte(i + 1)
	}
	return sig
}

func TestSignatureRoundTrip(t *testing.T) {
	tests := []Signature{
		{},
		testSignature(),
		SignatureFromBytes(bytes64Of(0xff)),
	}

	for _, want := range tests {
		encoded := want.String()
		got, err := SignatureFromBase58(encoded)
		if err != nil {
			t.Fatalf("SignatureFromBase58(%q): %v", encoded, err)
		}
		if got != want {
			t.Fatalf("round trip mismatch: got %x, want %x", got, want)
		}
	}
}

func TestSignatureFromBytes(t *testing.T) {
	in := bytes64Of(7)
	sig := SignatureFromBytes(in)
	in[0] = 9
	if sig[0] != 7 {
		t.Fatal("SignatureFromBytes retained the input slice")
	}

	// Short input fills the prefix and leaves the rest zero.
	short := SignatureFromBytes([]byte{1, 2, 3})
	if short[0] != 1 || short[2] != 3 || short[3] != 0 || short[63] != 0 {
		t.Fatalf("SignatureFromBytes short input mismatch: %x", short)
	}

	// Overlong input is truncated.
	long := SignatureFromBytes(append(bytes64Of(5), 0xaa))
	if long != SignatureFromBytes(bytes64Of(5)) {
		t.Fatalf("SignatureFromBytes overlong input mismatch: %x", long)
	}
}

func TestSignatureFromBase58RejectsInvalidInput(t *testing.T) {
	tests := []string{
		"",
		"abc",
		strings.Repeat("0", 88),
		strings.Repeat("z", 100),
		strings.Repeat("1", 63), // too short for 64 bytes
	}

	for _, input := range tests {
		got, err := SignatureFromBase58(input)
		if err == nil {
			t.Errorf("SignatureFromBase58(%q) unexpectedly succeeded", input)
		}
		if !got.IsZero() {
			t.Errorf("SignatureFromBase58(%q) returned data on error: %x", input, got)
		}
	}
}

func TestMustSignatureFromBase58Panics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("MustSignatureFromBase58 accepted invalid input")
		}
	}()
	MustSignatureFromBase58("not-base58")
}

func TestSignatureJSONRoundTrip(t *testing.T) {
	want := testSignature()
	data, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `"`+want.String()+`"` {
		t.Fatalf("MarshalJSON() = %s, want quoted %q", data, want.String())
	}

	var got Signature
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("JSON round trip mismatch: got %x, want %x", got, want)
	}
}

func TestSignatureUnmarshalJSONPreservesValueOnError(t *testing.T) {
	want := testSignature()
	tests := [][]byte{
		[]byte(`"invalid"`),
		[]byte(`{"not":"a string"}`),
		[]byte(`"unterminated`),
	}

	for _, data := range tests {
		got := want
		if err := json.Unmarshal(data, &got); err == nil {
			t.Errorf("json.Unmarshal(%s) unexpectedly succeeded", data)
		}
		if got != want {
			t.Errorf("json.Unmarshal(%s) changed receiver to %x", data, got)
		}
	}
}

func TestSignatureVerify(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	msg := []byte("solana test message")
	sig := SignatureFromBytes(ed25519.Sign(priv, msg))
	pubkey := PublicKeyFromBytes(pub)

	if !sig.Verify(pubkey, msg) {
		t.Fatal("Verify rejected a valid signature")
	}
	if sig.Verify(pubkey, []byte("other message")) {
		t.Fatal("Verify accepted a signature for a different message")
	}
	if testSignature().Verify(pubkey, msg) {
		t.Fatal("Verify accepted a bogus signature")
	}
}

func TestSignatureZero(t *testing.T) {
	var zero Signature
	if !zero.IsZero() {
		t.Fatal("zero Signature is not zero")
	}
	if testSignature().IsZero() {
		t.Fatal("non-zero Signature is zero")
	}
	if !zero.Equals(Signature{}) || zero.Equals(testSignature()) {
		t.Fatal("Equals mismatch")
	}
}

func bytes64Of(value byte) []byte {
	out := make([]byte, SignatureLength)
	for i := range out {
		out[i] = value
	}
	return out
}

var (
	benchmarkSignature       = testSignature()
	benchmarkSignatureString string
	benchmarkSignatureJSON   []byte
)

func BenchmarkSignatureString(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchmarkSignatureString = benchmarkSignature.String()
	}
}

func BenchmarkSignatureFromBase58(b *testing.B) {
	encoded := benchmarkSignature.String()
	b.ReportAllocs()
	for b.Loop() {
		if _, err := SignatureFromBase58(encoded); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSignatureMarshalJSON(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var err error
		benchmarkSignatureJSON, err = benchmarkSignature.MarshalJSON()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSignatureUnmarshalJSON(b *testing.B) {
	data, err := benchmarkSignature.MarshalJSON()
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		var sig Signature
		if err := sig.UnmarshalJSON(data); err != nil {
			b.Fatal(err)
		}
	}
}
