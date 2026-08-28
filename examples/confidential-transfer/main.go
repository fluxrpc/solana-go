package main

import (
	"fmt"
	"log"

	solana "github.com/fluxrpc/solana-go"
	token2022 "github.com/fluxrpc/solana-go/programs/token-2022"
)

func main() {
	service := token2022.ConfidentialTransferService{}
	service.Start()

	authority := solana.NewWallet()
	mint := solana.NewWallet().PublicKey()
	source := solana.NewWallet().PublicKey()
	destination := solana.NewWallet().PublicKey()
	seed := service.PDAWalletPublicSeed(token2022.ProgramID, authority.PublicKey(), mint, source)
	keys, err := service.DeriveConfidentialKeys(authority.PrivateKey, seed[:])
	if err != nil {
		log.Fatal(err)
	}
	sourceKeypair, err := service.ElGamalKeypairFromSecret(keys.ElGamalSecretKey)
	if err != nil {
		log.Fatal(err)
	}
	destinationKeypair, err := service.GenerateElGamalKeypair()
	if err != nil {
		log.Fatal(err)
	}

	initialBalance := uint64(1_000_000)
	transferAmount := uint64(10_000)
	availableBalance, _, err := service.EncryptElGamal(sourceKeypair.PublicKey, initialBalance)
	if err != nil {
		log.Fatal(err)
	}
	decryptableBalance, err := service.EncryptAEAmount(keys.AEKey, initialBalance)
	if err != nil {
		log.Fatal(err)
	}
	account := token2022.ConfidentialTransferAccount{
		Approved:                    true,
		ElGamalPubkey:               sourceKeypair.PublicKey,
		AvailableBalance:            availableBalance,
		DecryptableAvailableBalance: decryptableBalance,
	}

	plan, err := service.ConfidentialTransferWithProofs(
		source,
		mint,
		destination,
		authority.PublicKey(),
		nil,
		transferAmount,
		account,
		sourceKeypair,
		keys.AEKey,
		destinationKeypair.PublicKey,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	proofs := []token2022.ZKProofData{plan.Proofs.Equality, plan.Proofs.CiphertextValidity, plan.Proofs.Range}
	for _, proof := range proofs {
		if err := service.VerifyZKProofData(proof); err != nil {
			log.Fatal(err)
		}
	}
	remainingBalance, err := service.DecryptAEAmount(keys.AEKey, plan.NewDecryptableAvailableBalance)
	if err != nil {
		log.Fatal(err)
	}
	computeUnitLimit := uint32(1_400_000)
	recentBlockhash := solana.Hash{1}
	transaction, err := solana.NewTransactionV1(
		plan.Instructions,
		recentBlockhash,
		solana.TransactionConfig{ComputeUnitLimit: &computeUnitLimit},
		solana.TransactionPayer(authority.PublicKey()),
	)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := transaction.Sign(authority.PrivateKeyFor); err != nil {
		log.Fatal(err)
	}
	raw, err := transaction.MarshalBinary()
	if err != nil {
		log.Fatal(err)
	}
	encoded := solana.Base64(raw).String()

	fmt.Printf("prepared %d instructions for a confidential transfer of %d tokens\n", len(plan.Instructions), transferAmount)
	fmt.Printf("source balance: %d -> %d\n", initialBalance, remainingBalance)
	fmt.Printf("signed v1 transaction: %d wire bytes, %d base64 bytes\n", len(raw), len(encoded))
	for index, instruction := range plan.Instructions {
		data, err := instruction.Data()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("instruction %d: program=%s data=%d bytes\n", index, instruction.ProgramID(), len(data))
	}
}
