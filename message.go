package solana_go

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/bytedance/sonic"
	"github.com/fluxrpc/base58"
)

type MessageHeader struct {
	// The total number of signatures required to make the transaction valid.
	// The signatures must match the first `numRequiredSignatures` of `message.account_keys`.
	NumRequiredSignatures uint8 `json:"numRequiredSignatures"`

	// The last numReadonlySignedAccounts of the signed keys are read-only accounts.
	NumReadonlySignedAccounts uint8 `json:"numReadonlySignedAccounts"`

	// The last `numReadonlyUnsignedAccounts` of the unsigned keys are read-only accounts.
	NumReadonlyUnsignedAccounts uint8 `json:"numReadonlyUnsignedAccounts"`
}

type MessageVersion int

const (
	MessageVersionLegacy MessageVersion = 0 // default
	MessageVersionV0     MessageVersion = 1 // v0
)

// messageVersionPrefix is the high bit mask used to indicate a versioned
// message. If the first byte has this bit set, the message is versioned; the
// remaining 7 bits encode the version number (0 for V0, 1 for V1, etc.).
const messageVersionPrefix = 0x80

// Uint8SliceAsNum is a slice of uint8s that is JSON-encoded as an array of
// numbers instead of a base64 string.
type Uint8SliceAsNum []uint8

func (slice Uint8SliceAsNum) MarshalJSON() ([]byte, error) {
	return appendUint8Array(make([]byte, 0, len(slice)*4+2), slice), nil
}

func (slice *Uint8SliceAsNum) UnmarshalJSON(data []byte) error {
	var values []uint16
	if err := sonic.Unmarshal(data, &values); err != nil {
		return err
	}
	out := make(Uint8SliceAsNum, len(values))
	for i, v := range values {
		if v > 0xFF {
			return fmt.Errorf("value %d at index %d exceeds uint8 range", v, i)
		}
		out[i] = uint8(v)
	}
	*slice = out
	return nil
}

type MessageAddressTableLookup struct {
	AccountKey      PublicKey       `json:"accountKey"` // The account key of the address table.
	WritableIndexes Uint8SliceAsNum `json:"writableIndexes"`
	ReadonlyIndexes Uint8SliceAsNum `json:"readonlyIndexes"`
}

type MessageAddressTableLookupSlice []MessageAddressTableLookup

// NumLookups returns the number of accounts from all the lookups.
func (lookups MessageAddressTableLookupSlice) NumLookups() int {
	count := 0
	for i := range lookups {
		count += len(lookups[i].WritableIndexes)
		count += len(lookups[i].ReadonlyIndexes)
	}
	return count
}

// NumWritableLookups returns the number of writable accounts
// across all the lookups (all the address tables).
func (lookups MessageAddressTableLookupSlice) NumWritableLookups() int {
	count := 0
	for i := range lookups {
		count += len(lookups[i].WritableIndexes)
	}
	return count
}

type Message struct {
	version MessageVersion

	// List of base-58 encoded public keys used by the transaction,
	// including by the instructions and for signatures.
	// The first `message.header.numRequiredSignatures` public keys
	// must sign the transaction.
	AccountKeys []PublicKey `json:"accountKeys"`

	// Details the account types and signatures required by the transaction.
	Header MessageHeader `json:"header"`

	// A base-58 encoded hash of a recent block in the ledger used to
	// prevent transaction duplication and to give transactions lifetimes.
	RecentBlockhash Hash `json:"recentBlockhash"`

	// List of program instructions that will be executed in sequence
	// and committed in one atomic transaction if all succeed.
	Instructions []CompiledInstruction `json:"instructions"`

	// List of address table lookups used to load additional accounts
	// for this transaction. Only present in versioned (V0+) messages.
	AddressTableLookups MessageAddressTableLookupSlice `json:"addressTableLookups"`
}

// GetVersion returns the message version.
func (mx *Message) GetVersion() MessageVersion {
	return mx.version
}

