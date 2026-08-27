package token2022

import (
	"crypto/rand"
	"crypto/sha3"
	"fmt"

	"github.com/oasisprotocol/curve25519-voi/curve"
	"github.com/oasisprotocol/curve25519-voi/curve/scalar"
)

type ElGamalKeypair struct {
	PublicKey ElGamalPubkey
	SecretKey ElGamalSecretKey
}

type PedersenOpening [32]byte

type PedersenCommitment [32]byte

type DecryptHandle [32]byte

type GroupedElGamalCiphertext2Handles [96]byte

type GroupedElGamalCiphertext3Handles [128]byte

func (service ConfidentialTransferService) GenerateElGamalKeypair() (ElGamalKeypair, error) {
	secretScalar, err := scalar.New().SetRandom(rand.Reader)
	if err != nil {
		return ElGamalKeypair{}, fmt.Errorf("generate ElGamal keypair: %w", err)
	}
	secret, err := service.elGamalSecretKey(secretScalar)
	if err != nil {
		return ElGamalKeypair{}, fmt.Errorf("generate ElGamal keypair: %w", err)
	}
	return service.ElGamalKeypairFromSecret(secret)
}

func (service ConfidentialTransferService) ElGamalKeypairFromSecret(secret ElGamalSecretKey) (ElGamalKeypair, error) {
	publicKey, err := service.DeriveElGamalPubkey(secret)
	if err != nil {
		return ElGamalKeypair{}, err
	}
	return ElGamalKeypair{PublicKey: publicKey, SecretKey: secret}, nil
}

func (service ConfidentialTransferService) ElGamalSecretKeyFromBytes(data []byte) (ElGamalSecretKey, error) {
	value, err := service.scalarFromBytes(data, "ElGamal secret key")
	if err != nil {
		return ElGamalSecretKey{}, err
	}
	if value.Equal(scalar.New()) == 1 {
		return ElGamalSecretKey{}, fmt.Errorf("decode ElGamal secret key: zero scalar")
	}
	return service.elGamalSecretKey(value)
}

func (service ConfidentialTransferService) ElGamalPubkeyFromBytes(data []byte) (ElGamalPubkey, error) {
	if _, err := service.ristrettoPoint(data, "ElGamal public key"); err != nil {
		return ElGamalPubkey{}, err
	}
	publicKey := ElGamalPubkey{}
	copy(publicKey[:], data)
	return publicKey, nil
}

func (service ConfidentialTransferService) GeneratePedersenOpening() (PedersenOpening, error) {
	value, err := scalar.New().SetRandom(rand.Reader)
	if err != nil {
		return PedersenOpening{}, fmt.Errorf("generate Pedersen opening: %w", err)
	}
	return service.pedersenOpening(value)
}

func (service ConfidentialTransferService) PedersenOpeningFromBytes(data []byte) (PedersenOpening, error) {
	value, err := service.scalarFromBytes(data, "Pedersen opening")
	if err != nil {
		return PedersenOpening{}, err
	}
	return service.pedersenOpening(value)
}

func (service ConfidentialTransferService) PedersenCommitmentFromBytes(data []byte) (PedersenCommitment, error) {
	if _, err := service.ristrettoPoint(data, "Pedersen commitment"); err != nil {
		return PedersenCommitment{}, err
	}
	commitment := PedersenCommitment{}
	copy(commitment[:], data)
	return commitment, nil
}

func (service ConfidentialTransferService) CommitPedersen(amount uint64, opening PedersenOpening) (PedersenCommitment, error) {
	openingScalar, err := service.scalarFromBytes(opening[:], "Pedersen opening")
	if err != nil {
		return PedersenCommitment{}, err
	}
	basepoint, err := service.pedersenOpeningBasepoint()
	if err != nil {
		return PedersenCommitment{}, fmt.Errorf("commit Pedersen: %w", err)
	}
	point := service.pedersenCommitmentPoint(amount, openingScalar, basepoint)
	return service.pedersenCommitment(point), nil
}

func (service ConfidentialTransferService) EncryptElGamal(publicKey ElGamalPubkey, amount uint64) (ElGamalCiphertext, PedersenOpening, error) {
	opening, err := service.GeneratePedersenOpening()
	if err != nil {
		return ElGamalCiphertext{}, PedersenOpening{}, err
	}
	ciphertext, err := service.EncryptElGamalWithOpening(publicKey, amount, opening)
	return ciphertext, opening, err
}

