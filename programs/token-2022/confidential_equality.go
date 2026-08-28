package token2022

import (
	"crypto/rand"
	"fmt"
	"io"

	"github.com/gtank/merlin"
	"github.com/oasisprotocol/curve25519-voi/curve"
	"github.com/oasisprotocol/curve25519-voi/curve/scalar"
)

type CiphertextCiphertextEqualityProof [224]byte

type CiphertextCommitmentEqualityProof [192]byte

func (service ConfidentialTransferService) GenerateCiphertextCiphertextEqualityProof(firstKeypair ElGamalKeypair, secondPubkey ElGamalPubkey, firstCiphertext, secondCiphertext ElGamalCiphertext, secondOpening PedersenOpening, amount uint64) (ZKProofData, error) {
	return service.generateCiphertextCiphertextEqualityProofWithReader(firstKeypair, secondPubkey, firstCiphertext, secondCiphertext, secondOpening, amount, rand.Reader)
}

func (service ConfidentialTransferService) generateCiphertextCiphertextEqualityProofWithReader(firstKeypair ElGamalKeypair, secondPubkey ElGamalPubkey, firstCiphertext, secondCiphertext ElGamalCiphertext, secondOpening PedersenOpening, amount uint64, random io.Reader) (ZKProofData, error) {
	secret, firstPubkey, err := service.elGamalKeypairValues(firstKeypair)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate ciphertext equality proof: %w", err)
	}
	secondPubkeyPoint, err := service.ristrettoPoint(secondPubkey[:], "second ElGamal public key")
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate ciphertext equality proof: %w", err)
	}
	firstCommitment, firstHandle, err := service.elGamalCiphertextPoints(firstCiphertext)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate ciphertext equality proof: %w", err)
	}
	if !service.elGamalDecryptionMatchesAmount(secret, firstCommitment, firstHandle, amount) {
		return ZKProofData{}, fmt.Errorf("generate ciphertext equality proof: first ciphertext amount mismatch")
	}
	expectedSecondCiphertext, err := service.EncryptElGamalWithOpening(secondPubkey, amount, secondOpening)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate ciphertext equality proof: %w", err)
	}
	if expectedSecondCiphertext != secondCiphertext {
		return ZKProofData{}, fmt.Errorf("generate ciphertext equality proof: second ciphertext amount mismatch")
	}
	secondOpeningScalar, err := service.scalarFromBytes(secondOpening[:], "second Pedersen opening")
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate ciphertext equality proof: %w", err)
	}
	secretNonce, amountNonce, openingNonce, err := service.equalityProofNonces(random)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate ciphertext equality proof: %w", err)
	}
	openingBasepoint, err := service.pedersenOpeningBasepoint()
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate ciphertext equality proof: %w", err)
	}
	commitments := [4]PedersenCommitment{
		service.pedersenCommitment(curve.NewRistrettoPoint().Mul(firstPubkey, secretNonce)),
		service.pedersenCommitment(curve.NewRistrettoPoint().MultiscalarMul(
			[]*scalar.Scalar{amountNonce, secretNonce},
			[]*curve.RistrettoPoint{curve.RISTRETTO_BASEPOINT_POINT, firstHandle},
		)),
		service.pedersenCommitment(curve.NewRistrettoPoint().MultiscalarMul(
			[]*scalar.Scalar{amountNonce, openingNonce},
			[]*curve.RistrettoPoint{curve.RISTRETTO_BASEPOINT_POINT, openingBasepoint},
		)),
		service.pedersenCommitment(curve.NewRistrettoPoint().Mul(secondPubkeyPoint, openingNonce)),
	}
	context := make([]byte, 192)
	copy(context[:32], firstKeypair.PublicKey[:])
	copy(context[32:64], secondPubkey[:])
	copy(context[64:128], firstCiphertext[:])
	copy(context[128:], secondCiphertext[:])
	transcript := service.ciphertextCiphertextEqualityTranscript(context)
	for index := range commitments {
		service.appendTranscriptPoint(transcript, service.equalityCommitmentLabel(index), commitments[index][:])
	}
	challenge, err := service.transcriptChallengeScalar(transcript, "c")
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate ciphertext equality proof: %w", err)
	}
	responses := [3]*scalar.Scalar{
		scalar.New().Add(scalar.New().Mul(challenge, secret), secretNonce),
		scalar.New().Add(scalar.New().Mul(challenge, scalar.New().SetUint64(amount)), amountNonce),
		scalar.New().Add(scalar.New().Mul(challenge, secondOpeningScalar), openingNonce),
	}
	proof, err := service.ciphertextCiphertextEqualityProofBytes(commitments, responses, transcript)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate ciphertext equality proof: %w", err)
	}
	return ZKProofData{Discriminator: 2, Context: context, Proof: proof[:]}, nil
}

