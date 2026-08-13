package solana_go

import (
	"encoding/base64"
	"fmt"

	"github.com/bytedance/sonic"
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

// Data is content plus the encoding it travels in, JSON-encoded as the RPC
// tuple ["<encoded content>", "<encoding>"].
// NOTE: base64+zstd data cannot be decoded by this SDK (kept out to avoid a
// compression dependency); request base64 instead.
type Data struct {
	Content  []byte
	Encoding EncodingType
}

func (t Data) MarshalJSON() ([]byte, error) {
	// Refuse to silently drop content: String() can only render the
	// encodings this SDK supports. The encoding is also echoed into the
	// output verbatim, so unknown values (only reachable by constructing a
	// Data directly) must not slip through unescaped.
	switch t.Encoding {
	case EncodingBase58, EncodingBase64:
	case EncodingBase64Zstd, "":
		if len(t.Content) > 0 {
			return nil, fmt.Errorf("cannot marshal Data with unsupported encoding %q", t.Encoding)
		}
	default:
		return nil, fmt.Errorf("cannot marshal Data with unsupported encoding %q", t.Encoding)
	}

	// ["<content>","<encoding>"] built directly; both halves are escape-free.
	content := t.String()
	buf := make([]byte, 0, len(content)+len(t.Encoding)+8)
	buf = append(buf, '[', '"')
	buf = append(buf, content...)
	buf = append(buf, `","`...)
	buf = append(buf, t.Encoding...)
	buf = append(buf, '"', ']')
	return buf, nil
}

func (t *Data) UnmarshalJSON(data []byte) error {
	var in []string
	if err := sonic.Unmarshal(data, &in); err != nil {
		return err
	}
	if len(in) != 2 {
		return fmt.Errorf("invalid length for Data, expected 2, found %d", len(in))
	}

	// Validate the encoding even for empty content: it is echoed verbatim by
	// MarshalJSON, so unknown values must never be stored. The empty
	// encoding is tolerated for empty content only, because that is how a
	// zero Data value marshals.
	t.Encoding = EncodingType(in[1])
	switch t.Encoding {
	case EncodingBase58, EncodingBase64, EncodingBase64Zstd:
	case "":
		if in[0] != "" {
			return fmt.Errorf("unsupported encoding %s", in[1])
		}
	default:
		return fmt.Errorf("unsupported encoding %s", in[1])
	}

	if in[0] == "" {
		t.Content = []byte{}
		return nil
	}

	var err error
	switch t.Encoding {
	case EncodingBase58:
		t.Content, err = base58.Decode(in[0])
	case EncodingBase64:
		t.Content, err = base64.StdEncoding.DecodeString(in[0])
	default: // EncodingBase64Zstd
		err = fmt.Errorf("base64+zstd data is not supported by this SDK; request base64 instead")
	}
	return err
}

// String returns the encoded form of the content per the Data's encoding.
func (t Data) String() string {
	switch t.Encoding {
	case EncodingBase58:
		return base58.Encode(t.Content)
	case EncodingBase64:
		return base64.StdEncoding.EncodeToString(t.Content)
	default:
		return ""
	}
}

// GetBinary returns the raw decoded content bytes.
func (t Data) GetBinary() []byte {
	return t.Content
}
