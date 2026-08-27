package token2022

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/binary"
	"fmt"
	"io"
	"math/bits"

	"github.com/gtank/merlin"
	"github.com/oasisprotocol/curve25519-voi/curve"
	"github.com/oasisprotocol/curve25519-voi/curve/scalar"
)

type PercentageWithCapProof [256]byte

type percentageWithCapProofValues struct {
	yMax     *curve.RistrettoPoint
	zMax     *scalar.Scalar
	cMax     *scalar.Scalar
	yDelta   *curve.RistrettoPoint
	yClaimed *curve.RistrettoPoint
	zX       *scalar.Scalar
	zDelta   *scalar.Scalar
	zClaimed *scalar.Scalar
}

func (service ConfidentialTransferService) GeneratePercentageWithCapProof(percentageCommitment PedersenCommitment, percentageOpening PedersenOpening, percentageAmount uint64, deltaCommitment PedersenCommitment, deltaOpening PedersenOpening, deltaAmount uint64, claimedCommitment PedersenCommitment, claimedOpening PedersenOpening, maxValue uint64) (ZKProofData, error) {
	return service.generatePercentageWithCapProofWithReader(percentageCommitment, percentageOpening, percentageAmount, deltaCommitment, deltaOpening, deltaAmount, claimedCommitment, claimedOpening, maxValue, rand.Reader)
}

func (service ConfidentialTransferService) generatePercentageWithCapProofWithReader(percentageCommitment PedersenCommitment, percentageOpening PedersenOpening, percentageAmount uint64, deltaCommitment PedersenCommitment, deltaOpening PedersenOpening, deltaAmount uint64, claimedCommitment PedersenCommitment, claimedOpening PedersenOpening, maxValue uint64, random io.Reader) (ZKProofData, error) {
	expectedPercentage, err := service.CommitPedersen(percentageAmount, percentageOpening)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate percentage with cap proof: %w", err)
	}
	if expectedPercentage != percentageCommitment {
		return ZKProofData{}, fmt.Errorf("generate percentage with cap proof: percentage commitment mismatch")
	}
	expectedClaimed, err := service.CommitPedersen(deltaAmount, claimedOpening)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate percentage with cap proof: %w", err)
	}
	if expectedClaimed != claimedCommitment {
		return ZKProofData{}, fmt.Errorf("generate percentage with cap proof: claimed commitment mismatch")
	}
	if percentageAmount > maxValue {
		return ZKProofData{}, fmt.Errorf("generate percentage with cap proof: capped amount exceeds maximum")
	}
	expectedDelta, err := service.CommitPedersen(deltaAmount, deltaOpening)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate percentage with cap proof: %w", err)
	}
	deltaMismatch := expectedDelta != deltaCommitment
	if percentageAmount < maxValue && deltaMismatch {
		return ZKProofData{}, fmt.Errorf("generate percentage with cap proof: delta commitment mismatch for percentage branch")
	}
	percentageOpeningScalar, err := service.scalarFromBytes(percentageOpening[:], "percentage opening")
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate percentage with cap proof: %w", err)
	}
	deltaOpeningScalar, err := service.scalarFromBytes(deltaOpening[:], "delta opening")
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate percentage with cap proof: %w", err)
	}
	claimedOpeningScalar, err := service.scalarFromBytes(claimedOpening[:], "claimed opening")
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate percentage with cap proof: %w", err)
	}
	percentagePoint, err := service.ristrettoPoint(percentageCommitment[:], "percentage commitment")
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate percentage with cap proof: %w", err)
	}
	deltaPoint, err := service.ristrettoPoint(deltaCommitment[:], "delta commitment")
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate percentage with cap proof: %w", err)
	}
	claimedPoint, err := service.ristrettoPoint(claimedCommitment[:], "claimed commitment")
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate percentage with cap proof: %w", err)
	}
	openingBasepoint, err := service.pedersenOpeningBasepoint()
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate percentage with cap proof: %w", err)
	}
	context := service.percentageWithCapContext(percentageCommitment, deltaCommitment, claimedCommitment, maxValue)
	above, err := service.percentageWithCapAboveProof(percentageOpeningScalar, deltaPoint, claimedPoint, openingBasepoint, context, random)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate percentage with cap proof: %w", err)
	}
	below, err := service.percentageWithCapBelowProof(percentagePoint, deltaOpeningScalar, deltaAmount, claimedOpeningScalar, maxValue, openingBasepoint, context, random)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate percentage with cap proof: %w", err)
	}
	aboveBytes, err := service.percentageWithCapProofBytes(above)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate percentage with cap proof: %w", err)
	}
	belowBytes, err := service.percentageWithCapProofBytes(below)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate percentage with cap proof: %w", err)
	}
	proof := PercentageWithCapProof{}
	_, belowMax := bits.Sub64(percentageAmount, maxValue, 0)
	for i := range proof {
		proof[i] = byte(subtle.ConstantTimeSelect(int(belowMax), int(belowBytes[i]), int(aboveBytes[i])))
	}
	return ZKProofData{Discriminator: 5, Context: context, Proof: proof[:]}, nil
}

