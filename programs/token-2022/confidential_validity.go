package token2022

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/gtank/merlin"
	"github.com/oasisprotocol/curve25519-voi/curve"
	"github.com/oasisprotocol/curve25519-voi/curve/scalar"
)

type GroupedCiphertext2HandlesValidityProof [160]byte

type BatchedGroupedCiphertext2HandlesValidityProof [160]byte

type GroupedCiphertext3HandlesValidityProof [192]byte

type BatchedGroupedCiphertext3HandlesValidityProof [192]byte

func (service ConfidentialTransferService) GenerateGroupedCiphertext2HandlesValidityProof(publicKeys [2]ElGamalPubkey, ciphertext GroupedElGamalCiphertext2Handles, amount uint64, opening PedersenOpening) (ZKProofData, error) {
	return service.generateGroupedCiphertext2HandlesValidityProofWithReader(publicKeys, ciphertext, amount, opening, rand.Reader)
}

func (service ConfidentialTransferService) generateGroupedCiphertext2HandlesValidityProofWithReader(publicKeys [2]ElGamalPubkey, ciphertext GroupedElGamalCiphertext2Handles, amount uint64, opening PedersenOpening, random io.Reader) (ZKProofData, error) {
	return service.generateGroupedCiphertextValidityProof(publicKeys[:], [][]byte{ciphertext[:]}, []uint64{amount}, []PedersenOpening{opening}, 9, "grouped-ciphertext-validity-2-handles-instruction", random)
}

func (service ConfidentialTransferService) VerifyGroupedCiphertext2HandlesValidityProof(data ZKProofData) error {
	return service.verifyGroupedCiphertextValidityProof(data, 2, 1, 9, "grouped-ciphertext-validity-2-handles-instruction")
}

func (service ConfidentialTransferService) GenerateBatchedGroupedCiphertext2HandlesValidityProof(publicKeys [2]ElGamalPubkey, ciphertexts [2]GroupedElGamalCiphertext2Handles, amounts [2]uint64, openings [2]PedersenOpening) (ZKProofData, error) {
	return service.generateBatchedGroupedCiphertext2HandlesValidityProofWithReader(publicKeys, ciphertexts, amounts, openings, rand.Reader)
}

func (service ConfidentialTransferService) generateBatchedGroupedCiphertext2HandlesValidityProofWithReader(publicKeys [2]ElGamalPubkey, ciphertexts [2]GroupedElGamalCiphertext2Handles, amounts [2]uint64, openings [2]PedersenOpening, random io.Reader) (ZKProofData, error) {
	return service.generateGroupedCiphertextValidityProof(publicKeys[:], [][]byte{ciphertexts[0][:], ciphertexts[1][:]}, amounts[:], openings[:], 10, "batched-grouped-ciphertext-validity-2-handles-instruction", random)
}

func (service ConfidentialTransferService) VerifyBatchedGroupedCiphertext2HandlesValidityProof(data ZKProofData) error {
	return service.verifyGroupedCiphertextValidityProof(data, 2, 2, 10, "batched-grouped-ciphertext-validity-2-handles-instruction")
}

func (service ConfidentialTransferService) GenerateGroupedCiphertext3HandlesValidityProof(publicKeys [3]ElGamalPubkey, ciphertext GroupedElGamalCiphertext3Handles, amount uint64, opening PedersenOpening) (ZKProofData, error) {
	return service.generateGroupedCiphertext3HandlesValidityProofWithReader(publicKeys, ciphertext, amount, opening, rand.Reader)
}

func (service ConfidentialTransferService) generateGroupedCiphertext3HandlesValidityProofWithReader(publicKeys [3]ElGamalPubkey, ciphertext GroupedElGamalCiphertext3Handles, amount uint64, opening PedersenOpening, random io.Reader) (ZKProofData, error) {
	return service.generateGroupedCiphertextValidityProof(publicKeys[:], [][]byte{ciphertext[:]}, []uint64{amount}, []PedersenOpening{opening}, 11, "grouped-ciphertext-validity-3-handles-instruction", random)
}

func (service ConfidentialTransferService) VerifyGroupedCiphertext3HandlesValidityProof(data ZKProofData) error {
	return service.verifyGroupedCiphertextValidityProof(data, 3, 1, 11, "grouped-ciphertext-validity-3-handles-instruction")
}

