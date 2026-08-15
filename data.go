package solana_go

import (
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/fluxrpc/base58"
	"github.com/klauspost/compress/zstd"
)

// Base58 is a byte slice that is JSON-encoded as a base58 string.
type Base58 []byte

// MarshalJSON implements json.Marshaler, encoding the bytes as a base58
// JSON string.
func (t Base58) MarshalJSON() ([]byte, error) {
	// Base58 characters never need JSON escaping, so write the quoted string
	// directly instead of going through json.Marshal.
	buf := make([]byte, 0, len(t)*2+2)
	buf = append(buf, '"')
	buf = base58.AppendEncode(buf, t)
	buf = append(buf, '"')
	return buf, nil
}

// UnmarshalJSON implements json.Unmarshaler, decoding a base58 JSON
// string.
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

// String returns the base58 representation of the bytes.
func (t Base58) String() string {
	return base58.Encode(t)
}

// Base64 is a byte slice that is JSON-encoded as a standard base64 string.
type Base64 []byte

// MarshalJSON implements json.Marshaler, encoding the bytes as a standard
// base64 JSON string.
func (t Base64) MarshalJSON() ([]byte, error) {
	// The standard base64 alphabet never needs JSON escaping.
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(t))+2)
	buf[0] = '"'
	base64.StdEncoding.Encode(buf[1:], t)
	buf[len(buf)-1] = '"'
	return buf, nil
}

// UnmarshalJSON implements json.Unmarshaler, decoding a standard base64
// JSON string.
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

// String returns the standard base64 representation of the bytes.
func (t Base64) String() string {
	return base64.StdEncoding.EncodeToString(t)
}

// Data is content plus the encoding it travels in, JSON-encoded as the RPC
// tuple ["<encoded content>", "<encoding>"]. All spec encodings are
// supported, including base64+zstd.
type Data struct {
	Content  []byte
	Encoding EncodingType
}

// Package-level zstd codecs, lazily initialized. Both are safe for
// concurrent EncodeAll/DecodeAll use; the decoder's default 64MiB memory
// limit comfortably bounds any legitimate account payload.
var (
	zstdDecoder     *zstd.Decoder
	zstdDecoderOnce sync.Once
	zstdEncoder     *zstd.Encoder
	zstdEncoderOnce sync.Once
)

func getZstdDecoder() *zstd.Decoder {
	zstdDecoderOnce.Do(func() {
		zstdDecoder, _ = zstd.NewReader(nil, zstd.WithDecoderConcurrency(0))
	})
	return zstdDecoder
}

func getZstdEncoder() *zstd.Encoder {
	zstdEncoderOnce.Do(func() {
		zstdEncoder, _ = zstd.NewWriter(nil, zstd.WithEncoderConcurrency(1))
	})
	return zstdEncoder
}

// MarshalJSON writes ["<content>","<encoding>"] directly into one buffer,
// appending the encoded content in place. It refuses encodings it cannot
// render (only reachable by constructing a Data directly), rather than
// silently dropping content or echoing unknown strings unescaped.
func (t Data) MarshalJSON() ([]byte, error) {
	switch t.Encoding {
	case EncodingBase58:
		buf := make([]byte, 0, len(t.Content)*2+len(t.Encoding)+8)
		buf = append(buf, '[', '"')
		buf = base58.AppendEncode(buf, t.Content)
		buf = append(buf, `","`...)
		buf = append(buf, t.Encoding...)
		return append(buf, '"', ']'), nil
	case EncodingBase64:
		n := base64.StdEncoding.EncodedLen(len(t.Content))
		buf := make([]byte, 0, n+len(t.Encoding)+8)
		buf = append(buf, '[', '"')
		off := len(buf)
		buf = buf[:off+n]
		base64.StdEncoding.Encode(buf[off:], t.Content)
		buf = append(buf, `","`...)
		buf = append(buf, t.Encoding...)
		return append(buf, '"', ']'), nil
	case EncodingBase64Zstd:
		if len(t.Content) == 0 {
			return []byte(`["","` + EncodingBase64Zstd + `"]`), nil
		}
		compressed := getZstdEncoder().EncodeAll(t.Content, nil)
		n := base64.StdEncoding.EncodedLen(len(compressed))
		buf := make([]byte, 0, n+len(t.Encoding)+8)
		buf = append(buf, '[', '"')
		off := len(buf)
		buf = buf[:off+n]
		base64.StdEncoding.Encode(buf[off:], compressed)
		buf = append(buf, `","`...)
		buf = append(buf, t.Encoding...)
		return append(buf, '"', ']'), nil
	case "":
		// The zero value is representable only with no content.
		if len(t.Content) > 0 {
			return nil, fmt.Errorf("cannot marshal Data with unsupported encoding %q", t.Encoding)
		}
		return []byte(`["",""]`), nil
	default:
		return nil, fmt.Errorf("cannot marshal Data with unsupported encoding %q", t.Encoding)
	}
}

