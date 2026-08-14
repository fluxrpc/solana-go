package benchcmp

// WebSocket notification throughput: each iteration dials the local push
// server, subscribes, and receives a fixed batch of account notifications.
// The server floods the batch as fast as the socket allows, so the number
// measures the full client pipeline: frame reading, envelope routing and
// typed decoding.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	fluxws "github.com/fluxrpc/solana-go/ws"
	gaglrpc "github.com/gagliardetto/solana-go/rpc"
	gaglws "github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

const wsBenchBatch = 20000

func newWSPushServer(b *testing.B) string {
	notification := []byte(`{"jsonrpc":"2.0","method":"accountNotification","params":{"result":{"context":{"slot":341197053},"value":` + rpcAccountFixture + `},"subscription":1}}`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			return
		}
		defer conn.Close()

		data, err := wsutil.ReadClientText(conn)
		if err != nil {
			return
		}
		var req struct {
			ID uint64 `json:"id"`
		}
		if err := json.Unmarshal(data, &req); err != nil {
			return
		}
		ack, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": 1})
		if err := wsutil.WriteServerText(conn, ack); err != nil {
			return
		}
		for i := 0; i < wsBenchBatch; i++ {
			if err := wsutil.WriteServerText(conn, notification); err != nil {
				return
			}
		}
		// Hold the connection until the client is done reading.
		wsutil.ReadClientText(conn)
	}))
	b.Cleanup(server.Close)
	return "ws" + strings.TrimPrefix(server.URL, "http")
}

func BenchmarkWs_AccountNotifications(b *testing.B) {
	url := newWSPushServer(b)
	ctx := context.Background()

	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			client, err := fluxws.ConnectWithOptions(ctx, url, &fluxws.Options{SubscriptionBuffer: wsBenchBatch})
			if err != nil {
				b.Fatal(err)
			}
			sub, err := client.AccountSubscribe(ctx, fluxPdaMint, "")
			if err != nil {
				b.Fatal(err)
			}
			for i := 0; i < wsBenchBatch; i++ {
				got, err := sub.Recv(ctx)
				if err != nil {
					b.Fatal(err)
				}
				if got.Value.Lamports == 0 {
					b.Fatal("bad notification")
				}
			}
			if dropped := sub.Dropped(); dropped != 0 {
				b.Fatalf("dropped %d notifications", dropped)
			}
			client.Close()
		}
		b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/wsBenchBatch, "ns/msg")
	})

	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			client, err := gaglws.Connect(ctx, url)
			if err != nil {
				b.Fatal(err)
			}
			sub, err := client.AccountSubscribe(gaglPdaMint, "")
			if err != nil {
				b.Fatal(err)
			}
			for i := 0; i < wsBenchBatch; i++ {
				got, err := sub.Recv(ctx)
				if err != nil {
					b.Fatal(err)
				}
				if got.Value.Lamports == 0 {
					b.Fatal("bad notification")
				}
			}
			sub.Unsubscribe()
			client.Close()
		}
		b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/wsBenchBatch, "ns/msg")
	})

	_ = gaglrpc.CommitmentFinalized
}