func (service ConfidentialTransferService) GenerateBatchedGroupedCiphertext3HandlesValidityProof(publicKeys [3]ElGamalPubkey, ciphertexts [2]GroupedElGamalCiphertext3Handles, amounts [2]uint64, openings [2]PedersenOpening) (ZKProofData, error) {
	return service.generateBatchedGroupedCiphertext3HandlesValidityProofWithReader(publicKeys, ciphertexts, amounts, openings, rand.Reader)
}

func (service ConfidentialTransferService) generateBatchedGroupedCiphertext3HandlesValidityProofWithReader(publicKeys [3]ElGamalPubkey, ciphertexts [2]GroupedElGamalCiphertext3Handles, amounts [2]uint64, openings [2]PedersenOpening, random io.Reader) (ZKProofData, error) {
	return service.generateGroupedCiphertextValidityProof(publicKeys[:], [][]byte{ciphertexts[0][:], ciphertexts[1][:]}, amounts[:], openings[:], 12, "batched-grouped-ciphertext-validity-3-handles-instruction", random)
}

func (service ConfidentialTransferService) VerifyBatchedGroupedCiphertext3HandlesValidityProof(data ZKProofData) error {
	return service.verifyGroupedCiphertextValidityProof(data, 3, 2, 12, "batched-grouped-ciphertext-validity-3-handles-instruction")
}

func (service ConfidentialTransferService) generateGroupedCiphertextValidityProof(publicKeys []ElGamalPubkey, ciphertexts [][]byte, amounts []uint64, openings []PedersenOpening, discriminator uint8, instructionLabel string, random io.Reader) (ZKProofData, error) {
	for index := range ciphertexts {
		expected, err := service.encryptGroupedElGamal(publicKeys, amounts[index], openings[index])
		if err != nil {
			return ZKProofData{}, fmt.Errorf("generate grouped ciphertext validity proof: %w", err)
		}
		if !bytes.Equal(expected, ciphertexts[index]) {
			return ZKProofData{}, fmt.Errorf("generate grouped ciphertext validity proof: ciphertext %d mismatch", index)
		}
	}
	context := make([]byte, len(publicKeys)*32)
	for index := range publicKeys {
		copy(context[index*32:], publicKeys[index][:])
	}
	for index := range ciphertexts {
		context = append(context, ciphertexts[index]...)
	}
	transcript := service.groupedCiphertextValidityTranscript(context, len(publicKeys), len(ciphertexts), instructionLabel)
	amountScalar := scalar.New().SetUint64(amounts[0])
	openingScalar, err := service.scalarFromBytes(openings[0][:], "Pedersen opening")
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate grouped ciphertext validity proof: %w", err)
	}
	if len(ciphertexts) == 2 {
		service.appendBatchedValidityDomainSeparator(transcript, len(publicKeys))
		batching, err := service.transcriptChallengeScalar(transcript, "t")
		if err != nil {
			return ZKProofData{}, fmt.Errorf("generate grouped ciphertext validity proof: %w", err)
		}
		highOpening, err := service.scalarFromBytes(openings[1][:], "high Pedersen opening")
		if err != nil {
			return ZKProofData{}, fmt.Errorf("generate grouped ciphertext validity proof: %w", err)
		}
		amountScalar = scalar.New().Add(amountScalar, scalar.New().Mul(scalar.New().SetUint64(amounts[1]), batching))
		openingScalar = scalar.New().Add(openingScalar, scalar.New().Mul(highOpening, batching))
	}
	proof, err := service.generateGroupedCiphertextValidityProofDirect(publicKeys, amountScalar, openingScalar, transcript, random)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate grouped ciphertext validity proof: %w", err)
	}
	return ZKProofData{Discriminator: discriminator, Context: context, Proof: proof}, nil
}