func (service ConfidentialTransferService) EncryptElGamalWithOpening(publicKey ElGamalPubkey, amount uint64, opening PedersenOpening) (ElGamalCiphertext, error) {
	grouped, err := service.encryptGroupedElGamal([]ElGamalPubkey{publicKey}, amount, opening)
	if err != nil {
		return ElGamalCiphertext{}, err
	}
	ciphertext := ElGamalCiphertext{}
	copy(ciphertext[:], grouped)
	return ciphertext, nil
}

func (service ConfidentialTransferService) ElGamalCiphertextFromBytes(data []byte) (ElGamalCiphertext, error) {
	if err := service.validateRistrettoPoints(data, 2, "ElGamal ciphertext"); err != nil {
		return ElGamalCiphertext{}, err
	}
	ciphertext := ElGamalCiphertext{}
	copy(ciphertext[:], data)
	return ciphertext, nil
}

func (service ConfidentialTransferService) DecryptElGamalTarget(secret ElGamalSecretKey, ciphertext ElGamalCiphertext) (PedersenCommitment, error) {
	secretScalar, err := service.scalarFromBytes(secret[:], "ElGamal secret key")
	if err != nil {
		return PedersenCommitment{}, err
	}
	commitment, handle, err := service.elGamalCiphertextPoints(ciphertext)
	if err != nil {
		return PedersenCommitment{}, err
	}
	target := curve.NewRistrettoPoint().Sub(commitment, curve.NewRistrettoPoint().Mul(handle, secretScalar))
	return service.pedersenCommitment(target), nil
}

func (service ConfidentialTransferService) AddElGamalCiphertexts(left, right ElGamalCiphertext) (ElGamalCiphertext, error) {
	return service.combineElGamalCiphertexts(left, right, false)
}

func (service ConfidentialTransferService) SubtractElGamalCiphertexts(left, right ElGamalCiphertext) (ElGamalCiphertext, error) {
	return service.combineElGamalCiphertexts(left, right, true)
}

func (service ConfidentialTransferService) AddElGamalAmount(ciphertext ElGamalCiphertext, amount uint64) (ElGamalCiphertext, error) {
	return service.combineElGamalAmount(ciphertext, amount, false)
}

func (service ConfidentialTransferService) SubtractElGamalAmount(ciphertext ElGamalCiphertext, amount uint64) (ElGamalCiphertext, error) {
	return service.combineElGamalAmount(ciphertext, amount, true)
}

func (service ConfidentialTransferService) EncryptGroupedElGamal2(publicKeys [2]ElGamalPubkey, amount uint64) (GroupedElGamalCiphertext2Handles, PedersenOpening, error) {
	opening, err := service.GeneratePedersenOpening()
	if err != nil {
		return GroupedElGamalCiphertext2Handles{}, PedersenOpening{}, err
	}
	ciphertext, err := service.EncryptGroupedElGamal2WithOpening(publicKeys, amount, opening)
	return ciphertext, opening, err
}

func (service ConfidentialTransferService) EncryptGroupedElGamal2WithOpening(publicKeys [2]ElGamalPubkey, amount uint64, opening PedersenOpening) (GroupedElGamalCiphertext2Handles, error) {
	grouped, err := service.encryptGroupedElGamal(publicKeys[:], amount, opening)
	if err != nil {
		return GroupedElGamalCiphertext2Handles{}, err
	}
	ciphertext := GroupedElGamalCiphertext2Handles{}
	copy(ciphertext[:], grouped)
	return ciphertext, nil
}

func (service ConfidentialTransferService) GroupedElGamalCiphertext2FromBytes(data []byte) (GroupedElGamalCiphertext2Handles, error) {
	if err := service.validateRistrettoPoints(data, 3, "grouped ElGamal ciphertext with 2 handles"); err != nil {
		return GroupedElGamalCiphertext2Handles{}, err
	}
	ciphertext := GroupedElGamalCiphertext2Handles{}
	copy(ciphertext[:], data)
	return ciphertext, nil
}

func (service ConfidentialTransferService) EncryptGroupedElGamal3(publicKeys [3]ElGamalPubkey, amount uint64) (GroupedElGamalCiphertext3Handles, PedersenOpening, error) {
	opening, err := service.GeneratePedersenOpening()
	if err != nil {
		return GroupedElGamalCiphertext3Handles{}, PedersenOpening{}, err
	}
	ciphertext, err := service.EncryptGroupedElGamal3WithOpening(publicKeys, amount, opening)
	return ciphertext, opening, err
}

