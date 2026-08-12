package solana_go

import "github.com/bytedance/sonic"

// jsonUnquote extracts the value of a JSON-encoded string. Base58 and base64
// payloads never contain characters that need JSON escaping, so it takes a
// fast path that skips the general JSON string decoder, while still falling
// back to it for escaped input.
func jsonUnquote(data []byte) (string, error) {
	if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
		raw := data[1 : len(data)-1]
		clean := true
		for _, c := range raw {
			if c == '\\' || c == '"' || c < 0x20 {
				clean = false
				break
			}
		}
		if clean {
			return string(raw), nil
		}
	}
	var s string
	err := sonic.Unmarshal(data, &s)
	return s, err
}