// VerifyPercentageWithCapProof must be coupled with a range proof on the percentage commitment to prevent scalar-field wrapping.
func (service ConfidentialTransferService) VerifyPercentageWithCapProof(data ZKProofData) error {
	if data.Discriminator != 5 || len(data.Context) != 104 || len(data.Proof) != 256 {
		return fmt.Errorf("verify percentage with cap proof: invalid proof data")
	}
	percentage, err := service.ristrettoPoint(data.Context[:32], "percentage commitment")
	if err != nil {
		return fmt.Errorf("verify percentage with cap proof: %w", err)
	}
	delta, err := service.ristrettoPoint(data.Context[32:64], "delta commitment")
	if err != nil {
		return fmt.Errorf("verify percentage with cap proof: %w", err)
	}
	claimed, err := service.ristrettoPoint(data.Context[64:96], "claimed commitment")
	if err != nil {
		return fmt.Errorf("verify percentage with cap proof: %w", err)
	}
	if percentage.IsIdentity() || delta.IsIdentity() || claimed.IsIdentity() {
		return fmt.Errorf("verify percentage with cap proof: identity context point")
	}
	proof, err := service.parsePercentageWithCapProof(data.Proof)
	if err != nil {
		return fmt.Errorf("verify percentage with cap proof: %w", err)
	}
	transcript := service.percentageWithCapTranscript(data.Context)
	service.appendTranscriptPoint(transcript, "Y_max_proof", data.Proof[:32])
	service.appendTranscriptPoint(transcript, "Y_delta", data.Proof[96:128])
	service.appendTranscriptPoint(transcript, "Y_claimed", data.Proof[128:160])
	challenge, err := service.transcriptChallengeScalar(transcript, "c")
	if err != nil {
		return fmt.Errorf("verify percentage with cap proof: %w", err)
	}
	equalityChallenge := scalar.New().Sub(challenge, proof.cMax)
	service.appendTranscriptScalar(transcript, "z_max", data.Proof[32:64])
	service.appendTranscriptScalar(transcript, "c_max_proof", data.Proof[64:96])
	service.appendTranscriptScalar(transcript, "z_x", data.Proof[160:192])
	service.appendTranscriptScalar(transcript, "z_delta_real", data.Proof[192:224])
	service.appendTranscriptScalar(transcript, "z_claimed", data.Proof[224:])
	batching, err := service.transcriptChallengeScalar(transcript, "w")
	if err != nil {
		return fmt.Errorf("verify percentage with cap proof: %w", err)
	}
	openingBasepoint, err := service.pedersenOpeningBasepoint()
	if err != nil {
		return fmt.Errorf("verify percentage with cap proof: %w", err)
	}
	maxValue := scalar.New().SetUint64(binary.LittleEndian.Uint64(data.Context[96:]))
	batchingSquared := scalar.New().Mul(batching, batching)
	check := curve.NewRistrettoPoint().MultiscalarMulVartime(
		[]*scalar.Scalar{
			proof.cMax,
			scalar.New().Neg(scalar.New().Mul(proof.cMax, maxValue)),
			scalar.New().Neg(proof.zMax),
			scalar.New().One(),
			scalar.New().Mul(batching, proof.zX),
			scalar.New().Mul(batching, proof.zDelta),
			scalar.New().Neg(scalar.New().Mul(batching, equalityChallenge)),
			scalar.New().Neg(batching),
			scalar.New().Mul(batchingSquared, proof.zX),
			scalar.New().Mul(batchingSquared, proof.zClaimed),
			scalar.New().Neg(scalar.New().Mul(batchingSquared, equalityChallenge)),
			scalar.New().Neg(batchingSquared),
		},
		[]*curve.RistrettoPoint{
			percentage, curve.RISTRETTO_BASEPOINT_POINT, openingBasepoint, proof.yMax,
			curve.RISTRETTO_BASEPOINT_POINT, openingBasepoint, delta, proof.yDelta,
			curve.RISTRETTO_BASEPOINT_POINT, openingBasepoint, claimed, proof.yClaimed,
		},
	)
	if !check.IsIdentity() {
		return fmt.Errorf("verify percentage with cap proof: algebraic relation")
	}
	return nil
}