// SetVersion sets the message version.
// This method forces the message to be encoded in the specified version.
// NOTE: if you set lookups, the version will default to V0.
func (mx *Message) SetVersion(version MessageVersion) (*Message, error) {
	switch version {
	case MessageVersionLegacy, MessageVersionV0:
	default:
		return nil, fmt.Errorf("invalid message version: %d", version)
	}
	mx.version = version
	return mx, nil
}

// SetAddressTableLookups (re)sets the lookups used by this message,
// and sets the message version to V0.
func (mx *Message) SetAddressTableLookups(lookups []MessageAddressTableLookup) *Message {
	mx.AddressTableLookups = lookups
	mx.version = MessageVersionV0
	return mx
}

// AddAddressTableLookup adds a new lookup to the message,
// and sets the message version to V0.
func (mx *Message) AddAddressTableLookup(lookup MessageAddressTableLookup) *Message {
	mx.AddressTableLookups = append(mx.AddressTableLookups, lookup)
	mx.version = MessageVersionV0
	return mx
}

// Signers returns the pubkeys of all accounts that are signers:
// always the first `NumRequiredSignatures` account keys.
func (mx *Message) Signers() []PublicKey {
	numSigners := int(mx.Header.NumRequiredSignatures)
	if numSigners > len(mx.AccountKeys) {
		numSigners = len(mx.AccountKeys)
	}
	out := make([]PublicKey, numSigners)
	copy(out, mx.AccountKeys[:numSigners])
	return out
}

// IsSigner reports whether the given account is a signer of the message.
func (mx *Message) IsSigner(account PublicKey) bool {
	for idx, key := range mx.AccountKeys {
		if key == account {
			return idx < int(mx.Header.NumRequiredSignatures)
		}
	}
	return false
}

// MarshalJSON writes the message JSON directly into one buffer: every value
// is a base58 string or a small integer, so the reflection-based encoder has
// nothing to offer here. Legacy messages omit the addressTableLookups key.
func (mx Message) MarshalJSON() ([]byte, error) {
	buf := make([]byte, 0, 256+len(mx.AccountKeys)*(base58.EncodedMaxLen32+3))

	buf = append(buf, `{"accountKeys":[`...)
	for i := range mx.AccountKeys {
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = append(buf, '"')
		buf = base58.AppendEncode32(buf, (*[32]byte)(&mx.AccountKeys[i]))
		buf = append(buf, '"')
	}

	buf = append(buf, `],"header":{"numRequiredSignatures":`...)
	buf = strconv.AppendUint(buf, uint64(mx.Header.NumRequiredSignatures), 10)
	buf = append(buf, `,"numReadonlySignedAccounts":`...)
	buf = strconv.AppendUint(buf, uint64(mx.Header.NumReadonlySignedAccounts), 10)
	buf = append(buf, `,"numReadonlyUnsignedAccounts":`...)
	buf = strconv.AppendUint(buf, uint64(mx.Header.NumReadonlyUnsignedAccounts), 10)

	buf = append(buf, `},"recentBlockhash":"`...)
	buf = base58.AppendEncode32(buf, (*[32]byte)(&mx.RecentBlockhash))

	buf = append(buf, `","instructions":[`...)
	for i := range mx.Instructions {
		if i > 0 {
			buf = append(buf, ',')
		}
		ins := &mx.Instructions[i]
		buf = append(buf, `{"programIdIndex":`...)
		buf = strconv.AppendUint(buf, uint64(ins.ProgramIDIndex), 10)
		buf = append(buf, `,"accounts":[`...)
		for j, account := range ins.Accounts {
			if j > 0 {
				buf = append(buf, ',')
			}
			buf = strconv.AppendUint(buf, uint64(account), 10)
		}
		buf = append(buf, `],"data":"`...)
		buf = base58.AppendEncode(buf, ins.Data)
		buf = append(buf, `"}`...)
	}
	buf = append(buf, ']')

	if mx.version != MessageVersionLegacy {
		buf = append(buf, `,"addressTableLookups":[`...)
		for i := range mx.AddressTableLookups {
			if i > 0 {
				buf = append(buf, ',')
			}
			lookup := &mx.AddressTableLookups[i]
			buf = append(buf, `{"accountKey":"`...)
			buf = base58.AppendEncode32(buf, (*[32]byte)(&lookup.AccountKey))
			buf = append(buf, `","writableIndexes":`...)
			buf = appendUint8Array(buf, lookup.WritableIndexes)
			buf = append(buf, `,"readonlyIndexes":`...)
			buf = appendUint8Array(buf, lookup.ReadonlyIndexes)
			buf = append(buf, '}')
		}
		buf = append(buf, ']')
	}

	return append(buf, '}'), nil
}

