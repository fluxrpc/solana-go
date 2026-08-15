package rpc

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	solana "github.com/fluxrpc/solana-go"
)

// DataBytesOrJSON holds account (or transaction) data that the RPC returned
// either as an encoded-binary tuple ["<content>","<encoding>"] or, for the
// jsonParsed encoding, as a free-form JSON object.
type DataBytesOrJSON struct {
	rawDataEncoding solana.EncodingType
	asDecodedBinary solana.Data
	asJSON          json.RawMessage
}

// DataBytesOrJSONFromBase64 creates a new DataBytesOrJSON from the provided
// base64-encoded string.
func DataBytesOrJSONFromBase64(stringBase64 string) (*DataBytesOrJSON, error) {
	decodedData, err := base64.StdEncoding.DecodeString(stringBase64)
	if err != nil {
		return nil, err
	}
	return DataBytesOrJSONFromBytes(decodedData), nil
}

// DataBytesOrJSONFromBytes creates a new DataBytesOrJSON from the provided bytes.
func DataBytesOrJSONFromBytes(data []byte) *DataBytesOrJSON {
	return &DataBytesOrJSON{
		rawDataEncoding: solana.EncodingBase64,
		asDecodedBinary: solana.Data{
			Encoding: solana.EncodingBase64,
			Content:  data,
		},
	}
}

// MarshalJSON implements json.Marshaler, encoding json/jsonParsed data as
// its raw JSON payload (null when absent) and binary data as the
// ["<content>","<encoding>"] tuple.
func (dt DataBytesOrJSON) MarshalJSON() ([]byte, error) {
	if dt.rawDataEncoding == solana.EncodingJSONParsed || dt.rawDataEncoding == solana.EncodingJSON {
		if dt.asJSON == nil {
			return []byte("null"), nil
		}
		return dt.asJSON, nil
	}
	return dt.asDecodedBinary.MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler, accepting either the
// ["<content>","<encoding>"] binary tuple or a jsonParsed object, whose raw
// bytes are kept for GetRawJSON. A null input is ignored.
func (wrap *DataBytesOrJSON) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	switch data[0] {
	// A JSON array is the ["<content>","<encoding>"] binary tuple.
	case '[':
		if err := wrap.asDecodedBinary.UnmarshalJSON(data); err != nil {
			return err
		}
		wrap.rawDataEncoding = wrap.asDecodedBinary.Encoding
	// A JSON object is a jsonParsed payload; keep the raw bytes and let the
	// caller decode them on request.
	case '{':
		wrap.asJSON = append(json.RawMessage(nil), data...)
		wrap.rawDataEncoding = solana.EncodingJSONParsed
	default:
		return fmt.Errorf("unknown kind: %v", data)
	}

	return nil
}

// GetBinary returns the decoded bytes if the encoding is "base58" or "base64".
func (dt *DataBytesOrJSON) GetBinary() []byte {
	if dt == nil {
		return nil
	}
	return dt.asDecodedBinary.Content
}

// GetRawJSON returns a json.RawMessage when the data encoding is "jsonParsed".
func (dt *DataBytesOrJSON) GetRawJSON() json.RawMessage {
	return dt.asJSON
}
