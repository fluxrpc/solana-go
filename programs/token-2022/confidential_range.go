package token2022

import (
	"crypto/rand"
	"fmt"
	"io"

	"github.com/gtank/merlin"
	"github.com/oasisprotocol/curve25519-voi/curve"
	"github.com/oasisprotocol/curve25519-voi/curve/scalar"
)

type batchedRangeProof struct {
	a            PedersenCommitment
	s            PedersenCommitment
	t1           PedersenCommitment
	t2           PedersenCommitment
	tx           *scalar.Scalar
	txBlinding   *scalar.Scalar
	eBlinding    *scalar.Scalar
	innerProduct rangeInnerProductProof
}

func (service ConfidentialTransferService) GenerateBatchedRangeProof(commitments []PedersenCommitment, amounts []uint64, bitLengths []uint8, openings []PedersenOpening) (ZKProofData, error) {
	return service.generateBatchedRangeProofWithReader(commitments, amounts, bitLengths, openings, rand.Reader)
}

func (service ConfidentialTransferService) generateBatchedRangeProofWithReader(commitments []PedersenCommitment, amounts []uint64, bitLengths []uint8, openings []PedersenOpening, random io.Reader) (ZKProofData, error) {
	discriminator, proofSize, totalBits, err := service.batchedRangeProofParameters(bitLengths)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate batched range proof: %w", err)
	}
	context, openingScalars, lengths, err := service.batchedRangeProofContext(commitments, amounts, bitLengths, openings)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate batched range proof: %w", err)
	}
	transcript := service.batchedRangeProofTranscript(context)
	proof, err := service.newBatchedRangeProof(amounts, lengths, openingScalars, totalBits, transcript, random)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate batched range proof: %w", err)
	}
	encoded, err := service.marshalBatchedRangeProof(proof)
	if err != nil {
		return ZKProofData{}, fmt.Errorf("generate batched range proof: %w", err)
	}
	if len(encoded) != proofSize {
		return ZKProofData{}, fmt.Errorf("generate batched range proof: invalid proof length %d", len(encoded))
	}
	return ZKProofData{Discriminator: discriminator, Context: context, Proof: encoded}, nil
}

func (service ConfidentialTransferService) VerifyBatchedRangeProof(data ZKProofData) error {
	proofSize, totalBits, err := service.batchedRangeVerificationParameters(data.Discriminator)
	if err != nil || len(data.Context) != 264 || len(data.Proof) != proofSize {
		return fmt.Errorf("verify batched range proof: invalid proof data")
	}
	commitments, bitLengths, err := service.parseBatchedRangeProofContext(data.Context)
	if err != nil {
		return fmt.Errorf("verify batched range proof: %w", err)
	}
	sum := 0
	for _, bitLength := range bitLengths {
		sum += bitLength
	}
	if sum != totalBits {
		return fmt.Errorf("verify batched range proof: invalid commitment length")
	}
	proof, err := service.parseBatchedRangeProof(data.Proof)
	if err != nil {
		return fmt.Errorf("verify batched range proof: %w", err)
	}
	transcript := service.batchedRangeProofTranscript(data.Context)
	if err := service.verifyBatchedRangeProof(proof, commitments, bitLengths, totalBits, transcript); err != nil {
		return fmt.Errorf("verify batched range proof: %w", err)
	}
	return nil
}

func (ConfidentialTransferService) batchedRangeProofParameters(bitLengths []uint8) (uint8, int, int, error) {
	total := 0
	for _, bitLength := range bitLengths {
		total += int(bitLength)
	}
	switch total {
	case 64:
		return 6, 672, 64, nil
	case 128:
		return 7, 736, 128, nil
	case 256:
		return 8, 800, 256, nil
	default:
		return 0, 0, 0, fmt.Errorf("bit lengths must total 64, 128, or 256")
	}
}

func (ConfidentialTransferService) batchedRangeVerificationParameters(discriminator uint8) (int, int, error) {
	switch discriminator {
	case 6:
		return 672, 64, nil
	case 7:
		return 736, 128, nil
	case 8:
		return 800, 256, nil
	default:
		return 0, 0, fmt.Errorf("invalid discriminator")
	}
}

