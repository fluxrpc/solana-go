package rpc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bytedance/sonic"
	solana "github.com/fluxrpc/solana-go"
)

// ErrNotFound is returned when the queried item (account, transaction,
// block) does not exist or is not available on the queried node.
var ErrNotFound = errors.New("not found")

// Client is a Solana JSON-RPC HTTP client over the types in this package.
// It is safe for concurrent use; throughput comes from connection reuse,
// pooled response buffers and a single sonic pass over each response.
type Client struct {
	url  string
	http *http.Client
	id   atomic.Uint64

	headersMu sync.RWMutex
	headers   http.Header

	// cache is the optional account cache; nil when disabled. See
	// EnableCache.
	cache atomic.Pointer[accountCache]
}

// New creates a client for the given endpoint with a transport tuned for
// high request concurrency against a single host.
func New(url string) *Client {
	return NewWithClient(url, &http.Client{
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        128,
			MaxIdleConnsPerHost: 128,
			IdleConnTimeout:     90 * time.Second,
			ForceAttemptHTTP2:   true,
		},
	})
}

// NewWithClient creates a client using the provided http.Client.
func NewWithClient(url string, httpClient *http.Client) *Client {
	return &Client{url: url, http: httpClient}
}

// SetHeader sets a header sent with every request (e.g. authentication).
func (c *Client) SetHeader(key, value string) {
	c.headersMu.Lock()
	defer c.headersMu.Unlock()
	if c.headers == nil {
		c.headers = http.Header{}
	}
	c.headers.Set(key, value)
}

// bodyPool recycles response read buffers across calls.
var bodyPool = sync.Pool{
	New: func() any { return bytes.NewBuffer(make([]byte, 0, 64<<10)) },
}

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      uint64 `json:"id"`
	Method  string `json:"method"`
	Params  []any  `json:"params,omitempty"`
}

// do performs a JSON-RPC request and returns the response body via handle,
// which must consume it before returning (the buffer is recycled).
func (c *Client) do(ctx context.Context, method string, params []any, handle func(body []byte) error) error {
	payload, err := sonic.Marshal(rpcRequest{
		JSONRPC: "2.0",
		ID:      c.id.Add(1),
		Method:  method,
		Params:  params,
	})
	if err != nil {
		return fmt.Errorf("%s: encoding request: %w", method, err)
	}

	resp, err := c.post(ctx, payload)
	if err != nil {
		return fmt.Errorf("%s: %w", method, err)
	}
	defer resp.Body.Close()

	buf := bodyPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bodyPool.Put(buf)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return fmt.Errorf("%s: reading response: %w", method, err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: http %d: %.200s", method, resp.StatusCode, buf.Bytes())
	}
	if err := handle(buf.Bytes()); err != nil {
		return fmt.Errorf("%s: %w", method, err)
	}
	return nil
}

func (c *Client) post(ctx context.Context, payload []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	c.headersMu.RLock()
	headers := c.headers.Clone()
	c.headersMu.RUnlock()
	for key, values := range headers {
		req.Header[key] = values
	}
	return c.http.Do(req)
}

// call performs a JSON-RPC request and decodes its result into result. The
// caller must pass a pointer. Keeping the decoder on Client makes ownership
// of transport and response handling explicit while still decoding
// the whole response envelope in one sonic pass.
func (c *Client) call(ctx context.Context, method string, result any, params ...any) error {
	envelope := struct {
		Result any       `json:"result"`
		Error  *RPCError `json:"error"`
	}{Result: result}
	return c.do(ctx, method, params, func(body []byte) error {
		if err := sonic.Unmarshal(body, &envelope); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
		if envelope.Error != nil {
			return envelope.Error
		}
		return nil
	})
}

// callNullable is call for methods where a JSON null result means the item
// does not exist; it maps that to ErrNotFound.
func (c *Client) callNullable(ctx context.Context, method string, result any, params ...any) error {
	if err := c.call(ctx, method, result, params...); err != nil {
		return err
	}
	value := reflect.ValueOf(result)
	if value.Kind() != reflect.Pointer || value.IsNil() || value.Elem().Kind() != reflect.Pointer {
		return errors.New("nullable RPC result must be a non-nil pointer to a pointer")
	}
	if value.Elem().IsNil() {
		return ErrNotFound
	}
	return nil
}

// GetProgramAccountsStream invokes getProgramAccounts and decodes accounts
// incrementally off the response body as it downloads (see
// StreamProgramAccounts). The context is always requested so the response
// carries a slot even in streaming form.
func (c *Client) GetProgramAccountsStream(ctx context.Context, program solana.PublicKey, opts *GetProgramAccountsOpts, fn func(*KeyedAccount) error) (*Context, error) {
	withCtx := true
	requestOpts := GetProgramAccountsOpts{WithContext: &withCtx}
	if opts != nil {
		requestOpts = *opts
		requestOpts.WithContext = &withCtx
	}

	payload, err := sonic.Marshal(rpcRequest{
		JSONRPC: "2.0",
		ID:      c.id.Add(1),
		Method:  "getProgramAccounts",
		Params:  []any{program, &requestOpts},
	})
	if err != nil {
		return nil, fmt.Errorf("getProgramAccounts: encoding request: %w", err)
	}

	resp, err := c.post(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("getProgramAccounts: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("getProgramAccounts: http %d: %.200s", resp.StatusCode, body)
	}
	return StreamProgramAccounts(resp.Body, fn)
}
