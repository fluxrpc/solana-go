package ws_test

import (
	"context"
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/rpc"
	"github.com/fluxrpc/solana-go/ws"
)

func ExampleConnect() {
	ctx := context.Background()
	client, err := ws.Connect(ctx, "wss://your-endpoint")
	if err != nil {
		panic(err)
	}
	defer client.Close()

	usdc := solana.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")
	sub, err := client.AccountSubscribe(ctx, usdc, rpc.CommitmentConfirmed)
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe(ctx)

	for {
		update, err := sub.Recv(ctx)
		if err != nil {
			panic(err) // wraps ws.ErrSubscriptionClosed on connection loss
		}
		fmt.Println("slot", update.Context.Slot, "lamports", update.Value.Lamports)
	}
}