func (service ConfidentialTransferService) batchedRangeProofContext(commitments []PedersenCommitment, amounts []uint64, bitLengths []uint8, openings []PedersenOpening) ([]byte, []*scalar.Scalar, []int, error) {
	count := len(commitments)
	if count == 0 || count > 8 || len(amounts) != count || len(bitLengths) != count || len(openings) != count {
		return nil, nil, nil, fmt.Errorf("illegal commitment length")
	}
	context := make([]byte, 264)
	openingScalars := make([]*scalar.Scalar, count)
	lengths := make([]int, count)
	for index := 0; index < count; index++ {
		point, err := service.ristrettoPoint(commitments[index][:], "range proof commitment")
		if err != nil || point.IsIdentity() {
			return nil, nil, nil, fmt.Errorf("invalid commitment")
		}
		if bitLengths[index] == 0 || bitLengths[index] > 64 {
			return nil, nil, nil, fmt.Errorf("illegal amount bit length")
		}
		openingScalars[index], err = service.scalarFromBytes(openings[index][:], "range proof opening")
		if err != nil {
			return nil, nil, nil, err
		}
		copy(context[index*32:(index+1)*32], commitments[index][:])
		context[256+index] = bitLengths[index]
		lengths[index] = int(bitLengths[index])
	}
	return context, openingScalars, lengths, nil
}

func (service ConfidentialTransferService) parseBatchedRangeProofContext(context []byte) ([]*curve.RistrettoPoint, []int, error) {
	commitments := make([]*curve.RistrettoPoint, 0, 8)
	lengths := make([]int, 0, 8)
	padding := false
	for index := 0; index < 8; index++ {
		encoded := context[index*32 : (index+1)*32]
		zero := true
		for _, value := range encoded {
			zero = zero && value == 0
		}
		if zero {
			padding = true
			if context[256+index] != 0 {
				return nil, nil, fmt.Errorf("invalid range proof context padding")
			}
			continue
		}
		if padding {
			return nil, nil, fmt.Errorf("invalid range proof context padding")
		}
		point, err := service.ristrettoPoint(encoded, "range proof commitment")
		if err != nil || point.IsIdentity() {
			return nil, nil, fmt.Errorf("invalid range proof commitment")
		}
		bitLength := int(context[256+index])
		if bitLength == 0 || bitLength > 64 {
			return nil, nil, fmt.Errorf("illegal amount bit length")
		}
		commitments = append(commitments, point)
		lengths = append(lengths, bitLength)
	}
	if len(commitments) == 0 {
		return nil, nil, fmt.Errorf("illegal commitment length")
	}
	return commitments, lengths, nil
}

func (service ConfidentialTransferService) batchedRangeProofTranscript(context []byte) *merlin.Transcript {
	transcript := service.newZKElGamalTranscript("batched-range-proof-instruction")
	transcript.AppendMessage([]byte("commitments"), context[:256])
	transcript.AppendMessage([]byte("bit-lengths"), context[256:])
	return transcript
}