func (service ConfidentialTransferService) percentageWithCapAboveProof(percentageOpening *scalar.Scalar, delta, claimed, openingBasepoint *curve.RistrettoPoint, context []byte, random io.Reader) (percentageWithCapProofValues, error) {
	zX, zDelta, zClaimed, equalityChallenge, yMaxScalar, err := service.percentageWithCapScalars(random)
	if err != nil {
		return percentageWithCapProofValues{}, err
	}
	yDelta := curve.NewRistrettoPoint().MultiscalarMul(
		[]*scalar.Scalar{zX, zDelta, scalar.New().Neg(equalityChallenge)},
		[]*curve.RistrettoPoint{curve.RISTRETTO_BASEPOINT_POINT, openingBasepoint, delta},
	)
	yClaimed := curve.NewRistrettoPoint().MultiscalarMul(
		[]*scalar.Scalar{zX, zClaimed, scalar.New().Neg(equalityChallenge)},
		[]*curve.RistrettoPoint{curve.RISTRETTO_BASEPOINT_POINT, openingBasepoint, claimed},
	)
	yMax := curve.NewRistrettoPoint().Mul(openingBasepoint, yMaxScalar)
	transcript := service.percentageWithCapTranscript(context)
	service.appendPercentageWithCapCommitments(transcript, yMax, yDelta, yClaimed)
	challenge, err := service.transcriptChallengeScalar(transcript, "c")
	if err != nil {
		return percentageWithCapProofValues{}, err
	}
	cMax := scalar.New().Sub(challenge, equalityChallenge)
	zMax := scalar.New().Add(scalar.New().Mul(cMax, percentageOpening), yMaxScalar)
	return percentageWithCapProofValues{yMax: yMax, zMax: zMax, cMax: cMax, yDelta: yDelta, yClaimed: yClaimed, zX: zX, zDelta: zDelta, zClaimed: zClaimed}, nil
}

func (service ConfidentialTransferService) percentageWithCapBelowProof(percentage *curve.RistrettoPoint, deltaOpening *scalar.Scalar, deltaAmount uint64, claimedOpening *scalar.Scalar, maxValue uint64, openingBasepoint *curve.RistrettoPoint, context []byte, random io.Reader) (percentageWithCapProofValues, error) {
	zMax, cMax, yX, yDeltaScalar, yClaimedScalar, err := service.percentageWithCapScalars(random)
	if err != nil {
		return percentageWithCapProofValues{}, err
	}
	yMax := curve.NewRistrettoPoint().MultiscalarMul(
		[]*scalar.Scalar{zMax, scalar.New().Neg(cMax), scalar.New().Mul(cMax, scalar.New().SetUint64(maxValue))},
		[]*curve.RistrettoPoint{openingBasepoint, percentage, curve.RISTRETTO_BASEPOINT_POINT},
	)
	yDelta := curve.NewRistrettoPoint().MultiscalarMul(
		[]*scalar.Scalar{yX, yDeltaScalar},
		[]*curve.RistrettoPoint{curve.RISTRETTO_BASEPOINT_POINT, openingBasepoint},
	)
	yClaimed := curve.NewRistrettoPoint().MultiscalarMul(
		[]*scalar.Scalar{yX, yClaimedScalar},
		[]*curve.RistrettoPoint{curve.RISTRETTO_BASEPOINT_POINT, openingBasepoint},
	)
	transcript := service.percentageWithCapTranscript(context)
	service.appendPercentageWithCapCommitments(transcript, yMax, yDelta, yClaimed)
	challenge, err := service.transcriptChallengeScalar(transcript, "c")
	if err != nil {
		return percentageWithCapProofValues{}, err
	}
	equalityChallenge := scalar.New().Sub(challenge, cMax)
	zX := scalar.New().Add(scalar.New().Mul(equalityChallenge, scalar.New().SetUint64(deltaAmount)), yX)
	zDelta := scalar.New().Add(scalar.New().Mul(equalityChallenge, deltaOpening), yDeltaScalar)
	zClaimed := scalar.New().Add(scalar.New().Mul(equalityChallenge, claimedOpening), yClaimedScalar)
	return percentageWithCapProofValues{yMax: yMax, zMax: zMax, cMax: cMax, yDelta: yDelta, yClaimed: yClaimed, zX: zX, zDelta: zDelta, zClaimed: zClaimed}, nil
}

func (ConfidentialTransferService) percentageWithCapScalars(random io.Reader) (*scalar.Scalar, *scalar.Scalar, *scalar.Scalar, *scalar.Scalar, *scalar.Scalar, error) {
	values := [5]*scalar.Scalar{}
	for i := range values {
		value, err := scalar.New().SetRandom(random)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		values[i] = value
	}
	return values[0], values[1], values[2], values[3], values[4], nil
}