func appendUint8Array(buf []byte, values []uint8) []byte {
	buf = append(buf, '[')
	for i, v := range values {
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = strconv.AppendUint(buf, uint64(v), 10)
	}
	return append(buf, ']')
}

// UnmarshalJSON decodes the message from JSON and determines its version.
// The Solana RPC emits `addressTableLookups` only for versioned (V0+)
// messages; its presence in the JSON is what distinguishes V0 from legacy,
// since the private `version` field has no wire representation.
func (mx *Message) UnmarshalJSON(data []byte) error {
	// Decode `addressTableLookups` via a RawMessage pointer so presence of
	// the key can be detected in a single parse. A non-nil pointer means the
	// key was present with a non-null value, which selects V0; an absent key
	// or an explicit null both decode to nil and select legacy (matching
	// upstream behavior).
	aux := struct {
		AccountKeys         []PublicKey           `json:"accountKeys"`
		Header              MessageHeader         `json:"header"`
		RecentBlockhash     Hash                  `json:"recentBlockhash"`
		Instructions        []CompiledInstruction `json:"instructions"`
		AddressTableLookups *json.RawMessage      `json:"addressTableLookups"`
	}{}
	if err := sonic.Unmarshal(data, &aux); err != nil {
		return err
	}
	mx.AccountKeys = aux.AccountKeys
	mx.Header = aux.Header
	mx.RecentBlockhash = aux.RecentBlockhash
	mx.Instructions = aux.Instructions

	if aux.AddressTableLookups == nil {
		mx.version = MessageVersionLegacy
		mx.AddressTableLookups = nil
		return nil
	}
	mx.version = MessageVersionV0
	return sonic.Unmarshal(*aux.AddressTableLookups, &mx.AddressTableLookups)
}

func (mx *Message) MarshalBinary() ([]byte, error) {
	switch mx.version {
	case MessageVersionLegacy:
		buf := make([]byte, 0, mx.bodySize())
		return mx.appendBody(buf), nil
	case MessageVersionV0:
		return mx.marshalV0()
	default:
		return nil, fmt.Errorf("invalid message version: %d", mx.version)
	}
}

func (mx *Message) marshalV0() ([]byte, error) {
	// The actual Solana version number is the Go enum value minus 1
	// (MessageVersionV0=1 maps to Solana version 0).
	solanaVersion := byte(mx.version - 1)
	if solanaVersion > 0x7F {
		return nil, fmt.Errorf("invalid message version: %d", mx.version)
	}

	size := 1 + mx.bodySize() + shortvecLen(len(mx.AddressTableLookups))
	for i := range mx.AddressTableLookups {
		lookup := &mx.AddressTableLookups[i]
		size += PublicKeyLength +
			shortvecLen(len(lookup.WritableIndexes)) + len(lookup.WritableIndexes) +
			shortvecLen(len(lookup.ReadonlyIndexes)) + len(lookup.ReadonlyIndexes)
	}

	buf := make([]byte, 0, size)
	buf = append(buf, messageVersionPrefix|solanaVersion)
	buf = mx.appendBody(buf)

	buf = appendShortvecLen(buf, len(mx.AddressTableLookups))
	for i := range mx.AddressTableLookups {
		lookup := &mx.AddressTableLookups[i]
		buf = append(buf, lookup.AccountKey[:]...)
		buf = appendShortvecLen(buf, len(lookup.WritableIndexes))
		buf = append(buf, lookup.WritableIndexes...)
		buf = appendShortvecLen(buf, len(lookup.ReadonlyIndexes))
		buf = append(buf, lookup.ReadonlyIndexes...)
	}
	return buf, nil
}

