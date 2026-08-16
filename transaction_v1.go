package solana_go

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"math/bits"
)

// SIMD-0385 v1 transaction constants. Multi-byte integers are little-endian
// (explicit in the SIMD for config values; the universal Solana wire
// convention for the remaining fields).
const (
	// TransactionV1VersionByte is the leading byte identifying a v1
	// transaction (128 | 1).
	TransactionV1VersionByte = 129

	// TransactionV1MaxSize is the maximum serialized size of a v1
	// transaction, up from 1232 for legacy/v0.
	TransactionV1MaxSize = 4096

	transactionV1MaxSignatures   = 12
	transactionV1MaxAddresses    = 64
	transactionV1MaxInstructions = 64

	transactionV1HeapSizeMin  = 32 * 1024
	transactionV1HeapSizeMax  = 256 * 1024
	transactionV1HeapSizeUnit = 1024
)

// Config mask bits (SIMD-0385): each set bit contributes one 4-byte slot to
// the ConfigValues array, in ascending bit order.
const (
	configBitPriorityFeeLo uint32 = 1 << 0 // 8-byte LE u64 spans bits 0 and 1
	configBitPriorityFeeHi uint32 = 1 << 1
	configBitComputeUnits  uint32 = 1 << 2 // 4-byte LE u32
	configBitLoadedDataSz  uint32 = 1 << 3 // 4-byte LE u32
	configBitHeapSize      uint32 = 1 << 4 // 4-byte LE u32

	configKnownBits = configBitPriorityFeeLo | configBitPriorityFeeHi |
		configBitComputeUnits | configBitLoadedDataSz | configBitHeapSize
)

// TransactionConfig carries the resource and fee requests of a v1
// transaction. It replaces the ComputeBudgetProgram instructions of
// legacy/v0 transactions: the runtime reads these values from the
// transaction header itself. A nil field leaves its mask bit unset and the
// protocol minimum applies (priority fee 0, compute-unit limit 0, loaded
// accounts data size limit 0, heap 32KiB).
type TransactionConfig struct {
	// PriorityFeeLamports is the total priority fee in lamports
	// (mask bits 0-1, 8-byte value).
	PriorityFeeLamports *uint64 `json:"priorityFeeLamports,omitempty"`
	// ComputeUnitLimit is the requested compute-unit limit (mask bit 2).
	ComputeUnitLimit *uint32 `json:"computeUnitLimit,omitempty"`
	// LoadedAccountsDataSizeLimit is the requested loaded-accounts data
	// size limit in bytes (mask bit 3).
	LoadedAccountsDataSizeLimit *uint32 `json:"loadedAccountsDataSizeLimit,omitempty"`
	// HeapSize is the requested heap size in bytes (mask bit 4). Must be a
	// multiple of 1KiB in [32KiB, 256KiB].
	HeapSize *uint32 `json:"heapSize,omitempty"`
}

// mask returns the TransactionConfigMask for the set fields.
func (c *TransactionConfig) mask() uint32 {
	var m uint32
	if c.PriorityFeeLamports != nil {
		m |= configBitPriorityFeeLo | configBitPriorityFeeHi
	}
	if c.ComputeUnitLimit != nil {
		m |= configBitComputeUnits
	}
	if c.LoadedAccountsDataSizeLimit != nil {
		m |= configBitLoadedDataSz
	}
	if c.HeapSize != nil {
		m |= configBitHeapSize
	}
	return m
}

// validate checks the config constraints of SIMD-0385.
func (c *TransactionConfig) validate() error {
	if c.HeapSize != nil {
		h := *c.HeapSize
		if h < transactionV1HeapSizeMin || h > transactionV1HeapSizeMax || h%transactionV1HeapSizeUnit != 0 {
			return fmt.Errorf("invalid heap size %d: must be a multiple of %d in [%d, %d]",
				h, transactionV1HeapSizeUnit, transactionV1HeapSizeMin, transactionV1HeapSizeMax)
		}
	}
	return nil
}

