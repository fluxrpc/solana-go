package rpc

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestDataBytesOrJSONBinaryTuple(t *testing.T) {
	for _, tc := range []struct {
		fixture []byte
		want    []byte
	}{
		{[]byte(`["dGVzdCBkYXRh","base64"]`), []byte("test data")},
		{[]byte(`["3yZe7d","base58"]`), []byte("test")},
	} {
		got := jsonRoundTrip[*DataBytesOrJSON](t, tc.fixture)
		if !bytes.Equal(got.GetBinary(), tc.want) {
			t.Fatalf("GetBinary() = %q, want %q", got.GetBinary(), tc.want)
		}
		if got.GetRawJSON() != nil {
			t.Fatal("GetRawJSON() is non-nil for binary data")
		}
	}
}

func TestDataBytesOrJSONParsed(t *testing.T) {
	fixture := []byte(`{"parsed":{"info":{"decimals":6},"type":"mint"},"program":"spl-token","space":82}`)
	got := jsonRoundTrip[*DataBytesOrJSON](t, fixture)
	if got.GetRawJSON() == nil {
		t.Fatal("GetRawJSON() is nil for jsonParsed data")
	}
	if got.GetBinary() != nil {
		t.Fatalf("GetBinary() = %x, want nil", got.GetBinary())
	}
	var parsed struct {
		Program string `json:"program"`
	}
	if err := json.Unmarshal(got.GetRawJSON(), &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.Program != "spl-token" {
		t.Fatalf("program = %q", parsed.Program)
	}
}

func TestDataBytesOrJSONRejectsInvalidInput(t *testing.T) {
	tests := []string{
		`["dGVzdA==","base64+zstd"]`, // unsupported by design
		`"just a string"`,
		`42`,
	}
	for _, input := range tests {
		var got DataBytesOrJSON
		if err := json.Unmarshal([]byte(input), &got); err == nil {
			t.Errorf("json.Unmarshal(%s) unexpectedly succeeded", input)
		}
	}

	// null leaves the value untouched rather than erroring.
	var got DataBytesOrJSON
	if err := json.Unmarshal([]byte(`null`), &got); err != nil {
		t.Fatal(err)
	}
	if got.GetBinary() != nil || got.GetRawJSON() != nil {
		t.Fatalf("null decoded to %+v", got)
	}
}

func TestDataBytesOrJSONConstructors(t *testing.T) {
	fromBytes := DataBytesOrJSONFromBytes([]byte("test data"))
	if !bytes.Equal(fromBytes.GetBinary(), []byte("test data")) {
		t.Fatalf("GetBinary() = %q", fromBytes.GetBinary())
	}
	data, err := json.Marshal(fromBytes)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `["dGVzdCBkYXRh","base64"]` {
		t.Fatalf("MarshalJSON() = %s", data)
	}

	fromBase64, err := DataBytesOrJSONFromBase64("dGVzdCBkYXRh")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(fromBase64.GetBinary(), []byte("test data")) {
		t.Fatalf("GetBinary() = %q", fromBase64.GetBinary())
	}

	if _, err := DataBytesOrJSONFromBase64("!!!"); err == nil {
		t.Fatal("DataBytesOrJSONFromBase64 accepted invalid base64")
	}
}
