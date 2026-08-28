package token2022

import (
	"errors"
	"math"
	"sync"
	"testing"
)

type confidentialProofFixture struct {
	service     ConfidentialTransferService
	source      ElGamalKeypair
	destination ElGamalKeypair
	auditor     ElGamalKeypair
	withheld    ElGamalKeypair
	key         AEKey
	account     ConfidentialTransferAccount
}

func (fixture *confidentialProofFixture) start(test testing.TB, balance uint64) {
	test.Helper()
	fixture.service.Start()
	keypairs := []*ElGamalKeypair{&fixture.source, &fixture.destination, &fixture.auditor, &fixture.withheld}
	for index, keypair := range keypairs {
		var err error
		*keypair, err = fixture.service.ElGamalKeypairFromSecret(ElGamalSecretKey{byte(index + 1)})
		if err != nil {
			test.Fatal(err)
		}
	}
	fixture.key = AEKey{5}
	available, err := fixture.service.EncryptElGamalWithOpening(fixture.source.PublicKey, balance, PedersenOpening{6})
	if err != nil {
		test.Fatal(err)
	}
	decryptable, err := fixture.service.EncryptAEAmountWithNonce(fixture.key, balance, [12]byte{7})
	if err != nil {
		test.Fatal(err)
	}
	fixture.account = ConfidentialTransferAccount{ElGamalPubkey: fixture.source.PublicKey, AvailableBalance: available, DecryptableAvailableBalance: decryptable}
}

func (fixture confidentialProofFixture) verifyAmountProofs(proofs ConfidentialAmountProofBundle) error {
	for _, proof := range []ZKProofData{proofs.Equality, proofs.CiphertextValidity, proofs.Range} {
		if err := fixture.service.VerifyZKProofData(proof); err != nil {
			return err
		}
	}
	return nil
}

type confidentialProofConcurrencyTest struct {
	fixture confidentialProofFixture
	wait    sync.WaitGroup
	errors  chan error
}

func (test *confidentialProofConcurrencyTest) generate(amount uint64) {
	defer test.wait.Done()
	proofs, _, err := test.fixture.service.GenerateConfidentialTransferProofs(test.fixture.account, amount, test.fixture.source, test.fixture.key, test.fixture.destination.PublicKey, &test.fixture.auditor.PublicKey)
	if err == nil {
		err = test.fixture.verifyAmountProofs(proofs)
	}
	test.errors <- err
}

func TestConfidentialTransferProofBoundaries(t *testing.T) {
	fixture := confidentialProofFixture{}
	fixture.start(t, uint64(1)<<48-1)
	for _, amount := range []uint64{0, 1<<16 - 1, 1 << 16, 1<<48 - 1} {
		proofs, remaining, err := fixture.service.GenerateConfidentialTransferProofs(fixture.account, amount, fixture.source, fixture.key, fixture.destination.PublicKey, &fixture.auditor.PublicKey)
		if err != nil {
			t.Fatalf("amount %d: %v", amount, err)
		}
		if remaining != uint64(1)<<48-1-amount {
			t.Fatalf("amount %d: remaining = %d", amount, remaining)
		}
		if err := fixture.verifyAmountProofs(proofs); err != nil {
			t.Fatalf("amount %d: %v", amount, err)
		}
	}
}

