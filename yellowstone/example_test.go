package yellowstone_test

import (
	"context"
	"fmt"

	"github.com/fluxrpc/solana-go/yellowstone"
	pb "github.com/rpcpool/yellowstone-grpc/examples/golang/proto"
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
	req := yellowstone.NewRequest(pb.CommitmentLevel_CONFIRMED)
	yellowstone.AddAccounts(req, "token", yellowstone.AccountsByOwner("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"))
	yellowstone.AddSlots(req, "slots", yellowstone.Slots())

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
			converted := yellowstone.ConvertAccount(acct)
			fmt.Println("account", converted.Pubkey, "lamports", converted.Lamports)
		}
	}
}
