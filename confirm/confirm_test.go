package confirm

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"context"

	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/rpc"
	fluxws "github.com/fluxrpc/solana-go/ws"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// signedTestTx builds a minimal signed transaction.
func signedTestTx(t *testing.T) *solana.Transaction {
	t.Helper()
	key, err := solana.NewRandomPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	pub := key.PublicKey()
	tx, err := solana.NewTransaction([]solana.Instruction{
		solana.NewInstruction(solana.MemoProgramID, solana.AccountMetaSlice{
			{PublicKey: pub, IsSigner: true, IsWritable: true},
		}, []byte("hi")),
	}, solana.Hash{1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Sign(func(k solana.PublicKey) *solana.PrivateKey {
		if k == pub {
			return &key
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return tx
}

// mockRPC serves sendTransaction and getSignatureStatuses. statusResponses
// is consumed one JSON value per getSignatureStatuses call; the last entry
// repeats.
type mockRPC struct {
	sendCalls   atomic.Int64
	statusCalls atomic.Int64
	// onSend is invoked while handling sendTransaction.
	onSend          func()
	statusResponses []string
	server          *httptest.Server
}

func newMockRPC(t *testing.T, statusResponses ...string) *mockRPC {
	m := &mockRPC{statusResponses: statusResponses}
	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     uint64            `json:"id"`
			Method string            `json:"method"`
			Params []json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("bad rpc request: %v", err)
			return
		}
		switch req.Method {
		case "sendTransaction":
			m.sendCalls.Add(1)
			if m.onSend != nil {
				m.onSend()
			}
			// Echo any signature; the client only decodes it.
			var b64 string
			_ = json.Unmarshal(req.Params[0], &b64)
			tx, err := solana.TransactionFromBase64(b64)
			if err != nil {
				t.Errorf("undecodable sent transaction: %v", err)
				return
			}
			fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%d,"result":"%s"}`, req.ID, tx.Signatures[0])
		case "getSignatureStatuses":
			n := m.statusCalls.Add(1)
			idx := int(n) - 1
			if idx >= len(m.statusResponses) {
				idx = len(m.statusResponses) - 1
			}
			fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%d,"result":{"context":{"slot":1},"value":[%s]}}`, req.ID, m.statusResponses[idx])
		default:
			t.Errorf("unexpected rpc method %s", req.Method)
		}
	}))
	t.Cleanup(m.server.Close)
	return m
}

// mockWS acks the first subscription and pushes the given notification
// value after pushSignal fires (or immediately if pushSignal is nil).
type mockWS struct {
	url           string
	subscribed    atomic.Bool
	subscribeSeen chan struct{}
}

func newMockWS(t *testing.T, notificationValue string, pushSignal chan struct{}) *mockWS {
	m := &mockWS{subscribeSeen: make(chan struct{}, 1)}
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
			ID     uint64 `json:"id"`
			Method string `json:"method"`
		}
		if err := json.Unmarshal(data, &req); err != nil || req.Method != "signatureSubscribe" {
			t.Errorf("unexpected ws request: %s", data)
			return
		}
		m.subscribed.Store(true)
		m.subscribeSeen <- struct{}{}
		if err := wsutil.WriteServerText(conn, fmt.Appendf(nil, `{"jsonrpc":"2.0","id":%d,"result":1}`, req.ID)); err != nil {
			return
		}
		if notificationValue != "" {
			if pushSignal != nil {
				<-pushSignal
			}
			notification := `{"jsonrpc":"2.0","method":"signatureNotification","params":{"result":` + notificationValue + `,"subscription":1}}`
			if err := wsutil.WriteServerText(conn, []byte(notification)); err != nil {
				return
			}
		}
		// Serve the unsubscribe (or wait for hangup).
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
	t.Cleanup(server.Close)
	m.url = "ws" + strings.TrimPrefix(server.URL, "http")
	return m
}

func TestSendAndConfirmSubscribesBeforeSend(t *testing.T) {
	tx := signedTestTx(t)
	pushSignal := make(chan struct{})
	wsMock := newMockWS(t, `{"context":{"slot":5},"value":{"err":null}}`, pushSignal)
	rpcMock := newMockRPC(t)
	rpcMock.onSend = func() {
		if !wsMock.subscribed.Load() {
			t.Error("sendTransaction reached the RPC before the signature subscription was armed")
		}
		close(pushSignal)
	}

	ctx := context.Background()
	wsClient, err := fluxws.Connect(ctx, wsMock.url)
	if err != nil {
		t.Fatal(err)
	}
	defer wsClient.Close()

	sig, err := SendAndConfirm(ctx, rpc.New(rpcMock.server.URL), wsClient, tx)
	if err != nil {
		t.Fatal(err)
	}
	if sig != tx.Signatures[0] {
		t.Fatalf("sig = %s, want %s", sig, tx.Signatures[0])
	}
	if rpcMock.sendCalls.Load() != 1 {
		t.Fatalf("sendTransaction calls = %d", rpcMock.sendCalls.Load())
	}
}