func (service ConfidentialTransferService) newBatchedRangeProof(amounts []uint64, bitLengths []int, openings []*scalar.Scalar, totalBits int, transcript *merlin.Transcript, random io.Reader) (batchedRangeProof, error) {
	if len(amounts) != len(bitLengths) || len(amounts) != len(openings) || totalBits == 0 || totalBits&(totalBits-1) != 0 {
		return batchedRangeProof{}, fmt.Errorf("invalid vector length")
	}
	generators, err := service.rangeGenerators(totalBits)
	if err != nil {
		return batchedRangeProof{}, err
	}
	service.appendRangeProofDomainSeparator(transcript, uint64(totalBits))
	openingBasepoint, err := service.pedersenOpeningBasepoint()
	if err != nil {
		return batchedRangeProof{}, err
	}
	aBlinding, err := scalar.New().SetRandom(random)
	if err != nil {
		return batchedRangeProof{}, err
	}
	aPoint := curve.NewRistrettoPoint().Mul(openingBasepoint, aBlinding)
	position := 0
	for amountIndex, amount := range amounts {
		for bitIndex := 0; bitIndex < bitLengths[amountIndex]; bitIndex++ {
			if amount>>bitIndex&1 == 1 {
				aPoint.Add(aPoint, generators.g[position])
			} else {
				aPoint.Sub(aPoint, generators.h[position])
			}
			position++
		}
	}
	aCommitment := service.pedersenCommitment(aPoint)
	sLeft, err := service.randomScalars(totalBits, random)
	if err != nil {
		return batchedRangeProof{}, err
	}
	sRight, err := service.randomScalars(totalBits, random)
	if err != nil {
		return batchedRangeProof{}, err
	}
	sBlinding, err := scalar.New().SetRandom(random)
	if err != nil {
		return batchedRangeProof{}, err
	}
	sScalars := make([]*scalar.Scalar, 0, totalBits*2+1)
	sScalars = append(sScalars, sBlinding)
	sScalars = append(sScalars, sLeft...)
	sScalars = append(sScalars, sRight...)
	sPoints := make([]*curve.RistrettoPoint, 0, totalBits*2+1)
	sPoints = append(sPoints, openingBasepoint)
	sPoints = append(sPoints, generators.g...)
	sPoints = append(sPoints, generators.h...)
	sCommitment := service.pedersenCommitment(curve.NewRistrettoPoint().MultiscalarMul(sScalars, sPoints))
	service.appendTranscriptPoint(transcript, "A", aCommitment[:])
	service.appendTranscriptPoint(transcript, "S", sCommitment[:])
	y, err := service.transcriptChallengeScalar(transcript, "y")
	if err != nil {
		return batchedRangeProof{}, err
	}
	z, err := service.transcriptChallengeScalar(transcript, "z")
	if err != nil {
		return batchedRangeProof{}, err
	}
	leftPolynomial := service.newRangeVectorPolynomial(totalBits)
	rightPolynomial := service.newRangeVectorPolynomial(totalBits)
	position = 0
	zPower := scalar.New().Mul(z, z)
	yPower := scalar.New().One()
	for amountIndex, amount := range amounts {
		twoPower := scalar.New().One()
		for bitIndex := 0; bitIndex < bitLengths[amountIndex]; bitIndex++ {
			bit := scalar.New().SetUint64(amount >> bitIndex & 1)
			rightBit := scalar.New().Sub(bit, scalar.New().One())
			leftPolynomial.constant[position] = scalar.New().Sub(bit, z)
			leftPolynomial.linear[position] = sLeft[position]
			rightPolynomial.constant[position] = scalar.New().Add(
				scalar.New().Mul(yPower, scalar.New().Add(rightBit, z)),
				scalar.New().Mul(zPower, twoPower),
			)
			rightPolynomial.linear[position] = scalar.New().Mul(yPower, sRight[position])
			yPower.Mul(yPower, y)
			twoPower.Add(twoPower, twoPower)
			position++
		}
		zPower.Mul(zPower, z)
	}
	t0, t1, t2, err := service.rangeVectorPolynomialInnerProduct(leftPolynomial, rightPolynomial)
	if err != nil {
		return batchedRangeProof{}, err
	}
	t1Blinding, err := scalar.New().SetRandom(random)
	if err != nil {
		return batchedRangeProof{}, err
	}
	t2Blinding, err := scalar.New().SetRandom(random)
	if err != nil {
		return batchedRangeProof{}, err
	}
	t1Commitment := service.pedersenCommitment(curve.NewRistrettoPoint().MultiscalarMul(
		[]*scalar.Scalar{t1, t1Blinding},
		[]*curve.RistrettoPoint{curve.RISTRETTO_BASEPOINT_POINT, openingBasepoint},
	))
	t2Commitment := service.pedersenCommitment(curve.NewRistrettoPoint().MultiscalarMul(
		[]*scalar.Scalar{t2, t2Blinding},
		[]*curve.RistrettoPoint{curve.RISTRETTO_BASEPOINT_POINT, openingBasepoint},
	))
	service.appendTranscriptPoint(transcript, "T_1", t1Commitment[:])
	service.appendTranscriptPoint(transcript, "T_2", t2Commitment[:])
	x, err := service.transcriptChallengeScalar(transcript, "x")
	if err != nil {
		return batchedRangeProof{}, err
	}
	aggregatedOpening := scalar.New()
	zPower.Set(z)
	for _, opening := range openings {
		zPower.Mul(zPower, z)
		aggregatedOpening.Add(aggregatedOpening, scalar.New().Mul(zPower, opening))
	}
	tx := service.evaluateRangeScalarPolynomial(t0, t1, t2, x)
	txBlinding := service.evaluateRangeScalarPolynomial(aggregatedOpening, t1Blinding, t2Blinding, x)
	service.appendRangeScalar(transcript, "t_x", tx)
	service.appendRangeScalar(transcript, "t_x_blinding", txBlinding)
	eBlinding := scalar.New().Add(aBlinding, scalar.New().Mul(sBlinding, x))
	service.appendRangeScalar(transcript, "e_blinding", eBlinding)
	w, err := service.transcriptChallengeScalar(transcript, "w")
	if err != nil {
		return batchedRangeProof{}, err
	}
	q := curve.NewRistrettoPoint().Mul(curve.RISTRETTO_BASEPOINT_POINT, w)
	gFactors := make([]*scalar.Scalar, totalBits)
	for index := range gFactors {
		gFactors[index] = scalar.New().One()
	}
	hFactors := service.scalarPowers(scalar.New().Invert(y), totalBits)
	if _, err := service.transcriptChallengeScalar(transcript, "c"); err != nil {
		return batchedRangeProof{}, err
	}
	innerProduct, err := service.newRangeInnerProductProof(
		q,
		gFactors,
		hFactors,
		generators,
		service.evaluateRangeVectorPolynomial(leftPolynomial, x),
		service.evaluateRangeVectorPolynomial(rightPolynomial, x),
		transcript,
	)
	if err != nil {
		return batchedRangeProof{}, err
	}
	service.appendRangeScalar(transcript, "ipp_a", innerProduct.a)
	service.appendRangeScalar(transcript, "ipp_b", innerProduct.b)
	if _, err := service.transcriptChallengeScalar(transcript, "d"); err != nil {
		return batchedRangeProof{}, err
	}
	return batchedRangeProof{
		a:            aCommitment,
		s:            sCommitment,
		t1:           t1Commitment,
		t2:           t2Commitment,
		tx:           tx,
		txBlinding:   txBlinding,
		eBlinding:    eBlinding,
		innerProduct: innerProduct,
	}, nil
}

