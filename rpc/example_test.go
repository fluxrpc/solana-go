package rpc_test

import (
	"context"
	"fmt"
	"strings"

	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/rpc"
)

func ExampleClient() {
	client := rpc.New("https://your-endpoint")
	client.SetHeader("X-Api-Key", "...") // if the endpoint is authenticated
	ctx := context.Background()

	usdc := solana.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")
	account, err := client.GetAccountInfo(ctx, usdc)
	if err != nil {
		panic(err) // rpc.ErrNotFound when the account does not exist
	}
	fmt.Println(account.Value.Owner, len(account.GetBinary()))
}

func ExampleClient_GetProgramAccountsStream() {
	client := rpc.New("https://your-endpoint")
	ctx := context.Background()

	// Accounts are decoded and delivered while the response body is still
	// downloading; memory stays bounded by the largest single account.
	c, err := client.GetProgramAccountsStream(ctx, solana.TokenProgramID, nil,
		func(ka *rpc.KeyedAccount) error {
			fmt.Println(ka.Pubkey, ka.Account.Lamports)
			return nil
		})
	if err != nil {
		panic(err)
	}
	fmt.Println("as of slot", c.Slot)
}

func ExampleStreamProgramAccounts() {
	// StreamProgramAccounts consumes a raw getProgramAccounts JSON-RPC
	// response body from any io.Reader.
	body := strings.NewReader(`{"jsonrpc":"2.0","result":{"context":{"slot":341197053},"value":[` +
		`{"pubkey":"SysvarC1ock11111111111111111111111111111111","account":{"lamports":1169280,"owner":"Sysvar1111111111111111111111111111111111111","data":["","base64"],"executable":false,"rentEpoch":361,"space":40}}` +
		`]},"id":1}`)

	c, err := rpc.StreamProgramAccounts(body, func(ka *rpc.KeyedAccount) error {
		fmt.Println(ka.Pubkey, ka.Account.Lamports)
		return nil
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("slot", c.Slot)
	// Output:
	// SysvarC1ock11111111111111111111111111111111 1169280
	// slot 341197053
}
