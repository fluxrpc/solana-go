package yellowstone

import (
	"context"
	"net"
	"sync"
	"testing"

	pb "github.com/rpcpool/yellowstone-grpc/examples/golang/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

// mockGeyser is an in-process pb.GeyserServer serving canned responses. It
// records the x-token metadata and requests of every call.
type mockGeyser struct {
	pb.UnimplementedGeyserServer

	mu       sync.Mutex
	tokens   []string
	requests []any

	// subscribe drives the Subscribe stream when set.
	subscribe func(stream pb.Geyser_SubscribeServer) error
}

func (m *mockGeyser) record(ctx context.Context, req any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	md, _ := metadata.FromIncomingContext(ctx)
	token := ""
	if values := md.Get("x-token"); len(values) > 0 {
		token = values[0]
	}
	m.tokens = append(m.tokens, token)
	m.requests = append(m.requests, req)
}

func (m *mockGeyser) lastToken(t *testing.T) string {
	t.Helper()
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.tokens) == 0 {
		t.Fatal("no calls recorded")
	}
	return m.tokens[len(m.tokens)-1]
}

func (m *mockGeyser) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PongResponse, error) {
	m.record(ctx, req)
	return &pb.PongResponse{Count: req.Count}, nil
}

func (m *mockGeyser) GetVersion(ctx context.Context, req *pb.GetVersionRequest) (*pb.GetVersionResponse, error) {
	m.record(ctx, req)
	return &pb.GetVersionResponse{Version: `{"version":"mock"}`}, nil
}

func (m *mockGeyser) GetSlot(ctx context.Context, req *pb.GetSlotRequest) (*pb.GetSlotResponse, error) {
	m.record(ctx, req)
	return &pb.GetSlotResponse{Slot: 341197053}, nil
}

func (m *mockGeyser) GetLatestBlockhash(ctx context.Context, req *pb.GetLatestBlockhashRequest) (*pb.GetLatestBlockhashResponse, error) {
	m.record(ctx, req)
	return &pb.GetLatestBlockhashResponse{
		Slot:                 341197053,
		Blockhash:            "EkSnNWid2cvwEVnVx9aBqawnmiCNiDgp3gUdkDPTKN1N",
		LastValidBlockHeight: 150,
	}, nil
}

func (m *mockGeyser) GetBlockHeight(ctx context.Context, req *pb.GetBlockHeightRequest) (*pb.GetBlockHeightResponse, error) {
	m.record(ctx, req)
	return &pb.GetBlockHeightResponse{BlockHeight: 100}, nil
}

func (m *mockGeyser) IsBlockhashValid(ctx context.Context, req *pb.IsBlockhashValidRequest) (*pb.IsBlockhashValidResponse, error) {
	m.record(ctx, req)
	return &pb.IsBlockhashValidResponse{Slot: 341197053, Valid: true}, nil
}

func (m *mockGeyser) Subscribe(stream pb.Geyser_SubscribeServer) error {
	m.record(stream.Context(), nil)
	if m.subscribe == nil {
		return nil
	}
	return m.subscribe(stream)
}

// newTestClient connects a Client to a bufconn-backed mockGeyser.
func newTestClient(t testing.TB, opts ...Option) (*Client, *mockGeyser) {
	t.Helper()
	listener := bufconn.Listen(1 << 20)
	mock := &mockGeyser{}
	server := grpc.NewServer()
	pb.RegisterGeyserServer(server, mock)
	go server.Serve(listener)
	t.Cleanup(server.Stop)

	opts = append(opts, WithGRPCOptions(grpc.WithContextDialer(
		func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.DialContext(ctx)
		},
	)))
	client, err := Connect(context.Background(), "passthrough:///bufnet", opts...)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { client.Close() })
	return client, mock
}

func TestClientTokenMetadata(t *testing.T) {
	client, mock := newTestClient(t, WithToken("secret"))
	ctx := context.Background()

	if _, err := client.Ping(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if token := mock.lastToken(t); token != "secret" {
		t.Fatalf("unary x-token = %q", token)
	}

	// Keep the server-side handler alive until the client hangs up, so the
	// client's initial Send never races the handler returning.
	mock.subscribe = func(stream pb.Geyser_SubscribeServer) error {
		for {
			if _, err := stream.Recv(); err != nil {
				return nil
			}
		}
	}
	stream, err := client.Subscribe(ctx, NewRequest(pb.CommitmentLevel_PROCESSED))
	if err != nil {
		t.Fatal(err)
	}
	stream.Close()
	if token := mock.lastToken(t); token != "secret" {
		t.Fatalf("stream x-token = %q", token)
	}
}

func TestClientNoTokenMetadata(t *testing.T) {
	client, mock := newTestClient(t)
	if _, err := client.Ping(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if token := mock.lastToken(t); token != "" {
		t.Fatalf("x-token = %q, want none", token)
	}
}

func TestClientUnaryWrappers(t *testing.T) {
	client, mock := newTestClient(t)
	ctx := context.Background()

	pong, err := client.Ping(ctx, 42)
	if err != nil {
		t.Fatal(err)
	}
	if pong.Count != 42 {
		t.Fatalf("pong count = %d", pong.Count)
	}

	version, err := client.GetVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if version.Version != `{"version":"mock"}` {
		t.Fatalf("version = %q", version.Version)
	}

	slot, err := client.GetSlot(ctx, pb.CommitmentLevel_FINALIZED)
	if err != nil {
		t.Fatal(err)
	}
	if slot.Slot != 341197053 {
		t.Fatalf("slot = %d", slot.Slot)
	}
	slotReq := mock.requests[len(mock.requests)-1].(*pb.GetSlotRequest)
	if slotReq.GetCommitment() != pb.CommitmentLevel_FINALIZED {
		t.Fatalf("commitment = %v", slotReq.GetCommitment())
	}

	blockhash, err := client.GetLatestBlockhash(ctx, pb.CommitmentLevel_CONFIRMED)
	if err != nil {
		t.Fatal(err)
	}
	if blockhash.Blockhash != "EkSnNWid2cvwEVnVx9aBqawnmiCNiDgp3gUdkDPTKN1N" || blockhash.LastValidBlockHeight != 150 {
		t.Fatalf("blockhash = %+v", blockhash)
	}
	if _, err := ConvertBlockhash(blockhash.Blockhash); err != nil {
		t.Fatal(err)
	}

	height, err := client.GetBlockHeight(ctx, pb.CommitmentLevel_PROCESSED)
	if err != nil {
		t.Fatal(err)
	}
	if height.BlockHeight != 100 {
		t.Fatalf("height = %d", height.BlockHeight)
	}

	valid, err := client.IsBlockhashValid(ctx, blockhash.Blockhash, pb.CommitmentLevel_CONFIRMED)
	if err != nil {
		t.Fatal(err)
	}
	if !valid.Valid {
		t.Fatal("expected valid blockhash")
	}
	validReq := mock.requests[len(mock.requests)-1].(*pb.IsBlockhashValidRequest)
	if validReq.Blockhash != blockhash.Blockhash || validReq.GetCommitment() != pb.CommitmentLevel_CONFIRMED {
		t.Fatalf("request = %+v", validReq)
	}
}
