package rpc

import (
	"encoding/json"
	"testing"
)

func TestContextJSON(t *testing.T) {
	var result struct {
		RPCContext
		Value uint64 `json:"value"`
	}
	payload := []byte(`{"context":{"slot":100208911,"apiVersion":"2.0.15"},"value":42}`)
	if err := json.Unmarshal(payload, &result); err != nil {
		t.Fatal(err)
	}
	if result.Context.Slot != 100208911 || result.Value != 42 {
		t.Fatalf("got %+v", result)
	}
	if result.Context.ApiVersion == nil || *result.Context.ApiVersion != "2.0.15" {
		t.Fatalf("ApiVersion = %v", result.Context.ApiVersion)
	}

	// Context without apiVersion omits the key on re-marshal.
	data, err := json.Marshal(Context{Slot: 5})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"slot":5}` {
		t.Fatalf("MarshalJSON() = %s", data)
	}
}
