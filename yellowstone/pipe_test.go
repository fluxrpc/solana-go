package yellowstone

import (
	"context"
	"sync"
	"testing"
	"time"

	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/rpc"
	pb "github.com/fluxrpc/solana-go/yellowstone/proto"
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
	go func() { done <- stream.PipeAccounts(sink) }()

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

type fakeCacheSink struct {
	fakeSink
	slots []struct {
		commitment rpc.CommitmentType
		slot       uint64
	}
	heights     []uint64
	blockhashes []struct {
		hash                 solana.Hash
		lastValidBlockHeight uint64
		slot                 uint64
		commitment           rpc.CommitmentType
	}
}

func (f *fakeCacheSink) CacheStoreSlot(commitment rpc.CommitmentType, slot uint64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.slots = append(f.slots, struct {
		commitment rpc.CommitmentType
		slot       uint64
	}{commitment, slot})
}

func (f *fakeCacheSink) CacheStoreBlockHeight(commitment rpc.CommitmentType, height uint64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.heights = append(f.heights, height)
}

func (f *fakeCacheSink) CacheStoreLatestBlockhash(commitment rpc.CommitmentType, hash solana.Hash, lastValidBlockHeight, slot uint64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.blockhashes = append(f.blockhashes, struct {
		hash                 solana.Hash
		lastValidBlockHeight uint64
		slot                 uint64
		commitment           rpc.CommitmentType
	}{hash, lastValidBlockHeight, slot, commitment})
}

func TestPipe(t *testing.T) {
	var pubkey, owner solana.PublicKey
	pubkey[0] = 7
	owner[0] = 9
	blockhash := solana.Hash{4, 2}

	client, mock := newTestClient(t)
	mock.subscribe = func(stream pb.Geyser_SubscribeServer) error {
		// Slot updates at each commitment level; lifecycle statuses ignored.
		stream.Send(&pb.SubscribeUpdate{UpdateOneof: &pb.SubscribeUpdate_Slot{
			Slot: &pb.SubscribeUpdateSlot{Slot: 100, Status: pb.SlotStatus_SLOT_PROCESSED},
		}})
		stream.Send(&pb.SubscribeUpdate{UpdateOneof: &pb.SubscribeUpdate_Slot{
			Slot: &pb.SubscribeUpdateSlot{Slot: 98, Status: pb.SlotStatus_SLOT_FINALIZED},
		}})
		stream.Send(&pb.SubscribeUpdate{UpdateOneof: &pb.SubscribeUpdate_Slot{
			Slot: &pb.SubscribeUpdateSlot{Slot: 99, Status: pb.SlotStatus_SLOT_DEAD},
		}})
		// Block metadata carrying blockhash + height.
		stream.Send(&pb.SubscribeUpdate{UpdateOneof: &pb.SubscribeUpdate_BlockMeta{
			BlockMeta: &pb.SubscribeUpdateBlockMeta{
				Slot:        100,
				Blockhash:   blockhash.String(),
				BlockHeight: &pb.BlockHeight{BlockHeight: 90},
			},
		}})
		// Account updates still flow.
		stream.Send(&pb.SubscribeUpdate{UpdateOneof: &pb.SubscribeUpdate_Account{
			Account: &pb.SubscribeUpdateAccount{
				Slot: 100,
				Account: &pb.SubscribeUpdateAccountInfo{
					Pubkey:   pubkey[:],
					Owner:    owner[:],
					Lamports: 55,
				},
			},
		}})
		return nil
	}

	stream, err := client.Subscribe(context.Background(), NewRequest(pb.CommitmentLevel_CONFIRMED))
	if err != nil {
		t.Fatal(err)
	}

	sink := &fakeCacheSink{}
	if err := stream.Pipe(rpc.CommitmentConfirmed, sink); err != nil {
		t.Fatal(err)
	}

	if len(sink.slots) != 2 {
		t.Fatalf("slots = %+v (dead status must be ignored)", sink.slots)
	}
	if sink.slots[0].commitment != rpc.CommitmentProcessed || sink.slots[0].slot != 100 {
		t.Fatalf("slot[0] = %+v", sink.slots[0])
	}
	if sink.slots[1].commitment != rpc.CommitmentFinalized || sink.slots[1].slot != 98 {
		t.Fatalf("slot[1] = %+v", sink.slots[1])
	}

	if len(sink.heights) != 1 || sink.heights[0] != 90 {
		t.Fatalf("heights = %v", sink.heights)
	}
	if len(sink.blockhashes) != 1 {
		t.Fatalf("blockhashes = %+v", sink.blockhashes)
	}
	bh := sink.blockhashes[0]
	if bh.hash != blockhash || bh.lastValidBlockHeight != 90+maxBlockhashAge || bh.slot != 100 || bh.commitment != rpc.CommitmentConfirmed {
		t.Fatalf("blockhash = %+v", bh)
	}

	if len(sink.snapshot()) != 1 || sink.snapshot()[0].data.Lamports != 55 {
		t.Fatalf("accounts = %+v", sink.snapshot())
	}
}
