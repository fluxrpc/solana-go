package ws

import (
	"context"
	"errors"
	"testing"

	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/rpc"
)

func TestSubscriptionRelease(t *testing.T) {
	ts := newWSTestServer(t)
	client := testConnect(t, ts, nil)
	ctx := context.Background()

	sig := solana.Signature{1}
	sub, err := client.SignatureSubscribe(ctx, sig, rpc.CommitmentFinalized)
	if err != nil {
		t.Fatal(err)
	}
	ts.nextReq(t)

	// Deliver the single-shot notification, then release locally: no
	// unsubscribe request may reach the server.
	ts.push(1, "signatureNotification", `{"context":{"slot":5},"value":{"err":null}}`)
	if _, err := sub.Recv(ctx); err != nil {
		t.Fatal(err)
	}
	sub.Release()
	sub.Release() // idempotent

	if _, err := sub.Recv(ctx); !errors.Is(err, ErrSubscriptionClosed) {
		t.Fatalf("Recv after Release: %v", err)
	}

	// A later subscription must be able to reuse resources normally, and
	// releasing the old subscription again must not affect it. The mock
	// server hands out increasing sub ids, so force the old id by releasing
	// after the new subscription exists.
	sub2, err := client.SlotSubscribe(ctx)
	if err != nil {
		t.Fatal(err)
	}
	ts.nextReq(t)
	sub.Release() // stale release, different id: no effect on sub2

	ts.push(2, "slotNotification", `{"parent":1,"root":1,"slot":42}`)
	got, err := sub2.Recv(ctx)
	if err != nil || got.Slot != 42 {
		t.Fatalf("sub2.Recv = %+v, %v", got, err)
	}

	// No unsubscribe request was ever sent.
	select {
	case req := <-ts.reqs:
		t.Fatalf("unexpected request after Release: %s", req.Method)
	default:
	}
}
