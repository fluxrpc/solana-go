package ws

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/rpc"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

const wsAccountValue = `{"lamports":88849814690250,"owner":"11111111111111111111111111111111","data":["dGVzdCBkYXRh","base64"],"executable":false,"rentEpoch":361,"space":9}`

type wsRequest struct {
	ID     uint64          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

// wsTestServer speaks the pubsub wire protocol: it auto-acks *Subscribe
// requests with incrementing subscription IDs, replies true to
// *Unsubscribe, records every request, and lets tests push notifications.
type wsTestServer struct {
	t    *testing.T
	url  string
	reqs chan wsRequest

	mu            sync.Mutex
	conn          net.Conn
	nextSub       uint64
	ready         chan struct{}
	failNext      atomic.Bool
	burstAfterAck atomic.Int64
}

func newWSTestServer(t *testing.T) *wsTestServer {
	t.Helper()
	ts := &wsTestServer{t: t, reqs: make(chan wsRequest, 64), ready: make(chan struct{})}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			return
		}
		ts.mu.Lock()
		ts.conn = conn
		ts.mu.Unlock()
		close(ts.ready)
		for {
			data, err := wsutil.ReadClientText(conn)
			if err != nil {
				return
			}
			var req wsRequest
			if err := json.Unmarshal(data, &req); err != nil {
				continue
			}
			ts.reqs <- req
			switch {
			case ts.failNext.CompareAndSwap(true, false):
				ts.write(`{"jsonrpc":"2.0","id":` + uitoa(req.ID) + `,"error":{"code":-32602,"message":"invalid params"}}`)
			case strings.HasSuffix(req.Method, "Unsubscribe"):
				ts.write(`{"jsonrpc":"2.0","id":` + uitoa(req.ID) + `,"result":true}`)
			case strings.HasSuffix(req.Method, "Subscribe"):
				ts.mu.Lock()
				ts.nextSub++
				sub := ts.nextSub
				ts.mu.Unlock()
				ts.write(`{"jsonrpc":"2.0","id":` + uitoa(req.ID) + `,"result":` + uitoa(sub) + `}`)
				if n := ts.burstAfterAck.Swap(0); n > 0 {
					for i := int64(0); i < n; i++ {
						ts.push(sub, "slotNotification", `{"parent":1,"root":1,"slot":`+uitoa(uint64(100+i))+`}`)
					}
				}
			}
		}
	}))
	t.Cleanup(server.Close)
	ts.url = "ws" + strings.TrimPrefix(server.URL, "http")
	return ts
}

func uitoa(v uint64) string {
	if v == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}

func (ts *wsTestServer) write(frame string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if err := wsutil.WriteServerText(ts.conn, []byte(frame)); err != nil {
		ts.t.Logf("server write: %v", err)
	}
}

// push delivers a notification for the given subscription.
func (ts *wsTestServer) push(subID uint64, method, result string) {
	ts.write(`{"jsonrpc":"2.0","method":"` + method + `","params":{"result":` + result + `,"subscription":` + uitoa(subID) + `}}`)
}

func (ts *wsTestServer) closeConn() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.conn.Close()
}

func (ts *wsTestServer) nextReq(t *testing.T) wsRequest {
	t.Helper()
	select {
	case req := <-ts.reqs:
		return req
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for request")
		return wsRequest{}
	}
}

func testConnect(t *testing.T, ts *wsTestServer, opts *Options) *Client {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client, err := ConnectWithOptions(ctx, ts.url, opts)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { client.Close() })
	return client
}