func (service ConfidentialTransferService) VerifyCiphertextCiphertextEqualityProof(data ZKProofData) (err error) {
	defer service.classifyInvalidProof(&err)
	if data.Discriminator != 2 || len(data.Context) != 192 || len(data.Proof) != 224 {
		return fmt.Errorf("verify ciphertext equality proof: invalid proof data")
	}
	firstPubkey, err := service.ristrettoPoint(data.Context[:32], "first ElGamal public key")
	if err != nil {
		return fmt.Errorf("verify ciphertext equality proof: %w", err)
	}
	secondPubkey, err := service.ristrettoPoint(data.Context[32:64], "second ElGamal public key")
	if err != nil {
		return fmt.Errorf("verify ciphertext equality proof: %w", err)
	}
	firstCommitment, err := service.ristrettoPoint(data.Context[64:96], "first ciphertext commitment")
	if err != nil {
		return fmt.Errorf("verify ciphertext equality proof: %w", err)
	}
	firstHandle, err := service.ristrettoPoint(data.Context[96:128], "first ciphertext handle")
	if err != nil {
		return fmt.Errorf("verify ciphertext equality proof: %w", err)
	}
	secondCommitment, err := service.ristrettoPoint(data.Context[128:160], "second ciphertext commitment")
	if err != nil {
		return fmt.Errorf("verify ciphertext equality proof: %w", err)
	}
	secondHandle, err := service.ristrettoPoint(data.Context[160:], "second ciphertext handle")
	if err != nil {
		return fmt.Errorf("verify ciphertext equality proof: %w", err)
	}
	if firstPubkey.IsIdentity() || secondPubkey.IsIdentity() || firstCommitment.IsIdentity() || firstHandle.IsIdentity() {
		return fmt.Errorf("verify ciphertext equality proof: identity context point")
	}
	commitments, responses, err := service.parseCiphertextCiphertextEqualityProof(data.Proof)
	if err != nil {
		return fmt.Errorf("verify ciphertext equality proof: %w", err)
	}
	transcript := service.ciphertextCiphertextEqualityTranscript(data.Context)
	for index := range commitments {
		service.appendTranscriptPoint(transcript, service.equalityCommitmentLabel(index), data.Proof[index*32:(index+1)*32])
	}
	challenge, err := service.transcriptChallengeScalar(transcript, "c")
	if err != nil {
		return fmt.Errorf("verify ciphertext equality proof: %w", err)
	}
	batching, err := service.appendEqualityResponses(transcript, responses)
	if err != nil {
		return fmt.Errorf("verify ciphertext equality proof: %w", err)
	}
	openingBasepoint, err := service.pedersenOpeningBasepoint()
	if err != nil {
		return fmt.Errorf("verify ciphertext equality proof: %w", err)
	}
	batchingSquared := scalar.New().Mul(batching, batching)
	batchingCubed := scalar.New().Mul(batching, batchingSquared)
	negativeChallenge := scalar.New().Neg(challenge)
	negativeOne := scalar.New().Neg(scalar.New().One())
	negativeBatching := scalar.New().Neg(batching)
	negativeBatchingSquared := scalar.New().Neg(batchingSquared)
	negativeBatchingCubed := scalar.New().Neg(batchingCubed)
	check := curve.NewRistrettoPoint().MultiscalarMul(
		[]*scalar.Scalar{
			responses[0], negativeChallenge, negativeOne,
			scalar.New().Mul(batching, responses[1]), scalar.New().Mul(batching, responses[0]), scalar.New().Mul(negativeBatching, challenge), negativeBatching,
			scalar.New().Mul(batchingSquared, responses[1]), scalar.New().Mul(batchingSquared, responses[2]), scalar.New().Mul(negativeBatchingSquared, challenge), negativeBatchingSquared,
			scalar.New().Mul(batchingCubed, responses[2]), scalar.New().Mul(negativeBatchingCubed, challenge), negativeBatchingCubed,
		},
		[]*curve.RistrettoPoint{
			firstPubkey, openingBasepoint, commitments[0],
			curve.RISTRETTO_BASEPOINT_POINT, firstHandle, firstCommitment, commitments[1],
			curve.RISTRETTO_BASEPOINT_POINT, openingBasepoint, secondCommitment, commitments[2],
			secondPubkey, secondHandle, commitments[3],
		},
	)
	if !check.IsIdentity() {
		return fmt.Errorf("verify ciphertext equality proof: algebraic relation")
	}
	return nil
}

