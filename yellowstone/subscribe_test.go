package yellowstone

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	pb "github.com/fluxrpc/solana-go/yellowstone/proto"
)

func TestSubscribe(t *testing.T) {
	client, mock := newTestClient(t)

	requests := make(chan *pb.SubscribeRequest, 2)
	done := make(chan error, 1)
	mock.subscribe = func(stream pb.Geyser_SubscribeServer) error {
		// The initial request must arrive before any updates flow.
		initial, err := stream.Recv()
		if err != nil {
			done <- err
			return err
		}
		requests <- initial

		script := []*pb.SubscribeUpdate{
			{Filters: []string{"slots"}, UpdateOneof: &pb.SubscribeUpdate_Slot{
				Slot: &pb.SubscribeUpdateSlot{Slot: 100, Status: pb.SlotStatus_SLOT_CONFIRMED},
			}},
			{UpdateOneof: &pb.SubscribeUpdate_Account{
				Account: &pb.SubscribeUpdateAccount{
					Slot:    100,
					Account: &pb.SubscribeUpdateAccountInfo{Lamports: 5, Data: []byte{1, 2, 3}},
				},
			}},
			{UpdateOneof: &pb.SubscribeUpdate_Transaction{
				Transaction: &pb.SubscribeUpdateTransaction{Slot: 100},
			}},
			{UpdateOneof: &pb.SubscribeUpdate_Ping{Ping: &pb.SubscribeUpdatePing{}}},
		}
		for _, update := range script {
			if err := stream.Send(update); err != nil {
				done <- err
				return err
			}
		}

		// A live filter change arrives as a second request; acknowledge it
		// with a pong so the client can observe the round trip.
		updated, err := stream.Recv()
		if err != nil {
			done <- err
			return err
		}
		requests <- updated
		if err := stream.Send(&pb.SubscribeUpdate{UpdateOneof: &pb.SubscribeUpdate_Pong{
			Pong: &pb.SubscribeUpdatePong{Id: 7},
		}}); err != nil {
			done <- err
			return err
		}

		// Close on the client half-closes the send side.
		if _, err := stream.Recv(); err != io.EOF {
			done <- err
			return err
		}
		done <- nil
		return nil
	}

	req := NewRequest(pb.CommitmentLevel_CONFIRMED)
	req.AllSlots("slots")
	stream, err := client.Subscribe(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	initial := <-requests
	if initial.GetCommitment() != pb.CommitmentLevel_CONFIRMED {
		t.Fatalf("commitment = %v", initial.GetCommitment())
	}
	if _, ok := initial.Slots["slots"]; !ok {
		t.Fatalf("slots filter missing: %+v", initial.Slots)
	}

	slot, err := stream.Recv()
	if err != nil {
		t.Fatal(err)
	}
	if slot.GetSlot().GetSlot() != 100 || slot.Filters[0] != "slots" {
		t.Fatalf("slot update = %+v", slot)
	}
	account, err := stream.Recv()
	if err != nil {
		t.Fatal(err)
	}
	if account.GetAccount().GetAccount().GetLamports() != 5 {
		t.Fatalf("account update = %+v", account)
	}
	transaction, err := stream.Recv()
	if err != nil {
		t.Fatal(err)
	}
	if transaction.GetTransaction().GetSlot() != 100 {
		t.Fatalf("transaction update = %+v", transaction)
	}
	if _, err := stream.Recv(); err != nil {
		t.Fatal(err) // ping
	}

	change := NewRequest(pb.CommitmentLevel_CONFIRMED)
	change.AccountsByOwner("usdc", "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
	if err := stream.Update(change); err != nil {
		t.Fatal(err)
	}
	updated := <-requests
	filter, ok := updated.Accounts["usdc"]
	if !ok || filter.Owner[0] != "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA" {
		t.Fatalf("updated request = %+v", updated)
	}
	pong, err := stream.Recv()
	if err != nil {
		t.Fatal(err)
	}
	if pong.GetPong().GetId() != 7 {
		t.Fatalf("pong = %+v", pong)
	}

	if err := stream.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("server: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server did not observe stream close")
	}
	if _, err := stream.Recv(); err == nil {
		t.Fatal("Recv after Close should fail")
	}
}

func TestSubscribeSendError(t *testing.T) {
	client, mock := newTestClient(t)
	mock.subscribe = func(stream pb.Geyser_SubscribeServer) error {
		return errors.New("rejected")
	}

	// The server tears the stream down immediately; the initial send (or the
	// first Recv) must surface an error rather than hang.
	stream, err := client.Subscribe(context.Background(), NewRequest(pb.CommitmentLevel_PROCESSED))
	if err != nil {
		return
	}
	if _, err := stream.Recv(); err == nil {
		t.Fatal("expected error from rejected subscription")
	}
	stream.Close()
}

func TestNewRequestInitializesAllFilterMaps(t *testing.T) {
	req := NewRequest(pb.CommitmentLevel_FINALIZED)
	if req.Accounts == nil || req.Slots == nil || req.Transactions == nil ||
		req.TransactionsStatus == nil || req.Blocks == nil || req.BlocksMeta == nil || req.Entry == nil {
		t.Fatalf("uninitialized filter map: %+v", req)
	}
	if req.GetCommitment() != pb.CommitmentLevel_FINALIZED {
		t.Fatalf("commitment = %v", req.GetCommitment())
	}
}

func TestFilterBuilders(t *testing.T) {
	req := NewRequest(pb.CommitmentLevel_PROCESSED).
		AccountsByOwner("owner", "o1", "o2").
		AccountsByKey("keys", "k1").
		TransactionsByAccount("include", "a1", "a2").
		TransactionsByAccountRequired("required", "a1").
		TransactionStatusesByAccount("status", "a2").
		BlocksIncluding("blocks", "a1").
		AllSlots("slots").
		AllBlocksMeta("block-meta").
		AllEntries("entries")

	owner := req.Accounts["owner"]
	if len(owner.Owner) != 2 || owner.Owner[0] != "o1" || len(owner.Account) != 0 {
		t.Fatalf("AccountsByOwner = %+v", owner)
	}

	keys := req.Accounts["keys"]
	if len(keys.Account) != 1 || keys.Account[0] != "k1" || len(keys.Owner) != 0 {
		t.Fatalf("AccountsByKey = %+v", keys)
	}

	include := req.Transactions["include"]
	if len(include.AccountInclude) != 2 || len(include.AccountRequired) != 0 {
		t.Fatalf("TransactionsByAccount = %+v", include)
	}

	required := req.Transactions["required"]
	if len(required.AccountRequired) != 1 || len(required.AccountInclude) != 0 {
		t.Fatalf("TransactionsByAccountRequired = %+v", required)
	}
	if status := req.TransactionsStatus["status"]; len(status.AccountInclude) != 1 || status.AccountInclude[0] != "a2" {
		t.Fatalf("TransactionStatusesByAccount = %+v", status)
	}

	blocks := req.Blocks["blocks"]
	if len(blocks.AccountInclude) != 1 || blocks.AccountInclude[0] != "a1" {
		t.Fatalf("Blocks = %+v", blocks)
	}

	if len(req.Slots) != 1 || len(req.BlocksMeta) != 1 || len(req.Entry) != 1 {
		t.Fatalf("Add helpers misregistered: %+v", req)
	}
}