func TestAccountSubscribeLifecycle(t *testing.T) {
	ts := newWSTestServer(t)
	client := testConnect(t, ts, nil)
	ctx := context.Background()

	account := solana.MustPublicKeyFromBase58("SysvarC1ock11111111111111111111111111111111")
	sub, err := client.AccountSubscribe(ctx, account, rpc.CommitmentFinalized)
	if err != nil {
		t.Fatal(err)
	}

	req := ts.nextReq(t)
	if req.Method != "accountSubscribe" {
		t.Fatalf("method = %s", req.Method)
	}
	var params []json.RawMessage
	if err := json.Unmarshal(req.Params, &params); err != nil {
		t.Fatal(err)
	}
	if string(params[0]) != `"`+account.String()+`"` {
		t.Fatalf("params[0] = %s", params[0])
	}
	var opts map[string]string
	if err := json.Unmarshal(params[1], &opts); err != nil {
		t.Fatal(err)
	}
	if opts["commitment"] != "finalized" || opts["encoding"] != "base64" {
		t.Fatalf("opts = %v", opts)
	}

	for i := 0; i < 2; i++ {
		ts.push(1, "accountNotification", `{"context":{"slot":100},"value":`+wsAccountValue+`}`)
	}
	for i := 0; i < 2; i++ {
		got, err := sub.Recv(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if got.Context.Slot != 100 || got.Value.Lamports != 88849814690250 {
			t.Fatalf("notification = %+v", got)
		}
	}

	if err := sub.Unsubscribe(ctx); err != nil {
		t.Fatal(err)
	}
	if req := ts.nextReq(t); req.Method != "accountUnsubscribe" {
		t.Fatalf("method = %s", req.Method)
	}
	if _, err := sub.Recv(ctx); !errors.Is(err, ErrSubscriptionClosed) {
		t.Fatalf("Recv after unsubscribe: %v", err)
	}
}

func TestNotificationRouting(t *testing.T) {
	ts := newWSTestServer(t)
	client := testConnect(t, ts, nil)
	ctx := context.Background()

	accountSub, err := client.AccountSubscribe(ctx, solana.PublicKey{1}, "")
	if err != nil {
		t.Fatal(err)
	}
	slotSub, err := client.SlotSubscribe(ctx)
	if err != nil {
		t.Fatal(err)
	}
	ts.nextReq(t)
	ts.nextReq(t)

	// Interleave notifications for both subscriptions.
	ts.push(2, "slotNotification", `{"parent":74,"root":72,"slot":75}`)
	ts.push(1, "accountNotification", `{"context":{"slot":75},"value":`+wsAccountValue+`}`)
	ts.push(2, "slotNotification", `{"parent":75,"root":73,"slot":76}`)

	slot1, err := slotSub.Recv(ctx)
	if err != nil || slot1.Slot != 75 {
		t.Fatalf("slot1 = %+v, %v", slot1, err)
	}
	acct, err := accountSub.Recv(ctx)
	if err != nil || acct.Context.Slot != 75 {
		t.Fatalf("account = %+v, %v", acct, err)
	}
	slot2, err := slotSub.Recv(ctx)
	if err != nil || slot2.Slot != 76 {
		t.Fatalf("slot2 = %+v, %v", slot2, err)
	}
}

func TestDroppedNotifications(t *testing.T) {
	ts := newWSTestServer(t)
	client := testConnect(t, ts, &Options{SubscriptionBuffer: 2})
	ctx := context.Background()

	sub, err := client.SlotSubscribe(ctx)
	if err != nil {
		t.Fatal(err)
	}
	ts.nextReq(t)

	for i := 0; i < 5; i++ {
		ts.push(1, "slotNotification", `{"parent":1,"root":1,"slot":`+uitoa(uint64(10+i))+`}`)
	}
	// Wait for the read loop to process all frames: the drop counter
	// reaches 3 once the buffer (2) is full and 3 more were discarded.
	deadline := time.Now().Add(5 * time.Second)
	for sub.Dropped() != 3 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if got := sub.Dropped(); got != 3 {
		t.Fatalf("Dropped() = %d, want 3", got)
	}

	// The two buffered notifications are the oldest ones.
	first, err := sub.Recv(ctx)
	if err != nil || first.Slot != 10 {
		t.Fatalf("first = %+v, %v", first, err)
	}
}

func TestSubscribeError(t *testing.T) {
	ts := newWSTestServer(t)
	ts.failNext.Store(true)
	client := testConnect(t, ts, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.SlotSubscribe(ctx)
	var rpcErr *rpc.RPCError
	if !errors.As(err, &rpcErr) || rpcErr.Code != -32602 {
		t.Fatalf("err = %v", err)
	}
}

func TestConnectionLossClosesSubscriptions(t *testing.T) {
	ts := newWSTestServer(t)
	client := testConnect(t, ts, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sub, err := client.SlotSubscribe(ctx)
	if err != nil {
		t.Fatal(err)
	}
	ts.nextReq(t)
	ts.closeConn()

	if _, err := sub.Recv(ctx); err == nil {
		t.Fatal("Recv succeeded after connection loss")
	}
	if client.Err() == nil {
		t.Fatal("client.Err() is nil after connection loss")
	}
	if _, err := client.SlotSubscribe(ctx); err == nil {
		t.Fatal("subscribe succeeded on dead connection")
	}
}

func TestSignatureReceivedNotification(t *testing.T) {
	ts := newWSTestServer(t)
	client := testConnect(t, ts, nil)
	ctx := context.Background()

	sub, err := client.SignatureSubscribeWithOpts(ctx, solana.Signature{1}, rpc.CommitmentProcessed, true)
	if err != nil {
		t.Fatal(err)
	}
	ts.nextReq(t)

	ts.push(1, "signatureNotification", `{"context":{"slot":5},"value":"receivedSignature"}`)
	ts.push(1, "signatureNotification", `{"context":{"slot":6},"value":{"err":null}}`)

	received, err := sub.Recv(ctx)
	if err != nil || !received.Value.Received {
		t.Fatalf("received = %+v, %v", received, err)
	}
	processed, err := sub.Recv(ctx)
	if err != nil || processed.Value.Received || processed.Value.Err != nil {
		t.Fatalf("processed = %+v, %v", processed, err)
	}
}

func TestOtherNotificationDecodes(t *testing.T) {
	ts := newWSTestServer(t)
	client := testConnect(t, ts, nil)
	ctx := context.Background()

	logsSub, err := client.LogsSubscribeMentions(ctx, solana.TokenProgramID, rpc.CommitmentConfirmed)
	if err != nil {
		t.Fatal(err)
	}
	req := ts.nextReq(t)
	if !strings.Contains(string(req.Params), `"mentions"`) {
		t.Fatalf("params = %s", req.Params)
	}
	ts.push(1, "logsNotification", `{"context":{"slot":9},"value":{"signature":"5yUSwqQqeZLEEYKxnG4JC4XhaaBpV3RS4nQbK8bQTyjLX5btVq9A1Ja5nuJzV7Z3Zq8G6EVKFvN4DKUL6PSAxmTk","err":null,"logs":["Program log: hello"]}}`)
	logs, err := logsSub.Recv(ctx)
	if err != nil || len(logs.Value.Logs) != 1 || logs.Value.Err != nil {
		t.Fatalf("logs = %+v, %v", logs, err)
	}

	rootSub, err := client.RootSubscribe(ctx)
	if err != nil {
		t.Fatal(err)
	}
	ts.nextReq(t)
	ts.push(2, "rootNotification", `42`)
	root, err := rootSub.Recv(ctx)
	if err != nil || *root != 42 {
		t.Fatalf("root = %v, %v", root, err)
	}

	updatesSub, err := client.SlotsUpdatesSubscribe(ctx)
	if err != nil {
		t.Fatal(err)
	}
	ts.nextReq(t)
	ts.push(3, "slotsUpdatesNotification", `{"slot":100,"timestamp":1625081266243,"type":"frozen","stats":{"numTransactionEntries":10,"numSuccessfulTransactions":50,"numFailedTransactions":1,"maxTransactionsPerEntry":8}}`)
	update, err := updatesSub.Recv(ctx)
	if err != nil || update.Type != "frozen" || update.Stats.NumSuccessfulTransactions != 50 {
		t.Fatalf("update = %+v, %v", update, err)
	}
}

// TestNoNotificationLossAfterSubscribe covers the flood case: the server
// sends notifications immediately behind the subscribe ack; the read loop
// registers the channel while handling the ack, so none may be lost.
func TestNoNotificationLossAfterSubscribe(t *testing.T) {
	ts := newWSTestServer(t)
	// Prime the server to follow the next subscribe ack with an immediate
	// burst before the client's subscribe() call even returns.
	ts.burstAfterAck.Store(3)
	client := testConnect(t, ts, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sub, err := client.SlotSubscribe(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		got, err := sub.Recv(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if got.Slot != uint64(100+i) {
			t.Fatalf("slot[%d] = %+v", i, got)
		}
	}
	if sub.Dropped() != 0 {
		t.Fatalf("dropped = %d", sub.Dropped())
	}
}
