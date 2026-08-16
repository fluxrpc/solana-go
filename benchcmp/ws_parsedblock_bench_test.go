package benchcmp

// Parsed-block notification throughput: same shape as the account
// notification pipeline benchmark, but with the heavier jsonParsed block
// payloads produced by blockSubscribe.

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

const wsParsedBlockBatch = 2000

// A parsed block carrying three parsed transactions.
var parsedBlockValue = `{"slot":112301554,"block":{"blockhash":"DZDb5RGDJp9DHkAVFF7WrujcaAAdMFRcaWSBv1PsPdvp","previousBlockhash":"C1qgvE2C5sjkTz1nQhNSbg6ZYK4hhWFbSD9CzjHYBSTS","parentSlot":112301553,"blockTime":1639926816,"blockHeight":101210751,"transactions":[` +
	`{"transaction":` + rpcParsedTransactionFixture + `,"meta":null},` +
	`{"transaction":` + rpcParsedTransactionFixture + `,"meta":null},` +
	`{"transaction":` + rpcParsedTransactionFixture + `,"meta":null}]}}`

func newWSParsedBlockPushServer(b *testing.B) string {
	notification := []byte(`{"jsonrpc":"2.0","method":"blockNotification","params":{"result":{"context":{"slot":112301554},"value":` + parsedBlockValue + `},"subscription":1}}`)

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
		for i := 0; i < wsParsedBlockBatch; i++ {
			if err := wsutil.WriteServerText(conn, notification); err != nil {
				return
			}
		}
		wsutil.ReadClientText(conn)
	}))
	b.Cleanup(server.Close)
	return "ws" + strings.TrimPrefix(server.URL, "http")
}

func BenchmarkWs_ParsedBlockNotifications(b *testing.B) {
	url := newWSParsedBlockPushServer(b)
	ctx := context.Background()

	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			client, err := fluxws.ConnectWithOptions(ctx, url, &fluxws.Options{SubscriptionBuffer: wsParsedBlockBatch})
			if err != nil {
				b.Fatal(err)
			}
			sub, err := client.ParsedBlockSubscribe(ctx, "")
			if err != nil {
				b.Fatal(err)
			}
			for i := 0; i < wsParsedBlockBatch; i++ {
				got, err := sub.Recv(ctx)
				if err != nil {
					b.Fatal(err)
				}
				if got.Value.Block == nil || len(got.Value.Block.Transactions) != 3 {
					b.Fatal("bad notification")
				}
			}
			if dropped := sub.Dropped(); dropped != 0 {
				b.Fatalf("dropped %d notifications", dropped)
			}
			client.Close()
		}
		b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/wsParsedBlockBatch, "ns/msg")
	})

	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			client, err := gaglws.Connect(ctx, url)
			if err != nil {
				b.Fatal(err)
			}
			sub, err := client.ParsedBlockSubscribe(
				gaglws.NewBlockSubscribeFilterAll(),
				&gaglws.BlockSubscribeOpts{Commitment: gaglrpc.CommitmentFinalized},
			)
			if err != nil {
				b.Fatal(err)
			}
			for i := 0; i < wsParsedBlockBatch; i++ {
				got, err := sub.Recv(ctx)
				if err != nil {
					b.Fatal(err)
				}
				if got.Value.Block == nil || len(got.Value.Block.Transactions) != 3 {
					b.Fatal("bad notification")
				}
			}
			client.Close()
		}
		b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/wsParsedBlockBatch, "ns/msg")
	})
}