// appendValues appends the ConfigValues slots in ascending bit order.
func (c *TransactionConfig) appendValues(buf []byte) []byte {
	if c.PriorityFeeLamports != nil {
		buf = binary.LittleEndian.AppendUint64(buf, *c.PriorityFeeLamports)
	}
	if c.ComputeUnitLimit != nil {
		buf = binary.LittleEndian.AppendUint32(buf, *c.ComputeUnitLimit)
	}
	if c.LoadedAccountsDataSizeLimit != nil {
		buf = binary.LittleEndian.AppendUint32(buf, *c.LoadedAccountsDataSizeLimit)
	}
	if c.HeapSize != nil {
		buf = binary.LittleEndian.AppendUint32(buf, *c.HeapSize)
	}
	return buf
}

// TransactionV1 is a SIMD-0385 v1 transaction: version byte 129, up to
// 4096 serialized bytes, resource requests in the header instead of
// ComputeBudgetProgram instructions, signatures trailing the signed
// payload, and no address lookup tables.
//
// The JSON form is SDK-defined (the RPC spec has not standardized a v1
// JSON representation yet) and mirrors the struct layout.
type TransactionV1 struct {
	// Header is the legacy header: signature and readonly account counts.
	Header MessageHeader `json:"header"`

	// Config is the transaction's fee/resource requests
	// (TransactionConfigMask + ConfigValues on the wire).
	Config TransactionConfig `json:"config"`

	// LifetimeSpecifier is the recent blockhash (renamed by SIMD-0385
	// without changing its meaning).
	LifetimeSpecifier Hash `json:"lifetimeSpecifier"`

	// AccountKeys lists every address the transaction references, ordered
	// exactly as in prior formats: writable signers first (fee payer at
	// index 0), then readonly signers, writable non-signers, readonly
	// non-signers. No duplicates.
	AccountKeys []PublicKey `json:"accountKeys"`

	// Instructions are the compiled instructions. Account indexes and the
	// program index must fit in a u8 and be < len(AccountKeys);
	// instruction data must fit in a u16 length.
	Instructions []CompiledInstruction `json:"instructions"`

	// Signatures holds Header.NumRequiredSignatures signatures over the
	// serialized transaction prefix (everything before the signatures);
	// Signatures[i] is by AccountKeys[i].
	Signatures []Signature `json:"signatures"`
}

// sanitize enforces the SIMD-0385 sanitization rules that do not depend on
// the serialized size.
func (tx *TransactionV1) sanitize() error {
	h := tx.Header
	if h.NumRequiredSignatures < 1 {
		return errors.New("v1 transaction: NumRequiredSignatures must be >= 1")
	}
	if h.NumRequiredSignatures > transactionV1MaxSignatures {
		return fmt.Errorf("v1 transaction: %d signatures exceeds the maximum of %d", h.NumRequiredSignatures, transactionV1MaxSignatures)
	}
	if h.NumReadonlySignedAccounts >= h.NumRequiredSignatures {
		return errors.New("v1 transaction: NumReadonlySignedAccounts must be < NumRequiredSignatures (the fee payer must be writable)")
	}
	numAddresses := len(tx.AccountKeys)
	if numAddresses > transactionV1MaxAddresses {
		return fmt.Errorf("v1 transaction: %d addresses exceeds the maximum of %d", numAddresses, transactionV1MaxAddresses)
	}
	if numAddresses < int(h.NumRequiredSignatures)+int(h.NumReadonlyUnsignedAccounts) {
		return errors.New("v1 transaction: fewer addresses than the header requires")
	}
	seen := make(map[PublicKey]struct{}, numAddresses)
	for _, key := range tx.AccountKeys {
		if _, dup := seen[key]; dup {
			return fmt.Errorf("v1 transaction: duplicate address %s", key)
		}
		seen[key] = struct{}{}
	}
	if len(tx.Instructions) > transactionV1MaxInstructions {
		return fmt.Errorf("v1 transaction: %d instructions exceeds the maximum of %d", len(tx.Instructions), transactionV1MaxInstructions)
	}
	for i, ix := range tx.Instructions {
		if int(ix.ProgramIDIndex) >= numAddresses {
			return fmt.Errorf("v1 transaction: instruction %d program index %d out of range", i, ix.ProgramIDIndex)
		}
		if len(ix.Accounts) > 255 {
			return fmt.Errorf("v1 transaction: instruction %d has %d accounts, maximum is 255", i, len(ix.Accounts))
		}
		for _, idx := range ix.Accounts {
			if int(idx) >= numAddresses {
				return fmt.Errorf("v1 transaction: instruction %d account index %d out of range", i, idx)
			}
		}
		if len(ix.Data) > 0xFFFF {
			return fmt.Errorf("v1 transaction: instruction %d data is %d bytes, maximum is %d", i, len(ix.Data), 0xFFFF)
		}
	}
	return tx.Config.validate()
}