// bodySize returns the exact encoded size of the version-independent part of
// the message (header, account keys, blockhash and instructions).
func (mx *Message) bodySize() int {
	size := 3 +
		shortvecLen(len(mx.AccountKeys)) + PublicKeyLength*len(mx.AccountKeys) +
		len(mx.RecentBlockhash) +
		shortvecLen(len(mx.Instructions))
	for i := range mx.Instructions {
		ins := &mx.Instructions[i]
		size += 1 +
			shortvecLen(len(ins.Accounts)) + len(ins.Accounts) +
			shortvecLen(len(ins.Data)) + len(ins.Data)
	}
	return size
}

// appendBody appends the version-independent part of the message to buf.
func (mx *Message) appendBody(buf []byte) []byte {
	buf = append(buf,
		mx.Header.NumRequiredSignatures,
		mx.Header.NumReadonlySignedAccounts,
		mx.Header.NumReadonlyUnsignedAccounts,
	)

	buf = appendShortvecLen(buf, len(mx.AccountKeys))
	for i := range mx.AccountKeys {
		buf = append(buf, mx.AccountKeys[i][:]...)
	}

	buf = append(buf, mx.RecentBlockhash[:]...)

	buf = appendShortvecLen(buf, len(mx.Instructions))
	for i := range mx.Instructions {
		ins := &mx.Instructions[i]
		buf = append(buf, byte(ins.ProgramIDIndex))
		buf = appendShortvecLen(buf, len(ins.Accounts))
		for _, accountIdx := range ins.Accounts {
			buf = append(buf, byte(accountIdx))
		}
		buf = appendShortvecLen(buf, len(ins.Data))
		buf = append(buf, ins.Data...)
	}
	return buf
}

// UnmarshalBinary decodes a legacy or versioned (V0) message.
// Trailing bytes after the message are ignored.
//
// NOTE: instruction data aliases the input buffer to avoid copies; the
// caller must not mutate or reuse data while the message is alive.
func (mx *Message) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		return errors.New("message data is empty")
	}
	if data[0]&messageVersionPrefix != 0 {
		version := data[0] & 0x7F
		if version != 0 {
			return fmt.Errorf("unsupported message version: %d", version)
		}
		mx.version = MessageVersionV0
		rest, err := mx.unmarshalBody(data[1:])
		if err != nil {
			return err
		}
		return mx.unmarshalLookups(rest)
	}
	mx.version = MessageVersionLegacy
	mx.AddressTableLookups = nil
	_, err := mx.unmarshalBody(data)
	return err
}