func (service ConfidentialTransferService) EncryptGroupedElGamal3WithOpening(publicKeys [3]ElGamalPubkey, amount uint64, opening PedersenOpening) (GroupedElGamalCiphertext3Handles, error) {
	grouped, err := service.encryptGroupedElGamal(publicKeys[:], amount, opening)
	if err != nil {
		return GroupedElGamalCiphertext3Handles{}, err
	}
	ciphertext := GroupedElGamalCiphertext3Handles{}
	copy(ciphertext[:], grouped)
	return ciphertext, nil
}

func (service ConfidentialTransferService) GroupedElGamalCiphertext3FromBytes(data []byte) (GroupedElGamalCiphertext3Handles, error) {
	if err := service.validateRistrettoPoints(data, 4, "grouped ElGamal ciphertext with 3 handles"); err != nil {
		return GroupedElGamalCiphertext3Handles{}, err
	}
	ciphertext := GroupedElGamalCiphertext3Handles{}
	copy(ciphertext[:], data)
	return ciphertext, nil
}

func (service ConfidentialTransferService) GroupedElGamalCiphertext2Handle(ciphertext GroupedElGamalCiphertext2Handles, index int) (ElGamalCiphertext, error) {
	return service.groupedElGamalHandle(ciphertext[:], 2, index)
}

func (service ConfidentialTransferService) GroupedElGamalCiphertext3Handle(ciphertext GroupedElGamalCiphertext3Handles, index int) (ElGamalCiphertext, error) {
	return service.groupedElGamalHandle(ciphertext[:], 3, index)
}

func (ConfidentialTransferService) SplitAmount(amount uint64) (uint16, uint32, error) {
	if amount >= uint64(1)<<48 {
		return 0, 0, fmt.Errorf("split confidential amount: amount exceeds 48 bits")
	}
	return uint16(amount), uint32(amount >> 16), nil
}

func (ConfidentialTransferService) CombineAmount(low uint16, high uint32) uint64 {
	return uint64(low) | uint64(high)<<16
}

func (ConfidentialTransferService) scalarFromBytes(data []byte, name string) (*scalar.Scalar, error) {
	value, err := scalar.NewFromCanonicalBytes(data)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", name, err)
	}
	return value, nil
}

func (ConfidentialTransferService) ristrettoPoint(data []byte, name string) (*curve.RistrettoPoint, error) {
	point := curve.NewRistrettoPoint()
	if err := point.UnmarshalBinary(data); err != nil {
		return nil, fmt.Errorf("decode %s: %w", name, err)
	}
	return point, nil
}

func (service ConfidentialTransferService) validateRistrettoPoints(data []byte, count int, name string) error {
	if len(data) != count*32 {
		return fmt.Errorf("decode %s: invalid length %d", name, len(data))
	}
	for index := 0; index < count; index++ {
		if _, err := service.ristrettoPoint(data[index*32:(index+1)*32], name); err != nil {
			return err
		}
	}
	return nil
}

func (ConfidentialTransferService) pedersenOpeningBasepoint() (*curve.RistrettoPoint, error) {
	hash := sha3.Sum512(curve.RISTRETTO_BASEPOINT_COMPRESSED[:])
	return curve.NewRistrettoPoint().SetUniformBytes(hash[:])
}

func (ConfidentialTransferService) pedersenCommitmentPoint(amount uint64, opening *scalar.Scalar, basepoint *curve.RistrettoPoint) *curve.RistrettoPoint {
	return curve.NewRistrettoPoint().MultiscalarMul(
		[]*scalar.Scalar{scalar.New().SetUint64(amount), opening},
		[]*curve.RistrettoPoint{curve.RISTRETTO_BASEPOINT_POINT, basepoint},
	)
}

func (ConfidentialTransferService) pedersenOpening(value *scalar.Scalar) (PedersenOpening, error) {
	opening := PedersenOpening{}
	if err := value.ToBytes(opening[:]); err != nil {
		return PedersenOpening{}, fmt.Errorf("encode Pedersen opening: %w", err)
	}
	return opening, nil
}

func (ConfidentialTransferService) pedersenCommitment(point *curve.RistrettoPoint) PedersenCommitment {
	compressed := curve.NewCompressedRistretto().SetRistrettoPoint(point)
	commitment := PedersenCommitment{}
	copy(commitment[:], compressed[:])
	return commitment
}

