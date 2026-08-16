package benchcmp

// Full send-and-confirm pipeline: each iteration dials the local WS server,
// sends the transaction to the local RPC server, and waits for the
// signature notification. The WS server pushes the confirmation as soon as
// the RPC server has seen sendTransaction, so the number measures the
// client-side pipeline (connect, subscribe, send, notify, decode) without
// artificial delays.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	solana "github.com/fluxrpc/solana-go"
	fluxconfirm "github.com/fluxrpc/solana-go/confirm"
	fluxrpc "github.com/fluxrpc/solana-go/rpc"
	fluxws "github.com/fluxrpc/solana-go/ws"
	gagl "github.com/gagliardetto/solana-go"
	gaglrpc "github.com/gagliardetto/solana-go/rpc"
	gaglconfirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	gaglws "github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// confirmBenchServers starts an RPC server answering sendTransaction and a
// WS server that acks signatureSubscribe and pushes the success
// notification once sendTransaction has been observed (whichever order the
// client does those in).
func confirmBenchServers(b *testing.B, sigB58 string) (rpcURL, wsURL string) {
	sendSeen := make(chan struct{}, 1)

	rpcServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			// flux uses numeric ids, gagl uses uuid strings; echo verbatim.
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			b.Errorf("bad rpc request: %v", err)
			return
		}
		if req.Method != "sendTransaction" {
			b.Errorf("unexpected rpc method %s", req.Method)
			return
		}
		sendSeen <- struct{}{}
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"result":%q}`, req.ID, sigB58)
	}))
	b.Cleanup(rpcServer.Close)

	notification := []byte(`{"jsonrpc":"2.0","method":"signatureNotification","params":{"result":{"context":{"slot":5},"value":{"err":null}},"subscription":1}}`)
	wsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		if err := wsutil.WriteServerText(conn, fmt.Appendf(nil, `{"jsonrpc":"2.0","id":%d,"result":1}`, req.ID)); err != nil {
			return
		}
		<-sendSeen
		if err := wsutil.WriteServerText(conn, notification); err != nil {
			return
		}
		// Serve unsubscribes until the client hangs up.
		for {
			data, err := wsutil.ReadClientText(conn)
			if err != nil {
				return
			}
			var req struct {
				ID uint64 `json:"id"`
			}
			if json.Unmarshal(data, &req) == nil {
				_ = wsutil.WriteServerText(conn, fmt.Appendf(nil, `{"jsonrpc":"2.0","id":%d,"result":true}`, req.ID))
			}
		}
	}))
	b.Cleanup(wsServer.Close)

	return rpcServer.URL, "ws" + strings.TrimPrefix(wsServer.URL, "http")
}

// confirmBenchTx builds one deterministic signed transaction usable by both
// implementations.
func confirmBenchTx(b *testing.B) *solana.Transaction {
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i + 100)
	}
	key, err := solana.PrivateKeyFromSeed(seed)
	if err != nil {
		b.Fatal(err)
	}
	pub := key.PublicKey()
	tx, err := solana.NewTransaction([]solana.Instruction{
		solana.NewInstruction(solana.MemoProgramID, solana.AccountMetaSlice{
			{PublicKey: pub, IsSigner: true, IsWritable: true},
		}, []byte("confirm-bench")),
	}, solana.Hash{7})
	if err != nil {
		b.Fatal(err)
	}
	if _, err := tx.Sign(func(k solana.PublicKey) *solana.PrivateKey {
		if k == pub {
			return &key
		}
		return nil
	}); err != nil {
		b.Fatal(err)
	}
	return tx
}

func BenchmarkConfirm_SendAndConfirm(b *testing.B) {
	fluxTx := confirmBenchTx(b)
	raw, err := fluxTx.MarshalBinary()
	if err != nil {
		b.Fatal(err)
	}
	gaglTx, err := gagl.TransactionFromBytes(raw)
	if err != nil {
		b.Fatal(err)
	}
	sigB58 := fluxTx.Signatures[0].String()
	rpcURL, wsURL := confirmBenchServers(b, sigB58)
	ctx := context.Background()

	b.Run("flux", func(b *testing.B) {
		client := fluxrpc.New(rpcURL)
		b.ReportAllocs()
		for b.Loop() {
			wsClient, err := fluxws.Connect(ctx, wsURL)
			if err != nil {
				b.Fatal(err)
			}
			if _, err := fluxconfirm.SendAndConfirm(ctx, client, wsClient, fluxTx); err != nil {
				b.Fatal(err)
			}
			wsClient.Close()
		}
	})

	b.Run("gagl", func(b *testing.B) {
		client := gaglrpc.New(rpcURL)
		b.ReportAllocs()
		for b.Loop() {
			wsClient, err := gaglws.Connect(ctx, wsURL)
			if err != nil {
				b.Fatal(err)
			}
			if _, err := gaglconfirm.SendAndConfirmTransaction(ctx, client, wsClient, gaglTx); err != nil {
				b.Fatal(err)
			}
			wsClient.Close()
		}
	})
}
