package rpc

import "strconv"

// TransactionVersion is the version of a transaction as reported by the RPC:
// either the number of a versioned transaction (0, ...) or
// LegacyTransactionVersion for legacy transactions (JSON "legacy").
type TransactionVersion int

const (
	LegacyTransactionVersion TransactionVersion = -1
	legacyVersion                               = `"legacy"`
)

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

func (a TransactionVersion) MarshalJSON() ([]byte, error) {
	if a == LegacyTransactionVersion {
		return []byte(legacyVersion), nil
	}
	return []byte(strconv.Itoa(int(a))), nil
}
