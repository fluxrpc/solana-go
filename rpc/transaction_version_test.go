package rpc

import (
	"encoding/json"
	"testing"
)

func TestTransactionVersionUnmarshal(t *testing.T) {
	tests := []struct {
		in   string
		want TransactionVersion
	}{
		{`"legacy"`, LegacyTransactionVersion},
		{`null`, LegacyTransactionVersion},
		{`""`, LegacyTransactionVersion},
		{`0`, 0},
		{`1`, 1},
	}
	for _, tt := range tests {
		var v TransactionVersion
		if err := json.Unmarshal([]byte(tt.in), &v); err != nil {
			t.Fatalf("Unmarshal(%s): %v", tt.in, err)
		}
		if v != tt.want {
			t.Fatalf("Unmarshal(%s) = %d, want %d", tt.in, v, tt.want)
		}
	}

	var v TransactionVersion
	if err := json.Unmarshal([]byte(`"not-a-version"`), &v); err == nil {
		t.Fatal("Unmarshal accepted a garbage version")
	}
}

func TestTransactionVersionMarshal(t *testing.T) {
	tests := []struct {
		in   TransactionVersion
		want string
	}{
		{LegacyTransactionVersion, `"legacy"`},
		{0, `0`},
		{1, `1`},
		// Any other negative sentinel is marshaled as its number, only -1
		// is the legacy marker.
		{-2, `-2`},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt.in)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != tt.want {
			t.Fatalf("Marshal(%d) = %s, want %s", tt.in, data, tt.want)
		}
	}
}

func TestTransactionVersionMissing(t *testing.T) {
	// Pre-versioned-transactions RPC responses have no "version" key: the
	// field keeps its zero value (0), mirroring upstream behavior.
	var holder struct {
		Version TransactionVersion `json:"version"`
	}
	if err := json.Unmarshal([]byte(`{}`), &holder); err != nil {
		t.Fatal(err)
	}
	if holder.Version != 0 {
		t.Fatalf("Version = %d", holder.Version)
	}

	holder.Version = LegacyTransactionVersion
	data, err := json.Marshal(holder)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"version":"legacy"}` {
		t.Fatalf("Marshal = %s", data)
	}
}
