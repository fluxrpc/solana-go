package yellowstone

import (
	"context"
	"os"
	"testing"
	"time"

	pb "github.com/fluxrpc/solana-go/yellowstone/proto"
)

// TestLiveYellowstone runs against a real endpoint:
//
//	YELLOWSTONE_ENDPOINT=https://host:443 [YELLOWSTONE_TOKEN=...] go test -run TestLiveYellowstone
func TestLiveYellowstone(t *testing.T) {
	endpoint := os.Getenv("YELLOWSTONE_ENDPOINT")
	if endpoint == "" {
		t.Skip("YELLOWSTONE_ENDPOINT not set")
	}
	var opts []Option
	if token := os.Getenv("YELLOWSTONE_TOKEN"); token != "" {
		opts = append(opts, WithToken(token))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := Connect(ctx, endpoint, opts...)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	pong, err := client.Ping(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if pong.Count != 1 {
		t.Fatalf("pong count = %d", pong.Count)
	}

	version, err := client.GetVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if version.Version == "" {
		t.Fatal("empty version")
	}
	t.Logf("version: %s", version.Version)

	slot, err := client.GetSlot(ctx, pb.CommitmentLevel_PROCESSED)
	if err != nil {
		t.Fatal(err)
	}
	if slot.Slot == 0 {
		t.Fatal("slot = 0")
	}
	t.Logf("slot: %d", slot.Slot)

	req := NewRequest(pb.CommitmentLevel_PROCESSED)
	req.AllSlots("slots")
	stream, err := client.Subscribe(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()

	// Anything but pings counts; a slot update must arrive within the
	// context's 30s on any live chain.
	for {
		update, err := stream.Recv()
		if err != nil {
			t.Fatalf("no slot update within 30s: %v", err)
		}
		if slotUpdate := update.GetSlot(); slotUpdate != nil {
			t.Logf("slot update: %d (%s)", slotUpdate.Slot, slotUpdate.Status)
			return
		}
	}
}
