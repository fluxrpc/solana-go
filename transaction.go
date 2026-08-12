package solana_go

import (
	"encoding/base64"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/fluxrpc/base58"
)

type Transaction struct {
	// A list of base-58 encoded signatures applied to the transaction.
	// The list is always of length `message.header.numRequiredSignatures` and not empty.
	// The signature at index `i` corresponds to the public key at index
	// `i` in `message.account_keys`. The first one is used as the transaction id.
	Signatures []Signature `json:"signatures"`

	// Defines the content of the transaction.
	Message Message `json:"message"`
}

// TransactionFromBytes decodes a binary transaction.
func TransactionFromBytes(data []byte) (*Transaction, error) {
	tx := new(Transaction)
	if err := tx.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return tx, nil
}

// MustTransactionFromBytes decodes a binary transaction.
// Panics on error.
func MustTransactionFromBytes(data []byte) *Transaction {
	tx, err := TransactionFromBytes(data)
	if err != nil {
		panic(err)
	}
	return tx
}

// TransactionFromBase64 decodes a base64 encoded transaction.
func TransactionFromBase64(b64 string) (*Transaction, error) {
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, err
	}
	return TransactionFromBytes(data)
}

func (tx *Transaction) MarshalBinary() ([]byte, error) {
	messageContent, err := tx.Message.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to encode tx.Message to binary: %w", err)
	}

	// Missing signatures are encoded as zero-valued dummy signatures; without
	// them the serialized transaction would be invalid.
	numSigs := len(tx.Signatures)
	if required := int(tx.Message.Header.NumRequiredSignatures); numSigs < required {
		numSigs = required
	}

	buf := make([]byte, 0, shortvecLen(numSigs)+numSigs*SignatureLength+len(messageContent))
	buf = appendShortvecLen(buf, numSigs)
	for i := range tx.Signatures {
		buf = append(buf, tx.Signatures[i][:]...)
	}
	var zero Signature
	for i := len(tx.Signatures); i < numSigs; i++ {
		buf = append(buf, zero[:]...)
	}
	return append(buf, messageContent...), nil
}

func (tx *Transaction) UnmarshalBinary(data []byte) error {
	numSigs, n, err := decodeShortvecLen(data)
	if err != nil {
		return fmt.Errorf("signatures length: %w", err)
	}
	data = data[n:]
	// Bound the claimed length by the remaining input before allocating.
	if numSigs > len(data)/SignatureLength {
		return fmt.Errorf("signatures length %d too large for remaining %d bytes", numSigs, len(data))
	}
	tx.Signatures = make([]Signature, numSigs)
	for i := range tx.Signatures {
		copy(tx.Signatures[i][:], data[i*SignatureLength:])
	}
	data = data[numSigs*SignatureLength:]

	if err := tx.Message.UnmarshalBinary(data); err != nil {
		return fmt.Errorf("unable to decode tx.Message: %w", err)
	}
	return nil
}

// ToBase64 returns the base64 encoding of the binary transaction.
func (tx *Transaction) ToBase64() string {
	out, _ := tx.MarshalBinary()
	return base64.StdEncoding.EncodeToString(out)
}

// MarshalJSON writes the transaction JSON directly into one buffer,
// bypassing the reflection-based encoder (see Message.MarshalJSON).
func (tx Transaction) MarshalJSON() ([]byte, error) {
	message, err := tx.Message.MarshalJSON()
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 0, len(message)+32+len(tx.Signatures)*(base58.EncodedMaxLen64+3))
	buf = append(buf, `{"signatures":[`...)
	for i := range tx.Signatures {
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = append(buf, '"')
		buf = base58.AppendEncode64(buf, (*[64]byte)(&tx.Signatures[i]))
		buf = append(buf, '"')
	}
	buf = append(buf, `],"message":`...)
	buf = append(buf, message...)
	return append(buf, '}'), nil
}

// UnmarshalJSON decodes the transaction with sonic regardless of which JSON
// package the caller uses on the outer value.
func (tx *Transaction) UnmarshalJSON(data []byte) error {
	// The alias drops this method so sonic decodes the fields directly.
	type transactionAlias Transaction
	return sonic.Unmarshal(data, (*transactionAlias)(tx))
}

// privateKeyGetter looks up the private key of the given public key,
// returning nil if it is not available.
type privateKeyGetter func(key PublicKey) *PrivateKey

// PartialSign signs the message with the private keys made available by the
// getter, leaving the signatures of unavailable signers untouched (zero if
// not previously set), and returns the transaction's signatures.
func (tx *Transaction) PartialSign(getter privateKeyGetter) ([]Signature, error) {
	messageContent, err := tx.Message.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("unable to encode message for signing: %w", err)
	}
	signerKeys := tx.Message.Signers()

	// Ensure that the transaction has the correct number of signatures.
	if len(tx.Signatures) == 0 {
		tx.Signatures = make([]Signature, len(signerKeys))
	} else if len(tx.Signatures) != len(signerKeys) {
		return nil, fmt.Errorf("invalid signatures length, expected %d, actual %d", len(signerKeys), len(tx.Signatures))
	}

	for i, key := range signerKeys {
		privateKey := getter(key)
		if privateKey == nil {
			continue
		}
		s, err := privateKey.Sign(messageContent)
		if err != nil {
			return nil, fmt.Errorf("failed to sign with key %q: %w", key.String(), err)
		}
		tx.Signatures[i] = s
	}
	return tx.Signatures, nil
}

// Sign signs the message with the private keys made available by the getter,
// and returns the transaction's signatures.
// It errors if the getter cannot provide a key for any required signer.
func (tx *Transaction) Sign(getter privateKeyGetter) ([]Signature, error) {
	for _, key := range tx.Message.Signers() {
		if getter(key) == nil {
			return nil, fmt.Errorf("signer key %q not found. Ensure all the signer keys are in the vault", key.String())
		}
	}
	return tx.PartialSign(getter)
}

// VerifySignatures verifies all the signatures in the transaction
// against the pubkeys of the signers.
func (tx *Transaction) VerifySignatures() error {
	msg, err := tx.Message.MarshalBinary()
	if err != nil {
		return err
	}

	signers := tx.Message.Signers()
	if len(signers) != len(tx.Signatures) {
		return fmt.Errorf(
			"got %v signers, but %v signatures",
			len(signers),
			len(tx.Signatures),
		)
	}

	for i, sig := range tx.Signatures {
		if !sig.Verify(signers[i], msg) {
			return fmt.Errorf("invalid signature by %s", signers[i].String())
		}
	}
	return nil
}