func (service ConfidentialTransferService) GenerateCiphertextCommitmentEqualityProof(keypair ElGamalKeypair, ciphertext ElGamalCiphertext, commitment PedersenCommitment, opening PedersenOpening, amount uint64) (ZKProofData, error) {
	return service.generateCiphertextCommitmentEqualityProofWithReader(keypair, ciphertext, commitment, opening, amount, rand.Reader)
}

func (service ConfidentialTransferService) generateCiphertextCommitmentEqualityProofWithReader(keypair ElGamalKeypair, ciphertext ElGamalCiphertext, commitment PedersenCommitment, opening PedersenOpening, amount uint64, random io.Reader) (ZKProofData, error) {
	secret, publicKey, err := service.elGamalKeypairValues(keypair)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate ciphertext commitment equality proof: %w", err)
	}
	ciphertextCommitment, handle, err := service.elGamalCiphertextPoints(ciphertext)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate ciphertext commitment equality proof: %w", err)
	}
	if !service.elGamalDecryptionMatchesAmount(secret, ciphertextCommitment, handle, amount) {
		return ZKProofData{}, fmt.Errorf("generate ciphertext commitment equality proof: ciphertext amount mismatch")
	}
	expectedCommitment, err := service.CommitPedersen(amount, opening)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate ciphertext commitment equality proof: %w", err)
	}
	if expectedCommitment != commitment {
		return ZKProofData{}, fmt.Errorf("generate ciphertext commitment equality proof: commitment amount mismatch")
	}
	openingScalar, err := service.scalarFromBytes(opening[:], "Pedersen opening")
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate ciphertext commitment equality proof: %w", err)
	}
	secretNonce, amountNonce, openingNonce, err := service.equalityProofNonces(random)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate ciphertext commitment equality proof: %w", err)
	}
	openingBasepoint, err := service.pedersenOpeningBasepoint()
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate ciphertext commitment equality proof: %w", err)
	}
	commitments := [3]PedersenCommitment{
		service.pedersenCommitment(curve.NewRistrettoPoint().Mul(publicKey, secretNonce)),
		service.pedersenCommitment(curve.NewRistrettoPoint().MultiscalarMul(
			[]*scalar.Scalar{amountNonce, secretNonce},
			[]*curve.RistrettoPoint{curve.RISTRETTO_BASEPOINT_POINT, handle},
		)),
		service.pedersenCommitment(curve.NewRistrettoPoint().MultiscalarMul(
			[]*scalar.Scalar{amountNonce, openingNonce},
			[]*curve.RistrettoPoint{curve.RISTRETTO_BASEPOINT_POINT, openingBasepoint},
		)),
	}
	context := make([]byte, 128)
	copy(context[:32], keypair.PublicKey[:])
	copy(context[32:96], ciphertext[:])
	copy(context[96:], commitment[:])
	transcript := service.ciphertextCommitmentEqualityTranscript(context)
	for index := range commitments {
		service.appendTranscriptPoint(transcript, service.equalityCommitmentLabel(index), commitments[index][:])
	}
	challenge, err := service.transcriptChallengeScalar(transcript, "c")
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate ciphertext commitment equality proof: %w", err)
	}
	responses := [3]*scalar.Scalar{
		scalar.New().Add(scalar.New().Mul(challenge, secret), secretNonce),
		scalar.New().Add(scalar.New().Mul(challenge, scalar.New().SetUint64(amount)), amountNonce),
		scalar.New().Add(scalar.New().Mul(challenge, openingScalar), openingNonce),
	}
	proof, err := service.ciphertextCommitmentEqualityProofBytes(commitments, responses, transcript)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate ciphertext commitment equality proof: %w", err)
	}
	return ZKProofData{Discriminator: 3, Context: context, Proof: proof[:]}, nil
}

