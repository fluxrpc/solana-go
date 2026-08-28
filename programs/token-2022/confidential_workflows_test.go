package token2022

import (
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

func TestConfidentialAccountWorkflows(t *testing.T) {
	service := ConfidentialTransferService{}
	service.Start()
	keypair, err := service.ElGamalKeypairFromSecret(ElGamalSecretKey{1})
	if err != nil {
		t.Fatal(err)
	}
	configure, err := service.ConfigureConfidentialTransferAccountWithKeys(token2022Key(1), token2022Key(2), token2022Key(3), nil, keypair, AEKey{4}, 65_536)
	if err != nil {
		t.Fatal(err)
	}
	if len(configure) != 2 || configure[0].ProgramID() != ProgramID || configure[1].ProgramID() != service.ProofProgramID {
		t.Fatalf("configure = %#v", configure)
	}
	zero, err := service.EncryptElGamalWithOpening(keypair.PublicKey, 0, PedersenOpening{2})
	if err != nil {
		t.Fatal(err)
	}
	empty, err := service.EmptyConfidentialTransferAccountWithProof(token2022Key(1), token2022Key(3), nil, keypair, zero)
	if err != nil {
		t.Fatal(err)
	}
	if len(empty) != 2 || empty[1].ProgramID() != service.ProofProgramID {
		t.Fatalf("empty = %#v", empty)
	}
	registry, err := service.CreateElGamalRegistryWithKeypair(token2022Key(3), keypair)
	if err != nil {
		t.Fatal(err)
	}
	if len(registry) != 2 || registry[0].ProgramID() != service.RegistryProgramID {
		t.Fatalf("registry = %#v", registry)
	}
	funded, err := service.CreateFundedElGamalRegistryWithKeypair(token2022Key(4), token2022Key(3), 123, keypair)
	if err != nil {
		t.Fatal(err)
	}
	if len(funded) != 3 || funded[0].ProgramID() != solana.SystemProgramID {
		t.Fatalf("funded registry = %#v", funded)
	}
}
