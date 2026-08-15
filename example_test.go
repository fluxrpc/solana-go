package solana_go_test

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
)

func ExamplePublicKeyFromBase58() {
	key, err := solana.PublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
	if err != nil {
		panic(err)
	}
	fmt.Println(key)
	fmt.Println(key.IsOnCurve())
	// Output:
	// TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA
	// true
}

func ExampleFindAssociatedTokenAddress() {
	wallet := solana.MustPublicKeyFromBase58("G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo")
	usdc := solana.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")

	ata, bump, err := solana.FindAssociatedTokenAddress(wallet, usdc)
	if err != nil {
		panic(err)
	}
	fmt.Println(ata, bump)
	// Output:
	// Gkrr6Cr5bPLQhxSfJMBaU6BPuCWmwDyMpCRf5up41mXR 255
}

func ExampleTransactionFromBase64() {
	tx, err := solana.TransactionFromBase64("AfjEs3XhTc3hrxEvlnMPkm/cocvAUbFNbCl00qKnrFue6J53AhEqIFmcJJlJW3EDP5RmcMz+cNTTcZHW/WJYwAcBAAEDO8hh4VddzfcO5jbCt95jryl6y8ff65UcgukHNLWH+UQGgxCGGpgyfQVQV02EQYqm4QwzUt2qf9f1gVLM7rI4hwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA6ANIF55zOZWROWRkeh+lExxZBnKFqbvIxZDLE7EijjoBAgIAAQwCAAAAOTAAAAAAAAA=")
	if err != nil {
		panic(err)
	}
	fmt.Println(len(tx.Signatures), "signature")
	fmt.Println("fee payer:", tx.Message.AccountKeys[0])
	fmt.Println("blockhash:", tx.Message.RecentBlockhash)
	// Output:
	// 1 signature
	// fee payer: 52NGrUqh6tSGhr59ajGxsH3VnAaoRdSdTbAaV9G3UW35
	// blockhash: GcgVK9buRA7YepZh3zXuS399GJAESCisLnLDBCmR5Aoj
}

func ExampleTransaction_Sign() {
	payer, err := solana.NewRandomPrivateKey()
	if err != nil {
		panic(err)
	}

	tx := &solana.Transaction{
		Message: solana.Message{
			AccountKeys: []solana.PublicKey{payer.PublicKey(), {}, solana.SystemProgramID},
			Header:      solana.MessageHeader{NumRequiredSignatures: 1, NumReadonlyUnsignedAccounts: 1},
			Instructions: []solana.CompiledInstruction{{
				ProgramIDIndex: 2,
				Accounts:       []uint16{0, 1},
				Data:           solana.Base58{2, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0},
			}},
			RecentBlockhash: solana.Hash{}, // use a real recent blockhash
		},
	}

	if _, err := tx.Sign(func(pub solana.PublicKey) *solana.PrivateKey {
		if pub == payer.PublicKey() {
			return &payer
		}
		return nil
	}); err != nil {
		panic(err)
	}

	wire, err := tx.MarshalBinary()
	if err != nil {
		panic(err)
	}
	_ = wire // send via rpc.Client.SendRawTransactionWithOpts
}