func (service ConfidentialTransferService) VerifyCiphertextCommitmentEqualityProof(data ZKProofData) (err error) {
	defer service.classifyInvalidProof(&err)
	if data.Discriminator != 3 || len(data.Context) != 128 || len(data.Proof) != 192 {
		return fmt.Errorf("verify ciphertext commitment equality proof: invalid proof data")
	}
	publicKey, err := service.ristrettoPoint(data.Context[:32], "ElGamal public key")
	if err != nil {
		return fmt.Errorf("verify ciphertext commitment equality proof: %w", err)
	}
	ciphertextCommitment, err := service.ristrettoPoint(data.Context[32:64], "ciphertext commitment")
	if err != nil {
		return fmt.Errorf("verify ciphertext commitment equality proof: %w", err)
	}
	handle, err := service.ristrettoPoint(data.Context[64:96], "ciphertext handle")
	if err != nil {
		return fmt.Errorf("verify ciphertext commitment equality proof: %w", err)
	}
	commitment, err := service.ristrettoPoint(data.Context[96:], "Pedersen commitment")
	if err != nil {
		return fmt.Errorf("verify ciphertext commitment equality proof: %w", err)
	}
	if publicKey.IsIdentity() || ciphertextCommitment.IsIdentity() || handle.IsIdentity() || commitment.IsIdentity() {
		return fmt.Errorf("verify ciphertext commitment equality proof: identity context point")
	}
	commitments, responses, err := service.parseCiphertextCommitmentEqualityProof(data.Proof)
	if err != nil {
		return fmt.Errorf("verify ciphertext commitment equality proof: %w", err)
	}
	transcript := service.ciphertextCommitmentEqualityTranscript(data.Context)
	for index := range commitments {
		service.appendTranscriptPoint(transcript, service.equalityCommitmentLabel(index), data.Proof[index*32:(index+1)*32])
	}
	challenge, err := service.transcriptChallengeScalar(transcript, "c")
	if err != nil {
		return fmt.Errorf("verify ciphertext commitment equality proof: %w", err)
	}
	batching, err := service.appendEqualityResponses(transcript, responses)
	if err != nil {
		return fmt.Errorf("verify ciphertext commitment equality proof: %w", err)
	}
	openingBasepoint, err := service.pedersenOpeningBasepoint()
	if err != nil {
		return fmt.Errorf("verify ciphertext commitment equality proof: %w", err)
	}
	batchingSquared := scalar.New().Mul(batching, batching)
	negativeChallenge := scalar.New().Neg(challenge)
	negativeOne := scalar.New().Neg(scalar.New().One())
	negativeBatching := scalar.New().Neg(batching)
	negativeBatchingSquared := scalar.New().Neg(batchingSquared)
	check := curve.NewRistrettoPoint().MultiscalarMul(
		[]*scalar.Scalar{
			responses[0], negativeChallenge, negativeOne,
			scalar.New().Mul(batching, responses[1]), scalar.New().Mul(batching, responses[0]), scalar.New().Mul(negativeBatching, challenge), negativeBatching,
			scalar.New().Mul(batchingSquared, responses[1]), scalar.New().Mul(batchingSquared, responses[2]), scalar.New().Mul(negativeBatchingSquared, challenge), negativeBatchingSquared,
		},
		[]*curve.RistrettoPoint{
			publicKey, openingBasepoint, commitments[0],
			curve.RISTRETTO_BASEPOINT_POINT, handle, ciphertextCommitment, commitments[1],
			curve.RISTRETTO_BASEPOINT_POINT, openingBasepoint, commitment, commitments[2],
		},
	)
	if !check.IsIdentity() {
		return fmt.Errorf("verify ciphertext commitment equality proof: algebraic relation")
	}
	return nil
}