func (ConfidentialTransferService) randomScalars(size int, random io.Reader) ([]*scalar.Scalar, error) {
	values := make([]*scalar.Scalar, size)
	for index := range values {
		value, err := scalar.New().SetRandom(random)
		if err != nil {
			return nil, err
		}
		values[index] = value
	}
	return values, nil
}

func (ConfidentialTransferService) evaluateRangeScalarPolynomial(constant, linear, quadratic, x *scalar.Scalar) *scalar.Scalar {
	return scalar.New().Add(constant, scalar.New().Mul(x, scalar.New().Add(linear, scalar.New().Mul(x, quadratic))))
}

func (service ConfidentialTransferService) appendRangeProofDomainSeparator(transcript *merlin.Transcript, size uint64) {
	service.appendProofDomainSeparator(transcript, "range-proof")
	service.appendTranscriptUint64(transcript, "n", size)
}

func (ConfidentialTransferService) appendRangeScalar(transcript *merlin.Transcript, label string, value *scalar.Scalar) {
	encoded, _ := value.MarshalBinary()
	transcript.AppendMessage([]byte(label), encoded)
}

func (service ConfidentialTransferService) verifyBatchedRangeProof(proof batchedRangeProof, commitments []*curve.RistrettoPoint, bitLengths []int, totalBits int, transcript *merlin.Transcript) error {
	if len(commitments) != len(bitLengths) || totalBits == 0 || totalBits&(totalBits-1) != 0 {
		return fmt.Errorf("invalid vector length")
	}
	generators, err := service.rangeGenerators(totalBits)
	if err != nil {
		return err
	}
	service.appendRangeProofDomainSeparator(transcript, uint64(totalBits))
	proofPoints := make([]*curve.RistrettoPoint, 4)
	proofCommitments := []PedersenCommitment{proof.a, proof.s, proof.t1, proof.t2}
	for index := range proofPoints {
		proofPoints[index], err = service.ristrettoPoint(proofCommitments[index][:], "range proof point")
		if err != nil || proofPoints[index].IsIdentity() {
			return fmt.Errorf("invalid range proof point")
		}
	}
	service.appendTranscriptPoint(transcript, "A", proof.a[:])
	service.appendTranscriptPoint(transcript, "S", proof.s[:])
	y, err := service.transcriptChallengeScalar(transcript, "y")
	if err != nil {
		return err
	}
	z, err := service.transcriptChallengeScalar(transcript, "z")
	if err != nil {
		return err
	}
	service.appendTranscriptPoint(transcript, "T_1", proof.t1[:])
	service.appendTranscriptPoint(transcript, "T_2", proof.t2[:])
	x, err := service.transcriptChallengeScalar(transcript, "x")
	if err != nil {
		return err
	}
	service.appendRangeScalar(transcript, "t_x", proof.tx)
	service.appendRangeScalar(transcript, "t_x_blinding", proof.txBlinding)
	service.appendRangeScalar(transcript, "e_blinding", proof.eBlinding)
	w, err := service.transcriptChallengeScalar(transcript, "w")
	if err != nil {
		return err
	}
	if _, err := service.transcriptChallengeScalar(transcript, "c"); err != nil {
		return err
	}
	challengeSquares, inverseSquares, s, err := service.rangeInnerProductVerificationScalars(proof.innerProduct, totalBits, transcript)
	if err != nil {
		return err
	}
	service.appendRangeScalar(transcript, "ipp_a", proof.innerProduct.a)
	service.appendRangeScalar(transcript, "ipp_b", proof.innerProduct.b)
	d, err := service.transcriptChallengeScalar(transcript, "d")
	if err != nil {
		return err
	}
	zSquared := scalar.New().Mul(z, z)
	minusZ := scalar.New().Neg(z)
	zAndTwo := make([]*scalar.Scalar, 0, totalBits)
	zPower := scalar.New().One()
	for _, bitLength := range bitLengths {
		twoPower := scalar.New().One()
		for bitIndex := 0; bitIndex < bitLength; bitIndex++ {
			zAndTwo = append(zAndTwo, scalar.New().Mul(zPower, twoPower))
			twoPower.Add(twoPower, twoPower)
		}
		zPower.Mul(zPower, z)
	}
	gs := make([]*scalar.Scalar, totalBits)
	hs := make([]*scalar.Scalar, totalBits)
	yInversePowers := service.scalarPowers(scalar.New().Invert(y), totalBits)
	for index := 0; index < totalBits; index++ {
		gs[index] = scalar.New().Sub(minusZ, scalar.New().Mul(proof.innerProduct.a, s[index]))
		hs[index] = scalar.New().Add(
			z,
			scalar.New().Mul(
				yInversePowers[index],
				scalar.New().Sub(
					scalar.New().Mul(zSquared, zAndTwo[index]),
					scalar.New().Mul(proof.innerProduct.b, s[totalBits-1-index]),
				),
			),
		)
	}
	basepointScalar := scalar.New().Add(
		scalar.New().Mul(w, scalar.New().Sub(proof.tx, scalar.New().Mul(proof.innerProduct.a, proof.innerProduct.b))),
		scalar.New().Mul(d, scalar.New().Sub(service.rangeProofDelta(bitLengths, y, z), proof.tx)),
	)
	valueScalars := make([]*scalar.Scalar, len(commitments))
	zPower.SetUint64(1)
	for index := range valueScalars {
		valueScalars[index] = scalar.New().Mul(d, scalar.New().Mul(zSquared, zPower))
		zPower.Mul(zPower, z)
	}
	openingBasepoint, err := service.pedersenOpeningBasepoint()
	if err != nil {
		return err
	}
	xSquared := scalar.New().Mul(x, x)
	scalars := make([]*scalar.Scalar, 0, 6+len(challengeSquares)*2+totalBits*2+len(commitments))
	scalars = append(scalars,
		scalar.New().One(),
		x,
		scalar.New().Mul(d, x),
		scalar.New().Mul(d, xSquared),
		scalar.New().Neg(scalar.New().Add(proof.eBlinding, scalar.New().Mul(d, proof.txBlinding))),
		basepointScalar,
	)
	scalars = append(scalars, challengeSquares...)
	scalars = append(scalars, inverseSquares...)
	scalars = append(scalars, gs...)
	scalars = append(scalars, hs...)
	scalars = append(scalars, valueScalars...)
	points := make([]*curve.RistrettoPoint, 0, len(scalars))
	points = append(points, proofPoints...)
	points = append(points, openingBasepoint, curve.RISTRETTO_BASEPOINT_POINT)
	for _, encoded := range proof.innerProduct.left {
		point, err := service.ristrettoPoint(encoded[:], "range inner product L")
		if err != nil {
			return err
		}
		points = append(points, point)
	}
	for _, encoded := range proof.innerProduct.right {
		point, err := service.ristrettoPoint(encoded[:], "range inner product R")
		if err != nil {
			return err
		}
		points = append(points, point)
	}
	points = append(points, generators.g...)
	points = append(points, generators.h...)
	points = append(points, commitments...)
	if len(scalars) != len(points) {
		return fmt.Errorf("range proof multiscalar length mismatch")
	}
	if !curve.NewRistrettoPoint().MultiscalarMulVartime(scalars, points).IsIdentity() {
		return fmt.Errorf("algebraic relation")
	}
	return nil
}

