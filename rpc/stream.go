package rpc

import (
	"fmt"
	"io"

	"github.com/bytedance/sonic"
)

// RPCError is the error object of a JSON-RPC response envelope.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Error implements the error interface, formatting the JSON-RPC error code
// and message.
func (e *RPCError) Error() string {
	return fmt.Sprintf("rpc error %d: %s", e.Code, e.Message)
}

// StreamProgramAccounts incrementally decodes a getProgramAccounts JSON-RPC
// response body, invoking fn for each account as soon as its bytes arrive —
// decoding overlaps the download instead of waiting for the full (often very
// large) result to buffer, and memory stays bounded by the largest single
// account rather than the whole response. Both the plain and the withContext
// response shapes are handled; the context is returned when present.
//
// fn returning an error aborts the stream and returns that error. A JSON-RPC
// error envelope is returned as *RPCError. The *KeyedAccount is fully owned
// by the callback: every field copies out of the wire buffer during decode,
// so retaining it is safe.
func StreamProgramAccounts(r io.Reader, fn func(*KeyedAccount) error) (*Context, error) {
	s := &jsonScanner{r: r, buf: make([]byte, 0, 64<<10)}

	if err := s.expect('{'); err != nil {
		return nil, fmt.Errorf("response envelope: %w", err)
	}

	var ctx *Context
	err := s.forEachKey(func(key string) error {
		switch key {
		case "error":
			raw, err := s.scanValue()
			if err != nil {
				return err
			}
			rpcErr := new(RPCError)
			if err := sonic.Unmarshal(raw, rpcErr); err != nil {
				return fmt.Errorf("error object: %w", err)
			}
			return rpcErr
		case "result":
			c, err := s.peekNonSpace()
			if err != nil {
				return err
			}
			switch c {
			case '[':
				s.pos++
				return s.streamAccounts(fn)
			case '{':
				s.pos++
				return s.forEachKey(func(inner string) error {
					switch inner {
					case "context":
						raw, err := s.scanValue()
						if err != nil {
							return err
						}
						ctx = new(Context)
						if err := sonic.Unmarshal(raw, ctx); err != nil {
							return fmt.Errorf("context: %w", err)
						}
						return nil
					case "value":
						if err := s.expect('['); err != nil {
							return fmt.Errorf("result value: %w", err)
						}
						return s.streamAccounts(fn)
					default:
						_, err := s.scanValue()
						return err
					}
				})
			default: // null
				_, err := s.scanValue()
				return err
			}
		default: // jsonrpc, id, ...
			_, err := s.scanValue()
			return err
		}
	})
	if err != nil {
		return nil, err
	}
	return ctx, nil
}

// Batching bounds for streamAccounts: enough to amortize per-decode setup,
// small enough that delivery stays prompt and memory stays flat.
const (
	streamBatchElems = 32
	streamBatchBytes = 256 << 10
)

// streamAccounts consumes array elements after '['. Elements are batched
// into one decode call to amortize decoder setup, but a batch is flushed the
// moment the wire buffer runs dry — so on a slow link every account is
// delivered as soon as its bytes arrive, while a fast stream pays the
// decoder cost only once per 32 accounts.
func (s *jsonScanner) streamAccounts(fn func(*KeyedAccount) error) error {
	c, err := s.peekNonSpace()
	if err != nil {
		return err
	}
	if c == ']' {
		s.pos++
		return nil
	}

	batch := make([]byte, 0, 4096)
	count := 0
	var accounts []*KeyedAccount
	flush := func() error {
		if count == 0 {
			return nil
		}
		batch = append(batch, ']')
		accounts = accounts[:0]
		if err := sonic.Unmarshal(batch, &accounts); err != nil {
			return fmt.Errorf("decoding accounts: %w", err)
		}
		batch = batch[:0]
		count = 0
		for _, account := range accounts {
			if err := fn(account); err != nil {
				return err
			}
		}
		return nil
	}

	for {
		raw, err := s.scanValue()
		if err != nil {
			return fmt.Errorf("account element: %w", err)
		}
		if count == 0 {
			batch = append(batch, '[')
		} else {
			batch = append(batch, ',')
		}
		batch = append(batch, raw...)
		count++
		// s.pos == s.end means the next read would block on the network:
		// deliver what we have instead of sitting on it.
		if count >= streamBatchElems || len(batch) >= streamBatchBytes || s.pos >= s.end {
			if err := flush(); err != nil {
				return err
			}
		}

		c, err := s.nextNonSpace()
		if err != nil {
			return err
		}
		switch c {
		case ']':
			return flush()
		case ',':
		default:
			return fmt.Errorf("expected ',' or ']' after account, got %q", c)
		}
	}
}

// jsonScanner is a minimal incremental JSON splitter: it finds the byte
// boundaries of keys and values as data arrives from r, so values can be
// handed to a real decoder one at a time without buffering the whole body.
// buf[pos:end] is unconsumed data; mark is the earliest offset that must
// survive a refill (the start of the value being scanned).
type jsonScanner struct {
	r    io.Reader
	buf  []byte
	pos  int
	end  int
	mark int
}