// signedSize is the serialized size of the signed prefix (everything
// before the signatures).
func (tx *TransactionV1) signedSize() int {
	size := 1 + 3 + 4 + 32 + 1 + 1 // version, header, mask, lifetime, counts
	size += 32 * len(tx.AccountKeys)
	size += 4 * bits.OnesCount32(tx.Config.mask())
	size += 4 * len(tx.Instructions)
	for _, ix := range tx.Instructions {
		size += len(ix.Accounts) + len(ix.Data)
	}
	return size
}

// appendSigned appends the signed prefix.
func (tx *TransactionV1) appendSigned(buf []byte) []byte {
	buf = append(buf, TransactionV1VersionByte,
		tx.Header.NumRequiredSignatures,
		tx.Header.NumReadonlySignedAccounts,
		tx.Header.NumReadonlyUnsignedAccounts)
	buf = binary.LittleEndian.AppendUint32(buf, tx.Config.mask())
	buf = append(buf, tx.LifetimeSpecifier[:]...)
	buf = append(buf, uint8(len(tx.Instructions)), uint8(len(tx.AccountKeys)))
	for i := range tx.AccountKeys {
		buf = append(buf, tx.AccountKeys[i][:]...)
	}
	buf = tx.Config.appendValues(buf)
	for _, ix := range tx.Instructions {
		buf = append(buf, uint8(ix.ProgramIDIndex), uint8(len(ix.Accounts)))
		buf = binary.LittleEndian.AppendUint16(buf, uint16(len(ix.Data)))
	}
	for _, ix := range tx.Instructions {
		for _, idx := range ix.Accounts {
			buf = append(buf, uint8(idx))
		}
		buf = append(buf, ix.Data...)
	}
	return buf
}

// MarshalBinary encodes the transaction into the SIMD-0385 wire format in
// a single allocation, enforcing every sanitization rule including the
// 4096-byte size limit. Signatures must already be present (see Sign); the
// signature count must equal Header.NumRequiredSignatures.
func (tx *TransactionV1) MarshalBinary() ([]byte, error) {
	if err := tx.sanitize(); err != nil {
		return nil, err
	}
	if len(tx.Signatures) != int(tx.Header.NumRequiredSignatures) {
		return nil, fmt.Errorf("v1 transaction: %d signatures, header requires %d", len(tx.Signatures), tx.Header.NumRequiredSignatures)
	}
	size := tx.signedSize() + 64*len(tx.Signatures)
	if size > TransactionV1MaxSize {
		return nil, fmt.Errorf("v1 transaction: serialized size %d exceeds the maximum of %d", size, TransactionV1MaxSize)
	}
	buf := make([]byte, 0, size)
	buf = tx.appendSigned(buf)
	for i := range tx.Signatures {
		buf = append(buf, tx.Signatures[i][:]...)
	}
	return buf, nil
}