func TestConfidentialTransferProofErrors(t *testing.T) {
	fixture := confidentialProofFixture{}
	fixture.start(t, 1)
	_, _, err := fixture.service.GenerateConfidentialTransferProofs(fixture.account, 2, fixture.source, fixture.key, fixture.destination.PublicKey, nil)
	classified := &ConfidentialTransferError{}
	if !errors.As(err, &classified) || !classified.InsufficientFunds() {
		t.Fatalf("insufficient funds error = %v", err)
	}
	_, _, err = fixture.service.GenerateConfidentialTransferProofs(fixture.account, uint64(1)<<48, fixture.source, fixture.key, fixture.destination.PublicKey, nil)
	if !errors.As(err, &classified) || !classified.InvalidInput() {
		t.Fatalf("invalid amount error = %v", err)
	}
	_, _, err = fixture.service.CalculateConfidentialTransferFee(1, 10_001, 1)
	if !errors.As(err, &classified) || !classified.InvalidInput() {
		t.Fatalf("invalid fee error = %v", err)
	}
	if err := fixture.service.VerifyZKProofData(ZKProofData{}); !errors.As(err, &classified) || !classified.InvalidProof() {
		t.Fatalf("invalid proof error = %v", err)
	}
	if err := fixture.service.VerifyPubkeyValidityProof(ZKProofData{}); !errors.As(err, &classified) || !classified.InvalidProof() {
		t.Fatalf("specific invalid proof error = %v", err)
	}
	ciphertext, err := fixture.service.EncryptAEAmountWithNonce(fixture.key, 1, [12]byte{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = fixture.service.DecryptAEAmount(AEKey{9}, ciphertext)
	if !errors.As(err, &classified) || !classified.Decryption() {
		t.Fatalf("decryption error = %v", err)
	}
}

func TestConfidentialTransferFeeProofBoundaries(t *testing.T) {
	fixture := confidentialProofFixture{}
	fixture.start(t, uint64(1)<<48-1)
	tests := []struct {
		amount  uint64
		rate    uint16
		maximum uint64
		fee     uint64
	}{
		{amount: 0, rate: 0, maximum: 0, fee: 0},
		{amount: 100, rate: 5, maximum: 10, fee: 1},
		{amount: 65_535, rate: 5, maximum: 10, fee: 10},
		{amount: 65_535, rate: 5, maximum: 1, fee: 1},
		{amount: 65_536, rate: 5, maximum: 10, fee: 10},
		{amount: 65_536, rate: 5, maximum: 1, fee: 1},
		{amount: 1<<48 - 1, rate: 5, maximum: 10, fee: 10},
		{amount: 1<<48 - 1, rate: 5, maximum: 1, fee: 1},
		{amount: 10_001, rate: 100, maximum: math.MaxUint64, fee: 101},
	}
	for _, test := range tests {
		proofs, _, fee, err := fixture.service.GenerateConfidentialTransferWithFeeProofs(fixture.account, test.amount, fixture.source, fixture.key, fixture.destination.PublicKey, fixture.withheld.PublicKey, &fixture.auditor.PublicKey, test.rate, test.maximum)
		if err != nil {
			t.Fatal(err)
		}
		if fee != test.fee {
			t.Fatalf("maximum fee %d: fee = %d, want %d", test.maximum, fee, test.fee)
		}
		for _, proof := range []ZKProofData{proofs.Equality, proofs.TransferCiphertextValidity, proofs.FeeSigma, proofs.FeeCiphertextValidity, proofs.Range} {
			if err := fixture.service.VerifyZKProofData(proof); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestConfidentialTransferProofsConcurrent(t *testing.T) {
	test := confidentialProofConcurrencyTest{errors: make(chan error, 8)}
	test.fixture.start(t, 1_000_000)
	test.wait.Add(cap(test.errors))
	for index := 0; index < cap(test.errors); index++ {
		go test.generate(uint64(index + 1))
	}
	test.wait.Wait()
	close(test.errors)
	for err := range test.errors {
		if err != nil {
			t.Fatalf("concurrent proof generation: %v", err)
		}
	}
}

func BenchmarkGenerateConfidentialTransferProofs(b *testing.B) {
	fixture := confidentialProofFixture{}
	fixture.start(b, 1_000_000)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if _, _, err := fixture.service.GenerateConfidentialTransferProofs(fixture.account, 10_001, fixture.source, fixture.key, fixture.destination.PublicKey, &fixture.auditor.PublicKey); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConfidentialTransferServiceStart(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		service := ConfidentialTransferService{}
		service.Start()
	}
}

func BenchmarkGenerateConfidentialTransferWithFeeProofs(b *testing.B) {
	fixture := confidentialProofFixture{}
	fixture.start(b, 1_000_000)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if _, _, _, err := fixture.service.GenerateConfidentialTransferWithFeeProofs(fixture.account, 10_001, fixture.source, fixture.key, fixture.destination.PublicKey, fixture.withheld.PublicKey, &fixture.auditor.PublicKey, 100, 50); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVerifyConfidentialTransferProofs(b *testing.B) {
	fixture := confidentialProofFixture{}
	fixture.start(b, 1_000_000)
	proofs, _, err := fixture.service.GenerateConfidentialTransferProofs(fixture.account, 10_001, fixture.source, fixture.key, fixture.destination.PublicKey, &fixture.auditor.PublicKey)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err := fixture.verifyAmountProofs(proofs); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVerifyConfidentialTransferWithFeeProofs(b *testing.B) {
	fixture := confidentialProofFixture{}
	fixture.start(b, 1_000_000)
	proofs, _, _, err := fixture.service.GenerateConfidentialTransferWithFeeProofs(fixture.account, 10_001, fixture.source, fixture.key, fixture.destination.PublicKey, fixture.withheld.PublicKey, &fixture.auditor.PublicKey, 100, 50)
	if err != nil {
		b.Fatal(err)
	}
	data := []ZKProofData{proofs.Equality, proofs.TransferCiphertextValidity, proofs.FeeSigma, proofs.FeeCiphertextValidity, proofs.Range}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		for _, proof := range data {
			if err := fixture.service.VerifyZKProofData(proof); err != nil {
				b.Fatal(err)
			}
		}
	}
}
