package token2022

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/gtank/merlin"
	"github.com/oasisprotocol/curve25519-voi/curve"
	"github.com/oasisprotocol/curve25519-voi/curve/scalar"
)

type PubkeyValidityProof [64]byte

type ZeroCiphertextProof [96]byte

func (service ConfidentialTransferService) GeneratePubkeyValidityProof(keypair ElGamalKeypair) (ZKProofData, error) {
	return service.generatePubkeyValidityProofWithReader(keypair, rand.Reader)
}

func (service ConfidentialTransferService) generatePubkeyValidityProofWithReader(keypair ElGamalKeypair, random io.Reader) (ZKProofData, error) {
	secret, _, err := service.elGamalKeypairValues(keypair)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate public key validity proof: %w", err)
	}
	openingBasepoint, err := service.pedersenOpeningBasepoint()
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate public key validity proof: %w", err)
	}
	nonce, err := scalar.New().SetRandom(random)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate public key validity proof: %w", err)
	}
	commitment := curve.NewRistrettoPoint().Mul(openingBasepoint, nonce)
	commitmentBytes := service.pedersenCommitment(commitment)
	transcript := service.newZKElGamalTranscript("pubkey-validity-instruction")
	transcript.AppendMessage([]byte("pubkey"), keypair.PublicKey[:])
	service.appendProofDomainSeparator(transcript, "pubkey-proof")
	service.appendTranscriptPoint(transcript, "Y", commitmentBytes[:])
	challenge, err := service.transcriptChallengeScalar(transcript, "c")
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate public key validity proof: %w", err)
	}
	response := scalar.New().Add(scalar.New().Mul(challenge, scalar.New().Invert(secret)), nonce)
	responseBytes, err := response.MarshalBinary()
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate public key validity proof: %w", err)
	}
	proof := PubkeyValidityProof{}
	copy(proof[:32], commitmentBytes[:])
	copy(proof[32:], responseBytes)
	context := make([]byte, 32)
	copy(context, keypair.PublicKey[:])
	return ZKProofData{Discriminator: 4, Context: context, Proof: proof[:]}, nil
}

func (service ConfidentialTransferService) VerifyPubkeyValidityProof(data ZKProofData) (err error) {
	defer service.classifyInvalidProof(&err)
	if data.Discriminator != 4 || len(data.Context) != 32 || len(data.Proof) != 64 {
		return fmt.Errorf("verify public key validity proof: invalid proof data")
	}
	publicKey, err := service.ristrettoPoint(data.Context, "public key validity context")
	if err != nil {
		return fmt.Errorf("verify public key validity proof: %w", err)
	}
	if publicKey.IsIdentity() {
		return fmt.Errorf("verify public key validity proof: identity public key")
	}
	commitment, err := service.ristrettoPoint(data.Proof[:32], "public key validity commitment")
	if err != nil {
		return fmt.Errorf("verify public key validity proof: %w", err)
	}
	if commitment.IsIdentity() {
		return fmt.Errorf("verify public key validity proof: identity commitment")
	}
	response, err := service.scalarFromBytes(data.Proof[32:], "public key validity response")
	if err != nil {
		return fmt.Errorf("verify public key validity proof: %w", err)
	}
	transcript := service.newZKElGamalTranscript("pubkey-validity-instruction")
	transcript.AppendMessage([]byte("pubkey"), data.Context)
	service.appendProofDomainSeparator(transcript, "pubkey-proof")
	service.appendTranscriptPoint(transcript, "Y", data.Proof[:32])
	challenge, err := service.transcriptChallengeScalar(transcript, "c")
	if err != nil {
		return fmt.Errorf("verify public key validity proof: %w", err)
	}
	openingBasepoint, err := service.pedersenOpeningBasepoint()
	if err != nil {
		return fmt.Errorf("verify public key validity proof: %w", err)
	}
	check := curve.NewRistrettoPoint().MultiscalarMul(
		[]*scalar.Scalar{response, scalar.New().Neg(challenge), scalar.New().Neg(scalar.New().One())},
		[]*curve.RistrettoPoint{openingBasepoint, publicKey, commitment},
	)
	if !check.IsIdentity() {
		return fmt.Errorf("verify public key validity proof: algebraic relation")
	}
	return nil
}