func (service ConfidentialTransferService) verifyGroupedCiphertextValidityProof(data ZKProofData, handles, ciphertextCount int, discriminator uint8, instructionLabel string) error {
	ciphertextSize := (handles + 1) * 32
	contextSize := handles*32 + ciphertextCount*ciphertextSize
	proofSize := (handles + 3) * 32
	if data.Discriminator != discriminator || len(data.Context) != contextSize || len(data.Proof) != proofSize {
		return fmt.Errorf("verify grouped ciphertext validity proof: invalid proof data")
	}
	publicKeys := make([]*curve.RistrettoPoint, handles)
	for index := range publicKeys {
		publicKey, err := service.ristrettoPoint(data.Context[index*32:(index+1)*32], "grouped ciphertext public key")
		if err != nil {
			return fmt.Errorf("verify grouped ciphertext validity proof: %w", err)
		}
		publicKeys[index] = publicKey
	}
	if publicKeys[0].IsIdentity() || handles == 3 && publicKeys[1].IsIdentity() {
		return fmt.Errorf("verify grouped ciphertext validity proof: identity public key")
	}
	ciphertexts := make([][]*curve.RistrettoPoint, ciphertextCount)
	for ciphertextIndex := range ciphertexts {
		ciphertexts[ciphertextIndex] = make([]*curve.RistrettoPoint, handles+1)
		offset := handles*32 + ciphertextIndex*ciphertextSize
		for pointIndex := range ciphertexts[ciphertextIndex] {
			point, err := service.ristrettoPoint(data.Context[offset+pointIndex*32:offset+(pointIndex+1)*32], "grouped ciphertext")
			if err != nil {
				return fmt.Errorf("verify grouped ciphertext validity proof: %w", err)
			}
			ciphertexts[ciphertextIndex][pointIndex] = point
		}
		if ciphertexts[ciphertextIndex][0].IsIdentity() {
			return fmt.Errorf("verify grouped ciphertext validity proof: identity commitment")
		}
	}
	transcript := service.groupedCiphertextValidityTranscript(data.Context, handles, ciphertextCount, instructionLabel)
	batchedCiphertext := ciphertexts[0]
	if ciphertextCount == 2 {
		service.appendBatchedValidityDomainSeparator(transcript, handles)
		batching, err := service.transcriptChallengeScalar(transcript, "t")
		if err != nil {
			return fmt.Errorf("verify grouped ciphertext validity proof: %w", err)
		}
		batchedCiphertext = make([]*curve.RistrettoPoint, handles+1)
		for index := range batchedCiphertext {
			batchedCiphertext[index] = curve.NewRistrettoPoint().Add(
				ciphertexts[0][index],
				curve.NewRistrettoPoint().Mul(ciphertexts[1][index], batching),
			)
		}
	}
	if err := service.verifyGroupedCiphertextValidityProofDirect(publicKeys, batchedCiphertext, data.Proof, transcript); err != nil {
		return fmt.Errorf("verify grouped ciphertext validity proof: %w", err)
	}
	return nil
}

func (service ConfidentialTransferService) generateGroupedCiphertextValidityProofDirect(publicKeys []ElGamalPubkey, amount, opening *scalar.Scalar, transcript *merlin.Transcript, random io.Reader) ([]byte, error) {
	service.appendValidityDomainSeparator(transcript, len(publicKeys))
	openingNonce, err := scalar.New().SetRandom(random)
	if err != nil {
		return nil, err
	}
	amountNonce, err := scalar.New().SetRandom(random)
	if err != nil {
		return nil, err
	}
	openingBasepoint, err := service.pedersenOpeningBasepoint()
	if err != nil {
		return nil, err
	}
	commitments := make([]PedersenCommitment, len(publicKeys)+1)
	commitments[0] = service.pedersenCommitment(curve.NewRistrettoPoint().MultiscalarMul(
		[]*scalar.Scalar{openingNonce, amountNonce},
		[]*curve.RistrettoPoint{openingBasepoint, curve.RISTRETTO_BASEPOINT_POINT},
	))
	for index := range publicKeys {
		publicKey, err := service.ristrettoPoint(publicKeys[index][:], "grouped ciphertext public key")
		if err != nil {
			return nil, err
		}
		commitments[index+1] = service.pedersenCommitment(curve.NewRistrettoPoint().Mul(publicKey, openingNonce))
	}
	for index := range commitments {
		service.appendTranscriptPoint(transcript, service.equalityCommitmentLabel(index), commitments[index][:])
	}
	challenge, err := service.transcriptChallengeScalar(transcript, "c")
	if err != nil {
		return nil, err
	}
	responses := [2]*scalar.Scalar{
		scalar.New().Add(scalar.New().Mul(challenge, opening), openingNonce),
		scalar.New().Add(scalar.New().Mul(challenge, amount), amountNonce),
	}
	proof := make([]byte, (len(publicKeys)+3)*32)
	for index := range commitments {
		copy(proof[index*32:], commitments[index][:])
	}
	for index := range responses {
		encoded, err := responses[index].MarshalBinary()
		if err != nil {
			return nil, err
		}
		copy(proof[(len(commitments)+index)*32:], encoded)
		service.appendTranscriptScalar(transcript, service.validityResponseLabel(index), encoded)
	}
	if _, err := service.transcriptChallengeScalar(transcript, "w"); err != nil {
		return nil, err
	}
	return proof, nil
}

