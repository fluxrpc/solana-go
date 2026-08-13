package solana_go

// EncodingType is the data encoding requested from, or returned by, the RPC.
type EncodingType string

const (
	EncodingBase58     EncodingType = "base58"      // limited to account data of less than 129 bytes
	EncodingBase64     EncodingType = "base64"      // base64 encoded data of any size
	EncodingBase64Zstd EncodingType = "base64+zstd" // zstd-compressed then base64-encoded

	// EncodingJSONParsed attempts to use program-specific state parsers to
	// return more human-readable and explicit account state data. If a parser
	// cannot be found, the field falls back to base64 encoding.
	// Cannot be used if specifying dataSlice parameters (offset, length).
	EncodingJSONParsed EncodingType = "jsonParsed"

	EncodingJSON EncodingType = "json" // NOTE: you're probably looking for EncodingJSONParsed
)

// IsAnyOfEncodingType checks whether the provided candidate is any of the allowed.
func IsAnyOfEncodingType(candidate EncodingType, allowed ...EncodingType) bool {
	for _, v := range allowed {
		if candidate == v {
			return true
		}
	}
	return false
}