func (ConfidentialTransferService) elGamalDecryptionMatchesAmount(secret *scalar.Scalar, commitment, handle *curve.RistrettoPoint, amount uint64) bool {
	target := curve.NewRistrettoPoint().Sub(commitment, curve.NewRistrettoPoint().Mul(handle, secret))
	expected := curve.NewRistrettoPoint().Mul(curve.RISTRETTO_BASEPOINT_POINT, scalar.New().SetUint64(amount))
	return target.Equal(expected) == 1
}

func (ConfidentialTransferService) equalityProofNonces(random io.Reader) (*scalar.Scalar, *scalar.Scalar, *scalar.Scalar, error) {
	secretNonce, err := scalar.New().SetRandom(random)
	if err != nil {
		return nil, nil, nil, err
	}
	amountNonce, err := scalar.New().SetRandom(random)
	if err != nil {
		return nil, nil, nil, err
	}
	openingNonce, err := scalar.New().SetRandom(random)
	if err != nil {
		return nil, nil, nil, err
	}
	return secretNonce, amountNonce, openingNonce, nil
}

func (service ConfidentialTransferService) ciphertextCiphertextEqualityTranscript(context []byte) *merlin.Transcript {
	transcript := service.newZKElGamalTranscript("ciphertext-ciphertext-equality-instruction")
	transcript.AppendMessage([]byte("first-pubkey"), context[:32])
	transcript.AppendMessage([]byte("second-pubkey"), context[32:64])
	transcript.AppendMessage([]byte("first-ciphertext"), context[64:128])
	transcript.AppendMessage([]byte("second-ciphertext"), context[128:])
	service.appendProofDomainSeparator(transcript, "ciphertext-ciphertext-equality-proof")
	return transcript
}

func (service ConfidentialTransferService) ciphertextCommitmentEqualityTranscript(context []byte) *merlin.Transcript {
	transcript := service.newZKElGamalTranscript("ciphertext-commitment-equality-instruction")
	transcript.AppendMessage([]byte("pubkey"), context[:32])
	transcript.AppendMessage([]byte("ciphertext"), context[32:96])
	transcript.AppendMessage([]byte("commitment"), context[96:])
	service.appendProofDomainSeparator(transcript, "ciphertext-commitment-equality-proof")
	return transcript
}

func (service ConfidentialTransferService) ciphertextCiphertextEqualityProofBytes(commitments [4]PedersenCommitment, responses [3]*scalar.Scalar, transcript *merlin.Transcript) (CiphertextCiphertextEqualityProof, error) {
	proof := CiphertextCiphertextEqualityProof{}
	for index := range commitments {
		copy(proof[index*32:(index+1)*32], commitments[index][:])
	}
	for index, response := range responses {
		encoded, err := response.MarshalBinary()
		if err != nil {
			return CiphertextCiphertextEqualityProof{}, err
		}
		copy(proof[(index+4)*32:(index+5)*32], encoded)
		service.appendTranscriptScalar(transcript, service.equalityResponseLabel(index), encoded)
	}
	if _, err := service.transcriptChallengeScalar(transcript, "w"); err != nil {
		return CiphertextCiphertextEqualityProof{}, err
	}
	return proof, nil
}