// UnmarshalBinary decodes and fully sanitizes a SIMD-0385 v1 transaction.
//
// NOTE: instruction data aliases the input buffer to avoid copies; the
// caller must not mutate or reuse data while the transaction is alive.
func (tx *TransactionV1) UnmarshalBinary(data []byte) error {
	if len(data) > TransactionV1MaxSize {
		return fmt.Errorf("v1 transaction: %d bytes exceeds the maximum of %d", len(data), TransactionV1MaxSize)
	}
	if len(data) < 42 { // version, header, mask, lifetime, counts
		return errors.New("v1 transaction: truncated header")
	}
	if data[0] != TransactionV1VersionByte {
		return fmt.Errorf("v1 transaction: version byte %d, want %d", data[0], TransactionV1VersionByte)
	}
	tx.Header = MessageHeader{
		NumRequiredSignatures:       data[1],
		NumReadonlySignedAccounts:   data[2],
		NumReadonlyUnsignedAccounts: data[3],
	}
	mask := binary.LittleEndian.Uint32(data[4:8])
	tx.LifetimeSpecifier = Hash(data[8:40])
	numInstructions := int(data[40])
	numAddresses := int(data[41])
	rest := data[42:]

	if unknown := mask &^ configKnownBits; unknown != 0 {
		return fmt.Errorf("v1 transaction: unknown config mask bits %#x", unknown)
	}
	loBit, hiBit := mask&configBitPriorityFeeLo != 0, mask&configBitPriorityFeeHi != 0
	if loBit != hiBit {
		return errors.New("v1 transaction: priority-fee config requires both mask bits 0 and 1")
	}

	if len(rest) < 32*numAddresses {
		return errors.New("v1 transaction: truncated addresses")
	}
	tx.AccountKeys = make([]PublicKey, numAddresses)
	for i := range tx.AccountKeys {
		tx.AccountKeys[i] = PublicKey(rest[32*i : 32*i+32])
	}
	rest = rest[32*numAddresses:]

	configLen := 4 * bits.OnesCount32(mask)
	if len(rest) < configLen {
		return errors.New("v1 transaction: truncated config values")
	}
	tx.Config = TransactionConfig{}
	values := rest[:configLen]
	if loBit {
		fee := binary.LittleEndian.Uint64(values)
		tx.Config.PriorityFeeLamports = &fee
		values = values[8:]
	}
	if mask&configBitComputeUnits != 0 {
		v := binary.LittleEndian.Uint32(values)
		tx.Config.ComputeUnitLimit = &v
		values = values[4:]
	}
	if mask&configBitLoadedDataSz != 0 {
		v := binary.LittleEndian.Uint32(values)
		tx.Config.LoadedAccountsDataSizeLimit = &v
		values = values[4:]
	}
	if mask&configBitHeapSize != 0 {
		v := binary.LittleEndian.Uint32(values)
		tx.Config.HeapSize = &v
	}
	rest = rest[configLen:]

	if len(rest) < 4*numInstructions {
		return errors.New("v1 transaction: truncated instruction headers")
	}
	headers := rest[:4*numInstructions]
	rest = rest[4*numInstructions:]

	tx.Instructions = make([]CompiledInstruction, numInstructions)
	for i := range tx.Instructions {
		h := headers[4*i : 4*i+4]
		numAccounts := int(h[1])
		dataLen := int(binary.LittleEndian.Uint16(h[2:4]))
		if len(rest) < numAccounts+dataLen {
			return fmt.Errorf("v1 transaction: truncated payload for instruction %d", i)
		}
		accounts := make([]uint16, numAccounts)
		for j := 0; j < numAccounts; j++ {
			accounts[j] = uint16(rest[j])
		}
		tx.Instructions[i] = CompiledInstruction{
			ProgramIDIndex: uint16(h[0]),
			Accounts:       accounts,
			Data:           rest[numAccounts : numAccounts+dataLen : numAccounts+dataLen],
		}
		rest = rest[numAccounts+dataLen:]
	}

	numSignatures := int(tx.Header.NumRequiredSignatures)
	if len(rest) < 64*numSignatures {
		return errors.New("v1 transaction: truncated signatures")
	}
	tx.Signatures = make([]Signature, numSignatures)
	for i := range tx.Signatures {
		tx.Signatures[i] = Signature(rest[64*i : 64*i+64])
	}
	if len(rest) != 64*numSignatures {
		return errors.New("v1 transaction: trailing data after signatures")
	}

	return tx.sanitize()
}

// TransactionV1FromBytes decodes a binary v1 transaction.
//
// NOTE: instruction data aliases the input buffer; see UnmarshalBinary.
func TransactionV1FromBytes(data []byte) (*TransactionV1, error) {
	tx := new(TransactionV1)
	if err := tx.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return tx, nil
}

// TransactionV1FromBase64 decodes a base64 encoded v1 transaction.
func TransactionV1FromBase64(b64 string) (*TransactionV1, error) {
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("decode base64: %w", err)
	}
	return TransactionV1FromBytes(data)
}

// ToBase64 returns the base64 encoding of the binary transaction.
// Panics only if the transaction fails sanitization; use MarshalBinary to
// handle errors.
func (tx *TransactionV1) ToBase64() string {
	data, err := tx.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(data)
}