func (service ConfidentialTransferService) rangeProofDelta(bitLengths []int, y, z *scalar.Scalar) *scalar.Scalar {
	totalBits := 0
	for _, bitLength := range bitLengths {
		totalBits += bitLength
	}
	zSquared := scalar.New().Mul(z, z)
	delta := scalar.New().Mul(scalar.New().Sub(z, zSquared), service.scalarPowerSum(y, totalBits))
	zPower := scalar.New().Mul(zSquared, z)
	two := scalar.New().SetUint64(2)
	for _, bitLength := range bitLengths {
		delta.Sub(delta, scalar.New().Mul(zPower, service.scalarPowerSum(two, bitLength)))
		zPower.Mul(zPower, z)
	}
	return delta
}

func (service ConfidentialTransferService) marshalBatchedRangeProof(proof batchedRangeProof) ([]byte, error) {
	data := make([]byte, 0, 224+(len(proof.innerProduct.left)*2+2)*32)
	data = append(data, proof.a[:]...)
	data = append(data, proof.s[:]...)
	data = append(data, proof.t1[:]...)
	data = append(data, proof.t2[:]...)
	for _, value := range []*scalar.Scalar{proof.tx, proof.txBlinding, proof.eBlinding} {
		encoded, err := value.MarshalBinary()
		if err != nil {
			return nil, err
		}
		data = append(data, encoded...)
	}
	innerProduct, err := service.marshalRangeInnerProductProof(proof.innerProduct)
	if err != nil {
		return nil, err
	}
	return append(data, innerProduct...), nil
}

func (service ConfidentialTransferService) parseBatchedRangeProof(data []byte) (batchedRangeProof, error) {
	if len(data) < 224 || len(data)%32 != 0 {
		return batchedRangeProof{}, fmt.Errorf("decode batched range proof: invalid length")
	}
	proof := batchedRangeProof{}
	copy(proof.a[:], data[:32])
	copy(proof.s[:], data[32:64])
	copy(proof.t1[:], data[64:96])
	copy(proof.t2[:], data[96:128])
	var err error
	proof.tx, err = service.scalarFromBytes(data[128:160], "range proof t_x")
	if err != nil {
		return batchedRangeProof{}, err
	}
	proof.txBlinding, err = service.scalarFromBytes(data[160:192], "range proof t_x blinding")
	if err != nil {
		return batchedRangeProof{}, err
	}
	proof.eBlinding, err = service.scalarFromBytes(data[192:224], "range proof e blinding")
	if err != nil {
		return batchedRangeProof{}, err
	}
	proof.innerProduct, err = service.parseRangeInnerProductProof(data[224:])
	if err != nil {
		return batchedRangeProof{}, err
	}
	return proof, nil
}
