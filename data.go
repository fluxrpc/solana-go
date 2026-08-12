package solana_go

import (
	"encoding/base64"

	"github.com/fluxrpc/base58"
)

// Base58 is a byte slice that is JSON-encoded as a base58 string.
type Base58 []byte

func (t Base58) MarshalJSON() ([]byte, error) {
	// Base58 characters never need JSON escaping, so write the quoted string
	// directly instead of going through json.Marshal.
	buf := make([]byte, 0, len(t)*2+2)
	buf = append(buf, '"')
	buf = base58.AppendEncode(buf, t)
	buf = append(buf, '"')
	return buf, nil
}

func (t *Base58) UnmarshalJSON(data []byte) (err error) {
	s, err := jsonUnquote(data)
	if err != nil {
		return err
	}
	if s == "" {
		*t = Base58{}
		return nil
	}
	*t, err = base58.Decode(s)
	return err
}

func (t Base58) String() string {
	return base58.Encode(t)
}

// Base64 is a byte slice that is JSON-encoded as a standard base64 string.
type Base64 []byte

func (t Base64) MarshalJSON() ([]byte, error) {
	// The standard base64 alphabet never needs JSON escaping.
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(t))+2)
	buf[0] = '"'
	base64.StdEncoding.Encode(buf[1:], t)
	buf[len(buf)-1] = '"'
	return buf, nil
}

func (t *Base64) UnmarshalJSON(data []byte) (err error) {
	s, err := jsonUnquote(data)
	if err != nil {
		return err
	}
	if s == "" {
		*t = Base64{}
		return nil
	}
	*t, err = base64.StdEncoding.DecodeString(s)
	return err
}

func (t Base64) String() string {
	return base64.StdEncoding.EncodeToString(t)
}
