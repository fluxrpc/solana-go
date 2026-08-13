package rpc

import (
	"encoding/json"
	"reflect"
	"testing"
)

// jsonRoundTrip unmarshals fixture into T, re-marshals it, unmarshals the
// result again and requires both decoded values to be deeply equal. It
// returns the first decoded value.
func jsonRoundTrip[T any](t *testing.T, fixture []byte) T {
	t.Helper()
	var first T
	if err := json.Unmarshal(fixture, &first); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	data, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var second T
	if err := json.Unmarshal(data, &second); err != nil {
		t.Fatalf("unmarshal round trip: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("round trip mismatch:\nfirst:  %+v\nsecond: %+v\njson: %s", first, second, data)
	}
	return first
}