func (ConfidentialTransferService) elGamalSecretKey(value *scalar.Scalar) (ElGamalSecretKey, error) {
	secret := ElGamalSecretKey{}
	if err := value.ToBytes(secret[:]); err != nil {
		return ElGamalSecretKey{}, fmt.Errorf("encode ElGamal secret key: %w", err)
	}
	return secret, nil
}

func (service ConfidentialTransferService) encryptGroupedElGamal(publicKeys []ElGamalPubkey, amount uint64, opening PedersenOpening) ([]byte, error) {
	openingScalar, err := service.scalarFromBytes(opening[:], "Pedersen opening")
	if err != nil {
		return nil, err
	}
	basepoint, err := service.pedersenOpeningBasepoint()
	if err != nil {
		return nil, fmt.Errorf("encrypt grouped ElGamal: %w", err)
	}
	commitment := service.pedersenCommitment(service.pedersenCommitmentPoint(amount, openingScalar, basepoint))
	data := make([]byte, 32*(len(publicKeys)+1))
	copy(data, commitment[:])
	for index, publicKey := range publicKeys {
		point, err := service.ristrettoPoint(publicKey[:], "ElGamal public key")
		if err != nil {
			return nil, err
		}
		handle := service.pedersenCommitment(curve.NewRistrettoPoint().Mul(point, openingScalar))
		copy(data[(index+1)*32:], handle[:])
	}
	return data, nil
}

func (service ConfidentialTransferService) elGamalCiphertextPoints(ciphertext ElGamalCiphertext) (*curve.RistrettoPoint, *curve.RistrettoPoint, error) {
	commitment, err := service.ristrettoPoint(ciphertext[:32], "ElGamal ciphertext commitment")
	if err != nil {
		return nil, nil, err
	}
	handle, err := service.ristrettoPoint(ciphertext[32:], "ElGamal ciphertext handle")
	if err != nil {
		return nil, nil, err
	}
	return commitment, handle, nil
}

func (service ConfidentialTransferService) combineElGamalCiphertexts(left, right ElGamalCiphertext, subtract bool) (ElGamalCiphertext, error) {
	leftCommitment, leftHandle, err := service.elGamalCiphertextPoints(left)
	if err != nil {
		return ElGamalCiphertext{}, err
	}
	rightCommitment, rightHandle, err := service.elGamalCiphertextPoints(right)
	if err != nil {
		return ElGamalCiphertext{}, err
	}
	if subtract {
		return service.elGamalCiphertext(
			curve.NewRistrettoPoint().Sub(leftCommitment, rightCommitment),
			curve.NewRistrettoPoint().Sub(leftHandle, rightHandle),
		), nil
	}
	return service.elGamalCiphertext(
		curve.NewRistrettoPoint().Add(leftCommitment, rightCommitment),
		curve.NewRistrettoPoint().Add(leftHandle, rightHandle),
	), nil
}

func (service ConfidentialTransferService) combineElGamalAmount(ciphertext ElGamalCiphertext, amount uint64, subtract bool) (ElGamalCiphertext, error) {
	commitment, handle, err := service.elGamalCiphertextPoints(ciphertext)
	if err != nil {
		return ElGamalCiphertext{}, err
	}
	encodedAmount := curve.NewRistrettoPoint().Mul(curve.RISTRETTO_BASEPOINT_POINT, scalar.New().SetUint64(amount))
	if subtract {
		return service.elGamalCiphertext(curve.NewRistrettoPoint().Sub(commitment, encodedAmount), handle), nil
	}
	return service.elGamalCiphertext(curve.NewRistrettoPoint().Add(commitment, encodedAmount), handle), nil
}

func (service ConfidentialTransferService) elGamalCiphertext(commitment, handle *curve.RistrettoPoint) ElGamalCiphertext {
	commitmentBytes := service.pedersenCommitment(commitment)
	handleBytes := service.pedersenCommitment(handle)
	ciphertext := ElGamalCiphertext{}
	copy(ciphertext[:32], commitmentBytes[:])
	copy(ciphertext[32:], handleBytes[:])
	return ciphertext
}

func (ConfidentialTransferService) groupedElGamalHandle(data []byte, handles, index int) (ElGamalCiphertext, error) {
	if index < 0 || index >= handles {
		return ElGamalCiphertext{}, fmt.Errorf("grouped ElGamal handle: index %d out of bounds", index)
	}
	ciphertext := ElGamalCiphertext{}
	copy(ciphertext[:32], data[:32])
	copy(ciphertext[32:], data[(index+1)*32:(index+2)*32])
	return ciphertext, nil
}
