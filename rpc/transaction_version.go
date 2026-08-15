package rpc

import "strconv"

// TransactionVersion is the version of a transaction as reported by the RPC:
// either the number of a versioned transaction (0, ...) or
// LegacyTransactionVersion for legacy transactions (JSON "legacy").
type TransactionVersion int

const (
	// LegacyTransactionVersion is reported for legacy (unversioned)
	// transactions, serialized as the JSON string "legacy".
	LegacyTransactionVersion TransactionVersion = -1
	legacyVersion                               = `"legacy"`
)

// UnmarshalJSON implements json.Unmarshaler, decoding a version number;
// "legacy" (as well as null and "") decodes to LegacyTransactionVersion.
func (a *TransactionVersion) UnmarshalJSON(b []byte) error {
	// Ignore null, like in the main JSON package.
	s := string(b)
	if s == "null" || s == `""` || s == legacyVersion {
		*a = LegacyTransactionVersion
		return nil
	}

	v, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*a = TransactionVersion(v)
	return nil
}

// MarshalJSON implements json.Marshaler, encoding the version as a number,
// or as the string "legacy" for LegacyTransactionVersion.
func (a TransactionVersion) MarshalJSON() ([]byte, error) {
	if a == LegacyTransactionVersion {
		return []byte(legacyVersion), nil
	}
	return []byte(strconv.Itoa(int(a))), nil
}
