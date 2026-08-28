package token2022

import (
	"fmt"

	"github.com/gtank/merlin"
	"github.com/oasisprotocol/curve25519-voi/curve"
	"github.com/oasisprotocol/curve25519-voi/curve/scalar"
)

type rangeInnerProductProof struct {
	left  []PedersenCommitment
	right []PedersenCommitment
	a     *scalar.Scalar
	b     *scalar.Scalar
}

func (service ConfidentialTransferService) newRangeInnerProductProof(q *curve.RistrettoPoint, hFactors []*scalar.Scalar, generators rangeProofGenerators, a, b []*scalar.Scalar, transcript *merlin.Transcript) (rangeInnerProductProof, error) {
	size := len(generators.g)
	if len(generators.h) != size || len(a) != size || len(b) != size || len(hFactors) != size || size == 0 || size&(size-1) != 0 {
		return rangeInnerProductProof{}, fmt.Errorf("generate range inner product proof: invalid vector length")
	}
	service.appendRangeInnerProductDomainSeparator(transcript, uint64(size))
	g := service.clonePoints(generators.g)
	h := service.clonePoints(generators.h)
	a = service.cloneScalars(a)
	b = service.cloneScalars(b)
	proof := rangeInnerProductProof{
		left:  make([]PedersenCommitment, 0, service.rangeLog2(size)),
		right: make([]PedersenCommitment, 0, service.rangeLog2(size)),
	}
	first := true
	for size > 1 {
		half := size / 2
		aLeft, aRight := a[:half], a[half:size]
		bLeft, bRight := b[:half], b[half:size]
		gLeft, gRight := g[:half], g[half:size]
		hLeft, hRight := h[:half], h[half:size]
		cLeft, err := service.scalarInnerProduct(aLeft, bRight)
		if err != nil {
			return rangeInnerProductProof{}, err
		}
		cRight, err := service.scalarInnerProduct(aRight, bLeft)
		if err != nil {
			return rangeInnerProductProof{}, err
		}
		leftScalars := make([]*scalar.Scalar, 0, half*2+1)
		rightScalars := make([]*scalar.Scalar, 0, half*2+1)
		if first {
			for index := 0; index < half; index++ {
				leftScalars = append(leftScalars, aLeft[index])
				rightScalars = append(rightScalars, aRight[index])
			}
			for index := 0; index < half; index++ {
				leftScalars = append(leftScalars, scalar.New().Mul(bRight[index], hFactors[index]))
				rightScalars = append(rightScalars, scalar.New().Mul(bLeft[index], hFactors[half+index]))
			}
		} else {
			leftScalars = append(leftScalars, aLeft...)
			leftScalars = append(leftScalars, bRight...)
			rightScalars = append(rightScalars, aRight...)
			rightScalars = append(rightScalars, bLeft...)
		}
		leftScalars = append(leftScalars, cLeft)
		rightScalars = append(rightScalars, cRight)
		leftPoints := make([]*curve.RistrettoPoint, 0, half*2+1)
		leftPoints = append(leftPoints, gRight...)
		leftPoints = append(leftPoints, hLeft...)
		leftPoints = append(leftPoints, q)
		rightPoints := make([]*curve.RistrettoPoint, 0, half*2+1)
		rightPoints = append(rightPoints, gLeft...)
		rightPoints = append(rightPoints, hRight...)
		rightPoints = append(rightPoints, q)
		left := service.pedersenCommitment(curve.NewRistrettoPoint().MultiscalarMul(leftScalars, leftPoints))
		right := service.pedersenCommitment(curve.NewRistrettoPoint().MultiscalarMul(rightScalars, rightPoints))
		proof.left = append(proof.left, left)
		proof.right = append(proof.right, right)
		service.appendTranscriptPoint(transcript, "L", left[:])
		service.appendTranscriptPoint(transcript, "R", right[:])
		u, err := service.transcriptChallengeScalar(transcript, "u")
		if err != nil {
			return rangeInnerProductProof{}, err
		}
		uInverse := scalar.New().Invert(u)
		for index := 0; index < half; index++ {
			aLeft[index] = scalar.New().Add(scalar.New().Mul(aLeft[index], u), scalar.New().Mul(uInverse, aRight[index]))
			bLeft[index] = scalar.New().Add(scalar.New().Mul(bLeft[index], uInverse), scalar.New().Mul(u, bRight[index]))
			if first {
				gLeft[index] = curve.NewRistrettoPoint().MultiscalarMul(
					[]*scalar.Scalar{uInverse, u},
					[]*curve.RistrettoPoint{gLeft[index], gRight[index]},
				)
				hLeft[index] = curve.NewRistrettoPoint().MultiscalarMul(
					[]*scalar.Scalar{scalar.New().Mul(u, hFactors[index]), scalar.New().Mul(uInverse, hFactors[half+index])},
					[]*curve.RistrettoPoint{hLeft[index], hRight[index]},
				)
			} else {
				gLeft[index] = curve.NewRistrettoPoint().MultiscalarMul([]*scalar.Scalar{uInverse, u}, []*curve.RistrettoPoint{gLeft[index], gRight[index]})
				hLeft[index] = curve.NewRistrettoPoint().MultiscalarMul([]*scalar.Scalar{u, uInverse}, []*curve.RistrettoPoint{hLeft[index], hRight[index]})
			}
		}
		a, b, g, h = aLeft, bLeft, gLeft, hLeft
		size = half
		first = false
	}
	proof.a = a[0]
	proof.b = b[0]
	return proof, nil
}

