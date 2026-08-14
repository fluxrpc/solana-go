package ws

// Live test against a real WebSocket endpoint; runs only when WS_URL is set:
//
//	WS_URL=wss://... go test ./ws/ -run TestLive -v

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestLiveSlotSubscribe(t *testing.T) {
	url := os.Getenv("WS_URL")
	if url == "" {
		t.Skip("WS_URL not set; skipping live ws tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := Connect(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	sub, err := client.SlotSubscribe(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Mainnet produces a slot roughly every 400ms; two updates prove the
	// subscribe/notify/decode loop end to end.
	for i := 0; i < 2; i++ {
		slot, err := sub.Recv(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if slot.Slot == 0 {
			t.Fatalf("slot update = %+v", slot)
		}
		t.Logf("slot %d (parent %d, root %d)", slot.Slot, slot.Parent, slot.Root)
	}

	if err := sub.Unsubscribe(ctx); err != nil {
		t.Fatal(err)
	}
}
