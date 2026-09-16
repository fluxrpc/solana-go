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
	numLookups := mx.AddressTableLookups.NumLookups()
	if numLookups > 0 && len(mx.addressTables) == 0 {
		return nil, ErrAddressTablesNotSet
	}

	numWritable := mx.AddressTableLookups.NumWritableLookups()
	writable := make(PublicKeySlice, 0, numWritable)
	readonly := make(PublicKeySlice, 0, numLookups-numWritable)

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
			writable = append(writable, table[at])
		}
		for _, at := range lookup.ReadonlyIndexes {
			if int(at) >= len(table) {
				return nil, fmt.Errorf("%w: %d of %d in %s", ErrAddressTableIndexRange, at, len(table), lookup.AccountKey)
			}
			readonly = append(readonly, table[at])
		}
	}
	return append(writable, readonly...), nil
}

// ResolveLookups appends the looked-up accounts to AccountKeys. It is a no-op
// once resolved.
func (mx *Message) ResolveLookups() error {
	if mx.resolved {
		return nil
	}
	accounts, err := mx.GetAddressTableLookupAccounts()
	if err != nil {
		return err
	}
	mx.AccountKeys = append(mx.AccountKeys, accounts...)
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
	accounts, err := mx.GetAddressTableLookupAccounts()
	if err != nil {
		return nil, err
	}
	all := make(PublicKeySlice, len(mx.AccountKeys), len(mx.AccountKeys)+len(accounts))
	copy(all, mx.AccountKeys)
	return append(all, accounts...), nil
}