// PartialSign signs the transaction with every private key the getter
// returns, leaving the other signature slots untouched. The getter is
// called once per required signer (AccountKeys[0..NumRequiredSignatures));
// returning nil skips that signer.
func (tx *TransactionV1) PartialSign(getter privateKeyGetter) ([]Signature, error) {
	if err := tx.sanitize(); err != nil {
		return nil, err
	}
	numSigners := int(tx.Header.NumRequiredSignatures)
	payload := tx.appendSigned(make([]byte, 0, tx.signedSize()))
	if len(payload)+64*numSigners > TransactionV1MaxSize {
		return nil, fmt.Errorf("v1 transaction: serialized size %d exceeds the maximum of %d", len(payload)+64*numSigners, TransactionV1MaxSize)
	}
	if len(tx.Signatures) == 0 {
		tx.Signatures = make([]Signature, numSigners)
	} else if len(tx.Signatures) != numSigners {
		return nil, fmt.Errorf("v1 transaction: %d signatures, header requires %d", len(tx.Signatures), numSigners)
	}
	for i := 0; i < numSigners; i++ {
		key := getter(tx.AccountKeys[i])
		if key == nil {
			continue
		}
		sig, err := key.Sign(payload)
		if err != nil {
			return nil, fmt.Errorf("signing with %s: %w", tx.AccountKeys[i], err)
		}
		tx.Signatures[i] = sig
	}
	return tx.Signatures, nil
}

// Sign signs the transaction with every required signer. The getter must
// return a private key for each of the first NumRequiredSignatures account
// keys.
func (tx *TransactionV1) Sign(getter privateKeyGetter) ([]Signature, error) {
	sigs, err := tx.PartialSign(getter)
	if err != nil {
		return nil, err
	}
	var zero Signature
	for i, sig := range sigs {
		if sig == zero {
			return nil, fmt.Errorf("signer key %s: key not found. Ensure the getter returns every required signer", tx.AccountKeys[i])
		}
	}
	return sigs, nil
}

// VerifySignatures verifies every signature against its signer key and the
// signed payload.
func (tx *TransactionV1) VerifySignatures() error {
	if err := tx.sanitize(); err != nil {
		return err
	}
	if len(tx.Signatures) != int(tx.Header.NumRequiredSignatures) {
		return fmt.Errorf("v1 transaction: %d signatures, header requires %d", len(tx.Signatures), tx.Header.NumRequiredSignatures)
	}
	payload := tx.appendSigned(make([]byte, 0, tx.signedSize()))
	for i, sig := range tx.Signatures {
		if !sig.Verify(tx.AccountKeys[i], payload) {
			return fmt.Errorf("invalid signature by %s", tx.AccountKeys[i])
		}
	}
	return nil
}

// NewTransactionV1 compiles instructions into an unsigned v1 transaction.
// Account ordering, deduplication and fee-payer handling follow
// NewTransaction; the TransactionAddressTables option is rejected because
// the v1 format has no address lookup tables (at 4096 bytes the full
// address list fits directly). Resource requests travel in config instead
// of ComputeBudgetProgram instructions — v1 ignores those for
// configuration, so do not add them.
func NewTransactionV1(instructions []Instruction, lifetime Hash, config TransactionConfig, opts ...TransactionOption) (*TransactionV1, error) {
	var probe transactionOptions
	for _, opt := range opts {
		opt.apply(&probe)
	}
	if probe.addressTables != nil {
		return nil, errors.New("v1 transactions do not support address lookup tables")
	}

	compiled, err := NewTransaction(instructions, lifetime, opts...)
	if err != nil {
		return nil, err
	}
	tx := &TransactionV1{
		Header:            compiled.Message.Header,
		Config:            config,
		LifetimeSpecifier: lifetime,
		AccountKeys:       compiled.Message.AccountKeys,
		Instructions:      compiled.Message.Instructions,
	}
	if err := tx.sanitize(); err != nil {
		return nil, err
	}
	if size := tx.signedSize() + 64*int(tx.Header.NumRequiredSignatures); size > TransactionV1MaxSize {
		return nil, fmt.Errorf("v1 transaction: serialized size %d exceeds the maximum of %d", size, TransactionV1MaxSize)
	}
	return tx, nil
}
