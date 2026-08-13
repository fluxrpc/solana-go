package solana_go

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUnixTimeSeconds(t *testing.T) {
	ts := UnixTimeSeconds(1700000000)
	if got := ts.Time(); !got.Equal(time.Unix(1700000000, 0)) {
		t.Fatalf("Time() = %v", got)
	}

	data, err := json.Marshal(ts)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "1700000000" {
		t.Fatalf("MarshalJSON() = %s, want plain number", data)
	}

	var got UnixTimeSeconds
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got != ts {
		t.Fatalf("round trip = %d, want %d", got, ts)
	}
}
