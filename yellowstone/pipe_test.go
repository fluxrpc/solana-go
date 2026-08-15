package yellowstone

import (
	"context"
	"sync"
	"testing"
	"time"

	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/rpc"
	pb "github.com/rpcpool/yellowstone-grpc/examples/golang/proto"
)

type sinkUpdate struct {
	key  solana.PublicKey
	data *rpc.Account
	slot uint64
}

type fakeSink struct {
	mu      sync.Mutex
	updates []sinkUpdate
}

func (f *fakeSink) CacheStoreStreamed(account solana.PublicKey, data *rpc.Account, slot uint64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.updates = append(f.updates, sinkUpdate{key: account, data: data, slot: slot})
}

func (f *fakeSink) snapshot() []sinkUpdate {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]sinkUpdate(nil), f.updates...)
}

func TestPipeAccounts(t *testing.T) {
	var pubkey solana.PublicKey
	var owner solana.PublicKey
	pubkey[0] = 7
	owner[0] = 9
	payload := []byte("account-data")

	client, mock := newTestClient(t)
	mock.subscribe = func(stream pb.Geyser_SubscribeServer) error {
		// Slot updates and pings must be ignored by the pipe.
		stream.Send(&pb.SubscribeUpdate{UpdateOneof: &pb.SubscribeUpdate_Slot{
			Slot: &pb.SubscribeUpdateSlot{Slot: 41},
		}})
		stream.Send(&pb.SubscribeUpdate{UpdateOneof: &pb.SubscribeUpdate_Ping{Ping: &pb.SubscribeUpdatePing{}}})
		stream.Send(&pb.SubscribeUpdate{UpdateOneof: &pb.SubscribeUpdate_Account{
			Account: &pb.SubscribeUpdateAccount{
				Slot: 42,
				Account: &pb.SubscribeUpdateAccountInfo{
					Pubkey:     pubkey[:],
					Owner:      owner[:],
					Lamports:   1234,
					Data:       payload,
					Executable: false,
					RentEpoch:  361,
				},
			},
		}})
		return nil // clean shutdown -> io.EOF on the client
	}

	stream, err := client.Subscribe(context.Background(), NewRequest(pb.CommitmentLevel_PROCESSED))
	if err != nil {
		t.Fatal(err)
	}

	sink := &fakeSink{}
	done := make(chan error, 1)
	go func() { done <- PipeAccounts(stream, sink) }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("PipeAccounts: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("PipeAccounts did not finish")
	}

	updates := sink.snapshot()
	if len(updates) != 1 {
		t.Fatalf("got %d updates, want 1 (slot/ping must be ignored)", len(updates))
	}
	got := updates[0]
	if got.key != pubkey || got.slot != 42 {
		t.Fatalf("update = key %s slot %d", got.key, got.slot)
	}
	if got.data.Lamports != 1234 || got.data.Owner != owner || got.data.Space != uint64(len(payload)) {
		t.Fatalf("account = %+v", got.data)
	}
	if string(got.data.Data.GetBinary()) != "account-data" {
		t.Fatalf("data = %q", got.data.Data.GetBinary())
	}
	if got.data.RentEpoch.Uint64() != 361 {
		t.Fatalf("rentEpoch = %s", got.data.RentEpoch)
	}
}