func (service ConfidentialTransferService) GenerateZeroCiphertextProof(keypair ElGamalKeypair, ciphertext ElGamalCiphertext) (ZKProofData, error) {
	return service.generateZeroCiphertextProofWithReader(keypair, ciphertext, rand.Reader)
}

func (service ConfidentialTransferService) generateZeroCiphertextProofWithReader(keypair ElGamalKeypair, ciphertext ElGamalCiphertext, random io.Reader) (ZKProofData, error) {
	secret, publicKey, err := service.elGamalKeypairValues(keypair)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate zero ciphertext proof: %w", err)
	}
	commitment, handle, err := service.elGamalCiphertextPoints(ciphertext)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate zero ciphertext proof: %w", err)
	}
	if !curve.NewRistrettoPoint().Sub(commitment, curve.NewRistrettoPoint().Mul(handle, secret)).IsIdentity() {
		return ZKProofData{}, fmt.Errorf("generate zero ciphertext proof: ciphertext is not zero")
	}
	nonce, err := scalar.New().SetRandom(random)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate zero ciphertext proof: %w", err)
	}
	publicCommitment := service.pedersenCommitment(curve.NewRistrettoPoint().Mul(publicKey, nonce))
	handleCommitment := service.pedersenCommitment(curve.NewRistrettoPoint().Mul(handle, nonce))
	context := make([]byte, 96)
	copy(context[:32], keypair.PublicKey[:])
	copy(context[32:], ciphertext[:])
	transcript := service.newZKElGamalTranscript("zero-ciphertext-instruction")
	transcript.AppendMessage([]byte("pubkey"), context[:32])
	transcript.AppendMessage([]byte("ciphertext"), context[32:])
	service.appendProofDomainSeparator(transcript, "zero-ciphertext-proof")
	service.appendTranscriptPoint(transcript, "Y_P", publicCommitment[:])
	service.appendTranscriptPoint(transcript, "Y_D", handleCommitment[:])
	challenge, err := service.transcriptChallengeScalar(transcript, "c")
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate zero ciphertext proof: %w", err)
	}
	response := scalar.New().Add(scalar.New().Mul(challenge, secret), nonce)
	responseBytes, err := response.MarshalBinary()
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate zero ciphertext proof: %w", err)
	}
	service.appendTranscriptScalar(transcript, "z", responseBytes)
	if _, err := service.transcriptChallengeScalar(transcript, "w"); err != nil {
		return ZKProofData{}, fmt.Errorf("generate zero ciphertext proof: %w", err)
	}
	proof := ZeroCiphertextProof{}
	copy(proof[:32], publicCommitment[:])
	copy(proof[32:64], handleCommitment[:])
	copy(proof[64:], responseBytes)
	return ZKProofData{Discriminator: 1, Context: context, Proof: proof[:]}, nil
}