func TestSendAndConfirmExecutionError(t *testing.T) {
	tx := signedTestTx(t)
	wsMock := newMockWS(t, `{"context":{"slot":5},"value":{"err":{"InstructionError":[0,{"Custom":42}]}}}`, nil)
	rpcMock := newMockRPC(t)

	ctx := context.Background()
	wsClient, err := fluxws.Connect(ctx, wsMock.url)
	if err != nil {
		t.Fatal(err)
	}
	defer wsClient.Close()

	_, err = SendAndConfirm(ctx, rpc.New(rpcMock.server.URL), wsClient, tx)
	var execErr *ExecutionError
	if !errors.As(err, &execErr) {
		t.Fatalf("err = %v, want *ExecutionError", err)
	}
	if !strings.Contains(execErr.Error(), "InstructionError") {
		t.Fatalf("err = %v", execErr)
	}
}

func TestSendAndConfirmTimeout(t *testing.T) {
	tx := signedTestTx(t)
	wsMock := newMockWS(t, "", nil) // never notifies
	rpcMock := newMockRPC(t)

	ctx := context.Background()
	wsClient, err := fluxws.Connect(ctx, wsMock.url)
	if err != nil {
		t.Fatal(err)
	}
	defer wsClient.Close()

	_, err = SendAndConfirmWithOpts(ctx, rpc.New(rpcMock.server.URL), wsClient, tx, Opts{Timeout: 150 * time.Millisecond})
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("err = %v, want ErrTimeout", err)
	}
}

func TestSendAndConfirmPolling(t *testing.T) {
	tx := signedTestTx(t)
	rpcMock := newMockRPC(t,
		`null`,
		`{"slot":5,"confirmations":3,"err":null,"confirmationStatus":"confirmed","status":{"Ok":null}}`,
		`{"slot":5,"confirmations":null,"err":null,"confirmationStatus":"finalized","status":{"Ok":null}}`,
	)

	sig, err := SendAndConfirmWithOpts(context.Background(), rpc.New(rpcMock.server.URL), nil, tx, Opts{
		PollInterval: 5 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	if sig != tx.Signatures[0] {
		t.Fatalf("sig = %s", sig)
	}
	// null, confirmed (not yet finalized), finalized.
	if got := rpcMock.statusCalls.Load(); got != 3 {
		t.Fatalf("status calls = %d, want 3", got)
	}
}

func TestSendAndConfirmPollingLowerCommitment(t *testing.T) {
	tx := signedTestTx(t)
	rpcMock := newMockRPC(t,
		`{"slot":5,"confirmations":3,"err":null,"confirmationStatus":"confirmed","status":{"Ok":null}}`,
	)
	_, err := SendAndConfirmWithOpts(context.Background(), rpc.New(rpcMock.server.URL), nil, tx, Opts{
		Commitment:   rpc.CommitmentConfirmed,
		PollInterval: 5 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := rpcMock.statusCalls.Load(); got != 1 {
		t.Fatalf("status calls = %d, want 1", got)
	}
}

func TestSendAndConfirmPollingExecutionError(t *testing.T) {
	tx := signedTestTx(t)
	rpcMock := newMockRPC(t,
		`{"slot":5,"confirmations":null,"err":{"InstructionError":[0,"InvalidArgument"]},"confirmationStatus":"finalized","status":{"Err":{}}}`,
	)
	_, err := SendAndConfirmWithOpts(context.Background(), rpc.New(rpcMock.server.URL), nil, tx, Opts{
		PollInterval: 5 * time.Millisecond,
	})
	var execErr *ExecutionError
	if !errors.As(err, &execErr) {
		t.Fatalf("err = %v, want *ExecutionError", err)
	}
}

func TestSendAndConfirmPollingTimeout(t *testing.T) {
	tx := signedTestTx(t)
	rpcMock := newMockRPC(t, `null`)
	_, err := SendAndConfirmWithOpts(context.Background(), rpc.New(rpcMock.server.URL), nil, tx, Opts{
		Timeout:      100 * time.Millisecond,
		PollInterval: 10 * time.Millisecond,
	})
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("err = %v, want ErrTimeout", err)
	}
}

func TestSendAndConfirmUnsigned(t *testing.T) {
	if _, err := SendAndConfirm(context.Background(), nil, nil, &solana.Transaction{}); err == nil {
		t.Fatal("expected error for unsigned transaction")
	}
}

func TestStatusReaches(t *testing.T) {
	confirmations := uint64(3)
	cases := []struct {
		status rpc.ConfirmationStatusType
		conf   *uint64
		target rpc.CommitmentType
		want   bool
	}{
		{rpc.ConfirmationStatusProcessed, &confirmations, rpc.CommitmentProcessed, true},
		{rpc.ConfirmationStatusProcessed, &confirmations, rpc.CommitmentConfirmed, false},
		{rpc.ConfirmationStatusConfirmed, &confirmations, rpc.CommitmentConfirmed, true},
		{rpc.ConfirmationStatusConfirmed, &confirmations, rpc.CommitmentFinalized, false},
		{rpc.ConfirmationStatusFinalized, nil, rpc.CommitmentFinalized, true},
		{rpc.ConfirmationStatusFinalized, nil, "", true},
		{"", nil, "", true},             // rooted, no confirmationStatus field
		{"", &confirmations, "", false}, // not rooted, no confirmationStatus field
	}
	for i, tc := range cases {
		status := &rpc.SignatureStatusesResult{ConfirmationStatus: tc.status, Confirmations: tc.conf}
		if got := statusReaches(status, tc.target); got != tc.want {
			t.Errorf("case %d: statusReaches = %v, want %v", i, got, tc.want)
		}
	}
}
