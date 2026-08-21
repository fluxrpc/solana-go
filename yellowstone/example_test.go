package yellowstone_test

import (
	"context"
	"fmt"

	"github.com/fluxrpc/solana-go/yellowstone"
	pb "github.com/fluxrpc/solana-go/yellowstone/proto"
)

func ExampleConnect() {
	ctx := context.Background()
	client, err := yellowstone.Connect(ctx, "https://your-geyser:443",
		yellowstone.WithToken("your-x-token"))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// Subscribe to all accounts owned by the SPL Token program plus slot
	// updates.
	req := yellowstone.NewRequest(pb.CommitmentLevel_CONFIRMED).
		AccountsByOwner("token", "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA").
		AllSlots("slots")

	stream, err := client.Subscribe(ctx, req)
	if err != nil {
		panic(err)
	}
	defer stream.Close()

	for {
		update, err := stream.Recv()
		if err != nil {
			panic(err)
		}
		if acct := update.GetAccount(); acct != nil {
			converted := update.Account()
			fmt.Println("account", converted.Pubkey, "lamports", converted.Lamports)
		}
	}
}