func (service ConfidentialTransferService) verifyGroupedCiphertextValidityProofDirect(publicKeys, ciphertext []*curve.RistrettoPoint, proof []byte, transcript *merlin.Transcript) error {
	handles := len(publicKeys)
	service.appendValidityDomainSeparator(transcript, handles)
	commitments := make([]*curve.RistrettoPoint, handles+1)
	for index := range commitments {
		point, err := service.ristrettoPoint(proof[index*32:(index+1)*32], "grouped ciphertext validity proof commitment")
		if err != nil {
			return err
		}
		if index < handles && point.IsIdentity() {
			return fmt.Errorf("identity grouped ciphertext validity proof commitment")
		}
		commitments[index] = point
		service.appendTranscriptPoint(transcript, service.equalityCommitmentLabel(index), proof[index*32:(index+1)*32])
	}
	challenge, err := service.transcriptChallengeScalar(transcript, "c")
	if err != nil {
		return err
	}
	responses := [2]*scalar.Scalar{}
	for index := range responses {
		offset := (len(commitments) + index) * 32
		response, err := service.scalarFromBytes(proof[offset:offset+32], "grouped ciphertext validity proof response")
		if err != nil {
			return err
		}
		responses[index] = response
		service.appendTranscriptScalar(transcript, service.validityResponseLabel(index), proof[offset:offset+32])
	}
	batching, err := service.transcriptChallengeScalar(transcript, "w")
	if err != nil {
		return err
	}
	openingBasepoint, err := service.pedersenOpeningBasepoint()
	if err != nil {
		return err
	}
	negativeChallenge := scalar.New().Neg(challenge)
	negativeOne := scalar.New().Neg(scalar.New().One())
	scalars := []*scalar.Scalar{responses[0], responses[1], negativeChallenge, negativeOne}
	points := []*curve.RistrettoPoint{openingBasepoint, curve.RISTRETTO_BASEPOINT_POINT, ciphertext[0], commitments[0]}
	batchingPower := scalar.New().One()
	for index := range publicKeys {
		batchingPower = scalar.New().Mul(batchingPower, batching)
		negativePower := scalar.New().Neg(batchingPower)
		scalars = append(
			scalars,
			scalar.New().Mul(batchingPower, responses[0]),
			scalar.New().Mul(negativePower, challenge),
			negativePower,
		)
		points = append(points, publicKeys[index], ciphertext[index+1], commitments[index+1])
	}
	if !curve.NewRistrettoPoint().MultiscalarMul(scalars, points).IsIdentity() {
		return fmt.Errorf("algebraic relation")
	}
	return nil
}

func (service ConfidentialTransferService) groupedCiphertextValidityTranscript(context []byte, handles, ciphertextCount int, instructionLabel string) *merlin.Transcript {
	transcript := service.newZKElGamalTranscript(instructionLabel)
	for index := 0; index < handles; index++ {
		transcript.AppendMessage([]byte(service.validityPublicKeyLabel(index)), context[index*32:(index+1)*32])
	}
	offset := handles * 32
	if ciphertextCount == 1 {
		transcript.AppendMessage([]byte("grouped-ciphertext"), context[offset:])
		return transcript
	}
	ciphertextSize := (handles + 1) * 32
	transcript.AppendMessage([]byte("grouped-ciphertext-lo"), context[offset:offset+ciphertextSize])
	transcript.AppendMessage([]byte("grouped-ciphertext-hi"), context[offset+ciphertextSize:])
	return transcript
}

func (service ConfidentialTransferService) appendValidityDomainSeparator(transcript *merlin.Transcript, handles int) {
	service.appendProofDomainSeparator(transcript, "validity-proof")
	service.appendTranscriptUint64(transcript, "handles", uint64(handles))
}

func (service ConfidentialTransferService) appendBatchedValidityDomainSeparator(transcript *merlin.Transcript, handles int) {
	service.appendProofDomainSeparator(transcript, "batched-validity-proof")
	service.appendTranscriptUint64(transcript, "handles", uint64(handles))
}

func (ConfidentialTransferService) validityPublicKeyLabel(index int) string {
	switch index {
	case 0:
		return "first-pubkey"
	case 1:
		return "second-pubkey"
	default:
		return "third-pubkey"
	}
}

func (ConfidentialTransferService) validityResponseLabel(index int) string {
	if index == 0 {
		return "z_r"
	}
	return "z_x"
}