func (service ConfidentialTransferService) ciphertextCommitmentEqualityProofBytes(commitments [3]PedersenCommitment, responses [3]*scalar.Scalar, transcript *merlin.Transcript) (CiphertextCommitmentEqualityProof, error) {
	proof := CiphertextCommitmentEqualityProof{}
	for index := range commitments {
		copy(proof[index*32:(index+1)*32], commitments[index][:])
	}
	for index, response := range responses {
		encoded, err := response.MarshalBinary()
		if err != nil {
			return CiphertextCommitmentEqualityProof{}, err
		}
		copy(proof[(index+3)*32:(index+4)*32], encoded)
		service.appendTranscriptScalar(transcript, service.equalityResponseLabel(index), encoded)
	}
	if _, err := service.transcriptChallengeScalar(transcript, "w"); err != nil {
		return CiphertextCommitmentEqualityProof{}, err
	}
	return proof, nil
}

func (service ConfidentialTransferService) parseCiphertextCiphertextEqualityProof(data []byte) ([4]*curve.RistrettoPoint, [3]*scalar.Scalar, error) {
	commitments := [4]*curve.RistrettoPoint{}
	responses := [3]*scalar.Scalar{}
	for index := range commitments {
		point, err := service.ristrettoPoint(data[index*32:(index+1)*32], "ciphertext equality proof commitment")
		if err != nil {
			return commitments, responses, err
		}
		if point.IsIdentity() {
			return commitments, responses, fmt.Errorf("identity ciphertext equality proof commitment")
		}
		commitments[index] = point
	}
	for index := range responses {
		response, err := service.scalarFromBytes(data[(index+4)*32:(index+5)*32], "ciphertext equality proof response")
		if err != nil {
			return commitments, responses, err
		}
		responses[index] = response
	}
	return commitments, responses, nil
}

func (service ConfidentialTransferService) parseCiphertextCommitmentEqualityProof(data []byte) ([3]*curve.RistrettoPoint, [3]*scalar.Scalar, error) {
	commitments := [3]*curve.RistrettoPoint{}
	responses := [3]*scalar.Scalar{}
	for index := range commitments {
		point, err := service.ristrettoPoint(data[index*32:(index+1)*32], "ciphertext commitment equality proof commitment")
		if err != nil {
			return commitments, responses, err
		}
		if point.IsIdentity() {
			return commitments, responses, fmt.Errorf("identity ciphertext commitment equality proof commitment")
		}
		commitments[index] = point
	}
	for index := range responses {
		response, err := service.scalarFromBytes(data[(index+3)*32:(index+4)*32], "ciphertext commitment equality proof response")
		if err != nil {
			return commitments, responses, err
		}
		responses[index] = response
	}
	return commitments, responses, nil
}

func (service ConfidentialTransferService) appendEqualityResponses(transcript *merlin.Transcript, responses [3]*scalar.Scalar) (*scalar.Scalar, error) {
	for index, response := range responses {
		encoded, err := response.MarshalBinary()
		if err != nil {
			return nil, err
		}
		service.appendTranscriptScalar(transcript, service.equalityResponseLabel(index), encoded)
	}
	return service.transcriptChallengeScalar(transcript, "w")
}

func (ConfidentialTransferService) equalityCommitmentLabel(index int) string {
	switch index {
	case 0:
		return "Y_0"
	case 1:
		return "Y_1"
	case 2:
		return "Y_2"
	default:
		return "Y_3"
	}
}

func (ConfidentialTransferService) equalityResponseLabel(index int) string {
	switch index {
	case 0:
		return "z_s"
	case 1:
		return "z_x"
	default:
		return "z_r"
	}
}