// unmarshalBody decodes the version-independent part of the message and
// returns the unconsumed remainder of data.
func (mx *Message) unmarshalBody(data []byte) ([]byte, error) {
	if len(data) < 3 {
		return nil, fmt.Errorf("message too short for header: %d bytes", len(data))
	}
	mx.Header.NumRequiredSignatures = data[0]
	mx.Header.NumReadonlySignedAccounts = data[1]
	mx.Header.NumReadonlyUnsignedAccounts = data[2]
	data = data[3:]

	numKeys, n, err := decodeShortvecLen(data)
	if err != nil {
		return nil, fmt.Errorf("account keys length: %w", err)
	}
	data = data[n:]
	// Bound the claimed length by the remaining input before allocating.
	if numKeys > len(data)/PublicKeyLength {
		return nil, fmt.Errorf("account keys length %d too large for remaining %d bytes", numKeys, len(data))
	}
	mx.AccountKeys = make([]PublicKey, numKeys)
	for i := range mx.AccountKeys {
		copy(mx.AccountKeys[i][:], data[i*PublicKeyLength:])
	}
	data = data[numKeys*PublicKeyLength:]

	if len(data) < len(mx.RecentBlockhash) {
		return nil, fmt.Errorf("message too short for recent blockhash: %d bytes", len(data))
	}
	copy(mx.RecentBlockhash[:], data)
	data = data[len(mx.RecentBlockhash):]

	numInstructions, n, err := decodeShortvecLen(data)
	if err != nil {
		return nil, fmt.Errorf("instructions length: %w", err)
	}
	data = data[n:]
	if numInstructions > len(data) {
		return nil, fmt.Errorf("instructions length %d too large for remaining %d bytes", numInstructions, len(data))
	}
	mx.Instructions = make([]CompiledInstruction, numInstructions)
	for i := range mx.Instructions {
		ins := &mx.Instructions[i]
		if len(data) == 0 {
			return nil, fmt.Errorf("instruction %d: missing program ID index", i)
		}
		ins.ProgramIDIndex = uint16(data[0])
		data = data[1:]

		numAccounts, n, err := decodeShortvecLen(data)
		if err != nil {
			return nil, fmt.Errorf("instruction %d accounts length: %w", i, err)
		}
		data = data[n:]
		if numAccounts > len(data) {
			return nil, fmt.Errorf("instruction %d accounts length %d too large for remaining %d bytes", i, numAccounts, len(data))
		}
		ins.Accounts = make([]uint16, numAccounts)
		for j := range ins.Accounts {
			ins.Accounts[j] = uint16(data[j])
		}
		data = data[numAccounts:]

		dataLen, n, err := decodeShortvecLen(data)
		if err != nil {
			return nil, fmt.Errorf("instruction %d data length: %w", i, err)
		}
		data = data[n:]
		if dataLen > len(data) {
			return nil, fmt.Errorf("instruction %d data length %d too large for remaining %d bytes", i, dataLen, len(data))
		}
		// Subslice instead of copying; the instruction data aliases the input
		// buffer (capped capacity, so appends cannot clobber it).
		ins.Data = Base58(data[:dataLen:dataLen])
		data = data[dataLen:]
	}
	return data, nil
}

// unmarshalLookups decodes the address table lookups of a versioned message.
func (mx *Message) unmarshalLookups(data []byte) error {
	numLookups, n, err := decodeShortvecLen(data)
	if err != nil {
		return fmt.Errorf("address table lookups length: %w", err)
	}
	data = data[n:]
	mx.AddressTableLookups = nil
	if numLookups == 0 {
		return nil
	}
	// A lookup is at least 34 bytes: 32-byte key plus two length prefixes.
	if numLookups > len(data)/34 {
		return fmt.Errorf("address table lookups length %d too large for remaining %d bytes", numLookups, len(data))
	}
	mx.AddressTableLookups = make(MessageAddressTableLookupSlice, numLookups)
	for i := range mx.AddressTableLookups {
		lookup := &mx.AddressTableLookups[i]
		if len(data) < PublicKeyLength {
			return fmt.Errorf("lookup %d: missing account key", i)
		}
		copy(lookup.AccountKey[:], data)
		data = data[PublicKeyLength:]

		lookup.WritableIndexes, data, err = decodeLookupIndexes(data)
		if err != nil {
			return fmt.Errorf("lookup %d writable indexes: %w", i, err)
		}
		lookup.ReadonlyIndexes, data, err = decodeLookupIndexes(data)
		if err != nil {
			return fmt.Errorf("lookup %d readonly indexes: %w", i, err)
		}
	}
	return nil
}

func decodeLookupIndexes(data []byte) (Uint8SliceAsNum, []byte, error) {
	count, n, err := decodeShortvecLen(data)
	if err != nil {
		return nil, nil, err
	}
	data = data[n:]
	if count > len(data) {
		return nil, nil, fmt.Errorf("length %d too large for remaining %d bytes", count, len(data))
	}
	out := make(Uint8SliceAsNum, count)
	copy(out, data)
	return out, data[count:], nil
}

// ToBase64 returns the base64 encoding of the binary message.
func (mx Message) ToBase64() string {
	out, _ := mx.MarshalBinary()
	return base64.StdEncoding.EncodeToString(out)
}