// fill compacts the buffer (dropping everything before mark) and reads more
// data, growing the buffer when a value is larger than what fits.
func (s *jsonScanner) fill() error {
	if s.mark > 0 {
		copy(s.buf, s.buf[s.mark:s.end])
		s.pos -= s.mark
		s.end -= s.mark
		s.mark = 0
	}
	if s.end == cap(s.buf) {
		grown := make([]byte, cap(s.buf)*2)
		copy(grown, s.buf[:s.end])
		s.buf = grown
	}
	s.buf = s.buf[:cap(s.buf)]
	n, err := s.r.Read(s.buf[s.end:])
	s.end += n
	if n > 0 {
		return nil
	}
	if err == nil || err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	return err
}

func isJSONSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func (s *jsonScanner) peekNonSpace() (byte, error) {
	for {
		for s.pos < s.end {
			if c := s.buf[s.pos]; !isJSONSpace(c) {
				return c, nil
			}
			s.pos++
			s.mark = s.pos
		}
		if err := s.fill(); err != nil {
			return 0, err
		}
	}
}

func (s *jsonScanner) nextNonSpace() (byte, error) {
	c, err := s.peekNonSpace()
	if err != nil {
		return 0, err
	}
	s.pos++
	return c, nil
}

func (s *jsonScanner) expect(want byte) error {
	c, err := s.nextNonSpace()
	if err != nil {
		return err
	}
	if c != want {
		return fmt.Errorf("expected %q, got %q", want, c)
	}
	return nil
}

// forEachKey walks the keys of an object whose '{' is already consumed,
// invoking fn with the scanner positioned at each key's value.
func (s *jsonScanner) forEachKey(fn func(key string) error) error {
	first := true
	for {
		c, err := s.nextNonSpace()
		if err != nil {
			return err
		}
		if c == '}' {
			return nil
		}
		if !first {
			if c != ',' {
				return fmt.Errorf("expected ',' between object keys, got %q", c)
			}
			if c, err = s.nextNonSpace(); err != nil {
				return err
			}
		}
		first = false
		if c != '"' {
			return fmt.Errorf("expected object key, got %q", c)
		}
		key, err := s.scanStringRest()
		if err != nil {
			return err
		}
		if err := s.expect(':'); err != nil {
			return err
		}
		if err := fn(string(key)); err != nil {
			return err
		}
	}
}

// scanStringRest consumes a JSON string whose opening quote is already
// consumed and returns its raw contents (escapes unprocessed — fine for the
// envelope keys this scanner compares against).
func (s *jsonScanner) scanStringRest() ([]byte, error) {
	s.mark = s.pos - 1
	start := s.pos
	escaped := false
	for {
		for s.pos < s.end {
			c := s.buf[s.pos]
			s.pos++
			if escaped {
				escaped = false
				continue
			}
			switch c {
			case '\\':
				escaped = true
			case '"':
				out := s.buf[start : s.pos-1]
				s.mark = s.pos
				return out, nil
			}
		}
		before := s.mark
		if err := s.fill(); err != nil {
			return nil, err
		}
		start -= before - s.mark // fill shifted everything by (before - new mark)
	}
}

// scanValue consumes one complete JSON value and returns its raw bytes,
// valid until the next scanner call.
func (s *jsonScanner) scanValue() ([]byte, error) {
	if _, err := s.peekNonSpace(); err != nil {
		return nil, err
	}
	s.mark = s.pos
	start := s.pos
	depth := 0
	inString := false
	escaped := false
	for {
		for s.pos < s.end {
			c := s.buf[s.pos]
			if inString {
				s.pos++
				if escaped {
					escaped = false
				} else if c == '\\' {
					escaped = true
				} else if c == '"' {
					inString = false
					if depth == 0 {
						out := s.buf[start:s.pos]
						s.mark = s.pos
						return out, nil
					}
				}
				continue
			}
			switch c {
			case '"':
				inString = true
			case '{', '[':
				depth++
			case '}', ']':
				if depth == 0 {
					// Terminator of the enclosing container: primitive ends here.
					out := s.buf[start:s.pos]
					s.mark = s.pos
					return out, nil
				}
				depth--
				if depth == 0 {
					s.pos++
					out := s.buf[start:s.pos]
					s.mark = s.pos
					return out, nil
				}
			case ',':
				if depth == 0 {
					out := s.buf[start:s.pos]
					s.mark = s.pos
					return out, nil
				}
			}
			s.pos++
		}
		before := s.mark
		if err := s.fill(); err != nil {
			if err == io.ErrUnexpectedEOF && depth == 0 && !inString && s.pos > start {
				// A top-level primitive can legitimately end at EOF.
				out := s.buf[start:s.pos]
				s.mark = s.pos
				return out, nil
			}
			return nil, err
		}
		start -= before - s.mark
	}
}