func (service ConfidentialTransferService) percentageWithCapContext(percentage, delta, claimed PedersenCommitment, maxValue uint64) []byte {
	context := make([]byte, 104)
	copy(context[:32], percentage[:])
	copy(context[32:64], delta[:])
	copy(context[64:96], claimed[:])
	binary.LittleEndian.PutUint64(context[96:], maxValue)
	return context
}

func (service ConfidentialTransferService) percentageWithCapTranscript(context []byte) *merlin.Transcript {
	transcript := service.newZKElGamalTranscript("percentage-with-cap-instruction")
	transcript.AppendMessage([]byte("percentage-commitment"), context[:32])
	transcript.AppendMessage([]byte("delta-commitment"), context[32:64])
	transcript.AppendMessage([]byte("claimed-commitment"), context[64:96])
	service.appendTranscriptUint64(transcript, "max-value", binary.LittleEndian.Uint64(context[96:]))
	service.appendProofDomainSeparator(transcript, "percentage-with-cap-proof")
	return transcript
}

func (service ConfidentialTransferService) appendPercentageWithCapCommitments(transcript *merlin.Transcript, yMax, yDelta, yClaimed *curve.RistrettoPoint) {
	maxCommitment := service.pedersenCommitment(yMax)
	deltaCommitment := service.pedersenCommitment(yDelta)
	claimedCommitment := service.pedersenCommitment(yClaimed)
	service.appendTranscriptPoint(transcript, "Y_max_proof", maxCommitment[:])
	service.appendTranscriptPoint(transcript, "Y_delta", deltaCommitment[:])
	service.appendTranscriptPoint(transcript, "Y_claimed", claimedCommitment[:])
}

func (service ConfidentialTransferService) percentageWithCapProofBytes(values percentageWithCapProofValues) (PercentageWithCapProof, error) {
	proof := PercentageWithCapProof{}
	yMax := service.pedersenCommitment(values.yMax)
	yDelta := service.pedersenCommitment(values.yDelta)
	yClaimed := service.pedersenCommitment(values.yClaimed)
	copy(proof[:32], yMax[:])
	copy(proof[96:128], yDelta[:])
	copy(proof[128:160], yClaimed[:])
	scalars := [5]*scalar.Scalar{values.zMax, values.cMax, values.zX, values.zDelta, values.zClaimed}
	offsets := [5]int{32, 64, 160, 192, 224}
	for i, value := range scalars {
		if err := value.ToBytes(proof[offsets[i] : offsets[i]+32]); err != nil {
			return PercentageWithCapProof{}, err
		}
	}
	return proof, nil
}

func (service ConfidentialTransferService) parsePercentageWithCapProof(data []byte) (percentageWithCapProofValues, error) {
	yMax, err := service.ristrettoPoint(data[:32], "maximum proof commitment")
	if err != nil {
		return percentageWithCapProofValues{}, err
	}
	yDelta, err := service.ristrettoPoint(data[96:128], "delta proof commitment")
	if err != nil {
		return percentageWithCapProofValues{}, err
	}
	yClaimed, err := service.ristrettoPoint(data[128:160], "claimed proof commitment")
	if err != nil {
		return percentageWithCapProofValues{}, err
	}
	if yMax.IsIdentity() || yDelta.IsIdentity() || yClaimed.IsIdentity() {
		return percentageWithCapProofValues{}, fmt.Errorf("identity proof commitment")
	}
	zMax, err := service.scalarFromBytes(data[32:64], "maximum proof response")
	if err != nil {
		return percentageWithCapProofValues{}, err
	}
	cMax, err := service.scalarFromBytes(data[64:96], "maximum proof challenge")
	if err != nil {
		return percentageWithCapProofValues{}, err
	}
	zX, err := service.scalarFromBytes(data[160:192], "equality proof amount response")
	if err != nil {
		return percentageWithCapProofValues{}, err
	}
	zDelta, err := service.scalarFromBytes(data[192:224], "equality proof delta response")
	if err != nil {
		return percentageWithCapProofValues{}, err
	}
	zClaimed, err := service.scalarFromBytes(data[224:], "equality proof claimed response")
	if err != nil {
		return percentageWithCapProofValues{}, err
	}
	return percentageWithCapProofValues{yMax: yMax, zMax: zMax, cMax: cMax, yDelta: yDelta, yClaimed: yClaimed, zX: zX, zDelta: zDelta, zClaimed: zClaimed}, nil
}