func (service ConfidentialTransferService) rangeInnerProductVerificationScalars(proof rangeInnerProductProof, size int, transcript *merlin.Transcript) ([]*scalar.Scalar, []*scalar.Scalar, []*scalar.Scalar, error) {
	logSize := len(proof.left)
	if logSize == 0 || logSize != len(proof.right) || logSize >= 32 || size != 1<<logSize {
		return nil, nil, nil, fmt.Errorf("verify range inner product proof: invalid vector length")
	}
	service.appendRangeInnerProductDomainSeparator(transcript, uint64(size))
	challenges := make([]*scalar.Scalar, logSize)
	for index := range challenges {
		left, err := service.ristrettoPoint(proof.left[index][:], "range inner product L")
		if err != nil || left.IsIdentity() {
			return nil, nil, nil, fmt.Errorf("verify range inner product proof: invalid L point")
		}
		right, err := service.ristrettoPoint(proof.right[index][:], "range inner product R")
		if err != nil || right.IsIdentity() {
			return nil, nil, nil, fmt.Errorf("verify range inner product proof: invalid R point")
		}
		service.appendTranscriptPoint(transcript, "L", proof.left[index][:])
		service.appendTranscriptPoint(transcript, "R", proof.right[index][:])
		challenges[index], err = service.transcriptChallengeScalar(transcript, "u")
		if err != nil {
			return nil, nil, nil, err
		}
	}
	inverses := service.cloneScalars(challenges)
	allInverse := scalar.New().BatchInvert(inverses)
	challengeSquares := make([]*scalar.Scalar, logSize)
	inverseSquares := make([]*scalar.Scalar, logSize)
	for index := range challenges {
		challengeSquares[index] = scalar.New().Mul(challenges[index], challenges[index])
		inverseSquares[index] = scalar.New().Mul(inverses[index], inverses[index])
	}
	s := make([]*scalar.Scalar, size)
	s[0] = allInverse
	for index := 1; index < size; index++ {
		logIndex := service.rangeLog2(index)
		power := 1 << logIndex
		challenge := challengeSquares[logSize-1-logIndex]
		s[index] = scalar.New().Mul(s[index-power], challenge)
	}
	return challengeSquares, inverseSquares, s, nil
}

func (service ConfidentialTransferService) marshalRangeInnerProductProof(proof rangeInnerProductProof) ([]byte, error) {
	data := make([]byte, 0, (len(proof.left)*2+2)*32)
	for index := range proof.left {
		data = append(data, proof.left[index][:]...)
		data = append(data, proof.right[index][:]...)
	}
	a, err := proof.a.MarshalBinary()
	if err != nil {
		return nil, err
	}
	b, err := proof.b.MarshalBinary()
	if err != nil {
		return nil, err
	}
	data = append(data, a...)
	data = append(data, b...)
	return data, nil
}

func (service ConfidentialTransferService) parseRangeInnerProductProof(data []byte) (rangeInnerProductProof, error) {
	if len(data)%32 != 0 || len(data) < 64 || (len(data)/32-2)%2 != 0 {
		return rangeInnerProductProof{}, fmt.Errorf("decode range inner product proof: invalid length")
	}
	logSize := (len(data)/32 - 2) / 2
	if logSize >= 32 {
		return rangeInnerProductProof{}, fmt.Errorf("decode range inner product proof: invalid length")
	}
	proof := rangeInnerProductProof{
		left:  make([]PedersenCommitment, logSize),
		right: make([]PedersenCommitment, logSize),
	}
	for index := 0; index < logSize; index++ {
		copy(proof.left[index][:], data[index*64:index*64+32])
		copy(proof.right[index][:], data[index*64+32:index*64+64])
	}
	position := logSize * 64
	var err error
	proof.a, err = service.scalarFromBytes(data[position:position+32], "range inner product a")
	if err != nil {
		return rangeInnerProductProof{}, err
	}
	proof.b, err = service.scalarFromBytes(data[position+32:], "range inner product b")
	if err != nil {
		return rangeInnerProductProof{}, err
	}
	return proof, nil
}

func (ConfidentialTransferService) rangeLog2(value int) int {
	logarithm := 0
	for value > 1 {
		value >>= 1
		logarithm++
	}
	return logarithm
}

func (service ConfidentialTransferService) appendRangeInnerProductDomainSeparator(transcript *merlin.Transcript, size uint64) {
	service.appendProofDomainSeparator(transcript, "inner-product")
	service.appendTranscriptUint64(transcript, "n", size)
}