// UnmarshalJSON implements json.Unmarshaler, decoding the RPC data tuple
// and decompressing base64+zstd content.
func (t *Data) UnmarshalJSON(data []byte) error {
	// Fast path: parse the ["<content>","<encoding>"] tuple in place. Both
	// halves are escape-free in well-formed RPC output; anything unexpected
	// falls back to the general decoder.
	content, encoding, ok := parseStringPair(data)
	if !ok {
		var in []string
		if err := sonic.Unmarshal(data, &in); err != nil {
			return err
		}
		if len(in) != 2 {
			return fmt.Errorf("invalid length for Data, expected 2, found %d", len(in))
		}
		content, encoding = in[0], in[1]
	}

	// Validate and intern the encoding: it is echoed verbatim by
	// MarshalJSON, so unknown values must never be stored — and interning
	// means the stored name never aliases the input buffer. The empty
	// encoding is tolerated for empty content only, because that is how a
	// zero Data value marshals.
	switch encoding {
	case string(EncodingBase58):
		t.Encoding = EncodingBase58
	case string(EncodingBase64):
		t.Encoding = EncodingBase64
	case string(EncodingBase64Zstd):
		t.Encoding = EncodingBase64Zstd
	case "":
		if content != "" {
			return fmt.Errorf("unsupported encoding %q", encoding)
		}
		t.Encoding = ""
	default:
		return fmt.Errorf("unsupported encoding %q", encoding)
	}

	if content == "" {
		t.Content = []byte{}
		return nil
	}

	var err error
	switch t.Encoding {
	case EncodingBase58:
		t.Content, err = base58.Decode(content)
	case EncodingBase64:
		t.Content, err = base64.StdEncoding.DecodeString(content)
	default: // EncodingBase64Zstd
		var compressed []byte
		if compressed, err = base64.StdEncoding.DecodeString(content); err == nil {
			t.Content, err = getZstdDecoder().DecodeAll(compressed, nil)
		}
	}
	return err
}

// parseStringPair parses a two-element JSON array of escape-free strings,
// e.g. ["abc","base64"], returning views into data (not copies). It reports
// ok=false for any other shape (escapes, extra elements, non-strings), in
// which case the caller must re-parse with a general JSON decoder.
func parseStringPair(data []byte) (first, second string, ok bool) {
	i := skipJSONSpace(data, 0)
	if i >= len(data) || data[i] != '[' {
		return "", "", false
	}
	first, i, ok = parseSimpleJSONString(data, skipJSONSpace(data, i+1))
	if !ok {
		return "", "", false
	}
	i = skipJSONSpace(data, i)
	if i >= len(data) || data[i] != ',' {
		return "", "", false
	}
	second, i, ok = parseSimpleJSONString(data, skipJSONSpace(data, i+1))
	if !ok {
		return "", "", false
	}
	i = skipJSONSpace(data, i)
	if i >= len(data) || data[i] != ']' {
		return "", "", false
	}
	if skipJSONSpace(data, i+1) != len(data) {
		return "", "", false
	}
	return first, second, true
}

// parseSimpleJSONString parses an escape-free JSON string starting at
// data[i], returning its contents as a view into data and the index just
// past the closing quote.
func parseSimpleJSONString(data []byte, i int) (s string, next int, ok bool) {
	if i >= len(data) || data[i] != '"' {
		return "", 0, false
	}
	i++
	start := i
	for ; i < len(data); i++ {
		c := data[i]
		if c == '"' {
			return unsafeString(data[start:i]), i + 1, true
		}
		if c == '\\' || c < 0x20 {
			return "", 0, false
		}
	}
	return "", 0, false
}

func skipJSONSpace(data []byte, i int) int {
	for i < len(data) {
		switch data[i] {
		case ' ', '\t', '\n', '\r':
			i++
		default:
			return i
		}
	}
	return i
}

// String returns the encoded form of the content per the Data's encoding.
func (t Data) String() string {
	switch t.Encoding {
	case EncodingBase58:
		return base58.Encode(t.Content)
	case EncodingBase64:
		return base64.StdEncoding.EncodeToString(t.Content)
	case EncodingBase64Zstd:
		return base64.StdEncoding.EncodeToString(getZstdEncoder().EncodeAll(t.Content, nil))
	default:
		return ""
	}
}

// GetBinary returns the raw decoded content bytes.
func (t Data) GetBinary() []byte {
	return t.Content
}
