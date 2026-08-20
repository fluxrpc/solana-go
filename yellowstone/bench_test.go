package yellowstone

import (
	"context"
	"testing"

	solana "github.com/fluxrpc/solana-go"
	pb "github.com/rpcpool/yellowstone-grpc/examples/golang/proto"
	"google.golang.org/grpc"
)

// BenchmarkSubscribeThroughput measures end-to-end updates through a bufconn
// stream: the server marshals one account update ahead of time
// (grpc.PreparedMsg) and sends it in a tight loop; ns/op is the cost of one
// received update on the client.
func BenchmarkSubscribeThroughput(b *testing.B) {
	client, mock := newTestClient(b)
	mock.subscribe = func(stream pb.Geyser_SubscribeServer) error {
		if _, err := stream.Recv(); err != nil {
			return err
		}
		update := &pb.SubscribeUpdate{UpdateOneof: &pb.SubscribeUpdate_Account{
			Account: &pb.SubscribeUpdateAccount{
				Slot: 341197053,
				Account: &pb.SubscribeUpdateAccountInfo{
					Pubkey:       make([]byte, 32),
					Lamports:     500,
					Owner:        make([]byte, 32),
					RentEpoch:    361,
					Data:         make([]byte, 165), // SPL token account size
					WriteVersion: 42,
				},
			},
		}}
		prepared := &grpc.PreparedMsg{}
		if err := prepared.Encode(stream, update); err != nil {
			return err
		}
		for {
			if err := stream.SendMsg(prepared); err != nil {
				return nil // client closed
			}
		}
	}

	stream, err := client.Subscribe(context.Background(), NewRequest(pb.CommitmentLevel_PROCESSED))
	if err != nil {
		b.Fatal(err)
	}
	defer stream.Close()

	b.ReportAllocs()
	count := 0
	for b.Loop() {
		update, err := stream.Recv()
		if err != nil {
			b.Fatal(err)
		}
		if update.Account() == nil {
			b.Fatal("missing account update")
		}
		count++
	}
	b.ReportMetric(float64(count)/b.Elapsed().Seconds(), "updates/sec")
}

func BenchmarkUpdateTransaction(b *testing.B) {
	tx := testTransaction(b, true)
	computeUnits := uint64(4200)
	update := wrapTransactionUpdate(&pb.SubscribeUpdateTransaction{
		Slot: 341197053,
		Transaction: &pb.SubscribeUpdateTransactionInfo{
			Signature:   tx.Signatures[0].Bytes(),
			Transaction: txToProto(tx),
			Meta: &pb.TransactionStatusMeta{
				Fee:                  5000,
				PreBalances:          []uint64{100, 200},
				PostBalances:         []uint64{90, 210},
				LogMessages:          []string{"Program log: hi", "Program log: bye"},
				ComputeUnitsConsumed: &computeUnits,
			},
		},
	})

	b.ReportAllocs()
	for b.Loop() {
		if _, _, err := update.Transaction(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdateAccount(b *testing.B) {
	key, err := solana.NewRandomPrivateKey()
	if err != nil {
		b.Fatal(err)
	}
	update := wrapAccountUpdate(&pb.SubscribeUpdateAccount{
		Slot: 341197053,
		Account: &pb.SubscribeUpdateAccountInfo{
			Pubkey:       key.PublicKey().Bytes(),
			Lamports:     500,
			Owner:        solana.MustPublicKeyFromBase58(tokenProgram).Bytes(),
			RentEpoch:    361,
			Data:         make([]byte, 165),
			WriteVersion: 42,
		},
	})

	b.ReportAllocs()
	for b.Loop() {
		if update.Account() == nil {
			b.Fatal("nil update")
		}
	}
}
