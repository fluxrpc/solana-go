package solana_go

import (
	"errors"
	"fmt"
)

var (
	ErrAddressTablesAlreadySet = errors.New("address tables already set")
	ErrAddressTableNotFound    = errors.New("address table not found")
	ErrAddressTableIndexRange  = errors.New("address table index out of range")
)

// SetAddressTables supplies the table contents, keyed by table address.
func (mx *Message) SetAddressTables(tables map[PublicKey]PublicKeySlice) error {
	if mx.addressTables != nil {
		return ErrAddressTablesAlreadySet
	}
	mx.addressTables = tables
	return nil
}

func (mx *Message) GetAddressTables() map[PublicKey]PublicKeySlice {
	return mx.addressTables
}

func (mx *Message) GetAddressTableLookups() MessageAddressTableLookupSlice {
	return mx.AddressTableLookups
}

// GetAddressTableLookupAccounts returns every writable index across all
// lookups, then every readonly one. Compiled instructions index into that
// order, so it is not arbitrary.
func (mx *Message) GetAddressTableLookupAccounts() (PublicKeySlice, error) {
	return mx.lookupAccounts(nil)
}

// lookupAccounts fills a single allocation in [prefix][writable][readonly]
// order. Each table is fetched once, and no intermediate key slices are built.
func (mx *Message) lookupAccounts(prefix []PublicKey) (PublicKeySlice, error) {
	numLookups, numWritable := 0, 0
	for i := range mx.AddressTableLookups {
		lookup := &mx.AddressTableLookups[i]
		numWritable += len(lookup.WritableIndexes)
		numLookups += len(lookup.WritableIndexes) + len(lookup.ReadonlyIndexes)
	}
	if numLookups > 0 && len(mx.addressTables) == 0 {
		return nil, ErrAddressTablesNotSet
	}

	accounts := make(PublicKeySlice, len(prefix)+numLookups)
	copy(accounts, prefix)
	writable, readonly := len(prefix), len(prefix)+numWritable

	for index := range mx.AddressTableLookups {
		lookup := &mx.AddressTableLookups[index]
		table, ok := mx.addressTables[lookup.AccountKey]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrAddressTableNotFound, lookup.AccountKey)
		}
		for _, at := range lookup.WritableIndexes {
			if int(at) >= len(table) {
				return nil, fmt.Errorf("%w: %d of %d in %s", ErrAddressTableIndexRange, at, len(table), lookup.AccountKey)
			}
			accounts[writable] = table[at]
			writable++
		}
		for _, at := range lookup.ReadonlyIndexes {
			if int(at) >= len(table) {
				return nil, fmt.Errorf("%w: %d of %d in %s", ErrAddressTableIndexRange, at, len(table), lookup.AccountKey)
			}
			accounts[readonly] = table[at]
			readonly++
		}
	}
	return accounts, nil
}

// ResolveLookups appends the looked-up accounts to AccountKeys. It is a no-op
// once resolved.
func (mx *Message) ResolveLookups() error {
	if mx.resolved {
		return nil
	}
	numStatic := len(mx.AccountKeys)
	if len(mx.AddressTableLookups) > 0 {
		accounts, err := mx.lookupAccounts(mx.AccountKeys)
		if err != nil {
			return err
		}
		mx.AccountKeys = accounts
	}
	mx.numStaticAccounts = numStatic
	mx.resolved = true
	return nil
}

func (mx *Message) IsResolved() bool {
	return mx.resolved
}

// GetAllKeys returns the static keys followed by the looked-up ones, without
// mutating the message.
func (mx *Message) GetAllKeys() (PublicKeySlice, error) {
	if mx.resolved {
		return mx.AccountKeys, nil
	}
	return mx.lookupAccounts(mx.AccountKeys)
}

// staticAccountKeys keeps wire encoding and header-based permissions separate
// from the loaded accounts appended by ResolveLookups.
func (mx *Message) staticAccountKeys() []PublicKey {
	if mx.resolved {
		return mx.AccountKeys[:mx.numStaticAccounts]
	}
	return mx.AccountKeys
}

// invalidateLookups drops the loaded suffix before changing lookup descriptors.
func (mx *Message) invalidateLookups() {
	mx.AccountKeys = mx.staticAccountKeys()
	mx.resolved = false
	mx.numStaticAccounts = 0
}