func (service ConfidentialTransferService) VerifyZeroCiphertextProof(data ZKProofData) (err error) {
	defer service.classifyInvalidProof(&err)
	if data.Discriminator != 1 || len(data.Context) != 96 || len(data.Proof) != 96 {
		return fmt.Errorf("verify zero ciphertext proof: invalid proof data")
	}
	publicKey, err := service.ristrettoPoint(data.Context[:32], "zero ciphertext public key")
	if err != nil {
		return fmt.Errorf("verify zero ciphertext proof: %w", err)
	}
	commitment, err := service.ristrettoPoint(data.Context[32:64], "zero ciphertext commitment")
	if err != nil {
		return fmt.Errorf("verify zero ciphertext proof: %w", err)
	}
	handle, err := service.ristrettoPoint(data.Context[64:], "zero ciphertext handle")
	if err != nil {
		return fmt.Errorf("verify zero ciphertext proof: %w", err)
	}
	if publicKey.IsIdentity() || commitment.IsIdentity() || handle.IsIdentity() {
		return fmt.Errorf("verify zero ciphertext proof: identity context point")
	}
	publicCommitment, err := service.ristrettoPoint(data.Proof[:32], "zero ciphertext public commitment")
	if err != nil {
		return fmt.Errorf("verify zero ciphertext proof: %w", err)
	}
	if publicCommitment.IsIdentity() {
		return fmt.Errorf("verify zero ciphertext proof: identity public commitment")
	}
	handleCommitment, err := service.ristrettoPoint(data.Proof[32:64], "zero ciphertext handle commitment")
	if err != nil {
		return fmt.Errorf("verify zero ciphertext proof: %w", err)
	}
	response, err := service.scalarFromBytes(data.Proof[64:], "zero ciphertext response")
	if err != nil {
		return fmt.Errorf("verify zero ciphertext proof: %w", err)
	}
	transcript := service.newZKElGamalTranscript("zero-ciphertext-instruction")
	transcript.AppendMessage([]byte("pubkey"), data.Context[:32])
	transcript.AppendMessage([]byte("ciphertext"), data.Context[32:])
	service.appendProofDomainSeparator(transcript, "zero-ciphertext-proof")
	service.appendTranscriptPoint(transcript, "Y_P", data.Proof[:32])
	service.appendTranscriptPoint(transcript, "Y_D", data.Proof[32:64])
	challenge, err := service.transcriptChallengeScalar(transcript, "c")
	if err != nil {
		return fmt.Errorf("verify zero ciphertext proof: %w", err)
	}
	service.appendTranscriptScalar(transcript, "z", data.Proof[64:])
	batching, err := service.transcriptChallengeScalar(transcript, "w")
	if err != nil {
		return fmt.Errorf("verify zero ciphertext proof: %w", err)
	}
	openingBasepoint, err := service.pedersenOpeningBasepoint()
	if err != nil {
		return fmt.Errorf("verify zero ciphertext proof: %w", err)
	}
	negativeChallenge := scalar.New().Neg(challenge)
	negativeOne := scalar.New().Neg(scalar.New().One())
	batchedResponse := scalar.New().Mul(batching, response)
	negativeBatchedChallenge := scalar.New().Neg(scalar.New().Mul(batching, challenge))
	negativeBatching := scalar.New().Neg(batching)
	check := curve.NewRistrettoPoint().MultiscalarMul(
		[]*scalar.Scalar{response, negativeChallenge, negativeOne, batchedResponse, negativeBatchedChallenge, negativeBatching},
		[]*curve.RistrettoPoint{publicKey, openingBasepoint, publicCommitment, handle, commitment, handleCommitment},
	)
	if !check.IsIdentity() {
		return fmt.Errorf("verify zero ciphertext proof: algebraic relation")
	}
	return nil
}

func (service ConfidentialTransferService) elGamalKeypairValues(keypair ElGamalKeypair) (*scalar.Scalar, *curve.RistrettoPoint, error) {
	secret, err := service.scalarFromBytes(keypair.SecretKey[:], "ElGamal secret key")
	if err != nil {
		return nil, nil, err
	}
	if secret.Equal(scalar.New()) == 1 {
		return nil, nil, fmt.Errorf("zero ElGamal secret key")
	}
	derived, err := service.DeriveElGamalPubkey(keypair.SecretKey)
	if err != nil {
		return nil, nil, err
	}
	if derived != keypair.PublicKey {
		return nil, nil, fmt.Errorf("ElGamal keypair mismatch")
	}
	publicKey, err := service.ristrettoPoint(keypair.PublicKey[:], "ElGamal public key")
	if err != nil {
		return nil, nil, err
	}
	return secret, publicKey, nil
}

func (ConfidentialTransferService) newZKElGamalTranscript(label string) *merlin.Transcript {
	transcript := merlin.NewTranscript("solana-zk-elgamal-proof-program-v1")
	transcript.AppendMessage([]byte("dom-sep"), []byte(label))
	return transcript
}

func (ConfidentialTransferService) appendProofDomainSeparator(transcript *merlin.Transcript, label string) {
	transcript.AppendMessage([]byte("dom-sep"), []byte(label))
}

func (ConfidentialTransferService) appendTranscriptScalar(transcript *merlin.Transcript, label string, value []byte) {
	transcript.AppendMessage([]byte(label), value)
}

func (ConfidentialTransferService) appendTranscriptPoint(transcript *merlin.Transcript, label string, value []byte) {
	transcript.AppendMessage([]byte(label), value)
}

func (ConfidentialTransferService) appendTranscriptUint64(transcript *merlin.Transcript, label string, value uint64) {
	encoded := make([]byte, 8)
	binary.LittleEndian.PutUint64(encoded, value)
	transcript.AppendMessage([]byte(label), encoded)
}

func (ConfidentialTransferService) transcriptChallengeScalar(transcript *merlin.Transcript, label string) (*scalar.Scalar, error) {
	return scalar.NewFromBytesModOrderWide(transcript.ExtractBytes([]byte(label), 64))
}
